package handler

import (
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// PostHandler 处理前端公开说说请求。
type PostHandler struct {
	postService service.PostService
}

// NewPostHandler 创建前端公开说说处理器。
func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

// Get 查询公开说说详情。
// @Summary 前端查询说说详情
// @Description 根据说说 ID 查询公开说说详情
// @Tags post
// @Produce json
// @Param id path int true "说说ID"
// @Success 200 {object} dto.PostDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 404 {object} dto.ErrorResponse "说说不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /posts/{id} [get]
func (h *PostHandler) Get(c *gin.Context) {
	if h.postService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
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
	if result.Hide && !canViewHiddenPost(c) {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrPostNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// List 查询公开说说列表。
// @Summary 前端查询公开说说列表
// @Description 分页查询公开说说，支持按用户和标签过滤
// @Tags post
// @Produce json
// @Param user_id query int false "用户ID"
// @Param tag_id query int false "标签ID"
// @Param content query string false "内容关键字"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at_desc"
// @Success 200 {object} dto.PostPublicListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /posts [get]
func (h *PostHandler) List(c *gin.Context) {
	if h.postService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.PostListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	req.Order = strings.TrimSpace(req.Order)
	if !canViewHiddenPost(c) {
		publicOnly := false
		req.Hide = &publicOnly
	} else {
		req.Hide = nil
	}

	result, err := h.postService.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(mapPostErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	pinnedItems := make([]dto.PostDTO, 0, len(result.Items))
	normalItems := make([]dto.PostDTO, 0, len(result.Items))
	for _, item := range result.Items {
		if item.Pin != nil {
			pinnedItems = append(pinnedItems, item)
			continue
		}
		normalItems = append(normalItems, item)
	}

	c.JSON(http.StatusOK, dto.PostPublicListResult{
		PinnedItems: pinnedItems,
		Items:       normalItems,
		Total:       result.Total,
	})
}

func canViewHiddenPost(c *gin.Context) bool {
	if _, ok := middleware.GetUserID(c); !ok {
		return false
	}
	isAdmin, ok := middleware.GetUserAdmin(c)
	return ok && isAdmin
}
