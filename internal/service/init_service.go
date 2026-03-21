package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"easydrop/internal/dto"
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
	userService    UserService
	settingService SettingService
}

// NewInitService 创建系统初始化服务。
func NewInitService(userService UserService, settingService SettingService) InitService {
	return &initService{userService: userService, settingService: settingService}
}

func (s *initService) GetStatus(ctx context.Context) (*dto.InitStatusResult, error) {
	initialized, err := s.isInitialized(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.InitStatusResult{Initialized: initialized}, nil
}

func (s *initService) Initialize(ctx context.Context, input dto.InitInput) error {
	if s.userService == nil || s.settingService == nil {
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

	if _, err := s.userService.Create(ctx, dto.UserCreateInput{
		Username:      input.Username,
		Nickname:      input.Nickname,
		Email:         input.Email,
		Password:      input.Password,
		Admin:         true,
		Status:        nil,
		EmailVerified: true,
	}); err != nil {
		return err
	}

	if err := s.updateSiteSettings(ctx, input, allowRegister); err != nil {
		return err
	}

	v := "true"
	if err := s.settingService.UpdateItem(ctx, dto.SettingUpdateInput{Key: initSettingKey, Value: &v}); err != nil {
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

func (s *initService) updateSiteSettings(ctx context.Context, input dto.InitInput, allowRegister bool) error {
	allowRegisterValue := strconv.FormatBool(allowRegister)
	updates := []dto.SettingUpdateInput{
		{Key: "site.name", Value: ptrString(input.SiteName)},
		{Key: "site.url", Value: ptrString(input.SiteURL)},
		{Key: "site.announcement", Value: ptrString(input.SiteAnnouncement)},
		{Key: "site.allow_register", Value: &allowRegisterValue},
	}

	for i := range updates {
		if err := s.settingService.UpdateItem(ctx, updates[i]); err != nil {
			return err
		}
	}

	return nil
}

func ptrString(v string) *string {
	copyValue := v
	return &copyValue
}

var _ InitService = (*initService)(nil)
