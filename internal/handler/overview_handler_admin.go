package handler

import (
	"net/http"

	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// OverviewAdminHandler 处理管理端概览聚合请求。
type OverviewAdminHandler struct {
	overviewService service.AdminOverviewService
	errorResponder  ErrorResponder
}

// NewOverviewAdminHandler 创建管理端概览处理器。
func NewOverviewAdminHandler(overviewService service.AdminOverviewService, errorResponder ErrorResponder) *OverviewAdminHandler {
	return &OverviewAdminHandler{overviewService: overviewService, errorResponder: ensureErrorResponder(errorResponder)}
}

// Get 查询后台概览聚合数据（管理端）
// @Summary 管理端查询概览聚合
// @Description 返回后台概览页所需的汇总指标与最近 7 天趋势
// @Tags overview-admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AdminOverviewResult
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/overview [get]
func (h *OverviewAdminHandler) Get(c *gin.Context) {
	if !ensureServiceReady(c, h.overviewService) {
		return
	}

	result, err := h.overviewService.Get(c.Request.Context())
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
