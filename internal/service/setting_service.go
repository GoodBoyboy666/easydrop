package service

import (
	"context"
	"errors"
	"strings"

	"easydrop/internal/consts"
	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

// SettingService 提供站点动态配置读写能力。
type SettingService interface {
	GetValue(ctx context.Context, key string) (string, bool, error)
	ListItems(ctx context.Context, input dto.SettingListInput) (*dto.SettingListResult, error)
	UpdateItem(ctx context.Context, input dto.SettingUpdateInput) error
	GetPublicItems(ctx context.Context) (*dto.SettingPublicResult, error)
}

var (
	ErrSettingKeyRequired = errors.New("配置键不能为空")
)

type settingService struct {
	settingRepo repo.SettingRepo
	cache       cache.Cache
}

// NewSettingService 创建配置服务，并确保配置表与默认配置可用。
func NewSettingService(db *gorm.DB, settingRepo repo.SettingRepo, kvCache cache.Cache) (SettingService, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if settingRepo == nil {
		return nil, errors.New("setting repo is required")
	}
	if kvCache == nil {
		return nil, errors.New("cache is required")
	}

	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		return nil, err
	}
	if err := initDefaultSettings(settingRepo); err != nil {
		return nil, err
	}

	return &settingService{settingRepo: settingRepo, cache: kvCache}, nil
}

func (s *settingService) get(ctx context.Context, key string) (model.Setting, error) {
	cleanKey := strings.TrimSpace(key)
	if value, found, err := s.cache.Get(ctx, settingCacheKey(cleanKey)); err == nil && found {
		return model.Setting{Key: cleanKey, Value: value}, nil
	}

	setting, err := s.settingRepo.GetByKey(ctx, cleanKey)
	if err != nil {
		return model.Setting{}, err
	}
	_ = s.cache.Set(ctx, settingCacheKey(cleanKey), setting.Value, 0)
	return *setting, nil
}

func (s *settingService) GetValue(ctx context.Context, key string) (string, bool, error) {
	setting, err := s.get(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return setting.Value, true, nil
}

func (s *settingService) ListItems(ctx context.Context, input dto.SettingListInput) (*dto.SettingListResult, error) {
	page, size := normalizeServiceListPageSize(input.Page, input.Size)

	settings, total, err := s.settingRepo.List(ctx, repo.SettingFilter{
		Category: strings.TrimSpace(input.Category),
		Key:      strings.TrimSpace(input.Key),
	}, repo.ListOptions{
		Limit:  size,
		Offset: pageSizeToOffset(page, size),
		Order:  normalizeSettingListOrder(input.Order),
	})
	if err != nil {
		return nil, ErrInternal
	}

	items := make([]dto.SettingItem, 0, len(settings))
	for i := range settings {
		items = append(items, dto.SettingItem{
			Key:       settings[i].Key,
			Value:     settings[i].Value,
			Desc:      settings[i].Desc,
			Category:  settings[i].Category,
			Sensitive: settings[i].Sensitive,
			Public:    settings[i].Public,
		})
	}

	return &dto.SettingListResult{Items: items, Total: total}, nil
}

func (s *settingService) UpdateItem(ctx context.Context, input dto.SettingUpdateInput) error {
	cleanKey := strings.TrimSpace(input.Key)
	if cleanKey == "" {
		return ErrSettingKeyRequired
	}

	setting, err := s.settingRepo.GetByKey(ctx, cleanKey)
	found := true
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			found = false
			setting = &model.Setting{Key: cleanKey}
		} else {
			return ErrInternal
		}
	}

	if input.Value != nil {
		setting.Value = *input.Value
	} else if !found {
		setting.Value = ""
	}

	if err := s.settingRepo.UpsertByKey(ctx, setting); err != nil {
		return ErrInternal
	}

	_ = s.cache.Set(ctx, settingCacheKey(cleanKey), setting.Value, 0)
	return nil
}

func (s *settingService) GetPublicItems(ctx context.Context) (*dto.SettingPublicResult, error) {
	public := true
	settings, _, err := s.settingRepo.List(ctx, repo.SettingFilter{Public: &public}, repo.ListOptions{
		Limit:  100,
		Offset: 0,
		Order:  "key asc",
	})
	if err != nil {
		return nil, ErrInternal
	}

	items := make([]dto.SettingPublicItem, 0, len(settings))
	for i := range settings {
		items = append(items, dto.SettingPublicItem{
			Key:   settings[i].Key,
			Value: settings[i].Value,
		})
	}

	return &dto.SettingPublicResult{Items: items}, nil
}

func settingCacheKey(key string) string {
	return "setting:" + key
}

func normalizeSettingListOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "key_desc":
		return "key desc"
	case "key_asc", "":
		return "key asc"
	default:
		return "key asc"
	}
}

func initDefaultSettings(settingRepo repo.SettingRepo) error {
	defaults := []model.Setting{
		{
			Key:      consts.SiteNameSettingKey,
			Value:    "EasyDrop",
			Desc:     "站点名称",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteDescriptionSettingKey,
			Value:    "一个轻量级日志说说平台",
			Desc:     "站点描述",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteOwnerSettingKey,
			Value:    "Your Name",
			Desc:     "站长名称",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteOwnerDescriptionSettingKey,
			Value:    "Do what you want to do.",
			Desc:     "站长简介",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteURLSettingKey,
			Value:    "http://localhost:8080",
			Desc:     "站点访问地址",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteBackgroundSettingKey,
			Value:    "",
			Desc:     "网站背景",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteFaviconSettingKey,
			Value:    "",
			Desc:     "网站 favicon",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteAllowRegisterSettingKey,
			Value:    "true",
			Desc:     "是否允许注册",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.SiteAnnouncementSettingKey,
			Value:    "",
			Desc:     "站点公告",
			Category: "site",
			Public:   true,
		},
		{
			Key:      consts.AuthRequireEmailVerificationSettingKey,
			Value:    "false",
			Desc:     "登录前必须完成邮箱验证",
			Category: "auth",
			Public:   false,
		},
		{
			Key:      consts.StorageQuotaSettingKey,
			Value:    "10737418240",
			Desc:     "存储配额（字节，默认10GB）",
			Category: "storage",
			Public:   true,
		},
		{
			Key:      consts.StorageUploadSettingKey,
			Value:    "true",
			Desc:     "允许普通用户上传附件",
			Category: "storage",
			Public:   true,
		},
		{
			Key:      consts.AttachmentAllowedExtensionsSettingKey,
			Value:    consts.DefaultAttachmentAllowedExtensionsSettingValue,
			Desc:     "允许上传的附件扩展名，英文逗号分隔且不带点",
			Category: "storage",
			Public:   false,
		},
		{
			Key:      consts.UploadMaxRequestBodySettingKey,
			Value:    "52428800",
			Desc:     "上传接口最大请求体大小（字节，默认 50MB）",
			Category: "storage",
			Public:   false,
		},
		{
			Key:      consts.SystemInitializedSettingKey,
			Value:    "false",
			Desc:     "系统已初始化",
			Category: "system",
			Public:   false,
		},
	}

	return settingRepo.SyncDefaults(context.Background(), defaults)
}

var _ SettingService = (*settingService)(nil)
