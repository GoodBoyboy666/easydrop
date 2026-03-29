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
	"easydrop/internal/pkg/storage"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

var (
	ErrPostNotFound     = errors.New("说说不存在")
	ErrEmptyPostContent = errors.New("说说内容不能为空")
	ErrInvalidPostUser  = errors.New("用户不能为空")
	ErrTagNameTooLong   = errors.New("标签名称长度不能超过 50")
)

var tagPattern = regexp.MustCompile(`#([^#]+)#`)

type PostService interface {
	// Create 创建说说并自动关联正文中解析出的标签。
	Create(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error)
	// Get 按 ID 获取单条说说详情。
	Get(ctx context.Context, id uint) (*dto.PostDTO, error)
	// Update 更新说说内容并同步刷新标签关联。
	Update(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error)
	// Delete 删除说说并异步清理可能失效的标签。
	Delete(ctx context.Context, id uint) error
	// List 按条件查询说说列表。
	List(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error)
}

type postService struct {
	postRepo       repo.PostRepo
	commentRepo    repo.CommentRepo
	tagRepo        repo.TagRepo
	storageManager storage.Manager
}

// NewPostService 创建说说服务实例。
func NewPostService(postRepo repo.PostRepo, commentRepo repo.CommentRepo, tagRepo repo.TagRepo, storageManager storage.Manager) PostService {
	return &postService{
		postRepo:       postRepo,
		commentRepo:    commentRepo,
		tagRepo:        tagRepo,
		storageManager: storageManager,
	}
}

// Create 校验输入内容并创建新的说说记录。
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
		Content:        content,
		Hide:           input.Hide,
		DisableComment: input.DisableComment,
		Pin:            input.Pin,
		UserID:         input.UserID,
		Tags:           tags,
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		log.Printf("创建说说失败: %v", err)
		return nil, ErrInternal
	}

	createdPost, err := s.postRepo.GetByID(ctx, post.ID)
	if err != nil {
		log.Printf("查询已创建说说失败: %v", err)
		return nil, ErrInternal
	}
	postDTO, err := toPostDTO(ctx, createdPost, s.storageManager)
	if err != nil {
		log.Printf("解析说说头像失败: %v", err)
		return nil, ErrInternal
	}
	return postDTO, nil
}

// Get 根据说说 ID 查询详情。
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
	postDTO, err := toPostDTO(ctx, post, s.storageManager)
	if err != nil {
		log.Printf("解析说说头像失败: %v", err)
		return nil, ErrInternal
	}
	return postDTO, nil
}

// Update 更新说说内容，并在内容变化时重建标签关系。
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
	if input.Hide != nil {
		post.Hide = *input.Hide
	}
	if input.DisableComment != nil {
		post.DisableComment = *input.DisableComment
	}
	if input.ClearPin != nil && *input.ClearPin {
		post.Pin = nil
	} else if input.Pin != nil {
		post.Pin = input.Pin
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		log.Printf("更新说说失败: %v", err)
		return nil, ErrInternal
	}
	if len(oldTagIDs) > 0 {
		s.asyncCleanupOrphanTags(oldTagIDs)
	}
	postDTO, err := toPostDTO(ctx, post, s.storageManager)
	if err != nil {
		log.Printf("解析说说头像失败: %v", err)
		return nil, ErrInternal
	}
	return postDTO, nil
}

// Delete 删除说说并在后台尝试清理孤儿标签。
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
	if err := s.commentRepo.DeleteByPostID(ctx, id); err != nil {
		log.Printf("删除说说评论失败: %v", err)
		return ErrInternal
	}
	if err := s.postRepo.Delete(ctx, id); err != nil {
		log.Printf("删除说说失败: %v", err)
		return ErrInternal
	}
	if len(tagIDs) > 0 {
		s.asyncCleanupOrphanTags(tagIDs)
	}
	return nil
}

// List 根据用户、标签和分页条件返回说说列表。
func (s *postService) List(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
	page, size := normalizeServiceListPageSize(input.Page, input.Size)

	posts, total, err := s.postRepo.List(ctx, repo.PostFilter{
		UserID:  input.UserID,
		TagID:   input.TagID,
		Content: strings.TrimSpace(input.Content),
		Hide:    input.Hide,
	}, repo.ListOptions{
		Limit:  size,
		Offset: pageSizeToOffset(page, size),
		Order:  normalizePostListOrder(input.Order),
	})
	if err != nil {
		log.Printf("查询说说列表失败: %v", err)
		return nil, ErrInternal
	}

	items, err := toPostDTOs(ctx, posts, s.storageManager)
	if err != nil {
		log.Printf("解析说说列表头像失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.PostListResult{
		Items: items,
		Total: total,
	}, nil
}

// buildTagsFromContent 从正文中提取标签，不存在时自动创建。
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

// asyncCleanupOrphanTags 异步删除不再被任何说说引用的标签。
func (s *postService) asyncCleanupOrphanTags(tagIDs []uint) {
	ids := append([]uint(nil), tagIDs...)
	go func() {
		if _, err := s.tagRepo.CleanupOrphanTags(context.Background(), ids); err != nil {
			log.Printf("清理孤儿标签失败: %v", err)
		}
	}()
}

// extractTagNames 使用标签正则从正文中提取原始标签名。
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

// normalizeTagNames 清理空白并去重，得到规范化标签名列表。
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

// collectTagIDs 提取标签 ID 列表，忽略无效 ID。
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

// toPostDTO 将说说模型转换为单个 DTO。
func toPostDTO(ctx context.Context, post *model.Post, storageManager storage.Manager) (*dto.PostDTO, error) {
	if post == nil {
		return nil, nil
	}

	avatar, err := resolveUserAvatar(ctx, post.User.Avatar, post.User.Email, storageManager)
	if err != nil {
		return nil, err
	}

	return &dto.PostDTO{
		ID:             post.ID,
		Content:        post.Content,
		Hide:           post.Hide,
		DisableComment: post.DisableComment,
		Pin:            post.Pin,
		Author: dto.PostAuthorDTO{
			ID:       post.User.ID,
			Nickname: post.User.Nickname,
			Avatar:   avatar,
			Admin:    post.User.Admin,
		},
		Tags:      toTagDTOs(post.Tags),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

// toPostDTOs 将说说模型切片转换为 DTO 列表。
func toPostDTOs(ctx context.Context, posts []model.Post, storageManager storage.Manager) ([]dto.PostDTO, error) {
	if len(posts) == 0 {
		return nil, nil
	}
	items := make([]dto.PostDTO, 0, len(posts))
	for i := range posts {
		postDTO, err := toPostDTO(ctx, &posts[i], storageManager)
		if err != nil {
			return nil, err
		}
		if postDTO == nil {
			continue
		}
		items = append(items, *postDTO)
	}
	return items, nil
}

// toTagDTOs 将标签模型切片转换为 DTO 列表。
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
