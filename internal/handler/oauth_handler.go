package handler

import (
	"net/http"
	"strconv"

	"easydrop/internal/dto"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/oauth"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	oauthStateCookieName = "easydrop_oauth_state"
	oauthStateMaxAge     = 600
)

// OAuthHandler 处理社交登录相关的 HTTP 请求。
type OAuthHandler struct {
	oauthService   service.OAuthService
	authCookie     cookiepkg.AuthCookie
	errorResponder ErrorResponder
}

// NewOAuthHandler 创建社交登录处理器。
func NewOAuthHandler(oauthService service.OAuthService, authCookie cookiepkg.AuthCookie, errorResponder ErrorResponder) *OAuthHandler {
	return &OAuthHandler{
		oauthService:   oauthService,
		authCookie:     authCookie,
		errorResponder: ensureErrorResponder(errorResponder),
	}
}

// ListProviders 获取已启用的社交登录方式列表
// @Summary 获取社交登录方式
// @Description 返回当前已配置并启用的社交登录方式列表
// @Tags oauth
// @Produce json
// @Success 200 {object} object{providers=[]dto.OAuthProviderItem}
// @Router /auth/oauth/providers [get]
func (h *OAuthHandler) ListProviders(c *gin.Context) {
	if !ensureServiceReady(c, h.oauthService) {
		return
	}
	providers := h.oauthService.GetEnabledProviders()
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// Authorize 发起社交登录授权
// @Summary 发起社交登录
// @Description 生成 OAuth 授权 URL 并重定向到社交平台授权页面
// @Tags oauth
// @Param provider path string true "社交平台名称 (google/github/twitter/microsoft/apple)"
// @Success 302 {string} string "重定向到社交平台授权页"
// @Failure 400 {object} dto.ErrorResponse "未指定社交登录方式"
// @Failure 404 {object} dto.ErrorResponse "该社交登录方式未配置或未启用"
// @Router /auth/oauth/{provider} [get]
func (h *OAuthHandler) Authorize(c *gin.Context) {
	if !ensureServiceReady(c, h.oauthService) {
		return
	}
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "未指定社交登录方式"})
		return
	}

	state := oauth.GenerateState()
	setOAuthStateCookie(c, state)

	authURL, err := h.oauthService.GetAuthURL(c.Request.Context(), provider, state)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// Callback 处理社交登录回调
// @Summary 处理社交登录回调
// @Description 前端从 OAuth 提供方回跳 URL 中取得 code 与 state 后，POST 到此后端完成登录/注册，返回 JWT 并设置认证 Cookie
// @Tags oauth
// @Accept json
// @Produce json
// @Param provider path string true "社交平台名称 (google/github/twitter/microsoft/apple)"
// @Param input body dto.OAuthCallbackInput true "回调参数"
// @Success 200 {object} dto.AuthResult
// @Failure 400 {object} dto.ErrorResponse "参数格式错误或 state 校验失败"
// @Failure 404 {object} dto.ErrorResponse "该社交登录方式未配置或未启用"
// @Failure 409 {object} dto.ErrorResponse "邮箱已存在但未绑定"
// @Router /auth/oauth/{provider}/callback [post]
func (h *OAuthHandler) Callback(c *gin.Context) {
	if !ensureServiceReady(c, h.oauthService) {
		return
	}
	provider := c.Param("provider")

	var input dto.OAuthCallbackInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	stateFromCookie, _ := c.Cookie(oauthStateCookieName)
	clearOAuthStateCookie(c)

	result, err := h.oauthService.HandleCallback(c.Request.Context(), provider, input.Code, input.State, stateFromCookie)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	if h.authCookie != nil {
		h.authCookie.Set(c, result.AccessToken)
	}
	c.JSON(http.StatusOK, result)
}

// ListBindings 查看已绑定的社交账号
// @Summary 查看已绑定的社交账号
// @Description 返回当前用户已绑定的所有社交账号信息
// @Tags oauth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{binds=[]dto.OAuthBindDTO}
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Router /users/me/oauth-binds [get]
func (h *OAuthHandler) ListBindings(c *gin.Context) {
	if !ensureServiceReady(c, h.oauthService) {
		return
	}
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}
	binds, err := h.oauthService.GetUserBindings(c.Request.Context(), userID)
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"binds": binds})
}

// Unbind 解绑社交账号
// @Summary 解绑社交账号
// @Description 解除当前用户与指定社交账号的绑定
// @Tags oauth
// @Security BearerAuth
// @Param id path int true "绑定记录 ID"
// @Produce json
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "无效的绑定ID"
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 404 {object} dto.ErrorResponse "未找到绑定记录"
// @Router /users/me/oauth-binds/{id} [delete]
func (h *OAuthHandler) Unbind(c *gin.Context) {
	if !ensureServiceReady(c, h.oauthService) {
		return
	}
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}
	bindID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "无效的绑定ID"})
		return
	}
	if err := h.oauthService.Unbind(c.Request.Context(), userID, uint(bindID)); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "已解绑"})
}

// BindManually 手动绑定社交账号
// @Summary 手动绑定社交账号
// @Description 已登录用户主动绑定一个社交平台账户（用于解决邮箱冲突场景）
// @Tags oauth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param provider path string true "社交平台名称 (google/github/twitter/microsoft/apple)"
// @Param input body dto.OAuthBindManualInput true "绑定参数"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse "参数格式错误或 state 校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录"
// @Failure 409 {object} dto.ErrorResponse "该社交账号已绑定其他用户或已绑定该平台"
// @Router /users/me/oauth-binds/{provider} [post]
func (h *OAuthHandler) BindManually(c *gin.Context) {
	if !ensureServiceReady(c, h.oauthService) {
		return
	}
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录"})
		return
	}
	provider := c.Param("provider")
	var input dto.OAuthBindManualInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	stateFromCookie, _ := c.Cookie(oauthStateCookieName)
	clearOAuthStateCookie(c)

	if err := h.oauthService.BindManually(c.Request.Context(), userID, provider, input.Code, input.State, stateFromCookie); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "绑定成功"})
}

func setOAuthStateCookie(c *gin.Context, state string) {
	c.SetCookie(oauthStateCookieName, state, oauthStateMaxAge, "/api/v1", "", false, true)
}

func clearOAuthStateCookie(c *gin.Context) {
	c.SetCookie(oauthStateCookieName, "", -1, "/api/v1", "", false, true)
}

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := v.(uint)
	return uid, ok
}
