package handler

import (
	"errors"
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// InitHandler 处理系统初始化请求。
type InitHandler struct {
	initService service.InitService
}

// NewInitHandler 创建系统初始化处理器。
func NewInitHandler(initService service.InitService) *InitHandler {
	return &InitHandler{initService: initService}
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
		c.JSON(mapInitErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
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
		c.JSON(mapInitErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ErrorResponse{Message: "ok"})
}

func mapInitErrorStatus(err error) int {
	switch {
	case errors.Is(err, validator.ErrEmptyUsername),
		errors.Is(err, validator.ErrUsernameTooShort),
		errors.Is(err, validator.ErrUsernameTooLong),
		errors.Is(err, validator.ErrInvalidUsernameFormat),
		errors.Is(err, validator.ErrEmptyEmail),
		errors.Is(err, validator.ErrInvalidEmailFormat),
		errors.Is(err, validator.ErrEmptyPassword),
		errors.Is(err, validator.ErrPasswordTooShort),
		errors.Is(err, validator.ErrPasswordContainsSpace),
		errors.Is(err, validator.ErrPasswordMissingLetter),
		errors.Is(err, validator.ErrPasswordMissingNumber),
		errors.Is(err, service.ErrInvalidSiteSetting):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrAlreadyInitialized),
		errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrEmailExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
