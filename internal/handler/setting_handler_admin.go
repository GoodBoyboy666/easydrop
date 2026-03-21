package handler

import (
	"errors"
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingAdminHandler 处理管理端配置请求。
type SettingAdminHandler struct {
	settingService service.SettingService
}

type settingAdminKeyURIRequest struct {
	Key string `uri:"key" binding:"required"`
}

type settingAdminListQueryRequest struct {
	Category string `form:"category"`
	Key      string `form:"key"`
	Limit    *int   `form:"limit"`
	Offset   *int   `form:"offset"`
	Order    string `form:"order"`
}

type settingAdminUpdateRequest struct {
	Value *string `json:"value"`
}

// NewSettingAdminHandler 创建管理端配置处理器。
func NewSettingAdminHandler(settingService service.SettingService) *SettingAdminHandler {
	return &SettingAdminHandler{settingService: settingService}
}

// List 查询配置列表（管理端）
// @Summary 管理端查询配置列表
// @Description 分页查询配置列表，支持按分类和键名过滤
// @Tags setting-admin
// @Produce json
// @Security BearerAuth
// @Param category query string false "分类"
// @Param key query string false "键名（模糊匹配）"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 key_asc"
// @Success 200 {object} dto.SettingListResult
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "无管理员权限"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/admin/settings [get]
func (h *SettingAdminHandler) List(c *gin.Context) {
	if h.settingService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req settingAdminListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "查询参数不合法"})
		return
	}

	result, err := h.settingService.ListItems(c.Request.Context(), dto.SettingListInput{
		Category: strings.TrimSpace(req.Category),
		Key:      strings.TrimSpace(req.Key),
		Limit:    valueOrDefault(req.Limit, 0),
		Offset:   valueOrDefault(req.Offset, 0),
		Order:    strings.TrimSpace(req.Order),
	})
	if err != nil {
		c.JSON(mapSettingErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Public 查询公开配置（前端）
// @Summary 前端公开配置
// @Description 返回前端可公开读取的站点配置
// @Tags setting
// @Produce json
// @Success 200 {object} dto.SettingPublicResult
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/settings/public [get]
func (h *SettingAdminHandler) Public(c *gin.Context) {
	if h.settingService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	result, err := h.settingService.GetPublicItems(c.Request.Context())
	if err != nil {
		c.JSON(mapSettingErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update 按 key 更新配置（管理端）
// @Summary 管理端更新配置
// @Description 按配置 key 更新 value
// @Tags setting-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Param input body settingAdminUpdateRequest true "更新参数"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "无管理员权限"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/admin/settings/{key} [patch]
func (h *SettingAdminHandler) Update(c *gin.Context) {
	if h.settingService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	var uriReq settingAdminKeyURIRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "路径参数不合法"})
		return
	}

	var req settingAdminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "请求参数格式错误"})
		return
	}

	if err := h.settingService.UpdateItem(c.Request.Context(), dto.SettingUpdateInput{
		Key:   strings.TrimSpace(uriReq.Key),
		Value: req.Value,
	}); err != nil {
		c.JSON(mapSettingErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

func mapSettingErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrSettingKeyRequired):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
