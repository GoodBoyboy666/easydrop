package service

import (
	"context"
	"errors"
	"log"

	"easydrop/internal/model"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

// PostVisibilityPolicy 统一封装说说可见性规则。
type PostVisibilityPolicy interface {
	IncludeHiddenPosts(canViewHidden bool) bool
	EnsurePostReadable(ctx context.Context, postRepo repo.PostRepo, postID uint, canViewHidden bool) (*model.Post, error)
}

type postVisibilityPolicy struct{}

// NewPostVisibilityPolicy 创建说说可见性策略实例。
func NewPostVisibilityPolicy() PostVisibilityPolicy {
	return &postVisibilityPolicy{}
}

func (p *postVisibilityPolicy) IncludeHiddenPosts(canViewHidden bool) bool {
	return canViewHidden
}

func (p *postVisibilityPolicy) EnsurePostReadable(ctx context.Context, postRepo repo.PostRepo, postID uint, canViewHidden bool) (*model.Post, error) {
	post, err := postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		log.Printf("查询说说失败: %v", err)
		return nil, ErrInternal
	}
	if post.Hide && !canViewHidden {
		return nil, ErrPostNotFound
	}
	return post, nil
}
