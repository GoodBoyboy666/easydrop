package config

import (
	"context"
	"errors"

	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"
)

// DBConfig 管理存储在数据库中的配置。
type DBConfig interface {
	Get(ctx context.Context, key string) (model.Setting, error)
	GetValue(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string) error
	All(ctx context.Context) ([]model.Setting, error)
	ClearCache(ctx context.Context) error
	ClearCacheKey(ctx context.Context, key string) error
}

type dbConfig struct {
	db    *gorm.DB
	cache cache.Cache
}

// DBProviderSet 提供 DBConfig 的 Wire 注入入口。
var DBProviderSet = wire.NewSet(NewDBConfig)

// NewDBConfig 创建 DBConfig，并负责迁移与初始化默认配置。
func NewDBConfig(db *gorm.DB, kvCache cache.Cache) (DBConfig, error) {
	if db == nil {
		return nil, errors.New("db is required")
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

	return &dbConfig{db: db, cache: kvCache}, nil
}

func (c *dbConfig) Get(ctx context.Context, key string) (model.Setting, error) {
	if value, found, err := c.cache.Get(ctx, dbConfigCacheKey(key)); err == nil && found {
		return model.Setting{Key: key, Value: value}, nil
	}

	var setting model.Setting
	err := c.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err == nil {
		_ = c.cache.Set(ctx, dbConfigCacheKey(key), setting.Value, 0)
	}
	return setting, err
}

func (c *dbConfig) GetValue(ctx context.Context, key string) (string, bool, error) {
	setting, err := c.Get(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return setting.Value, true, nil
}

func (c *dbConfig) Set(ctx context.Context, key, value string) error {
	err := c.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&model.Setting{
		Key:   key,
		Value: value,
	}).Error
	if err != nil {
		return err
	}

	_ = c.cache.Set(ctx, dbConfigCacheKey(key), value, 0)
	return nil
}

func (c *dbConfig) All(ctx context.Context) ([]model.Setting, error) {
	var settings []model.Setting
	if err := c.db.WithContext(ctx).Order("key asc").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (c *dbConfig) ClearCache(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

func (c *dbConfig) ClearCacheKey(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, dbConfigCacheKey(key))
}

var _ DBConfig = (*dbConfig)(nil)

func dbConfigCacheKey(key string) string {
	return "setting:" + key
}

func initDefaultSettings(db *gorm.DB) error {
	defaults := []model.Setting{
		{
			Key:      "site.name",
			Value:    "EasyDrop",
			Desc:     "站点名称",
			Category: "site",
		},
		{
			Key:      "site.url",
			Value:    "http://localhost:8080",
			Desc:     "站点访问地址",
			Category: "site",
		},
		{
			Key:      "site.allow_register",
			Value:    "true",
			Desc:     "是否允许注册",
			Category: "site",
		},
		{
			Key:      "site.announcement",
			Value:    "",
			Desc:     "站点公告",
			Category: "site",
		},
		{
			Key:      "storage.quota",
			Value:    "10737418240",
			Desc:     "存储配额（字节，默认10GB）",
			Category: "storage",
		},
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}).Create(&defaults).Error
}
