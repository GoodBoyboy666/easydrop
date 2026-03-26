package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	OrdinaryMaxRequestBodyBytes      int64 = 2 * 1024 * 1024
	DefaultUploadMaxRequestBodyBytes int64 = 50 * 1024 * 1024
)

// RequestBodyLimit 提供请求体大小限制能力。
type RequestBodyLimit interface {
	Ordinary(c *gin.Context)
	Upload(c *gin.Context)
}

type requestBodyLimit struct {
	settings service.SettingService
}

// NewRequestBodyLimit 创建请求体大小限制中间件。
func NewRequestBodyLimit(settings service.SettingService) RequestBodyLimit {
	return &requestBodyLimit{
		settings: settings,
	}
}

// Ordinary 限制普通接口请求体大小为固定 2MB。
func (m *requestBodyLimit) Ordinary(c *gin.Context) {
	handleRequestBodyLimit(c, OrdinaryMaxRequestBodyBytes, true)
}

// Upload 按系统设置限制上传接口请求体大小。
func (m *requestBodyLimit) Upload(c *gin.Context) {
	var settings service.SettingService
	if m != nil {
		settings = m.settings
	}
	handleRequestBodyLimit(c, resolveUploadMaxRequestBodyBytes(c, settings), false)
}

func handleRequestBodyLimit(c *gin.Context, limit int64, bufferBody bool) {
	if c.Request == nil || c.Request.Body == nil {
		c.Next()
		return
	}

	if limit <= 0 {
		limit = OrdinaryMaxRequestBodyBytes
	}

	if c.Request.ContentLength > limit {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"message": "请求体过大"})
		return
	}

	if !bufferBody {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, limit))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"message": "请求体过大"})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
	c.Next()
}

func resolveUploadMaxRequestBodyBytes(c *gin.Context, settings service.SettingService) int64 {
	if settings == nil {
		return DefaultUploadMaxRequestBodyBytes
	}

	value, found, err := settings.GetValue(c.Request.Context(), service.UploadMaxRequestBodySettingKey)
	if err != nil {
		log.Printf("读取上传请求体大小限制失败，使用默认值: %v", err)
		return DefaultUploadMaxRequestBodyBytes
	}
	if !found {
		return DefaultUploadMaxRequestBodyBytes
	}

	limit, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || limit <= 0 {
		log.Printf("解析上传请求体大小限制失败，使用默认值: %v", err)
		return DefaultUploadMaxRequestBodyBytes
	}

	return limit
}

var _ RequestBodyLimit = (*requestBodyLimit)(nil)
