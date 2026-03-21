package handler

import (
	"io"
	"net/http"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// UserAdminHandler 处理管理端用户请求。
type UserAdminHandler struct {
	userService service.UserService
}

// NewUserAdminHandler 创建管理端用户处理器。
func NewUserAdminHandler(userService service.UserService) *UserAdminHandler {
	return &UserAdminHandler{userService: userService}
}

// List 查询用户列表（管理端）
// @Summary 管理端查询用户列表
// @Description 分页查询用户列表，支持按用户名、邮箱、状态过滤
// @Tags user-admin
// @Produce json
// @Security BearerAuth
// @Param username query string false "用户名（模糊匹配）"
// @Param email query string false "邮箱（模糊匹配）"
// @Param status query int false "用户状态"
// @Param limit query int false "分页大小"
// @Param offset query int false "偏移量"
// @Param order query string false "排序，如 id desc"
// @Success 200 {object} dto.UserListResult
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/users [get]
func (h *UserAdminHandler) List(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.UserListInput
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "查询参数不合法"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Order = strings.TrimSpace(req.Order)

	result, err := h.userService.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Create 创建用户（管理端）
// @Summary 管理端创建用户
// @Description 创建用户并返回用户信息
// @Tags user-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body dto.UserCreateInput true "用户创建参数"
// @Success 201 {object} dto.UserDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 409 {object} dto.ErrorResponse "用户名或邮箱冲突"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/users [post]
func (h *UserAdminHandler) Create(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var input dto.UserCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}

	result, err := h.userService.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Update 更新用户（管理端）
// @Summary 管理端更新用户
// @Description 根据用户 ID 更新用户信息
// @Tags user-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param input body dto.UserUpdateInput true "用户更新参数"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 409 {object} dto.ErrorResponse "用户名或邮箱冲突"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/users/{id} [patch]
func (h *UserAdminHandler) Update(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var uriReq dto.UserIDURIInput
	if err := c.ShouldBindUri(&uriReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	var input dto.UserUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "请求参数格式错误"})
		return
	}
	input.ID = uriReq.ID

	result, err := h.userService.Update(c.Request.Context(), input)
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete 删除用户（管理端）
// @Summary 管理端删除用户
// @Description 根据用户 ID 删除用户
// @Tags user-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/users/{id} [delete]
func (h *UserAdminHandler) Delete(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.UserIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	if err := h.userService.Delete(c.Request.Context(), req.ID); err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}

// UploadAvatar 上传用户头像（管理端）
// @Summary 管理端上传用户头像
// @Description 为指定用户上传并替换头像
// @Tags user-admin
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param avatar formData file true "头像文件"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限或配额不足"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/users/{id}/avatar [post]
func (h *UserAdminHandler) UploadAvatar(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.UserIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "avatar 不能为空"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "读取上传文件失败"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "读取上传文件失败"})
		return
	}

	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	result, err := h.userService.UploadAvatar(c.Request.Context(), dto.UserAvatarUploadInput{
		UserID:           req.ID,
		OriginalFilename: fileHeader.Filename,
		ContentType:      contentType,
		Content:          content,
	})
	if err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteAvatar 删除用户头像（管理端）
// @Summary 管理端删除用户头像
// @Description 删除指定用户头像并更新存储占用
// @Tags user-admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse "参数校验失败"
// @Failure 401 {object} dto.ErrorResponse "未登录或登录失效"
// @Failure 403 {object} dto.ErrorResponse "无管理员权限"
// @Failure 500 {object} dto.ErrorResponse "服务内部错误"
// @Router /admin/users/{id}/avatar [delete]
func (h *UserAdminHandler) DeleteAvatar(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
		return
	}

	var req dto.UserIDURIInput
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "路径参数不合法"})
		return
	}

	if err := h.userService.DeleteAvatar(c.Request.Context(), req.ID); err != nil {
		c.JSON(mapUserErrorStatus(err), dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ErrorResponse{Message: "ok"})
}
