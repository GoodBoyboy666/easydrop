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
	ErrAttachmentNotFound       = errors.New("附件不存在")
	ErrInvalidAttachmentBizType = errors.New("附件类型不合法")
	ErrInvalidFileSize          = errors.New("文件大小必须大于 0")
	ErrEmptyFileKey             = errors.New("文件 key 不能为空")
	ErrEmptyStorageType         = errors.New("存储类型不能为空")
)

// AttachmentService 提供附件管理相关的 CRUD 能力。
type AttachmentService interface {
	// Create 创建附件记录。
	Create(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error)
	// Get 根据 ID 获取附件详情。
	Get(ctx context.Context, id uint) (*dto.AttachmentDTO, error)
	// Update 更新附件信息（支持重分类）。
	Update(ctx context.Context, input dto.AttachmentUpdateInput) (*dto.AttachmentDTO, error)
	// Delete 删除附件记录。
	Delete(ctx context.Context, id uint) error
	// ListByUser 查询用户附件列表。
	ListByUser(ctx context.Context, input dto.AttachmentListInput) (*dto.AttachmentListResult, error)
}

type attachmentService struct {
	attachmentRepo repo.AttachmentRepo
	userRepo       repo.UserRepo
}

// NewAttachmentService 创建附件服务实例。
func NewAttachmentService(attachmentRepo repo.AttachmentRepo, userRepo repo.UserRepo) AttachmentService {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		userRepo:       userRepo,
	}
}

// Create 校验输入后创建附件记录。
func (s *attachmentService) Create(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error) {
	if input.UserID == 0 {
		return nil, ErrUserNotFound
	}

	if err := s.ensureUserExists(ctx, input.UserID); err != nil {
		return nil, err
	}

	storageType := strings.TrimSpace(input.StorageType)
	if storageType == "" {
		return nil, ErrEmptyStorageType
	}

	fileKey := strings.TrimSpace(input.FileKey)
	if fileKey == "" {
		return nil, ErrEmptyFileKey
	}

	if err := validateAttachmentBizType(input.BizType); err != nil {
		return nil, err
	}

	if input.FileSize <= 0 {
		return nil, ErrInvalidFileSize
	}

	attachment := &model.Attachment{
		UserID:      input.UserID,
		StorageType: storageType,
		FileKey:     fileKey,
		BizType:     input.BizType,
		FileSize:    input.FileSize,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		log.Printf("创建附件失败: %v", err)
		return nil, ErrInternal
	}

	d := toAttachmentDTO(attachment)
	return &d, nil
}

// Get 根据附件 ID 查询详情。
func (s *attachmentService) Get(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
	if id == 0 {
		return nil, ErrAttachmentNotFound
	}

	attachment, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttachmentNotFound
		}
		log.Printf("查询附件失败: %v", err)
		return nil, ErrInternal
	}

	d := toAttachmentDTO(attachment)
	return &d, nil
}

// Update 更新附件信息。
func (s *attachmentService) Update(ctx context.Context, input dto.AttachmentUpdateInput) (*dto.AttachmentDTO, error) {
	if input.ID == 0 {
		return nil, ErrAttachmentNotFound
	}

	attachment, err := s.attachmentRepo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttachmentNotFound
		}
		log.Printf("查询附件失败: %v", err)
		return nil, ErrInternal
	}

	if input.BizType != nil {
		if err := validateAttachmentBizType(*input.BizType); err != nil {
			return nil, err
		}
		attachment.BizType = *input.BizType
	}

	if err := s.attachmentRepo.Update(ctx, attachment); err != nil {
		log.Printf("更新附件失败: %v", err)
		return nil, ErrInternal
	}

	d := toAttachmentDTO(attachment)
	return &d, nil
}

// Delete 删除附件记录。
func (s *attachmentService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrAttachmentNotFound
	}

	if _, err := s.attachmentRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAttachmentNotFound
		}
		log.Printf("查询附件失败: %v", err)
		return ErrInternal
	}

	if err := s.attachmentRepo.Delete(ctx, id); err != nil {
		log.Printf("删除附件失败: %v", err)
		return ErrInternal
	}
	return nil
}

// ListByUser 查询用户附件列表。
func (s *attachmentService) ListByUser(ctx context.Context, input dto.AttachmentListInput) (*dto.AttachmentListResult, error) {
	attachments, total, err := s.attachmentRepo.List(ctx, repo.AttachmentFilter{
		UserID:  input.UserID,
		BizType: input.BizType,
	}, repo.ListOptions{
		Limit:  normalizeServiceListLimit(input.Limit),
		Offset: normalizeServiceListOffset(input.Offset),
		Order:  normalizeAttachmentListOrder(input.Order),
	})
	if err != nil {
		log.Printf("查询附件列表失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.AttachmentListResult{
		Items: toAttachmentDTOs(attachments),
		Total: total,
	}, nil
}

// ensureUserExists 确保用户存在。
func (s *attachmentService) ensureUserExists(ctx context.Context, userID uint) error {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("查询用户失败: %v", err)
		return ErrInternal
	}
	return nil
}

// validateAttachmentBizType 校验附件类型值是否在允许范围内。
func validateAttachmentBizType(bizType int) error {
	switch bizType {
	case model.AttachmentTypeImage, model.AttachmentTypeVideo, model.AttachmentTypeAudio, model.AttachmentTypeFile:
		return nil
	default:
		return ErrInvalidAttachmentBizType
	}
}

// toAttachmentDTO 将附件模型转换为 DTO。
func toAttachmentDTO(attachment *model.Attachment) dto.AttachmentDTO {
	return dto.AttachmentDTO{
		ID:          attachment.ID,
		UserID:      attachment.UserID,
		StorageType: attachment.StorageType,
		FileKey:     attachment.FileKey,
		BizType:     attachment.BizType,
		FileSize:    attachment.FileSize,
		CreatedAt:   attachment.CreatedAt,
	}
}

// toAttachmentDTOs 将附件模型切片转换为 DTO 列表。
func toAttachmentDTOs(attachments []model.Attachment) []dto.AttachmentDTO {
	if len(attachments) == 0 {
		return nil
	}
	items := make([]dto.AttachmentDTO, 0, len(attachments))
	for i := range attachments {
		items = append(items, toAttachmentDTO(&attachments[i]))
	}
	return items
}
