package repo

import (
	"context"
	"strings"

	"easydrop/internal/model"

	"gorm.io/gorm"
)

// TagRepo 定义标签仓储接口。
type TagRepo interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetByID(ctx context.Context, id uint) (*model.Tag, error)
	GetByName(ctx context.Context, userID uint, name string) (*model.Tag, error)
	Update(ctx context.Context, tag *model.Tag) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter TagFilter, opts ListOptions) ([]model.Tag, int64, error)
}

// TagFilter 标签查询过滤条件。
type TagFilter struct {
	UserID *uint
	Name   string
}

// GormTagRepo 基于 Gorm 的标签仓储实现。
type GormTagRepo struct {
	db *gorm.DB
}

// NewTagRepo 创建标签仓储实例。
func NewTagRepo(db *gorm.DB) *GormTagRepo {
	return &GormTagRepo{db: db}
}

func (r *GormTagRepo) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(withContext(ctx)).Create(tag).Error
}

func (r *GormTagRepo) GetByID(ctx context.Context, id uint) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(withContext(ctx)).First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *GormTagRepo) GetByName(ctx context.Context, userID uint, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(withContext(ctx)).
		Where("user_id = ? AND name = ?", userID, strings.TrimSpace(name)).
		First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *GormTagRepo) Update(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(withContext(ctx)).Save(tag).Error
}

func (r *GormTagRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Delete(&model.Tag{}, id).Error
}

func (r *GormTagRepo) List(ctx context.Context, filter TagFilter, opts ListOptions) ([]model.Tag, int64, error) {
	db := r.db.WithContext(withContext(ctx)).Model(&model.Tag{})

	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if name := strings.TrimSpace(filter.Name); name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tags []model.Tag
	db = applyListOptions(db, opts, "created_at desc")
	if err := db.Find(&tags).Error; err != nil {
		return nil, 0, err
	}
	return tags, total, nil
}
