package service

import (
	"context"
	"errors"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"
	"easydrop/internal/repo"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if err := initDefaultSettings(db); err != nil {
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
	settings, total, err := s.settingRepo.List(ctx, repo.SettingFilter{
		Category: strings.TrimSpace(input.Category),
		Key:      strings.TrimSpace(input.Key),
	}, repo.ListOptions{
		Limit:  normalizeServiceListLimit(input.Limit),
		Offset: normalizeServiceListOffset(input.Offset),
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

func initDefaultSettings(db *gorm.DB) error {
	defaults := []model.Setting{
		{
			Key:      "site.name",
			Value:    "EasyDrop",
			Desc:     "站点名称",
			Category: "site",
			Public:   true,
		},
		{
			Key:      "site.url",
			Value:    "http://localhost:8080",
			Desc:     "站点访问地址",
			Category: "site",
			Public:   true,
		},
		{
			Key:      "site.allow_register",
			Value:    "true",
			Desc:     "是否允许注册",
			Category: "site",
			Public:   true,
		},
		{
			Key:      "site.announcement",
			Value:    "",
			Desc:     "站点公告",
			Category: "site",
			Public:   true,
		},
		{
			Key:      "storage.quota",
			Value:    "10737418240",
			Desc:     "存储配额（字节，默认10GB）",
			Category: "storage",
			Public:   true,
		},
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}).Create(&defaults).Error
}

var _ SettingService = (*settingService)(nil)
