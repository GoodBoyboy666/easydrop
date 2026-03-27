package repo

import (
	"context"
	"time"

	"easydrop/internal/model"

	"gorm.io/gorm"
)

// OverviewRepo 定义后台概览聚合查询。
type OverviewRepo interface {
	GetSnapshot(ctx context.Context, since time.Time, until time.Time) (*OverviewSnapshot, error)
}

// OverviewSnapshot 表示后台概览原始聚合结果。
type OverviewSnapshot struct {
	UserTotal       int64
	PostTotal       int64
	CommentTotal    int64
	AttachmentTotal int64
	PostDaily       []OverviewDailyCount
	CommentDaily    []OverviewDailyCount
}

// OverviewDailyCount 表示单日聚合计数。
type OverviewDailyCount struct {
	Day   string `gorm:"column:day"`
	Total int64  `gorm:"column:total"`
}

// GormOverviewRepo 基于 Gorm 的后台概览仓储实现。
type GormOverviewRepo struct {
	db *gorm.DB
}

// NewOverviewRepo 创建后台概览仓储实例。
func NewOverviewRepo(db *gorm.DB) OverviewRepo {
	return &GormOverviewRepo{db: db}
}

func (r *GormOverviewRepo) GetSnapshot(ctx context.Context, since time.Time, until time.Time) (*OverviewSnapshot, error) {
	snapshot := &OverviewSnapshot{}

	if err := r.db.WithContext(withContext(ctx)).Model(&model.User{}).Count(&snapshot.UserTotal).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(withContext(ctx)).Model(&model.Post{}).Count(&snapshot.PostTotal).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(withContext(ctx)).Model(&model.Comment{}).Count(&snapshot.CommentTotal).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(withContext(ctx)).Model(&model.Attachment{}).Count(&snapshot.AttachmentTotal).Error; err != nil {
		return nil, err
	}

	postDaily, err := r.listDailyCounts(ctx, &model.Post{}, since, until)
	if err != nil {
		return nil, err
	}
	commentDaily, err := r.listDailyCounts(ctx, &model.Comment{}, since, until)
	if err != nil {
		return nil, err
	}

	snapshot.PostDaily = postDaily
	snapshot.CommentDaily = commentDaily
	return snapshot, nil
}

func (r *GormOverviewRepo) listDailyCounts(ctx context.Context, target any, since time.Time, until time.Time) ([]OverviewDailyCount, error) {
	var rows []OverviewDailyCount
	err := r.db.WithContext(withContext(ctx)).
		Model(target).
		Select("DATE(created_at) AS day, COUNT(*) AS total").
		Where("created_at >= ? AND created_at < ?", since, until).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
