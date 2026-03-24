package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

var (
	ErrCommentNotFound      = errors.New("评论不存在")
	ErrEmptyCommentContent  = errors.New("评论内容不能为空")
	ErrInvalidCommentPost   = errors.New("说说不能为空")
	ErrInvalidCommentUser   = errors.New("用户不能为空")
	ErrInvalidCommentParent = errors.New("回复目标评论不合法")
	ErrPostCommentDisabled  = errors.New("该说说已关闭评论")
)

// CommentService 提供评论相关的 CRUD 能力。
type CommentService interface {
	// Create 创建评论，并在有 parent_id 时按扁平接楼规则填充引用关系。
	Create(ctx context.Context, input dto.CommentCreateInput) (*dto.CommentDTO, error)
	// Get 根据 ID 获取评论详情。
	Get(ctx context.Context, id uint) (*dto.CommentDTO, error)
	// Update 更新评论内容。
	Update(ctx context.Context, input dto.CommentUpdateInput) (*dto.CommentDTO, error)
	// Delete 删除评论。
	Delete(ctx context.Context, id uint) error
	// ListByPost 查询指定说说下的评论列表。
	ListByPost(ctx context.Context, input dto.CommentListInput) (*dto.CommentListResult, error)
	// ListByUser 查询指定用户的评论列表。
	ListByUser(ctx context.Context, input dto.CommentUserListInput) (*dto.CommentListResult, error)
	// List 查询评论列表（管理端）。
	List(ctx context.Context, input dto.CommentAdminListInput) (*dto.CommentListResult, error)
}

type commentService struct {
	commentRepo repo.CommentRepo
	postRepo    repo.PostRepo
	userRepo    repo.UserRepo
}

// NewCommentService 创建评论服务实例。
func NewCommentService(commentRepo repo.CommentRepo, postRepo repo.PostRepo, userRepo repo.UserRepo) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

// Create 校验评论输入并创建。
func (s *commentService) Create(ctx context.Context, input dto.CommentCreateInput) (*dto.CommentDTO, error) {
	if input.PostID == 0 {
		return nil, ErrInvalidCommentPost
	}
	if input.UserID == 0 {
		return nil, ErrInvalidCommentUser
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrEmptyCommentContent
	}

	post, err := s.postRepo.GetByID(ctx, input.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		log.Printf("查询说说失败: %v", err)
		return nil, ErrInternal
	}
	if post.DisableComment {
		return nil, ErrPostCommentDisabled
	}

	if err := s.ensureUserExists(ctx, input.UserID); err != nil {
		return nil, err
	}

	var parentID *uint
	var rootID *uint
	var replyToUserID *uint
	if input.ParentID != nil {
		if *input.ParentID == 0 {
			return nil, ErrInvalidCommentParent
		}
		parentComment, err := s.commentRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrInvalidCommentParent
			}
			log.Printf("查询父评论失败: %v", err)
			return nil, ErrInternal
		}
		if parentComment.PostID != input.PostID {
			return nil, ErrInvalidCommentParent
		}

		pid := parentComment.ID
		parentID = &pid
		if parentComment.RootID != nil {
			root := *parentComment.RootID
			rootID = &root
		} else {
			root := parentComment.ID
			rootID = &root
		}
		replyUID := parentComment.UserID
		replyToUserID = &replyUID
	}

	comment := &model.Comment{
		PostID:        input.PostID,
		UserID:        input.UserID,
		Content:       content,
		ParentID:      parentID,
		RootID:        rootID,
		ReplyToUserID: replyToUserID,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		log.Printf("创建评论失败: %v", err)
		return nil, ErrInternal
	}

	createdComment, err := s.commentRepo.GetByID(ctx, comment.ID)
	if err != nil {
		log.Printf("查询已创建评论失败: %v", err)
		return nil, ErrInternal
	}

	d := toCommentDTO(createdComment)
	return &d, nil
}

// Get 根据评论 ID 查询详情。
func (s *commentService) Get(ctx context.Context, id uint) (*dto.CommentDTO, error) {
	if id == 0 {
		return nil, ErrCommentNotFound
	}

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		log.Printf("查询评论失败: %v", err)
		return nil, ErrInternal
	}

	d := toCommentDTO(comment)
	return &d, nil
}

// Update 更新评论内容。
func (s *commentService) Update(ctx context.Context, input dto.CommentUpdateInput) (*dto.CommentDTO, error) {
	if input.ID == 0 {
		return nil, ErrCommentNotFound
	}

	comment, err := s.commentRepo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		log.Printf("查询评论失败: %v", err)
		return nil, ErrInternal
	}

	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, ErrEmptyCommentContent
		}
		comment.Content = content
	}

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		log.Printf("更新评论失败: %v", err)
		return nil, ErrInternal
	}

	d := toCommentDTO(comment)
	return &d, nil
}

// Delete 删除评论。
func (s *commentService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrCommentNotFound
	}

	if _, err := s.commentRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		log.Printf("查询评论失败: %v", err)
		return ErrInternal
	}

	if err := s.commentRepo.Delete(ctx, id); err != nil {
		log.Printf("删除评论失败: %v", err)
		return ErrInternal
	}
	return nil
}

// ListByPost 查询指定说说的评论列表。
func (s *commentService) ListByPost(ctx context.Context, input dto.CommentListInput) (*dto.CommentListResult, error) {
	if input.PostID == 0 {
		return nil, ErrInvalidCommentPost
	}
	if err := s.ensurePostExists(ctx, input.PostID); err != nil {
		return nil, err
	}

	postID := input.PostID
	return s.list(ctx, repo.CommentFilter{PostID: &postID}, input.Limit, input.Offset, input.Order)
}

// ListByUser 查询指定用户的评论列表。
func (s *commentService) ListByUser(ctx context.Context, input dto.CommentUserListInput) (*dto.CommentListResult, error) {
	if input.UserID == 0 {
		return nil, ErrInvalidCommentUser
	}
	if input.PostID != nil && *input.PostID == 0 {
		return nil, ErrInvalidCommentPost
	}
	if input.PostID != nil {
		if err := s.ensurePostExists(ctx, *input.PostID); err != nil {
			return nil, err
		}
	}

	userID := input.UserID
	return s.list(ctx, repo.CommentFilter{PostID: input.PostID, UserID: &userID}, input.Limit, input.Offset, input.Order)
}

// List 查询评论列表（管理端）。
func (s *commentService) List(ctx context.Context, input dto.CommentAdminListInput) (*dto.CommentListResult, error) {
	if input.PostID != nil && *input.PostID == 0 {
		return nil, ErrInvalidCommentPost
	}
	if input.UserID != nil && *input.UserID == 0 {
		return nil, ErrInvalidCommentUser
	}
	if input.PostID != nil {
		if err := s.ensurePostExists(ctx, *input.PostID); err != nil {
			return nil, err
		}
	}

	return s.list(ctx, repo.CommentFilter{PostID: input.PostID, UserID: input.UserID}, input.Limit, input.Offset, input.Order)
}

func (s *commentService) list(ctx context.Context, filter repo.CommentFilter, limit, offset int, order string) (*dto.CommentListResult, error) {
	comments, total, err := s.commentRepo.List(ctx, repo.CommentFilter{
		PostID:   filter.PostID,
		UserID:   filter.UserID,
		RootID:   filter.RootID,
		ParentID: filter.ParentID,
	}, repo.ListOptions{
		Limit:  normalizeServiceListLimit(limit),
		Offset: normalizeServiceListOffset(offset),
		Order:  normalizeCommentListOrder(order),
	})
	if err != nil {
		log.Printf("查询评论列表失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.CommentListResult{
		Items: toCommentDTOs(comments),
		Total: total,
	}, nil
}

// ensurePostExists 确保评论绑定的说说存在。
func (s *commentService) ensurePostExists(ctx context.Context, postID uint) error {
	if _, err := s.postRepo.GetByID(ctx, postID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		log.Printf("查询说说失败: %v", err)
		return ErrInternal
	}
	return nil
}

// ensureUserExists 确保评论用户存在。
func (s *commentService) ensureUserExists(ctx context.Context, userID uint) error {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("查询用户失败: %v", err)
		return ErrInternal
	}
	return nil
}

// toCommentDTO 将评论模型转换为 DTO。
func toCommentDTO(comment *model.Comment) dto.CommentDTO {
	var replyToUser *dto.CommentAuthorDTO
	if comment.ReplyToUser != nil {
		replyToUser = &dto.CommentAuthorDTO{
			ID:       comment.ReplyToUser.ID,
			Nickname: comment.ReplyToUser.Nickname,
			Avatar:   comment.ReplyToUser.Avatar,
			Admin:    comment.ReplyToUser.Admin,
		}
	}

	return dto.CommentDTO{
		ID:     comment.ID,
		PostID: comment.PostID,
		Author: dto.CommentAuthorDTO{
			ID:       comment.User.ID,
			Nickname: comment.User.Nickname,
			Avatar:   comment.User.Avatar,
			Admin:    comment.User.Admin,
		},
		Content:     comment.Content,
		ParentID:    comment.ParentID,
		RootID:      comment.RootID,
		ReplyToUser: replyToUser,
		CreatedAt:   comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
	}
}

// toCommentDTOs 将评论模型切片转换为 DTO 列表。
func toCommentDTOs(comments []model.Comment) []dto.CommentDTO {
	if len(comments) == 0 {
		return nil
	}
	items := make([]dto.CommentDTO, 0, len(comments))
	for i := range comments {
		items = append(items, toCommentDTO(&comments[i]))
	}
	return items
}
