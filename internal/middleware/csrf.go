package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"easydrop/internal/config"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultCSRFCookieName = "easydrop_csrf_token"
	CSRFHeaderName        = "X-CSRF-Token"

	csrfHeaderFallbackName = "X-XSRF-Token"
	defaultCSRFCookiePath  = "/"
	defaultAuthCookieName  = "easydrop_access_token"
	csrfTokenBytes         = 32
)

var errCSRFTokenGenerateFailed = errors.New("生成 CSRF token 失败")

// CSRF 提供 CSRF 保护与登录成功后下发 Cookie 的中间件能力。
type CSRF interface {
	Protect(c *gin.Context)
	IssueCSRFCookieOnSuccess(c *gin.Context)
}

type csrf struct {
	protect                  gin.HandlerFunc
	issueCSRFCookieOnSuccess gin.HandlerFunc
}

// NewCSRF 创建 CSRF 中间件组件，供 DI 注入。
func NewCSRF(cfg *config.StaticConfig) CSRF {
	if cfg == nil {
		return &csrf{
			protect:                  passthroughCSRFMiddleware(),
			issueCSRFCookieOnSuccess: passthroughCSRFMiddleware(),
		}
	}

	options := CSRFOptions{
		AuthCookieName: cfg.AuthCookie.Name,
		CookieName:     DefaultCSRFCookieName,
		CookiePath:     cfg.AuthCookie.Path,
		CookieDomain:   cfg.AuthCookie.Domain,
		CookieSameSite: ParseSameSiteMode(cfg.AuthCookie.SameSite),
		CookieSecure:   cfg.Server.Mode == config.ServerModeProduction,
		CookieMaxAge:   cfg.JWT.Expire,
	}

	return &csrf{
		protect:                  NewDoubleSubmitCSRF(options),
		issueCSRFCookieOnSuccess: IssueCSRFCookieOnSuccess(options),
	}
}

func (m *csrf) Protect(c *gin.Context) {
	if m == nil || m.protect == nil {
		c.Next()
		return
	}
	m.protect(c)
}

func (m *csrf) IssueCSRFCookieOnSuccess(c *gin.Context) {
	if m == nil || m.issueCSRFCookieOnSuccess == nil {
		c.Next()
		return
	}
	m.issueCSRFCookieOnSuccess(c)
}

// CSRFOptions 定义双提交 Cookie CSRF 防护配置。
type CSRFOptions struct {
	AuthCookieName string
	CookieName     string
	CookiePath     string
	CookieDomain   string
	CookieSameSite http.SameSite
	CookieSecure   bool
	CookieMaxAge   time.Duration
}

type normalizedCSRFOptions struct {
	authCookieName string
	cookieName     string
	cookiePath     string
	cookieDomain   string
	cookieSameSite http.SameSite
	cookieSecure   bool
	cookieMaxAge   time.Duration
}

// NewDoubleSubmitCSRF 创建双提交 Cookie CSRF 防护中间件。
// 仅在以下条件同时满足时执行阻断校验：
// 1) 请求方法为写操作（POST/PUT/PATCH/DELETE）；
// 2) 无 Bearer Authorization 头；
// 3) 存在认证 Cookie。
func NewDoubleSubmitCSRF(options CSRFOptions) gin.HandlerFunc {
	config := normalizeCSRFOptions(options)

	return func(c *gin.Context) {
		if c == nil || c.Request == nil {
			c.Next()
			return
		}

		method := strings.ToUpper(strings.TrimSpace(c.Request.Method))
		if !isUnsafeHTTPMethod(method) {
			if err := ensureCSRFCookieForAuthRequest(c, config); err != nil {
				abortCSRFFailed(c, http.StatusInternalServerError, "CSRF token 生成失败")
				return
			}
			c.Next()
			return
		}

		if hasBearerAuthorization(c.GetHeader("Authorization")) {
			if err := ensureCSRFCookieForAuthRequest(c, config); err != nil {
				abortCSRFFailed(c, http.StatusInternalServerError, "CSRF token 生成失败")
				return
			}
			c.Next()
			return
		}

		if _, ok := getRequestCookie(c, config.authCookieName); !ok {
			c.Next()
			return
		}

		cookieToken, ok := getRequestCookie(c, config.cookieName)
		if !ok {
			if err := ensureCSRFCookieForAuthRequest(c, config); err != nil {
				abortCSRFFailed(c, http.StatusInternalServerError, "CSRF token 生成失败")
				return
			}
			abortCSRFFailed(c, http.StatusForbidden, "CSRF token 缺失")
			return
		}

		headerToken := readCSRFHeaderToken(c)
		if !constantTimeTokenEqual(cookieToken, headerToken) {
			abortCSRFFailed(c, http.StatusForbidden, "CSRF token 无效")
			return
		}

		c.Next()
	}
}

// IssueCSRFCookieOnSuccess 在请求成功后写入 CSRF Cookie。
// 适用于登录等会建立会话的接口，保证后续写请求可直接携带 CSRF 头。
func IssueCSRFCookieOnSuccess(options CSRFOptions) gin.HandlerFunc {
	config := normalizeCSRFOptions(options)

	return func(c *gin.Context) {
		c.Next()

		if c == nil || c.Writer == nil || c.Request == nil {
			return
		}
		if c.Writer.Status() >= http.StatusBadRequest {
			return
		}
		if _, ok := getRequestCookie(c, config.cookieName); ok {
			return
		}

		token, err := generateCSRFToken()
		if err != nil {
			return
		}

		writeCSRFCookie(c, config, token)
	}
}

// ParseSameSiteMode 将字符串配置解析为 http.SameSite。
func ParseSameSiteMode(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "default":
		return http.SameSiteDefaultMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

func isUnsafeHTTPMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func hasBearerAuthorization(value string) bool {
	parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
	if len(parts) != 2 {
		return false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return false
	}

	token := strings.TrimSpace(parts[1])
	return token != ""
}

func normalizeCSRFOptions(options CSRFOptions) normalizedCSRFOptions {
	authCookieName := strings.TrimSpace(options.AuthCookieName)
	if authCookieName == "" {
		authCookieName = defaultAuthCookieName
	}

	cookieName := strings.TrimSpace(options.CookieName)
	if cookieName == "" {
		cookieName = DefaultCSRFCookieName
	}

	cookiePath := strings.TrimSpace(options.CookiePath)
	if cookiePath == "" {
		cookiePath = defaultCSRFCookiePath
	}

	cookieSameSite := options.CookieSameSite
	if cookieSameSite == 0 {
		cookieSameSite = http.SameSiteLaxMode
	}

	cookieMaxAge := options.CookieMaxAge
	if cookieMaxAge <= 0 {
		cookieMaxAge = 24 * time.Hour
	}

	return normalizedCSRFOptions{
		authCookieName: authCookieName,
		cookieName:     cookieName,
		cookiePath:     cookiePath,
		cookieDomain:   strings.TrimSpace(options.CookieDomain),
		cookieSameSite: cookieSameSite,
		cookieSecure:   options.CookieSecure,
		cookieMaxAge:   cookieMaxAge,
	}
}

func ensureCSRFCookieForAuthRequest(c *gin.Context, options normalizedCSRFOptions) error {
	if _, ok := getRequestCookie(c, options.authCookieName); !ok {
		return nil
	}
	if _, ok := getRequestCookie(c, options.cookieName); ok {
		return nil
	}

	token, err := generateCSRFToken()
	if err != nil {
		return err
	}
	writeCSRFCookie(c, options, token)
	return nil
}

func getRequestCookie(c *gin.Context, name string) (string, bool) {
	if c == nil || strings.TrimSpace(name) == "" {
		return "", false
	}

	v, err := c.Cookie(name)
	if err != nil {
		return "", false
	}

	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}

	return v, true
}

func readCSRFHeaderToken(c *gin.Context) string {
	if c == nil {
		return ""
	}

	token := strings.TrimSpace(c.GetHeader(CSRFHeaderName))
	if token != "" {
		return token
	}

	return strings.TrimSpace(c.GetHeader(csrfHeaderFallbackName))
}

func constantTimeTokenEqual(cookieToken, headerToken string) bool {
	if cookieToken == "" || headerToken == "" {
		return false
	}
	if len(cookieToken) != len(headerToken) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) == 1
}

func generateCSRFToken() (string, error) {
	buf := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", errCSRFTokenGenerateFailed
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func writeCSRFCookie(c *gin.Context, options normalizedCSRFOptions, token string) {
	if c == nil || c.Writer == nil || strings.TrimSpace(token) == "" {
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     options.cookieName,
		Value:    token,
		Path:     options.cookiePath,
		Domain:   options.cookieDomain,
		HttpOnly: false,
		Secure:   options.cookieSecure,
		SameSite: options.cookieSameSite,
		MaxAge:   int(options.cookieMaxAge / time.Second),
		Expires:  time.Now().Add(options.cookieMaxAge),
	})
}

func abortCSRFFailed(c *gin.Context, status int, message string) {
	if c == nil {
		return
	}

	c.AbortWithStatusJSON(status, gin.H{"message": message})
}

func passthroughCSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

var _ CSRF = (*csrf)(nil)
