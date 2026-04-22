package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	avatarcfg "easydrop/internal/pkg/avatar"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/pkg/token"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidUserStatus   = errors.New("用户状态不合法")
	ErrInvalidStorageQuota = errors.New("存储配额不能为负数")
	ErrEmptyAvatarContent  = errors.New("头像内容不能为空")
	ErrEmptyAvatarFilename = errors.New("头像文件名不能为空")
	ErrInvalidEmailChange  = errors.New("邮箱变更凭证无效或已过期")
	ErrEmailChanged        = errors.New("当前邮箱与变更请求不一致")
)

const emailChangeTokenTTL = 30 * time.Minute

// UserService 提供用户管理相关的 CRUD 能力。
type UserService interface {
	// Create 创建用户。
	Create(ctx context.Context, input dto.UserCreateInput) (*dto.UserDTO, error)
	// Get 按 ID 获取用户详情。
	Get(ctx context.Context, id uint) (*dto.UserDTO, error)
	// UpdateProfile 修改用户 Profile（仅昵称）。
	UpdateProfile(ctx context.Context, input dto.UserProfileUpdateInput) (*dto.UserDTO, error)
	// ChangePassword 修改用户密码，需要校验旧密码。
	ChangePassword(ctx context.Context, input dto.UserChangePasswordInput) error
	// RequestEmailChange 发起邮箱变更请求，签发 token 并发送邮件。
	RequestEmailChange(ctx context.Context, input dto.UserChangeEmailInput) error
	// ConfirmEmailChange 验证邮箱变更 token 并更新。
	ConfirmEmailChange(ctx context.Context, input dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error)
	// Update 按输入字段更新用户信息。
	Update(ctx context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error)
	// UploadAvatar 上传或替换用户头像。
	UploadAvatar(ctx context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error)
	// DeleteAvatar 删除用户头像。
	DeleteAvatar(ctx context.Context, userID uint) error
	// Delete 删除指定用户。
	Delete(ctx context.Context, id uint) error
	// List 按筛选条件查询用户列表。
	List(ctx context.Context, input dto.UserListInput) (*dto.UserListResult, error)
}

type emailChangeTokenPayload struct {
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

type userService struct {
	userRepo        repo.UserRepo
	storageManager  storage.Manager
	settings        SettingService
	tokenManager    token.Manager
	emailService    EmailService
	gravatarBaseURL string
}

// NewUserService 创建用户服务实例。
func NewUserService(userRepo repo.UserRepo, storageManager storage.Manager, settings SettingService, tokenManager token.Manager, emailService EmailService, avatarConfig *avatarcfg.Config) UserService {
	gravatarBaseURL := ""
	if avatarConfig != nil {
		gravatarBaseURL = avatarConfig.GravatarBaseURL
	}

	return &userService{
		userRepo:        userRepo,
		storageManager:  storageManager,
		settings:        settings,
		tokenManager:    tokenManager,
		emailService:    emailService,
		gravatarBaseURL: normalizeGravatarBaseURL(gravatarBaseURL),
	}
}

// Create 校验输入后创建新用户，并对密码进行哈希。
func (s *userService) Create(ctx context.Context, input dto.UserCreateInput) (*dto.UserDTO, error) {
	// 先完成基础字段校验和规范化。
	username := strings.TrimSpace(input.Username)
	if err := validator.ValidateUsername(username); err != nil {
		return nil, err
	}

	email := strings.TrimSpace(input.Email)
	if err := validator.ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := validator.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	nickname := strings.TrimSpace(input.Nickname)
	if nickname == "" {
		nickname = username
	}

	status := 1
	if input.Status != nil {
		status = *input.Status
	}
	if err := validateUserStatus(status); err != nil {
		return nil, err
	}

	storageQuota, err := normalizeStorageQuota(input.StorageQuota)
	if err != nil {
		return nil, err
	}

	avatar := normalizeOptionalString(input.Avatar)

	// 校验用户名与邮箱唯一性。
	if err := s.ensureUsernameAvailable(ctx, username, 0); err != nil {
		return nil, err
	}
	if err := s.ensureEmailAvailable(ctx, email, 0); err != nil {
		return nil, err
	}

	// 生成密码哈希并写入用户记录。
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		return nil, ErrInternal
	}

	user := &model.User{
		Username:      username,
		Nickname:      nickname,
		Email:         email,
		Password:      string(hash),
		Admin:         input.Admin,
		Status:        status,
		Avatar:        avatar,
		EmailVerified: input.EmailVerified,
		StorageQuota:  storageQuota,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("创建用户失败: %v", err)
		return nil, ErrInternal
	}

	// 转换为对外 DTO。
	userDTO, err := toUserDTO(ctx, user, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户 DTO 失败: %v", err)
		return nil, ErrInternal
	}
	return &userDTO, nil
}

// Get 根据用户 ID 查询单个用户详情。
func (s *userService) Get(ctx context.Context, id uint) (*dto.UserDTO, error) {
	if id == 0 {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	userDTO, err := toUserDTO(ctx, user, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户 DTO 失败: %v", err)
		return nil, ErrInternal
	}
	return &userDTO, nil
}

// UpdateProfile 修改用户 Profile（仅昵称）。
func (s *userService) UpdateProfile(ctx context.Context, input dto.UserProfileUpdateInput) (*dto.UserDTO, error) {
	if input.UserID == 0 {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	if input.Nickname != nil {
		nickname := strings.TrimSpace(*input.Nickname)
		if nickname == "" {
			nickname = user.Username
		}
		user.Nickname = nickname

		if err := s.userRepo.Update(ctx, user); err != nil {
			log.Printf("更新用户资料失败: %v", err)
			return nil, ErrInternal
		}
	}

	userDTO, err := toUserDTO(ctx, user, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户 DTO 失败: %v", err)
		return nil, ErrInternal
	}
	return &userDTO, nil
}

// ChangePassword 修改用户密码，并强制校验旧密码。
func (s *userService) ChangePassword(ctx context.Context, input dto.UserChangePasswordInput) error {
	if input.UserID == 0 {
		return ErrUserNotFound
	}
	if strings.TrimSpace(input.OldPassword) == "" {
		return validator.ErrEmptyPassword
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		return ErrInvalidPassword
	}

	if err := validator.ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		return ErrInternal
	}

	user.Password = string(hash)
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("更新用户密码失败: %v", err)
		return ErrInternal
	}

	return nil
}

// RequestEmailChange 签发邮箱变更 token，并发送确认邮件。
func (s *userService) RequestEmailChange(ctx context.Context, input dto.UserChangeEmailInput) error {
	// 校验关键依赖与用户存在性。
	if input.UserID == 0 {
		return ErrUserNotFound
	}
	if s.tokenManager == nil || s.emailService == nil {
		return ErrInternal
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}

	// 校验当前密码与新邮箱合法性。
	currentPassword := strings.TrimSpace(input.CurrentPassword)
	if currentPassword == "" {
		return validator.ErrEmptyPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrInvalidPassword
	}

	newEmail := strings.TrimSpace(input.NewEmail)
	if err := validator.ValidateEmail(newEmail); err != nil {
		return err
	}
	if err := s.ensureEmailAvailable(ctx, newEmail, user.ID); err != nil {
		return err
	}

	// 组装 payload、签发 token 并发送确认邮件。
	payload, err := marshalEmailChangePayload(strings.TrimSpace(user.Email), newEmail)
	if err != nil {
		log.Printf("构建邮箱变更 payload 失败: %v", err)
		return ErrInternal
	}

	verifyToken, err := s.tokenManager.Issue(ctx, user.ID, token.KindChangeEmail, emailChangeTokenTTL, payload)
	if err != nil {
		log.Printf("签发邮箱变更 token 失败: %v", err)
		return ErrInternal
	}

	if err := s.emailService.SendChangeEmailEmail(ctx, newEmail, newEmail, verifyToken, emailChangeTokenTTL); err != nil {
		log.Printf("发送邮箱变更邮件失败: %v", err)
		return ErrInternal
	}

	return nil
}

// ConfirmEmailChange 校验邮箱变更 token 后更新邮箱。
func (s *userService) ConfirmEmailChange(ctx context.Context, input dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error) {
	// 消费邮箱变更 token，并映射为统一业务错误。
	if s.tokenManager == nil {
		return nil, ErrInternal
	}

	verifyToken := strings.TrimSpace(input.VerificationToken)
	record, err := s.tokenManager.Consume(ctx, token.KindChangeEmail, verifyToken)
	if err != nil {
		switch {
		case errors.Is(err, token.ErrEmptyToken),
			errors.Is(err, token.ErrTokenNotFound),
			errors.Is(err, token.ErrTokenMismatch),
			errors.Is(err, token.ErrTokenExpired):
			return nil, ErrInvalidEmailChange
		default:
			log.Printf("消费邮箱变更 token 失败: %v", err)
			return nil, ErrInternal
		}
	}

	// 解析 token 载荷并读取目标用户。
	payload, err := unmarshalEmailChangePayload(record.Payload)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, record.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	// 校验旧邮箱一致性与新邮箱可用性。
	if !strings.EqualFold(strings.TrimSpace(user.Email), payload.OldEmail) {
		return nil, ErrEmailChanged
	}

	if err := s.ensureEmailAvailable(ctx, payload.NewEmail, user.ID); err != nil {
		return nil, err
	}

	user.Email = payload.NewEmail
	user.EmailVerified = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("更新用户邮箱失败: %v", err)
		return nil, ErrInternal
	}

	// 返回更新后的用户 DTO。
	userDTO, err := toUserDTO(ctx, user, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户 DTO 失败: %v", err)
		return nil, ErrInternal
	}

	return &userDTO, nil
}

// Update 更新用户资料，并在需要时重做唯一性校验和密码哈希。
func (s *userService) Update(ctx context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error) {
	// 先读取用户实体，后续按可选字段增量更新。
	if input.ID == 0 {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	// 逐项应用更新字段，并做对应的业务校验。
	if input.Username != nil {
		username := strings.TrimSpace(*input.Username)
		if err := validator.ValidateUsername(username); err != nil {
			return nil, err
		}
		if err := s.ensureUsernameAvailable(ctx, username, user.ID); err != nil {
			return nil, err
		}
		user.Username = username
	}

	if input.Nickname != nil {
		nickname := strings.TrimSpace(*input.Nickname)
		if nickname == "" {
			nickname = user.Username
		}
		user.Nickname = nickname
	}

	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)
		if err := validator.ValidateEmail(email); err != nil {
			return nil, err
		}
		if err := s.ensureEmailAvailable(ctx, email, user.ID); err != nil {
			return nil, err
		}
		user.Email = email
	}

	if input.Password != nil {
		if err := validator.ValidatePassword(*input.Password); err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("生成密码哈希失败: %v", err)
			return nil, ErrInternal
		}
		user.Password = string(hash)
	}

	if input.Admin != nil {
		user.Admin = *input.Admin
	}

	if input.Status != nil {
		if err := validateUserStatus(*input.Status); err != nil {
			return nil, err
		}
		user.Status = *input.Status
	}

	if input.Avatar != nil {
		user.Avatar = normalizeOptionalString(input.Avatar)
	}

	if input.EmailVerified != nil {
		user.EmailVerified = *input.EmailVerified
	}

	if input.UseDefaultStorageQuota != nil && *input.UseDefaultStorageQuota {
		user.StorageQuota = nil
	} else if input.StorageQuota != nil {
		storageQuota, err := normalizeStorageQuota(input.StorageQuota)
		if err != nil {
			return nil, err
		}
		user.StorageQuota = storageQuota
	}

	// 持久化并返回 DTO。
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("更新用户失败: %v", err)
		return nil, ErrInternal
	}

	userDTO, err := toUserDTO(ctx, user, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户 DTO 失败: %v", err)
		return nil, ErrInternal
	}
	return &userDTO, nil
}

// UploadAvatar 上传并替换用户头像，同时维护用户存储占用。
func (s *userService) UploadAvatar(ctx context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
	// 校验上传输入与依赖。
	if input.UserID == 0 {
		return nil, ErrUserNotFound
	}
	if s.storageManager == nil {
		return nil, ErrInternal
	}
	if input.Content == nil || len(input.ContentSample) == 0 || input.FileSize <= 0 {
		return nil, ErrEmptyAvatarContent
	}
	if strings.TrimSpace(input.OriginalFilename) == "" {
		return nil, ErrEmptyAvatarFilename
	}
	if err := validateAvatarUpload(input.OriginalFilename, input.ContentType, input.ContentSample); err != nil {
		return nil, err
	}

	// 读取用户与旧头像占用。
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	oldAvatarKey := managedAvatarKey(user.Avatar)
	oldAvatarSize, err := s.getManagedObjectSize(ctx, oldAvatarKey)
	if err != nil {
		log.Printf("读取旧头像失败: %v", err)
		return nil, ErrInternal
	}

	defaultQuota, err := getDefaultStorageQuota(ctx, s.settings)
	if err != nil {
		return nil, err
	}

	// 上传新头像对象。
	newAvatarKey, err := s.storageManager.NewObjectKey(storage.CategoryAvatar, input.UserID, input.OriginalFilename)
	if err != nil {
		log.Printf("生成头像 key 失败: %v", err)
		return nil, ErrInternal
	}

	if err := s.storageManager.UploadStream(ctx, newAvatarKey, input.Content, input.FileSize, strings.TrimSpace(input.ContentType)); err != nil {
		log.Printf("上传头像失败: %v", err)
		return nil, ErrInternal
	}

	// 事务更新头像与存储占用，失败时回滚新对象。
	newAvatarValue := newAvatarKey
	updatedUser, err := s.userRepo.UpdateAvatarWithStorageUsedTx(ctx, input.UserID, &newAvatarValue, input.FileSize-oldAvatarSize, defaultQuota)
	if err != nil {
		if deleteErr := s.storageManager.Delete(ctx, newAvatarKey); deleteErr != nil {
			log.Printf("回滚新头像对象失败: %v", deleteErr)
		}
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repo.ErrUserAvatarQuotaExceeded):
			return nil, ErrStorageQuotaExceeded
		default:
			log.Printf("更新用户头像失败: %v", err)
			return nil, ErrInternal
		}
	}

	// 转换 DTO，并异步清理旧头像对象。
	userDTO, err := toUserDTO(ctx, updatedUser, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户 DTO 失败: %v", err)
		return nil, ErrInternal
	}

	if oldAvatarKey != "" && oldAvatarKey != newAvatarKey {
		if err := s.storageManager.Delete(ctx, oldAvatarKey); err != nil {
			log.Printf("删除旧头像失败: %v", err)
		}
	}

	return &userDTO, nil
}

// DeleteAvatar 删除用户头像，同时维护用户存储占用。
func (s *userService) DeleteAvatar(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}

	oldAvatarKey := managedAvatarKey(user.Avatar)
	oldAvatarSize, err := s.getManagedObjectSize(ctx, oldAvatarKey)
	if err != nil {
		log.Printf("读取旧头像失败: %v", err)
		return ErrInternal
	}

	_, err = s.userRepo.UpdateAvatarWithStorageUsedTx(ctx, userID, nil, -oldAvatarSize, 0)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("删除用户头像失败: %v", err)
		return ErrInternal
	}

	if oldAvatarKey != "" {
		if err := s.storageManager.Delete(ctx, oldAvatarKey); err != nil {
			log.Printf("删除旧头像失败: %v", err)
		}
	}

	return nil
}

// Delete 删除指定用户记录。
func (s *userService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrUserNotFound
	}

	if _, err := s.userRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		log.Printf("删除用户失败: %v", err)
		return ErrInternal
	}
	return nil
}

// List 根据过滤条件和分页参数返回用户列表。
func (s *userService) List(ctx context.Context, input dto.UserListInput) (*dto.UserListResult, error) {
	page, size := serviceListBounds.NormalizePageSize(input.Page, input.Size)

	users, total, err := s.userRepo.List(ctx, repo.UserFilter{
		ID:       input.ID,
		Username: strings.TrimSpace(input.Username),
		Email:    strings.TrimSpace(input.Email),
		Status:   input.Status,
	}, repo.ListOptions{
		Limit:  size,
		Offset: serviceListBounds.OffsetFromPage(page, size),
		Order:  resolveUserListOrder(input.Order),
	})
	if err != nil {
		log.Printf("查询用户列表失败: %v", err)
		return nil, ErrInternal
	}

	items, err := toUserDTOs(ctx, users, s.storageManager, s.gravatarBaseURL)
	if err != nil {
		log.Printf("构建用户列表 DTO 失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.UserListResult{
		Items: items,
		Total: total,
	}, nil
}

// ensureUsernameAvailable 确保用户名未被其他用户占用。
func (s *userService) ensureUsernameAvailable(ctx context.Context, username string, excludeID uint) error {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		if user.ID != excludeID {
			return ErrUsernameExists
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	log.Printf("按用户名查询失败: %v", err)
	return ErrInternal
}

// ensureEmailAvailable 确保邮箱未被其他用户占用。
func (s *userService) ensureEmailAvailable(ctx context.Context, email string, excludeID uint) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		if user.ID != excludeID {
			return ErrEmailExists
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	log.Printf("按邮箱查询失败: %v", err)
	return ErrInternal
}

// validateUserStatus 校验用户状态值是否在允许范围内。
func validateUserStatus(status int) error {
	switch status {
	case 1, 2, 3:
		return nil
	default:
		return ErrInvalidUserStatus
	}
}

// normalizeStorageQuota 规范化存储配额并校验其为非负数。
func normalizeStorageQuota(quota *int64) (*int64, error) {
	if quota == nil {
		return nil, nil
	}
	if *quota < 0 {
		return nil, ErrInvalidStorageQuota
	}
	value := *quota
	return &value, nil
}

// normalizeOptionalString 在指针存在时修剪首尾空白，并将空串折叠为 nil。
func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// managedAvatarKey 提取并返回系统托管头像对象键，非托管值返回空串。
func managedAvatarKey(avatar *string) string {
	if avatar == nil {
		return ""
	}

	trimmed := strings.TrimSpace(*avatar)
	if !isManagedAvatarKey(trimmed) {
		return ""
	}

	return trimmed
}

// getManagedObjectSize 获取托管对象大小；空键返回 0。
func (s *userService) getManagedObjectSize(ctx context.Context, objectKey string) (int64, error) {
	if objectKey == "" {
		return 0, nil
	}
	if s.storageManager == nil {
		return 0, ErrInternal
	}

	return s.storageManager.GetSize(ctx, objectKey)
}

// marshalEmailChangePayload 序列化邮箱变更 token 载荷。
func marshalEmailChangePayload(oldEmail, newEmail string) (string, error) {
	oldEmail = strings.TrimSpace(oldEmail)
	newEmail = strings.TrimSpace(newEmail)
	if oldEmail == "" || newEmail == "" {
		return "", ErrInvalidEmailChange
	}

	payload, err := json.Marshal(emailChangeTokenPayload{
		OldEmail: oldEmail,
		NewEmail: newEmail,
	})
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

// unmarshalEmailChangePayload 反序列化并校验邮箱变更 token 载荷。
func unmarshalEmailChangePayload(value string) (*emailChangeTokenPayload, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, ErrInvalidEmailChange
	}

	var payload emailChangeTokenPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		log.Printf("解析邮箱变更 payload 失败: %v", err)
		return nil, ErrInvalidEmailChange
	}

	payload.OldEmail = strings.TrimSpace(payload.OldEmail)
	payload.NewEmail = strings.TrimSpace(payload.NewEmail)
	if payload.OldEmail == "" || payload.NewEmail == "" {
		return nil, ErrInvalidEmailChange
	}

	if err := validator.ValidateEmail(payload.OldEmail); err != nil {
		return nil, ErrInvalidEmailChange
	}
	if err := validator.ValidateEmail(payload.NewEmail); err != nil {
		return nil, ErrInvalidEmailChange
	}

	return &payload, nil
}
