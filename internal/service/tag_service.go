package service

import (
	"context"
	"log"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/repo"
)

// TagService 提供标签查询能力。
type TagService interface {
	// List 按关键字、分页和排序条件查询标签列表。
	List(ctx context.Context, input dto.TagListInput) (*dto.TagListResult, error)
}

type tagService struct {
	tagRepo repo.TagRepo
}

// NewTagService 创建标签服务实例。
func NewTagService(tagRepo repo.TagRepo) TagService {
	return &tagService{tagRepo: tagRepo}
}

// List 返回全站标签列表，支持搜索、分页和热门排序。
func (s *tagService) List(ctx context.Context, input dto.TagListInput) (*dto.TagListResult, error) {
	items, total, err := s.tagRepo.List(ctx, repo.TagFilter{
		Name: strings.TrimSpace(input.Keyword),
	}, repo.ListOptions{
		Limit:  normalizeServiceListLimit(input.Limit),
		Offset: normalizeServiceListOffset(input.Offset),
		Order:  normalizeTagListOrder(input.Order),
	})
	if err != nil {
		log.Printf("查询标签列表失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.TagListResult{
		Items: toTagDTOs(items),
		Total: total,
	}, nil
}
