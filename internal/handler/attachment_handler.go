package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// AttachmentHandler 处理附件相关请求。
type AttachmentHandler struct {
	attachmentService service.AttachmentService
	settingService    service.SettingService
}

// NewAttachmentHandler 创建附件处理器。
func NewAttachmentHandler(attachmentService service.AttachmentService, settingService service.SettingService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: attachmentService, settingService: settingService}
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
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "存储配额不足"
// @Failure 404 {object} dto.ErrorResponse "用户不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /attachments [post]
func (h *AttachmentHandler) Upload(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	allowUpload, err := h.isUserUploadAllowed(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}
	if !allowUpload {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Message: "已关闭用户上传附件"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if isRequestTooLargeError(err) {
			c.JSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse{Message: "请求体过大"})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "file 不能为空"})
		return
	}

	file, sample, contentType, err := openUploadFile(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "读取上传文件失败"})
		return
	}
	defer func() {
		_ = file.Close()
	}()

	result, err := h.attachmentService.Create(c.Request.Context(), dto.AttachmentCreateInput{
		UserID:           userID,
		OriginalFilename: fileHeader.Filename,
		ContentType:      contentType,
		FileSize:         fileHeader.Size,
		Content:          file,
		ContentSample:    sample,
	})
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
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
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "附件不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /attachments/{id} [get]
func (h *AttachmentHandler) Get(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var req dto.AttachmentIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	result, err := h.attachmentService.Get(c.Request.Context(), req.ID)
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	if result.UserID != userID {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrAttachmentNotFound.Error()})
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
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /attachments [get]
func (h *AttachmentHandler) List(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var input dto.AttachmentListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	input.UserID = &userID
	input.Order = strings.TrimSpace(input.Order)

	result, err := h.attachmentService.ListByUser(c.Request.Context(), input)
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
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
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 404 {object} dto.ErrorResponse "附件不存在"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /attachments/{id} [delete]
func (h *AttachmentHandler) Delete(c *gin.Context) {
	if h.attachmentService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var req dto.AttachmentIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	attachment, err := h.attachmentService.Get(c.Request.Context(), req.ID)
	if err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	if attachment.UserID != userID {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: service.ErrAttachmentNotFound.Error()})
		return
	}

	if err := h.attachmentService.Delete(c.Request.Context(), req.ID); err != nil {
		status := mapAttachmentErrorStatus(err)
		c.JSON(status, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}

func mapAttachmentErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrInvalidAttachmentBizType),
		errors.Is(err, service.ErrInvalidFileSize),
		errors.Is(err, service.ErrEmptyAttachmentContent),
		errors.Is(err, service.ErrAttachmentExtensionsNotConfigured),
		errors.Is(err, service.ErrAttachmentExtensionNotAllowed),
		errors.Is(err, service.ErrAttachmentMIMETypeNotAllowed):
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

func (h *AttachmentHandler) isUserUploadAllowed(c *gin.Context) (bool, error) {
	if h == nil || h.settingService == nil {
		return true, nil
	}

	value, found, err := h.settingService.GetValue(c.Request.Context(), "storage.upload")
	if err != nil {
		return false, err
	}
	if !found {
		return true, nil
	}

	uploadEnabled, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, err
	}
	if uploadEnabled {
		return true, nil
	}

	isAdmin, _ := middleware.GetUserAdmin(c)
	return isAdmin, nil
}
