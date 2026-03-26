package cookie

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultAuthCookieName     = "easydrop_access_token"
	defaultAuthCookiePath     = "/"
	defaultAuthCookieMaxAge   = 24 * time.Hour
	defaultAuthCookieSameSite = http.SameSiteLaxMode
)

// AuthCookie 提供认证 Cookie 的读写能力。
type AuthCookie interface {
	Set(c *gin.Context, token string)
	Clear(c *gin.Context)
	Get(c *gin.Context) (string, bool)
}

// Config 是认证 Cookie 配置。
type Config struct {
	Name     string        `mapstructure:"name" yaml:"name"`
	Path     string        `mapstructure:"path" yaml:"path"`
	Domain   string        `mapstructure:"domain" yaml:"domain"`
	SameSite string        `mapstructure:"same_site" yaml:"same_site"`
	Secure   bool          `mapstructure:"-" yaml:"-"`
	MaxAge   time.Duration `mapstructure:"-" yaml:"-"`
}

type authCookie struct {
	name     string
	path     string
	domain   string
	maxAge   time.Duration
	sameSite http.SameSite
	secure   bool
}

// NewAuthCookie 创建认证 Cookie 组件。
func NewAuthCookie(cfg *Config) AuthCookie {
	options := &authCookie{
		name:     defaultAuthCookieName,
		path:     defaultAuthCookiePath,
		maxAge:   defaultAuthCookieMaxAge,
		sameSite: defaultAuthCookieSameSite,
	}
	if cfg == nil {
		return options
	}

	if name := strings.TrimSpace(cfg.Name); name != "" {
		options.name = name
	}
	if path := strings.TrimSpace(cfg.Path); path != "" {
		options.path = path
	}
	options.domain = strings.TrimSpace(cfg.Domain)
	if cfg.MaxAge > 0 {
		options.maxAge = cfg.MaxAge
	}
	options.sameSite = parseSameSiteMode(cfg.SameSite)
	options.secure = cfg.Secure

	return options
}

// Set 写入认证 Cookie。
func (c *authCookie) Set(ctx *gin.Context, token string) {
	if c == nil || ctx == nil || strings.TrimSpace(token) == "" {
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     c.name,
		Value:    token,
		Path:     c.path,
		Domain:   c.domain,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: c.sameSite,
		MaxAge:   int(c.maxAge / time.Second),
		Expires:  time.Now().Add(c.maxAge),
	})
}

// Clear 清除认证 Cookie。
func (c *authCookie) Clear(ctx *gin.Context) {
	if c == nil || ctx == nil {
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     c.name,
		Value:    "",
		Path:     c.path,
		Domain:   c.domain,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: c.sameSite,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// Get 读取认证 Cookie。
func (c *authCookie) Get(ctx *gin.Context) (string, bool) {
	if c == nil || ctx == nil || strings.TrimSpace(c.name) == "" {
		return "", false
	}

	token, err := ctx.Cookie(c.name)
	if err != nil {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}

	return token, true
}

func parseSameSiteMode(value string) http.SameSite {
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
		return defaultAuthCookieSameSite
	}
}

var _ AuthCookie = (*authCookie)(nil)
