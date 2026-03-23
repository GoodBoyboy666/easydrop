package repo

import (
	"context"

	"easydrop/internal/model"

	"gorm.io/gorm"
)

// PostRepo 定义帖子仓储接口。
type PostRepo interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id uint) (*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter PostFilter, opts ListOptions) ([]model.Post, int64, error)
}

// PostFilter 帖子查询过滤条件。
type PostFilter struct {
	UserID *uint
	TagID  *uint
	Hide   *bool
}

// GormPostRepo 基于 Gorm 的帖子仓储实现。
type GormPostRepo struct {
	db *gorm.DB
}

// NewPostRepo 创建帖子仓储实例。
func NewPostRepo(db *gorm.DB) PostRepo {
	return &GormPostRepo{db: db}
}

func (r *GormPostRepo) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(withContext(ctx)).Create(post).Error
}

func (r *GormPostRepo) GetByID(ctx context.Context, id uint) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(withContext(ctx)).Preload("Tags").Preload("User").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *GormPostRepo) Update(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tags").Save(post).Error; err != nil {
			return err
		}
		return tx.Model(post).Association("Tags").Replace(post.Tags)
	})
}

func (r *GormPostRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		post := model.Post{ID: id}
		if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
			return err
		}
		return tx.Delete(&model.Post{}, id).Error
	})
}

func (r *GormPostRepo) List(ctx context.Context, filter PostFilter, opts ListOptions) ([]model.Post, int64, error) {
	db := r.db.WithContext(withContext(ctx)).Model(&model.Post{})

	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.TagID != nil {
		db = db.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Where("post_tags.tag_id = ?", *filter.TagID)
	}
	if filter.Hide != nil {
		db = db.Where("hide = ?", *filter.Hide)
	}

	countDB := db
	if filter.TagID != nil {
		countDB = countDB.Distinct("posts.id")
	}

	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []model.Post
	db = applyListOptions(db, opts, "created_at desc")
	if filter.TagID != nil {
		db = db.Distinct("posts.id")
	}
	if err := db.Preload("Tags").Preload("User").Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}
