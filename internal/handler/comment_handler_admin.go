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
