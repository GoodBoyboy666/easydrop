package handler

import (
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// CommentAdminHandler 处理管理端评论请求。
type CommentAdminHandler struct {
	commentService service.CommentService
}

// NewCommentAdminHandler 创建管理端评论处理器。
func NewCommentAdminHandler(commentService service.CommentService) *CommentAdminHandler {
	return &CommentAdminHandler{commentService: commentService}
}

// List 查询评论列表（管理端）。
// @Summary 管理端查询评论列表
// @Description 分页查询评论列表，支持按说说和用户过滤
// @Tags comment-admin
// @Produce json
// @Security BearerAuth
// @Param post_id query int false "说说ID"
// @Param user_id query int false "用户ID"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at_desc 或 created_at_asc"
// @Success 200 {object} dto.CommentListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/comments [get]
func (h *CommentAdminHandler) List(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.CommentAdminListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.commentService.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Get 查询评论详情（管理端）。
// @Summary 管理端查询评论详情
// @Description 根据评论 ID 查询评论详情
// @Tags comment-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Success 200 {object} dto.CommentDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "评论不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/comments/{id} [get]
func (h *CommentAdminHandler) Get(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.CommentIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	result, err := h.commentService.Get(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update 更新评论（管理端）。
// @Summary 管理端更新评论
// @Description 根据评论 ID 更新评论内容
// @Tags comment-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Param input body dto.CommentUpdateInput true "评论更新参数"
// @Success 200 {object} dto.CommentDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "评论不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/comments/{id} [patch]
func (h *CommentAdminHandler) Update(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var uriReq dto.CommentIDURIInput
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	var input dto.CommentUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.ID = uriReq.ID

	result, err := h.commentService.Update(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete 删除评论（管理端）。
// @Summary 管理端删除评论
// @Description 根据评论 ID 删除评论
// @Tags comment-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "评论不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/comments/{id} [delete]
func (h *CommentAdminHandler) Delete(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.CommentIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	if err := h.commentService.Delete(c.Request.Context(), req.ID); err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}
