package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// AttachmentHandler 处理附件相关请求。
type AttachmentHandler struct {
	attachmentService service.AttachmentService
}

type attachmentURIRequest struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type attachmentListQueryRequest struct {
	ID      *uint  `form:"id" binding:"omitempty,gt=0"`
	BizType *int   `form:"biz_type"`
	Limit   *int   `form:"limit"`
	Offset  *int   `form:"offset"`
	Order   string `form:"order"`
}

// NewAttachmentHandler 创建附件处理器。
func NewAttachmentHandler(attachmentService service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: attachmentService}
}

// Upload 上传附件
// @Summary 上传附件
// @Description 以 multipart/form-data 方式上传附件并返回附件信息
// @Tags attachment
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "附件文件"
// @Success 201 {object} dto.AttachmentDTO
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "存储配额不足"
// @Failure 404 {object} MessageResponse "用户不存在"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/attachments [post]
func (h *AttachmentHandler) Upload(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "file 不能为空"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "读取上传文件失败"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "读取上传文件失败"})
		return
	}

	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	result, err := h.attachmentService.Create(c.Request.Context(), dto.AttachmentCreateInput{
		UserID:      userID,
		ContentType: contentType,
		Content:     content,
	})
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Get 查询附件详情
// @Summary 查询附件详情
// @Description 根据附件 ID 查询当前用户的附件详情
// @Tags attachment
// @Produce json
// @Security BearerAuth
// @Param id path int true "附件ID"
// @Success 200 {object} dto.AttachmentDTO
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 404 {object} MessageResponse "附件不存在"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/attachments/{id} [get]
func (h *AttachmentHandler) Get(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	var req attachmentURIRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "路径参数不合法"})
		return
	}

	result, err := h.attachmentService.Get(c.Request.Context(), req.ID)
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
		return
	}

	if result.UserID != userID {
		c.JSON(http.StatusNotFound, MessageResponse{Message: service.ErrAttachmentNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// List 查询当前用户附件列表
// @Summary 查询附件列表
// @Description 分页查询当前用户附件列表
// @Tags attachment
// @Produce json
// @Security BearerAuth
// @Param id query int false "附件ID"
// @Param biz_type query int false "附件业务类型"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 created_at desc"
// @Success 200 {object} dto.AttachmentListResult
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/attachments [get]
func (h *AttachmentHandler) List(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	var req attachmentListQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "查询参数不合法"})
		return
	}

	result, err := h.attachmentService.ListByUser(c.Request.Context(), dto.AttachmentListInput{
		ID:      req.ID,
		UserID:  &userID,
		BizType: req.BizType,
		Limit:   valueOrDefault(req.Limit, 0),
		Offset:  valueOrDefault(req.Offset, 0),
		Order:   strings.TrimSpace(req.Order),
	})
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete 删除附件
// @Summary 删除附件
// @Description 删除当前用户的附件
// @Tags attachment
// @Produce json
// @Security BearerAuth
// @Param id path int true "附件ID"
// @Success 200 {object} MessageResponse
// @Failure 40di0 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 404 {object} MessageResponse "附件不存在"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/attachments/{id} [delete]
func (h *AttachmentHandler) Delete(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	var req attachmentURIRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "路径参数不合法"})
		return
	}

	attachment, err := h.attachmentService.Get(c.Request.Context(), req.ID)
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
		return
	}

	if attachment.UserID != userID {
		c.JSON(http.StatusNotFound, MessageResponse{Message: service.ErrAttachmentNotFound.Error()})
		return
	}

	if err := h.attachmentService.Delete(c.Request.Context(), req.ID); err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

func valueOrDefault(v *int, fallback int) int {
	if v == nil {
		return fallback
	}
	return *v
}

func mapAttachmentErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrInvalidAttachmentBizType),
		errors.Is(err, service.ErrInvalidFileSize),
		errors.Is(err, service.ErrEmptyAttachmentContent):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrStorageQuotaExceeded):
		return http.StatusForbidden
	case errors.Is(err, service.ErrAttachmentNotFound),
		errors.Is(err, service.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
