package middleware

import (
	"bytes"
	"errors"
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
	if c == nil || c.Request == nil || c.Request.Body == nil {
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

	maxBytesBody := http.MaxBytesReader(c.Writer, c.Request.Body, limit)

	// 已知 Content-Length 的请求无需预读，避免一次完整内存拷贝。
	if !bufferBody || !shouldBufferRequestBody(c.Request) {
		c.Request.Body = maxBytesBody
		c.Next()
		return
	}

	body, err := io.ReadAll(maxBytesBody)
	if err != nil {
		if isRequestBodyTooLargeError(err) {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"message": "请求体过大"})
			return
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "请求体读取失败"})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
	c.Next()
}

func shouldBufferRequestBody(req *http.Request) bool {
	if req == nil {
		return false
	}

	// Content-Length = -1 表示未知长度（常见于 chunked），中间件需预读才能稳定返回 413。
	return req.ContentLength < 0
}

func isRequestBodyTooLargeError(err error) bool {
	if err == nil {
		return false
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "request body too large")
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
