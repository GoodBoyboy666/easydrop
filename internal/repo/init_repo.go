package repo

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"easydrop/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const initSettingKey = "system.initialized"

var (
	ErrInitAlreadyInitialized = errors.New("system already initialized")
	ErrInvalidInitState       = errors.New("invalid initialized state")
)

type SystemInitInput struct {
	AdminUser        model.User
	SiteName         string
	SiteURL          string
	SiteAnnouncement string
	AllowRegister    string
}

// InitRepo 定义系统初始化事务能力。
type InitRepo interface {
	Initialize(ctx context.Context, input SystemInitInput) error
}

// GormInitRepo 基于 Gorm 的初始化仓储实现。
type GormInitRepo struct {
	db *gorm.DB
}

// NewInitRepo 创建初始化仓储实例。
func NewInitRepo(db *gorm.DB) InitRepo {
	return &GormInitRepo{db: db}
}

func (r *GormInitRepo) Initialize(ctx context.Context, input SystemInitInput) error {
	return r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		initialized, err := loadInitializedForUpdate(tx)
		if err != nil {
			return err
		}
		if initialized {
			return ErrInitAlreadyInitialized
		}

		adminUser := input.AdminUser
		if err := tx.Create(&adminUser).Error; err != nil {
			return err
		}

		for _, setting := range buildInitSettingUpdates(input) {
			if err := updateInitSettingValue(tx, setting.key, setting.value); err != nil {
				return err
			}
		}

		return nil
	})
}

func loadInitializedForUpdate(tx *gorm.DB) (bool, error) {
	var setting model.Setting
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("key = ?", initSettingKey).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(setting.Value))
	if err != nil {
		return false, ErrInvalidInitState
	}
	return parsed, nil
}

type initSettingUpdate struct {
	key   string
	value string
}

func buildInitSettingUpdates(input SystemInitInput) []initSettingUpdate {
	return []initSettingUpdate{
		{key: "site.name", value: input.SiteName},
		{key: "site.url", value: input.SiteURL},
		{key: "site.announcement", value: input.SiteAnnouncement},
		{key: "site.allow_register", value: input.AllowRegister},
		{key: initSettingKey, value: "true"},
	}
}

func updateInitSettingValue(tx *gorm.DB, key, value string) error {
	result := tx.Model(&model.Setting{}).Where("key = ?", key).Update("value", value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

var _ InitRepo = (*GormInitRepo)(nil)
