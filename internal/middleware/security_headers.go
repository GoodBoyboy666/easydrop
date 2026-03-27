package middleware

import "github.com/gin-gonic/gin"

const (
	headerXContentTypeOptions = "X-Content-Type-Options"
	headerXFrameOptions       = "X-Frame-Options"
	headerReferrerPolicy      = "Referrer-Policy"
	headerPermissionsPolicy   = "Permissions-Policy"

	headerValueNoSniff                     = "nosniff"
	headerValueFrameDeny                   = "DENY"
	headerValueStrictOriginWhenCrossOrigin = "strict-origin-when-cross-origin"
	headerValueDefaultPermissionsPolicy    = "geolocation=(), microphone=(), camera=()"
)

// SecurityHeaders 提供统一安全响应头能力。
type SecurityHeaders interface {
	Apply(c *gin.Context)
}

type securityHeaders struct{}

// NewSecurityHeaders 创建安全头中间件。
func NewSecurityHeaders() SecurityHeaders {
	return &securityHeaders{}
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

	c.Next()
}

var _ SecurityHeaders = (*securityHeaders)(nil)
