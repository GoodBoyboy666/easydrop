package handler

import (
	"easydrop/internal/middleware"
	"errors"
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// PostAdminHandler 处理管理端说说请求。
type PostAdminHandler struct {
	postService service.PostService
}

// NewPostAdminHandler 创建管理端说说处理器。
func NewPostAdminHandler(postService service.PostService) *PostAdminHandler {
	return &PostAdminHandler{postService: postService}
}

// List 查询说说列表（管理端）
// @Summary 管理端查询说说列表
// @Description 分页查询说说列表，支持按用户和标签过滤
// @Tags post-admin
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID"
// @Param tag_id query int false "标签ID"
// @Param content query string false "内容关键字"
// @Param page query int false "页码，从 1 开始"
// @Param size query int false "每页条数，最大 100"
// @Param order query string false "排序，如 created_at desc"
// @Success 200 {object} dto.PostListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/posts [get]
func (h *PostAdminHandler) List(c *gin.Context) {
	if !ensureServiceReady(c, h.postService) {
		return
	}

	var req dto.PostListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.postService.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(mapPostErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Get 查询说说详情（管理端）
// @Summary 管理端查询说说详情
// @Description 根据说说 ID 查询详情
// @Tags post-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "说说ID"
// @Success 200 {object} dto.PostDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "说说不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/posts/{id} [get]
func (h *PostAdminHandler) Get(c *gin.Context) {
	if !ensureServiceReady(c, h.postService) {
		return
	}

	var req dto.PostIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	result, err := h.postService.Get(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(mapPostErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Create 创建说说（管理端）
// @Summary 管理端创建说说
// @Description 创建说说并返回详情
// @Tags post-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.PostCreateInput true "说说创建参数"
// @Success 201 {object} dto.PostDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/posts [post]
func (h *PostAdminHandler) Create(c *gin.Context) {
	if !ensureServiceReady(c, h.postService) {
		return
	}

	var input dto.PostCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}
	input.UserID = userID

	result, err := h.postService.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapPostErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Update 更新说说（管理端）
// @Summary 管理端更新说说
// @Description 根据说说 ID 更新说说内容
// @Tags post-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "说说ID"
// @Param input body dto.PostUpdateInput true "说说更新参数"
// @Success 200 {object} dto.PostDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "说说不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/posts/{id} [patch]
func (h *PostAdminHandler) Update(c *gin.Context) {
	if !ensureServiceReady(c, h.postService) {
		return
	}

	var uriReq dto.PostIDURIInput
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	var input dto.PostUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.ID = uriReq.ID

	result, err := h.postService.Update(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapPostErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete 删除说说（管理端）
// @Summary 管理端删除说说
// @Description 根据说说 ID 删除说说
// @Tags post-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "说说ID"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "说说不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/posts/{id} [delete]
func (h *PostAdminHandler) Delete(c *gin.Context) {
	if !ensureServiceReady(c, h.postService) {
		return
	}

	var req dto.PostIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	if err := h.postService.Delete(c.Request.Context(), req.ID); err != nil {
		c.JSON(mapPostErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}

func mapPostErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrEmptyPostContent),
		errors.Is(err, service.ErrInvalidPostUser),
		errors.Is(err, service.ErrTagNameTooLong):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrPostNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
