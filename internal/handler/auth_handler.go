package handler

import (
	"net/http"

	"easydrop/internal/dto"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 处理认证相关请求。
type AuthHandler struct {
	authService    service.AuthService
	userService    service.UserService
	authCookie     cookiepkg.AuthCookie
	errorResponder ErrorResponder
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(authService service.AuthService, userService service.UserService, authCookie cookiepkg.AuthCookie, errorResponder ErrorResponder) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		userService:    userService,
		authCookie:     authCookie,
		errorResponder: ensureErrorResponder(errorResponder),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户并发送邮箱验证邮件
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.RegisterInput true "注册信息"
// @Success 201 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或验证码缺失"
// @Failure 403 {object} dto.ErrorResponse "注册关闭"
// @Failure 409 {object} dto.ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	if !ensureServiceReady(c, h.authService) {
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
		h.errorResponder.Respond(c, err)
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
// @Failure 401 {object} dto.ErrorResponse "账号或密码错误"
// @Failure 403 {object} dto.ErrorResponse "用户状态异常"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	if !ensureServiceReady(c, h.authService) {
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
		h.errorResponder.Respond(c, err)
		return
	}

	if h.authCookie != nil {
		h.authCookie.Set(c, result.AccessToken)
	}
	c.JSON(http.StatusOK, result)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 清除认证 Cookie
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.MessageResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if h.authCookie != nil {
		h.authCookie.Clear(c)
	}
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "已退出登录"})
}

// RequestPasswordReset 发起密码重置请求。
// @Summary 忘记密码
// @Description 按邮箱发起密码重置并发送确认邮件
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.PasswordResetRequestInput true "重置密码请求参数"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或验证码缺失"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/password-reset/request [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	if !ensureServiceReady(c, h.authService) {
		return
	}

	var input dto.PasswordResetRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	if input.Captcha != nil {
		input.Captcha.RemoteIP = c.ClientIP()
	}

	if err := h.authService.RequestPasswordReset(c.Request.Context(), input); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

// ConfirmPasswordReset 校验重置 token 并更新密码。
// @Summary 重置密码
// @Description 使用邮件中的 token 完成密码重置
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.PasswordResetConfirmInput true "重置密码确认参数"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或 token 无效"
// @Failure 404 {object} dto.ErrorResponse "用户不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/password-reset/confirm [post]
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	if !ensureServiceReady(c, h.authService) {
		return
	}

	var input dto.PasswordResetConfirmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	if err := h.authService.ConfirmPasswordReset(c.Request.Context(), input); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

// ConfirmVerifyEmail 校验邮箱验证 token。
// @Summary 验证注册邮箱
// @Description 使用邮件中的 token 完成邮箱验证
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.EmailVerifyConfirmInput true "邮箱验证确认参数"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或 token 无效"
// @Failure 404 {object} dto.ErrorResponse "用户不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/verify-email/confirm [post]
func (h *AuthHandler) ConfirmVerifyEmail(c *gin.Context) {
	if !ensureServiceReady(c, h.authService) {
		return
	}

	var input dto.EmailVerifyConfirmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	if err := h.authService.ConfirmVerifyEmail(c.Request.Context(), input); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

// ConfirmEmailChange 校验邮箱修改 token 并完成邮箱修改。
// @Summary 确认邮箱修改
// @Description 使用邮件中的 token 完成邮箱修改
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.UserChangeEmailConfirmInput true "邮箱修改确认参数"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或 token 无效"
// @Failure 404 {object} dto.ErrorResponse "用户不存在"
// @Failure 409 {object} dto.ErrorResponse "邮箱冲突"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/email-change/confirm [post]
func (h *AuthHandler) ConfirmEmailChange(c *gin.Context) {
	if !ensureServiceReady(c, h.userService) {
		return
	}

	var input dto.UserChangeEmailConfirmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	result, err := h.userService.ConfirmEmailChange(c.Request.Context(), input)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
