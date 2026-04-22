package repo

import (
	"context"
	"strings"

	"easydrop/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SettingRepo 定义设置仓储接口。
type SettingRepo interface {
	Create(ctx context.Context, setting *model.Setting) error
	GetByKey(ctx context.Context, key string) (*model.Setting, error)
	SyncDefaults(ctx context.Context, defaults []model.Setting) error
	UpsertByKey(ctx context.Context, setting *model.Setting) error
	Update(ctx context.Context, setting *model.Setting) error
	DeleteByKey(ctx context.Context, key string) error
	All(ctx context.Context) ([]model.Setting, error)
	List(ctx context.Context, filter SettingFilter, opts ListOptions) ([]model.Setting, int64, error)
}

// SettingFilter 设置查询过滤条件。
type SettingFilter struct {
	Category string
	Key      string
	Public   *bool
}

// GormSettingRepo 基于 Gorm 的设置仓储实现。
type GormSettingRepo struct {
	db *gorm.DB
}

// NewSettingRepo 创建设置仓储实例。
func NewSettingRepo(db *gorm.DB) SettingRepo {
	return &GormSettingRepo{db: db}
}

func (r *GormSettingRepo) Create(ctx context.Context, setting *model.Setting) error {
	return r.db.WithContext(withContext(ctx)).Create(setting).Error
}

func (r *GormSettingRepo) GetByKey(ctx context.Context, key string) (*model.Setting, error) {
	var setting model.Setting
	err := r.db.WithContext(withContext(ctx)).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *GormSettingRepo) SyncDefaults(ctx context.Context, defaults []model.Setting) error {
	keys := make([]string, 0, len(defaults))
	for i := range defaults {
		keys = append(keys, defaults[i].Key)
	}

	return r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		if len(keys) > 0 {
			if err := tx.Where("key NOT IN ?", keys).Delete(&model.Setting{}).Error; err != nil {
				return err
			}
		}

		for i := range defaults {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoNothing: true,
			}).Create(&defaults[i]).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Setting{}).
				Where("key = ?", defaults[i].Key).
				Updates(map[string]any{
					"desc":     defaults[i].Desc,
					"category": defaults[i].Category,
					"public":   defaults[i].Public,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormSettingRepo) UpsertByKey(ctx context.Context, setting *model.Setting) error {
	return r.db.WithContext(withContext(ctx)).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"value",
			"public",
		}),
	}).Create(setting).Error
}

func (r *GormSettingRepo) Update(ctx context.Context, setting *model.Setting) error {
	return r.db.WithContext(withContext(ctx)).Save(setting).Error
}

func (r *GormSettingRepo) DeleteByKey(ctx context.Context, key string) error {
	return r.db.WithContext(withContext(ctx)).Where("key = ?", key).Delete(&model.Setting{}).Error
}

func (r *GormSettingRepo) All(ctx context.Context) ([]model.Setting, error) {
	var settings []model.Setting
	if err := r.db.WithContext(withContext(ctx)).Order("key asc").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *GormSettingRepo) List(ctx context.Context, filter SettingFilter, opts ListOptions) ([]model.Setting, int64, error) {
	db := r.db.WithContext(withContext(ctx)).Model(&model.Setting{})

	if category := strings.TrimSpace(filter.Category); category != "" {
		db = db.Where("category = ?", category)
	}
	if key := strings.TrimSpace(filter.Key); key != "" {
		db = db.Where("key LIKE ?", "%"+key+"%")
	}
	if filter.Public != nil {
		db = db.Where("public = ?", *filter.Public)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var settings []model.Setting
	db = db.Order(opts.Order).Limit(opts.Limit).Offset(opts.Offset)
	if err := db.Find(&settings).Error; err != nil {
		return nil, 0, err
	}
	return settings, total, nil
}
