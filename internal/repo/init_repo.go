package repo

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"easydrop/internal/consts"
	"easydrop/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInitAlreadyInitialized = errors.New("system already initialized")
	ErrInvalidInitState       = errors.New("invalid initialized state")
)

type SettingValueInput struct {
	Key   string
	Value string
}

type SystemInitInput struct {
	AdminUser model.User
	Settings  []SettingValueInput
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
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("key = ?", consts.SystemInitializedSettingKey).First(&setting).Error
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
	updates := make([]initSettingUpdate, 0, len(input.Settings))
	for _, setting := range input.Settings {
		updates = append(updates, initSettingUpdate{key: setting.Key, value: setting.Value})
	}
	return updates
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
