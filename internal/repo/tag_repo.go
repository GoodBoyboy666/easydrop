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
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	Update(ctx context.Context, tag *model.Tag) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter TagFilter, opts ListOptions) ([]model.Tag, int64, error)
	CleanupOrphanTags(ctx context.Context, tagIDs []uint) (int64, error)
}

// TagFilter 标签查询过滤条件。
type TagFilter struct {
	Name string
}

// GormTagRepo 基于 Gorm 的标签仓储实现。
type GormTagRepo struct {
	db *gorm.DB
}

// NewTagRepo 创建标签仓储实例。
func NewTagRepo(db *gorm.DB) TagRepo {
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

func (r *GormTagRepo) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(withContext(ctx)).
		Where("name = ?", strings.TrimSpace(name)).
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

func (r *GormTagRepo) CleanupOrphanTags(ctx context.Context, tagIDs []uint) (int64, error) {
	ids := uniqueUintIDs(tagIDs)
	if len(ids) == 0 {
		return 0, nil
	}

	var usedIDs []uint
	if err := r.db.WithContext(withContext(ctx)).
		Table("post_tags").
		Distinct("tag_id").
		Where("tag_id IN ?", ids).
		Pluck("tag_id", &usedIDs).Error; err != nil {
		return 0, err
	}

	used := make(map[uint]struct{}, len(usedIDs))
	for _, id := range usedIDs {
		used[id] = struct{}{}
	}

	orphans := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := used[id]; !ok {
			orphans = append(orphans, id)
		}
	}
	if len(orphans) == 0 {
		return 0, nil
	}

	result := r.db.WithContext(withContext(ctx)).Where("id IN ?", orphans).Delete(&model.Tag{})
	return result.RowsAffected, result.Error
}

func uniqueUintIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[uint]struct{}, len(ids))
	unique := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
