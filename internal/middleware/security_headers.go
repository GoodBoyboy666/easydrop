package middleware

import (
	"strings"

	"easydrop/internal/pkg/captcha"

	"github.com/gin-gonic/gin"
)

const (
	headerXContentTypeOptions   = "X-Content-Type-Options"
	headerXFrameOptions         = "X-Frame-Options"
	headerReferrerPolicy        = "Referrer-Policy"
	headerPermissionsPolicy     = "Permissions-Policy"
	headerContentSecurityPolicy = "Content-Security-Policy"

	headerValueNoSniff                     = "nosniff"
	headerValueFrameDeny                   = "DENY"
	headerValueStrictOriginWhenCrossOrigin = "strict-origin-when-cross-origin"
	headerValueDefaultPermissionsPolicy    = "geolocation=(), microphone=(), camera=()"

	swaggerRoutePrefix = "/api/swagger"
)

var captchaCSPSourcesByProvider = map[captcha.Provider][]string{
	captcha.ProviderTurnstile: {
		"https://challenges.cloudflare.com",
	},
	captcha.ProviderRecaptcha: {
		"https://www.google.com",
		"https://www.gstatic.com",
	},
	captcha.ProviderHCaptcha: {
		"https://js.hcaptcha.com",
		"https://hcaptcha.com",
		"https://*.hcaptcha.com",
	},
	captcha.ProviderGeetestV4: {
		"https://static.geetest.com",
		"https://gcaptcha4.geetest.com",
	},
}

// SecurityHeaders 提供统一安全响应头能力。
type SecurityHeaders interface {
	Apply(c *gin.Context)
}

type securityHeaders struct {
	captchaCfg *captcha.AllCaptchaConfig
}

// NewSecurityHeaders 创建安全头中间件。
func NewSecurityHeaders(captchaCfg *captcha.AllCaptchaConfig) SecurityHeaders {
	return &securityHeaders{captchaCfg: captchaCfg}
}

// Apply 为响应写入基础安全头。
func (m *securityHeaders) Apply(c *gin.Context) {
	if c == nil {
		return
	}

	headers := c.Writer.Header()
	headers.Set(headerXContentTypeOptions, headerValueNoSniff)
	headers.Set(headerXFrameOptions, headerValueFrameDeny)
	headers.Set(headerReferrerPolicy, headerValueStrictOriginWhenCrossOrigin)
	headers.Set(headerPermissionsPolicy, headerValueDefaultPermissionsPolicy)
	if !shouldSkipCSP(c) {
		headers.Set(headerContentSecurityPolicy, m.buildCSPPolicy())
	}

	c.Next()
}

func (m *securityHeaders) buildCSPPolicy() string {
	scriptSrc := []string{"'self'"}
	frameSrc := []string{"'self'"}
	connectSrc := []string{"'self'"}

	for _, source := range m.resolveCaptchaCSPSources() {
		scriptSrc = appendUniqueSource(scriptSrc, source)
		frameSrc = appendUniqueSource(frameSrc, source)
		connectSrc = appendUniqueSource(connectSrc, source)
	}

	return "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; script-src " + strings.Join(scriptSrc, " ") + "; frame-src " + strings.Join(frameSrc, " ") + "; connect-src " + strings.Join(connectSrc, " ")
}

func (m *securityHeaders) resolveCaptchaCSPSources() []string {
	if m == nil || m.captchaCfg == nil || !m.captchaCfg.Enabled {
		return nil
	}

	provider := captcha.Provider(strings.ToLower(strings.TrimSpace(string(m.captchaCfg.Provider))))
	sources, ok := captchaCSPSourcesByProvider[provider]
	if !ok {
		return nil
	}

	result := make([]string, 0, len(sources))
	for _, source := range sources {
		result = appendUniqueSource(result, source)
	}
	return result
}

func appendUniqueSource(sources []string, source string) []string {
	source = strings.TrimSpace(source)
	if source == "" {
		return sources
	}

	for _, current := range sources {
		if current == source {
			return sources
		}
	}

	return append(sources, source)
}

func shouldSkipCSP(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}

	requestPath := strings.TrimSpace(c.Request.URL.Path)
	if requestPath == "" {
		return false
	}

	return requestPath == swaggerRoutePrefix || strings.HasPrefix(requestPath, swaggerRoutePrefix+"/")
}

var _ SecurityHeaders = (*securityHeaders)(nil)
