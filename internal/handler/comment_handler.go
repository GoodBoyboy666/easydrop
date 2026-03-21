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
func (h *CommentHandler) Create(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var input dto.CommentCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.UserID = userID

	result, err := h.commentService.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapCommentErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Get 查询当前用户评论详情。
func (h *CommentHandler) Get(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
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
	if result.UserID != userID {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrCommentNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update 更新当前用户评论。
func (h *CommentHandler) Update(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
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
	if comment.UserID != userID {
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
func (h *CommentHandler) Delete(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
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
	if comment.UserID != userID {
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
func (h *CommentHandler) List(c *gin.Context) {
	if h.commentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
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
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.commentService.ListByUser(c.Request.Context(), req)
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
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
