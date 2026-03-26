package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const initSettingKey = "system.initialized"

var (
	ErrAlreadyInitialized = errors.New("系统已初始化")
)

// InitService 提供系统初始化能力。
type InitService interface {
	GetStatus(ctx context.Context) (*dto.InitStatusResult, error)
	Initialize(ctx context.Context, input dto.InitInput) error
}

type initService struct {
	userRepo       repo.UserRepo
	initRepo       repo.InitRepo
	settingService SettingService
	cache          cache.Cache
}

// NewInitService 创建系统初始化服务。
func NewInitService(userRepo repo.UserRepo, initRepo repo.InitRepo, settingService SettingService, kvCache cache.Cache) InitService {
	return &initService{
		userRepo:       userRepo,
		initRepo:       initRepo,
		settingService: settingService,
		cache:          kvCache,
	}
}

func (s *initService) GetStatus(ctx context.Context) (*dto.InitStatusResult, error) {
	initialized, err := s.isInitialized(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.InitStatusResult{Initialized: initialized}, nil
}

func (s *initService) Initialize(ctx context.Context, input dto.InitInput) error {
	if s.userRepo == nil || s.initRepo == nil || s.settingService == nil || s.cache == nil {
		return ErrInternal
	}

	initialized, err := s.isInitialized(ctx)
	if err != nil {
		return err
	}
	if initialized {
		return ErrAlreadyInitialized
	}

	allowRegister := true
	if input.AllowRegister != nil {
		allowRegister = *input.AllowRegister
	}

	adminUser, err := s.buildInitAdminUser(ctx, input)
	if err != nil {
		return err
	}

	if err := s.initRepo.Initialize(ctx, repo.SystemInitInput{
		AdminUser:        *adminUser,
		SiteName:         input.SiteName,
		SiteURL:          input.SiteURL,
		SiteAnnouncement: input.SiteAnnouncement,
		AllowRegister:    strconv.FormatBool(allowRegister),
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

	if err := s.syncInitSettingsCache(ctx, input, allowRegister); err != nil {
		log.Printf("同步初始化配置缓存失败: %v", err)
		return err
	}

	return nil
}

func (s *initService) isInitialized(ctx context.Context) (bool, error) {
	if s.settingService == nil {
		return false, ErrInternal
	}

	v, ok, err := s.settingService.GetValue(ctx, initSettingKey)
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

func (s *initService) buildInitAdminUser(ctx context.Context, input dto.InitInput) (*model.User, error) {
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

	if err := s.ensureInitUsernameAvailable(ctx, username); err != nil {
		return nil, err
	}
	if err := s.ensureInitEmailAvailable(ctx, email); err != nil {
		return nil, err
	}

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

func (s *initService) syncInitSettingsCache(ctx context.Context, input dto.InitInput, allowRegister bool) error {
	values := map[string]string{
		"site.name":           input.SiteName,
		"site.url":            input.SiteURL,
		"site.announcement":   input.SiteAnnouncement,
		"site.allow_register": strconv.FormatBool(allowRegister),
		initSettingKey:        "true",
	}

	for key, value := range values {
		if err := s.cache.Set(ctx, settingCacheKey(key), value, 0); err != nil {
			return ErrInternal
		}
	}
	return nil
}

var _ InitService = (*initService)(nil)
