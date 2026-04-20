package handler

import (
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// InitHandler 处理系统初始化请求。
type InitHandler struct {
	initService    service.InitService
	errorResponder ErrorResponder
}

// NewInitHandler 创建系统初始化处理器。
func NewInitHandler(initService service.InitService, errorResponder ErrorResponder) *InitHandler {
	return &InitHandler{initService: initService, errorResponder: ensureErrorResponder(errorResponder)}
}

// Status 查询系统初始化状态。
// @Summary 查询系统初始化状态
// @Description 返回系统是否已初始化
// @Tags init
// @Produce json
// @Success 200 {object} dto.InitStatusResult
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /init/status [get]
func (h *InitHandler) Status(c *gin.Context) {
	if !ensureServiceReady(c, h.initService) {
		return
	}

	result, err := h.initService.GetStatus(c.Request.Context())
	if err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Initialize 初始化系统并创建首个管理员。
// @Summary 初始化系统
// @Description 仅在未初始化时创建首个管理员并设置站点配置
// @Tags init
// @Accept json
// @Produce json
// @Param input body dto.InitInput true "初始化参数"
// @Success 201 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 403 {object} dto.ErrorResponse "init secret 无效"
// @Failure 409 {object} dto.ErrorResponse "系统已初始化"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /init [post]
func (h *InitHandler) Initialize(c *gin.Context) {
	if !ensureServiceReady(c, h.initService) {
		return
	}

	var input dto.InitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	if err := h.initService.Initialize(c.Request.Context(), input); err != nil {
		h.errorResponder.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ErrorResponse{Message: "ok"})
}
