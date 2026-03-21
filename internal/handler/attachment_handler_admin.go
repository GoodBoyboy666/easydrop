package handler

import (
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// AttachmentAdminHandler 处理管理端附件请求。
type AttachmentAdminHandler struct {
	attachmentService service.AttachmentService
}

// NewAttachmentAdminHandler 创建管理端附件处理器。
func NewAttachmentAdminHandler(attachmentService service.AttachmentService) *AttachmentAdminHandler {
	return &AttachmentAdminHandler{attachmentService: attachmentService}
}

// List 查询附件列表（管理端）
// @Summary 管理端查询附件列表
// @Description 分页查询附件列表，支持按用户、业务类型、上传时间区间过滤
// @Tags attachment-admin
// @Produce json
// @Security BearerAuth
// @Param id query int false "附件ID"
// @Param user_id query int false "用户ID"
// @Param biz_type query int false "附件业务类型"
// @Param created_from query int false "上传时间起点（Unix秒）"
// @Param created_to query int false "上传时间终点（Unix秒）"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at_desc"
// @Success 200 {object} dto.AttachmentListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/attachments [get]
func (h *AttachmentAdminHandler) List(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.AttachmentListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}

	if req.CreatedFrom != nil && req.CreatedTo != nil && *req.CreatedFrom > *req.CreatedTo {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "created_from 不能大于 created_to"})
		return
	}

	req.Order = strings.TrimSpace(req.Order)
	result, err := h.attachmentService.ListByUser(c.Request.Context(), req)
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete 删除附件（管理端）
// @Summary 管理端删除附件
// @Description 按附件 ID 删除任意用户附件
// @Tags attachment-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "附件ID"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 404 {object} dto.ErrorResponse "附件不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/attachments/{id} [delete]
func (h *AttachmentAdminHandler) Delete(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.AttachmentIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	if err := h.attachmentService.Delete(c.Request.Context(), req.ID); err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}

// BatchDelete 批量删除附件（管理端）
// @Summary 管理端批量删除附件
// @Description 批量删除附件，返回成功与失败明细
// @Tags attachment-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.AttachmentBatchDeleteInput true "批量删除请求"
// @Success 200 {object} dto.AttachmentBatchDeleteResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/attachments/batch-delete [post]
func (h *AttachmentAdminHandler) BatchDelete(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.AttachmentBatchDeleteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求体不合法"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "ids 不能为空"})
		return
	}

	resp := dto.AttachmentBatchDeleteResult{
		SuccessIDs: make([]uint, 0, len(req.IDs)),
		Failed:     make([]dto.AttachmentBatchDeleteFailedItem, 0),
	}

	for _, id := range req.IDs {
		if id == 0 {
			resp.Failed = append(resp.Failed, dto.AttachmentBatchDeleteFailedItem{ID: id, Message: "id 参数不合法"})
			continue
		}

		if err := h.attachmentService.Delete(c.Request.Context(), id); err != nil {
			resp.Failed = append(resp.Failed, dto.AttachmentBatchDeleteFailedItem{ID: id, Message: err.Error()})
			continue
		}

		resp.SuccessIDs = append(resp.SuccessIDs, id)
	}

	c.JSON(http.StatusOK, resp)
}
