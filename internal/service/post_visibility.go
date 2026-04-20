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
	// IncludeHiddenPosts 返回列表查询是否应包含隐藏说说。
	IncludeHiddenPosts(canViewHidden bool) bool
	// EnsurePostReadable 校验单条说说是否存在且对当前调用方可读。
	EnsurePostReadable(ctx context.Context, postRepo repo.PostRepo, postID uint, canViewHidden bool) (*model.Post, error)
}

type postVisibilityPolicy struct{}

// NewPostVisibilityPolicy 创建说说可见性策略实例。
func NewPostVisibilityPolicy() PostVisibilityPolicy {
	return &postVisibilityPolicy{}
}

// IncludeHiddenPosts 根据调用方权限决定列表查询是否包含隐藏说说。
func (p *postVisibilityPolicy) IncludeHiddenPosts(canViewHidden bool) bool {
	return canViewHidden
}

// EnsurePostReadable 校验说说存在性与可见性，并返回可读取的说说实体。
func (p *postVisibilityPolicy) EnsurePostReadable(ctx context.Context, postRepo repo.PostRepo, postID uint, canViewHidden bool) (*model.Post, error) {
	// 先读取说说，统一处理不存在与底层错误。
	post, err := postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		log.Printf("查询说说失败: %v", err)
		return nil, ErrInternal
	}
	// 再按权限检查可见性，避免泄露隐藏内容。
	if post.Hide && !canViewHidden {
		return nil, ErrPostNotFound
	}
	return post, nil
}
