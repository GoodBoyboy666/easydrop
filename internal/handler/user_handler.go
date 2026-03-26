package handler

import (
	"errors"
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

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
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	result, err := h.userService.Get(c.Request.Context(), userID)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
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
// @Param input body dto.UserProfileUpdateInput true "资料修改参数"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.ErrorResponse "请求参数格式错误"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/profile [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var input dto.UserProfileUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.UserID = userID

	result, err := h.userService.UpdateProfile(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
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
// @Param input body dto.UserChangePasswordInput true "密码修改参数"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "旧密码错误或未登录"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/password [patch]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var input dto.UserChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.UserID = userID

	err := h.userService.ChangePassword(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}

// RequestEmailChange 请求修改当前用户邮箱。
// @Summary 请求修改当前用户邮箱
// @Description 校验当前密码后发起邮箱修改请求并发送确认邮件
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.UserChangeEmailInput true "邮箱修改请求参数"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "密码错误或未登录"
// @Failure 409 {object} dto.ErrorResponse "邮箱已被占用"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/email-change [post]
func (h *UserHandler) RequestEmailChange(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	var input dto.UserChangeEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.UserID = userID

	err := h.userService.RequestEmailChange(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
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
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "存储配额不足"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		if isRequestTooLargeError(err) {
			c.JSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse{Message: "请求体过大"})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "avatar 不能为空"})
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

	result, err := h.userService.UploadAvatar(c.Request.Context(), dto.UserAvatarUploadInput{
		UserID:           userID,
		OriginalFilename: fileHeader.Filename,
		ContentType:      contentType,
		FileSize:         fileHeader.Size,
		Content:          file,
		ContentSample:    sample,
	})
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
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
// @Success 200 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /users/me/avatar [delete]
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "未登录或登录已失效"})
		return
	}

	if err := h.userService.DeleteAvatar(c.Request.Context(), userID); err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
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
		errors.Is(err, service.ErrEmptyAvatarFilename),
		errors.Is(err, service.ErrAvatarExtensionNotAllowed),
		errors.Is(err, service.ErrAvatarMIMETypeNotAllowed):
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
