package handler

import (
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// CaptchaHandler 处理验证码公开配置请求。
type CaptchaHandler struct {
	captchaService service.CaptchaConfigService
}

// NewCaptchaHandler 创建验证码公开配置处理器。
func NewCaptchaHandler(captchaService service.CaptchaConfigService) *CaptchaHandler {
	return &CaptchaHandler{captchaService: captchaService}
}

// GetConfig 查询验证码公开配置
// @Summary 验证码公开配置
// @Description 返回验证码是否开启、提供商及前端所需 site_key
// @Tags captcha
// @Produce json
// @Success 200 {object} dto.CaptchaConfigResult
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /captcha/config [get]
func (h *CaptchaHandler) GetConfig(c *gin.Context) {
	if h == nil || h.captchaService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	result := h.captchaService.GetConfig(c.Request.Context())
	if result == nil {
		result = &dto.CaptchaConfigResult{}
	}

	c.JSON(http.StatusOK, result)
}
