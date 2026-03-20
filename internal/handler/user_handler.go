package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type userProfileUpdateRequest struct {
	Nickname *string `json:"nickname"`
}

type userChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type userChangeEmailRequest struct {
	CurrentPassword string `json:"current_password"`
	NewEmail        string `json:"new_email"`
}

// UserHandler 处理当前登录用户的自助接口。
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile 获取当前用户资料。
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的资料信息
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserDTO
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	result, err := h.userService.Get(c.Request.Context(), userID)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateProfile 修改当前用户资料（仅昵称）。
// @Summary 修改当前用户资料
// @Description 修改当前登录用户资料（当前仅支持昵称）
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body userProfileUpdateRequest true "资料修改参数"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} MessageResponse "请求参数格式错误"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/users/me/profile [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	var req userProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "请求参数格式错误"})
		return
	}

	result, err := h.userService.UpdateProfile(c.Request.Context(), dto.UserProfileUpdateInput{
		UserID:   userID,
		Nickname: req.Nickname,
	})
	if err != nil {
		c.JSON(mapUserErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ChangePassword 修改当前用户密码。
// @Summary 修改当前用户密码
// @Description 使用旧密码校验后修改当前登录用户密码
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body userChangePasswordRequest true "密码修改参数"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "旧密码错误或未登录"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/users/me/password [patch]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	var req userChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "请求参数格式错误"})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), dto.UserChangePasswordInput{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		c.JSON(mapUserErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// RequestEmailChange 请求修改当前用户邮箱。
// @Summary 请求修改当前用户邮箱
// @Description 校验当前密码后发起邮箱修改请求并发送确认邮件
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body userChangeEmailRequest true "邮箱修改请求参数"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "密码错误或未登录"
// @Failure 409 {object} MessageResponse "邮箱已被占用"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/users/me/email-change [post]
func (h *UserHandler) RequestEmailChange(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	var req userChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "请求参数格式错误"})
		return
	}

	err := h.userService.RequestEmailChange(c.Request.Context(), dto.UserChangeEmailRequestInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewEmail:        req.NewEmail,
	})
	if err != nil {
		c.JSON(mapUserErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

// UploadAvatar 上传当前用户头像。
// @Summary 上传当前用户头像
// @Description 以 multipart/form-data 上传并替换当前登录用户头像
// @Tags user
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "头像文件"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} MessageResponse "参数校验失败"
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 403 {object} MessageResponse "存储配额不足"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: "avatar 不能为空"})
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

	result, err := h.userService.UploadAvatar(c.Request.Context(), dto.UserAvatarUploadInput{
		UserID:           userID,
		OriginalFilename: fileHeader.Filename,
		ContentType:      contentType,
		Content:          content,
	})
	if err != nil {
		c.JSON(mapUserErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteAvatar 删除当前用户头像。
// @Summary 删除当前用户头像
// @Description 删除当前登录用户头像并更新存储占用
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 401 {object} MessageResponse "未登录或登录失效"
// @Failure 500 {object} MessageResponse "服务内部错误"
// @Router /api/v1/users/me/avatar [delete]
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, MessageResponse{Message: "未登录或登录已失效"})
		return
	}

	if err := h.userService.DeleteAvatar(c.Request.Context(), userID); err != nil {
		c.JSON(mapUserErrorStatus(err), MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "ok"})
}

func mapUserErrorStatus(err error) int {
	switch {
	case errors.Is(err, validator.ErrEmptyPassword),
		errors.Is(err, validator.ErrPasswordTooShort),
		errors.Is(err, validator.ErrPasswordContainsSpace),
		errors.Is(err, validator.ErrPasswordMissingLetter),
		errors.Is(err, validator.ErrPasswordMissingNumber),
		errors.Is(err, validator.ErrEmptyEmail),
		errors.Is(err, validator.ErrInvalidEmailFormat),
		errors.Is(err, service.ErrInvalidEmailChange),
		errors.Is(err, service.ErrEmptyAvatarContent),
		errors.Is(err, service.ErrEmptyAvatarFilename):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrInvalidPassword):
		return http.StatusUnauthorized
	case errors.Is(err, service.ErrUserDisabled),
		errors.Is(err, service.ErrStorageQuotaExceeded):
		return http.StatusForbidden
	case errors.Is(err, service.ErrEmailExists),
		errors.Is(err, service.ErrUsernameExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
