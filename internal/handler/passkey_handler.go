package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"easydrop/internal/dto"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// PasskeyHandler 处理通行密钥 (WebAuthn/Passkey) 相关请求。
type PasskeyHandler struct {
	passkeyService service.PasskeyService
	jwtManager     jwt.Manager
	authCookie     cookiepkg.AuthCookie
	errorResponder ErrorResponder
}

// NewPasskeyHandler 创建通行密钥处理器。
func NewPasskeyHandler(passkeyService service.PasskeyService, jwtManager jwt.Manager, authCookie cookiepkg.AuthCookie, errorResponder ErrorResponder) *PasskeyHandler {
	return &PasskeyHandler{
		passkeyService: passkeyService,
		jwtManager:     jwtManager,
		authCookie:     authCookie,
		errorResponder: ensureErrorResponder(errorResponder),
	}
}

// BeginRegistration 开始注册通行密钥
// @Summary 开始注册通行密钥
// @Description 发起通行密钥注册流程，返回需传递给浏览器的创建选项和会话 ID。
// @Tags passkey
// @Accept json
// @Produce json
// @Success 200 {object} dto.PasskeyRegisterBeginResponse
// @Failure 400 {object} dto.ErrorResponse "数量已达上限"
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Security BearerAuth
// @Router /auth/passkey/register/begin [post]
func (h *PasskeyHandler) BeginRegistration(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}

	result, err := h.passkeyService.BeginRegistration(c.Request.Context(), userID)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// FinishRegistration 完成注册通行密钥
// @Summary 完成注册通行密钥
// @Description 验证浏览器返回的认证器响应，保存通行密钥凭证到数据库。
// @Description 新创建的通行密钥将自动命名为 "通行密钥 N"。
// @Tags passkey
// @Accept json
// @Produce json
// @Param input body dto.PasskeyRegisterFinishRequest true "注册完成请求，body 需包含 session_id 及完整的 WebAuthn 响应"
// @Success 201 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败、会话无效或数量已达上限"
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/passkey/register/finish [post]
func (h *PasskeyHandler) FinishRegistration(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "读取请求失败"})
		return
	}

	var req dto.PasskeyRegisterFinishRequest
	_ = json.Unmarshal(body, &req)

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "会话ID不能为空"})
		return
	}

	if err := h.passkeyService.FinishRegistration(c.Request.Context(), userID, req.SessionID, body); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.MessageResponse{Message: "通行密钥注册成功"})
}

// BeginLogin 开始通行密钥登录
// @Summary 开始通行密钥登录
// @Description 发起无用户名通行密钥登录流程，返回需传递给浏览器的断言选项和会话 ID。
// @Description 用户将在浏览器中选择要使用的通行密钥进行身份验证。
// @Tags passkey
// @Accept json
// @Produce json
// @Success 200 {object} dto.PasskeyLoginBeginResponse
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/passkey/login/begin [post]
func (h *PasskeyHandler) BeginLogin(c *gin.Context) {
	result, err := h.passkeyService.BeginLogin(c.Request.Context())
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// FinishLogin 完成通行密钥登录
// @Summary 完成通行密钥登录
// @Description 验证浏览器返回的断言响应，认证通过后签发 JWT 访问令牌。
// @Tags passkey
// @Accept json
// @Produce json
// @Param input body dto.PasskeyLoginFinishRequest true "登录完成请求，body 需包含 session_id 及完整的 WebAuthn 断言响应"
// @Success 200 {object} dto.AuthResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败或会话无效"
// @Failure 401 {object} dto.ErrorResponse "通行密钥验证失败"
// @Failure 403 {object} dto.ErrorResponse "用户状态异常"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /auth/passkey/login/finish [post]
func (h *PasskeyHandler) FinishLogin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "读取请求失败"})
		return
	}

	var req dto.PasskeyLoginFinishRequest
	_ = json.Unmarshal(body, &req)

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "会话ID不能为空"})
		return
	}

	user, err := h.passkeyService.FinishLogin(c.Request.Context(), req.SessionID, body)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	token, err := h.jwtManager.IssueAccessToken(user.ID, user.Username, user.Admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "签发令牌失败"})
		return
	}

	if h.authCookie != nil {
		h.authCookie.Set(c, token)
	}
	c.JSON(http.StatusOK, dto.AuthResult{AccessToken: token})
}

// List 列出当前用户的通行密钥
// @Summary 列出通行密钥
// @Description 获取当前用户所有已注册的通行密钥列表（仅元数据，不含密钥材料）。
// @Tags passkey
// @Accept json
// @Produce json
// @Success 200 {object} object{items=[]dto.PasskeyItem}
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Security BearerAuth
// @Router /users/me/passkeys [get]
func (h *PasskeyHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}

	items, err := h.passkeyService.List(c.Request.Context(), userID)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	if items == nil {
		items = []dto.PasskeyItem{}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Rename 重命名通行密钥
// @Summary 重命名通行密钥
// @Description 修改指定通行密钥的显示名称，名称长度限制为 1-15 个字符。
// @Tags passkey
// @Accept json
// @Produce json
// @Param id path int true "通行密钥 ID"
// @Param input body dto.PasskeyRenameRequest true "新名称"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 404 {object} dto.ErrorResponse "通行密钥不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Security BearerAuth
// @Router /users/me/passkeys/{id} [patch]
func (h *PasskeyHandler) Rename(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "通行密钥ID无效"})
		return
	}

	var req dto.PasskeyRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	if err := h.passkeyService.Rename(c.Request.Context(), userID, uint(id), req.Name); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "ok"})
}

// Delete 删除通行密钥
// @Summary 删除通行密钥
// @Description 删除指定的通行密钥，删除后该密钥将无法用于登录。
// @Tags passkey
// @Accept json
// @Produce json
// @Param id path int true "通行密钥 ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 404 {object} dto.ErrorResponse "通行密钥不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Security BearerAuth
// @Router /users/me/passkeys/{id} [delete]
func (h *PasskeyHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "通行密钥ID无效"})
		return
	}

	if err := h.passkeyService.Delete(c.Request.Context(), userID, uint(id)); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "已删除"})
}
