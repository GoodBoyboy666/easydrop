package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"easydrop/internal/consts"
	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"
	"easydrop/internal/pkg/initsecret"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrAlreadyInitialized = errors.New("系统已初始化")
)

// InitService 提供系统初始化能力。
type InitService interface {
	// GetStatus 返回当前系统初始化状态。
	GetStatus(ctx context.Context) (*dto.InitStatusResult, error)
	// Initialize 执行系统首次初始化。
	Initialize(ctx context.Context, input dto.InitInput) error
}

type initService struct {
	userRepo       repo.UserRepo
	initRepo       repo.InitRepo
	settingService SettingService
	cache          cache.Cache
	initSecret     initsecret.Guard
}

// NewInitService 创建系统初始化服务。
func NewInitService(userRepo repo.UserRepo, initRepo repo.InitRepo, settingService SettingService, kvCache cache.Cache, initSecret initsecret.Guard) InitService {
	return &initService{
		userRepo:       userRepo,
		initRepo:       initRepo,
		settingService: settingService,
		cache:          kvCache,
		initSecret:     initSecret,
	}
}

// GetStatus 返回系统是否已完成初始化。
func (s *initService) GetStatus(ctx context.Context) (*dto.InitStatusResult, error) {
	initialized, err := s.isInitialized(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.InitStatusResult{Initialized: initialized}, nil
}

// Initialize 执行系统初始化流程：校验状态、创建管理员、写入初始配置并同步缓存。
func (s *initService) Initialize(ctx context.Context, input dto.InitInput) error {
	// 先检查依赖是否完整，避免初始化进行到中途才失败。
	if s.userRepo == nil || s.initRepo == nil || s.settingService == nil || s.cache == nil || s.initSecret == nil {
		return ErrInternal
	}

	// 校验当前是否已初始化，以及初始化密钥是否合法。
	initialized, err := s.isInitialized(ctx)
	if err != nil {
		return err
	}
	if initialized {
		return ErrAlreadyInitialized
	}
	if err := s.initSecret.Validate(ctx, input.Secret); err != nil {
		if errors.Is(err, initsecret.ErrNotReady) {
			return ErrInternal
		}
		return err
	}

	// 解析初始化开关并构造管理员用户。
	allowRegister := true
	if input.AllowRegister != nil {
		allowRegister = *input.AllowRegister
	}

	adminUser, err := s.buildInitAdminUser(ctx, input)
	if err != nil {
		return err
	}

	// 在事务中写入管理员和系统初始化配置。
	if err := s.initRepo.Initialize(ctx, repo.SystemInitInput{
		AdminUser: *adminUser,
		Settings: []repo.SettingValueInput{
			{Key: consts.SiteNameSettingKey, Value: input.SiteName},
			{Key: consts.SiteURLSettingKey, Value: input.SiteURL},
			{Key: consts.SiteAnnouncementSettingKey, Value: input.SiteAnnouncement},
			{Key: consts.SiteAllowRegisterSettingKey, Value: strconv.FormatBool(allowRegister)},
			{Key: consts.SystemInitializedSettingKey, Value: "true"},
		},
	}); err != nil {
		switch {
		case errors.Is(err, repo.ErrInitAlreadyInitialized):
			return ErrAlreadyInitialized
		case errors.Is(err, repo.ErrInvalidInitState):
			return ErrInvalidSiteSetting
		default:
			log.Printf("执行系统初始化事务失败: %v", err)
			return ErrInternal
		}
	}

	// 将关键配置同步到缓存，保证初始化后立即可读。
	if err := s.syncInitSettingsCache(ctx, input, allowRegister); err != nil {
		log.Printf("同步初始化配置缓存失败: %v", err)
		return err
	}

	return nil
}

// isInitialized 读取系统初始化标记并转换为布尔状态。
func (s *initService) isInitialized(ctx context.Context) (bool, error) {
	if s.settingService == nil {
		return false, ErrInternal
	}

	v, ok, err := s.settingService.GetValue(ctx, consts.SystemInitializedSettingKey)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return false, ErrInvalidSiteSetting
	}

	return parsed, nil
}

// buildInitAdminUser 校验初始化管理员输入并构造管理员实体。
func (s *initService) buildInitAdminUser(ctx context.Context, input dto.InitInput) (*model.User, error) {
	// 校验管理员基础字段。
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

	// 校验用户名与邮箱唯一性。
	if err := s.ensureInitUsernameAvailable(ctx, username); err != nil {
		return nil, err
	}
	if err := s.ensureInitEmailAvailable(ctx, email); err != nil {
		return nil, err
	}

	// 生成密码哈希并组装管理员用户。
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成初始化管理员密码哈希失败: %v", err)
		return nil, ErrInternal
	}

	return &model.User{
		Username:      username,
		Nickname:      nickname,
		Email:         email,
		Password:      string(hash),
		Admin:         true,
		Status:        1,
		EmailVerified: true,
	}, nil
}

// ensureInitUsernameAvailable 确保初始化管理员用户名未被占用。
func (s *initService) ensureInitUsernameAvailable(ctx context.Context, username string) error {
	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return ErrUsernameExists
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	log.Printf("按用户名查询初始化用户失败: %v", err)
	return ErrInternal
}

// ensureInitEmailAvailable 确保初始化管理员邮箱未被占用。
func (s *initService) ensureInitEmailAvailable(ctx context.Context, email string) error {
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return ErrEmailExists
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	log.Printf("按邮箱查询初始化用户失败: %v", err)
	return ErrInternal
}

// syncInitSettingsCache 将初始化写入的关键配置同步到缓存层。
func (s *initService) syncInitSettingsCache(ctx context.Context, input dto.InitInput, allowRegister bool) error {
	values := map[string]string{
		consts.SiteNameSettingKey:          input.SiteName,
		consts.SiteURLSettingKey:           input.SiteURL,
		consts.SiteAnnouncementSettingKey:  input.SiteAnnouncement,
		consts.SiteAllowRegisterSettingKey: strconv.FormatBool(allowRegister),
		consts.SystemInitializedSettingKey: "true",
	}

	for key, value := range values {
		if err := s.cache.Set(ctx, settingCacheKey(key), value, 0); err != nil {
			return ErrInternal
		}
	}
	return nil
}

var _ InitService = (*initService)(nil)
