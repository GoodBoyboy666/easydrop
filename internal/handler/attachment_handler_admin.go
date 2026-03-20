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

type attachmentAdminURIRequest struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type attachmentAdminListQueryRequest struct {
	ID          *uint  `form:"id" binding:"omitempty,gt=0"`
	UserID      *uint  `form:"user_id" binding:"omitempty,gt=0"`
	BizType     *int   `form:"biz_type"`
	CreatedFrom *int64 `form:"created_from" binding:"omitempty,gte=0"`
	CreatedTo   *int64 `form:"created_to" binding:"omitempty,gte=0"`
	Limit       *int   `form:"limit"`
	Offset      *int   `form:"offset"`
	Order       string `form:"order"`
}

// NewAttachmentAdminHandler 创建管理端附件处理器。
func NewAttachmentAdminHandler(attachmentService service.AttachmentService) *AttachmentAdminHandler {
	return &AttachmentAdminHandler{attachmentService: attachmentService}
}

type attachmentBatchDeleteRequest struct {
	IDs []uint `json:"ids"`
}

type attachmentBatchDeleteFailedItem struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type attachmentBatchDeleteResponse struct {
	SuccessIDs []uint                            `json:"success_ids"`
	Failed     []attachmentBatchDeleteFailedItem `json:"failed"`
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
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "无管理员权限"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/admin/attachments [get]
func (h *AttachmentAdminHandler) List(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req attachmentAdminListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "查询参数不合法"})
		return
	}

	if req.CreatedFrom != nil && req.CreatedTo != nil && *req.CreatedFrom > *req.CreatedTo {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "created_from 不能大于 created_to"})
		return
	}

	result, err := h.attachmentService.ListByUser(c.Request.Context(), dto.AttachmentListInput{
		ID:          req.ID,
		UserID:      req.UserID,
		BizType:     req.BizType,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
		Limit:       valueOrDefault(req.Limit, 0),
		Offset:      valueOrDefault(req.Offset, 0),
		Order:       strings.TrimSpace(req.Order),
	})
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
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
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "无管理员权限"
// @Failure 404 {object} MessageResponse "附件不存在"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/admin/attachments/{id} [delete]
func (h *AttachmentAdminHandler) Delete(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req attachmentAdminURIRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "路径参数不合法"})
		return
	}

	if err := h.attachmentService.Delete(c.Request.Context(), req.ID); err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// BatchDelete 批量删除附件（管理端）
// @Summary 管理端批量删除附件
// @Description 批量删除附件，返回成功与失败明细
// @Tags attachment-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body attachmentBatchDeleteRequest true "批量删除请求"
// @Success 200 {object} attachmentBatchDeleteResponse
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "无管理员权限"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/admin/attachments/batch-delete [post]
func (h *AttachmentAdminHandler) BatchDelete(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req attachmentBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "请求体不合法"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "ids 不能为空"})
		return
	}

	resp := attachmentBatchDeleteResponse{
		SuccessIDs: make([]uint, 0, len(req.IDs)),
		Failed:     make([]attachmentBatchDeleteFailedItem, 0),
	}

	for _, id := range req.IDs {
		if id == 0 {
			resp.Failed = append(resp.Failed, attachmentBatchDeleteFailedItem{ID: id, Message: "id 参数不合法"})
			continue
		}

		if err := h.attachmentService.Delete(c.Request.Context(), id); err != nil {
			resp.Failed = append(resp.Failed, attachmentBatchDeleteFailedItem{ID: id, Message: err.Error()})
			continue
		}

		resp.SuccessIDs = append(resp.SuccessIDs, id)
	}

	c.JSON(http.StatusOK, resp)
}
