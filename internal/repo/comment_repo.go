package repo

import (
	"context"

	"easydrop/internal/model"

	"gorm.io/gorm"
)

// CommentRepo 定义评论仓储接口。
type CommentRepo interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id uint) (*model.Comment, error)
	Update(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter CommentFilter, opts ListOptions) ([]model.Comment, int64, error)
}

// CommentFilter 评论查询过滤条件。
type CommentFilter struct {
	PostID   uint
	UserID   *uint
	RootID   *uint
	ParentID *uint
}

// GormCommentRepo 基于 Gorm 的评论仓储实现。
type GormCommentRepo struct {
	db *gorm.DB
}

// NewCommentRepo 创建评论仓储实例。
func NewCommentRepo(db *gorm.DB) CommentRepo {
	return &GormCommentRepo{db: db}
}

func (r *GormCommentRepo) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(withContext(ctx)).Create(comment).Error
}

func (r *GormCommentRepo) GetByID(ctx context.Context, id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(withContext(ctx)).First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *GormCommentRepo) Update(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(withContext(ctx)).Save(comment).Error
}

func (r *GormCommentRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Delete(&model.Comment{}, id).Error
}

func (r *GormCommentRepo) List(ctx context.Context, filter CommentFilter, opts ListOptions) ([]model.Comment, int64, error) {
	db := r.db.WithContext(withContext(ctx)).Model(&model.Comment{}).
		Where("post_id = ?", filter.PostID)

	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.RootID != nil {
		db = db.Where("root_id = ?", *filter.RootID)
	}
	if filter.ParentID != nil {
		db = db.Where("parent_id = ?", *filter.ParentID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []model.Comment
	db = applyListOptions(db, opts, "created_at asc")
	if err := db.Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}
