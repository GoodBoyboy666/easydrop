package handler

import (
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// TagHandler 处理前端公开标签请求。
type TagHandler struct {
	tagService service.TagService
}

// NewTagHandler 创建前端公开标签处理器。
func NewTagHandler(tagService service.TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

// List 查询公开标签列表。
// @Summary 前端查询公开标签列表
// @Description 分页查询标签列表，支持按关键字过滤
// @Tags tag
// @Produce json
// @Param keyword query string false "标签关键字"
// @Param page query int false "页码，从 1 开始"
// @Param size query int false "每页条数，最大 100"
// @Param order query string false "排序，如 created_at_desc、hot_desc"
// @Success 200 {object} dto.TagListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /tags [get]
func (h *TagHandler) List(c *gin.Context) {
	if !ensureServiceReady(c, h.tagService) {
		return
	}

	var req dto.TagListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.Keyword = strings.TrimSpace(req.Keyword)
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.tagService.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
