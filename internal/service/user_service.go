package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidUserStatus   = errors.New("用户状态不合法")
	ErrInvalidStorageQuota = errors.New("存储配额不能为负数")
)

// UserService 提供用户管理相关的 CRUD 能力。
type UserService interface {
	// Create 创建用户。
	Create(ctx context.Context, input dto.UserCreateInput) (*dto.UserDTO, error)
	// Get 按 ID 获取用户详情。
	Get(ctx context.Context, id uint) (*dto.UserDTO, error)
	// Update 按输入字段更新用户信息。
	Update(ctx context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error)
	// Delete 删除指定用户。
	Delete(ctx context.Context, id uint) error
	// List 按筛选条件查询用户列表。
	List(ctx context.Context, input dto.UserListInput) (*dto.UserListResult, error)
}

type userService struct {
	userRepo repo.UserRepo
}

// NewUserService 创建用户服务实例。
func NewUserService(userRepo repo.UserRepo) UserService {
	return &userService{userRepo: userRepo}
}

// Create 校验输入后创建新用户，并对密码进行哈希。
func (s *userService) Create(ctx context.Context, input dto.UserCreateInput) (*dto.UserDTO, error) {
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

	if err := s.ensureUsernameAvailable(ctx, username, 0); err != nil {
		return nil, err
	}
	if err := s.ensureEmailAvailable(ctx, email, 0); err != nil {
		return nil, err
	}

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

	userDTO := toUserDTO(user)
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

	userDTO := toUserDTO(user)
	return &userDTO, nil
}

// Update 更新用户资料，并在需要时重做唯一性校验和密码哈希。
func (s *userService) Update(ctx context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error) {
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

	if input.StorageQuota != nil {
		storageQuota, err := normalizeStorageQuota(input.StorageQuota)
		if err != nil {
			return nil, err
		}
		user.StorageQuota = storageQuota
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("更新用户失败: %v", err)
		return nil, ErrInternal
	}

	userDTO := toUserDTO(user)
	return &userDTO, nil
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
	users, total, err := s.userRepo.List(ctx, repo.UserFilter{
		Username: strings.TrimSpace(input.Username),
		Email:    strings.TrimSpace(input.Email),
		Status:   input.Status,
	}, repo.ListOptions{
		Limit:  normalizeServiceListLimit(input.Limit),
		Offset: normalizeServiceListOffset(input.Offset),
		Order:  normalizeUserListOrder(input.Order),
	})
	if err != nil {
		log.Printf("查询用户列表失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.UserListResult{
		Items: toUserDTOs(users),
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
