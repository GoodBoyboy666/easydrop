package service

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

var (
	ErrPostNotFound     = errors.New("说说不存在")
	ErrEmptyPostContent = errors.New("说说内容不能为空")
	ErrInvalidPostUser  = errors.New("用户不能为空")
	ErrTagNameTooLong   = errors.New("标签名称长度不能超过 50")
)

var tagPattern = regexp.MustCompile(`#([^\s]+)`)

type PostService interface {
	Create(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error)
	Get(ctx context.Context, id uint) (*dto.PostDTO, error)
	Update(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error)
}

type postService struct {
	postRepo repo.PostRepo
	tagRepo  repo.TagRepo
}

func NewPostService(postRepo repo.PostRepo, tagRepo repo.TagRepo) PostService {
	return &postService{
		postRepo: postRepo,
		tagRepo:  tagRepo,
	}
}

func (s *postService) Create(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error) {
	if input.UserID == 0 {
		return nil, ErrInvalidPostUser
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrEmptyPostContent
	}

	tags, err := s.buildTagsFromContent(ctx, content)
	if err != nil {
		return nil, err
	}

	post := &model.Post{
		Content: content,
		UserID:  input.UserID,
		Tags:    tags,
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		log.Printf("创建说说失败: %v", err)
		return nil, ErrInternal
	}
	return toPostDTO(post), nil
}

func (s *postService) Get(ctx context.Context, id uint) (*dto.PostDTO, error) {
	if id == 0 {
		return nil, ErrPostNotFound
	}
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		log.Printf("获取说说失败: %v", err)
		return nil, ErrInternal
	}
	return toPostDTO(post), nil
}

func (s *postService) Update(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error) {
	if input.ID == 0 {
		return nil, ErrPostNotFound
	}
	post, err := s.postRepo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		log.Printf("获取说说失败: %v", err)
		return nil, ErrInternal
	}

	var oldTagIDs []uint
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, ErrEmptyPostContent
		}
		oldTagIDs = collectTagIDs(post.Tags)
		post.Content = content
		tags, err := s.buildTagsFromContent(ctx, content)
		if err != nil {
			return nil, err
		}
		post.Tags = tags
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		log.Printf("更新说说失败: %v", err)
		return nil, ErrInternal
	}
	if len(oldTagIDs) > 0 {
		s.asyncCleanupOrphanTags(oldTagIDs)
	}
	return toPostDTO(post), nil
}

func (s *postService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrPostNotFound
	}
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		log.Printf("获取说说失败: %v", err)
		return ErrInternal
	}
	tagIDs := collectTagIDs(post.Tags)
	if err := s.postRepo.Delete(ctx, id); err != nil {
		log.Printf("删除说说失败: %v", err)
		return ErrInternal
	}
	if len(tagIDs) > 0 {
		s.asyncCleanupOrphanTags(tagIDs)
	}
	return nil
}

func (s *postService) List(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
	posts, total, err := s.postRepo.List(ctx, repo.PostFilter{
		UserID: input.UserID,
		TagID:  input.TagID,
	}, repo.ListOptions{
		Limit:  input.Limit,
		Offset: input.Offset,
		Order:  input.Order,
	})
	if err != nil {
		log.Printf("查询说说列表失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.PostListResult{
		Items: toPostDTOs(posts),
		Total: total,
	}, nil
}

func (s *postService) buildTagsFromContent(ctx context.Context, content string) ([]model.Tag, error) {
	cleanNames := extractTagNames(content)
	if len(cleanNames) == 0 {
		return nil, nil
	}

	tags := make([]model.Tag, 0, len(cleanNames))
	for _, name := range cleanNames {
		if utf8.RuneCountInString(name) > 50 {
			return nil, ErrTagNameTooLong
		}
		tag, err := s.tagRepo.GetByName(ctx, name)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newTag := &model.Tag{Name: name}
				if err := s.tagRepo.Create(ctx, newTag); err != nil {
					log.Printf("创建标签失败: %v", err)
					return nil, ErrInternal
				}
				tags = append(tags, *newTag)
				continue
			}
			log.Printf("获取标签失败: %v", err)
			return nil, ErrInternal
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}

func (s *postService) asyncCleanupOrphanTags(tagIDs []uint) {
	ids := append([]uint(nil), tagIDs...)
	go func() {
		if _, err := s.tagRepo.CleanupOrphanTags(context.Background(), ids); err != nil {
			log.Printf("清理孤儿标签失败: %v", err)
		}
	}()
}

func extractTagNames(content string) []string {
	if content == "" {
		return nil
	}

	matches := tagPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	names := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		name := match[1]
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return normalizeTagNames(names)
}

func normalizeTagNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(names))
	cleaned := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		cleaned = append(cleaned, name)
	}
	return cleaned
}

func collectTagIDs(tags []model.Tag) []uint {
	if len(tags) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(tags))
	for _, tag := range tags {
		if tag.ID == 0 {
			continue
		}
		ids = append(ids, tag.ID)
	}
	return ids
}

func toPostDTO(post *model.Post) *dto.PostDTO {
	if post == nil {
		return nil
	}
	return &dto.PostDTO{
		ID:        post.ID,
		Content:   post.Content,
		UserID:    post.UserID,
		Tags:      toTagDTOs(post.Tags),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}

func toPostDTOs(posts []model.Post) []dto.PostDTO {
	if len(posts) == 0 {
		return nil
	}
	items := make([]dto.PostDTO, 0, len(posts))
	for i := range posts {
		postDTO := toPostDTO(&posts[i])
		if postDTO == nil {
			continue
		}
		items = append(items, *postDTO)
	}
	return items
}

func toTagDTOs(tags []model.Tag) []dto.TagDTO {
	if len(tags) == 0 {
		return nil
	}
	items := make([]dto.TagDTO, 0, len(tags))
	for _, tag := range tags {
		items = append(items, dto.TagDTO{
			ID:        tag.ID,
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt,
		})
	}
	return items
}
