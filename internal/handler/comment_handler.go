package handler

import (
	"errors"
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// CommentHandler 处理用户端评论请求。
type CommentHandler struct {
	commentService service.CommentService
}

// NewCommentHandler 创建用户端评论处理器。
func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// Create 创建评论。
// @Summary 用户创建评论
// @Description 当前登录用户创建评论
// @Tags comment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "说说ID"
// @Param input body dto.CommentCreateInput true "评论创建参数"
// @Success 201 {object} dto.CommentDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "说说或用户不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /posts/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var uriReq dto.PostIDURIInput
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	var input dto.CommentCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.UserID = userID
	input.PostID = uriReq.ID
	input.CanViewHidden = canViewHiddenPost(c)

	result, err := h.commentService.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Get 查询当前用户评论详情。
// @Summary 用户查询评论详情
// @Description 当前登录用户按评论 ID 查询自己的评论详情
// @Tags comment
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Success 200 {object} dto.CommentDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "评论不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/comments/{id} [get]
func (h *CommentHandler) Get(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
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
	if result.Author.ID != userID {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrCommentNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update 更新当前用户评论。
// @Summary 用户更新评论
// @Description 当前登录用户按评论 ID 更新自己的评论
// @Tags comment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Param input body dto.CommentUpdateInput true "评论更新参数"
// @Success 200 {object} dto.CommentDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "评论不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/comments/{id} [patch]
func (h *CommentHandler) Update(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var uriReq dto.CommentIDURIInput
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	comment, err := h.commentService.Get(c.Request.Context(), uriReq.ID)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}
	if comment.Author.ID != userID {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrCommentNotFound.Error()})
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

// Delete 删除当前用户评论。
// @Summary 用户删除评论
// @Description 当前登录用户按评论 ID 删除自己的评论
// @Tags comment
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论ID"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "评论不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var req dto.CommentIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	comment, err := h.commentService.Get(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}
	if comment.Author.ID != userID {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrCommentNotFound.Error()})
		return
	}

	if err := h.commentService.Delete(c.Request.Context(), req.ID); err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}

// List 查询当前用户评论列表。
// @Summary 用户查询评论列表
// @Description 分页查询当前登录用户自己的评论列表
// @Tags comment
// @Produce json
// @Security BearerAuth
// @Param post_id query int false "说说ID"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at_desc 或 created_at_asc"
// @Success 200 {object} dto.CommentListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "说说不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var req dto.CommentUserListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.UserID = userID
	req.CanViewHidden = canViewHiddenPost(c)
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.commentService.ListByUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListPublic 查询公开评论列表。
// @Summary 前端查询评论列表
// @Description 分页查询公开评论列表
// @Tags comment
// @Produce json
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at_desc 或 created_at_asc"
// @Success 200 {object} dto.CommentListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /comments [get]
func (h *CommentHandler) ListPublic(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	var req dto.CommentPublicListQueryInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.commentService.ListPublic(c.Request.Context(), dto.CommentPublicListInput{
		CanViewHidden: canViewHiddenPost(c),
		Limit:         req.Limit,
		Offset:        req.Offset,
		Order:         req.Order,
	})
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListByPost 查询指定说说的评论列表（公开）。
// @Summary 前端查询说说评论列表
// @Description 分页查询指定说说下的评论（扁平）
// @Tags comment
// @Produce json
// @Param id path int true "说说ID"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at_asc 或 created_at_desc"
// @Success 200 {object} dto.CommentListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 404 {object} dto.ErrorResponse "说说不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /posts/{id}/comments [get]
func (h *CommentHandler) ListByPost(c *gin.Context) {
	if !ensureServiceReady(c, h.commentService) {
		return
	}

	var uriReq dto.PostIDURIInput
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	var queryReq dto.CommentPostListQueryInput
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}

	result, err := h.commentService.ListByPost(c.Request.Context(), dto.CommentListInput{
		PostID:        uriReq.ID,
		CanViewHidden: canViewHiddenPost(c),
		Limit:         queryReq.Limit,
		Offset:        queryReq.Offset,
		Order:         strings.TrimSpace(queryReq.Order),
	})
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func mapCommentErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrEmptyCommentContent),
		errors.Is(err, service.ErrInvalidCommentPost),
		errors.Is(err, service.ErrInvalidCommentUser),
		errors.Is(err, service.ErrInvalidCommentParent):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrCommentNotFound),
		errors.Is(err, service.ErrPostNotFound),
		errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrPostCommentDisabled):
		return http.StatusForbidden
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
