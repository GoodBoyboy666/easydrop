package handler

import (
	"errors"
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 处理认证相关请求。
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户并返回登录态信息
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.RegisterInput true "注册信息"
// @Success 201 {object} dto.AuthResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或验证码缺失"
// @Failure 403 {object} dto.ErrorResponse "注册关闭"
// @Failure 409 {object} dto.ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var input dto.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	if input.Captcha != nil {
		input.Captcha.RemoteIP = c.ClientIP()
	}

	result, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		status := mapAuthErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名或邮箱登录并返回访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.LoginInput true "登录信息"
// @Success 200 {object} dto.AuthResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或验证码缺失"
// @Failure 401 {object} dto.ErrorResponse "账号不存在或密码错误"
// @Failure 403 {object} dto.ErrorResponse "用户状态异常"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	if input.Captcha != nil {
		input.Captcha.RemoteIP = c.ClientIP()
	}

	result, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		status := mapAuthErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func mapAuthErrorStatus(err error) int {
	switch {
	case errors.Is(err, validator.ErrEmptyUsername),
		errors.Is(err, validator.ErrUsernameTooShort),
		errors.Is(err, validator.ErrUsernameTooLong),
		errors.Is(err, validator.ErrInvalidUsernameFormat),
		errors.Is(err, validator.ErrEmptyPassword),
		errors.Is(err, validator.ErrPasswordTooShort),
		errors.Is(err, validator.ErrPasswordContainsSpace),
		errors.Is(err, validator.ErrPasswordMissingLetter),
		errors.Is(err, validator.ErrPasswordMissingNumber),
		errors.Is(err, validator.ErrEmptyEmail),
		errors.Is(err, validator.ErrInvalidEmailFormat),
		errors.Is(err, service.ErrEmptyAccount),
		errors.Is(err, service.ErrCaptchaRequired),
		errors.Is(err, service.ErrCaptchaFailed),
		errors.Is(err, service.ErrInvalidSiteSetting):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrInvalidPassword):
		return http.StatusUnauthorized
	case errors.Is(err, service.ErrRegisterClosed),
		errors.Is(err, service.ErrUserDisabled):
		return http.StatusForbidden
	case errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrEmailExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
