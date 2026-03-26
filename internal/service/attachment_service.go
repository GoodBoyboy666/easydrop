package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

var (
	ErrAttachmentNotFound       = errors.New("附件不存在")
	ErrInvalidAttachmentBizType = errors.New("附件类型不合法")
	ErrInvalidFileSize          = errors.New("文件大小必须大于 0")
	ErrEmptyAttachmentContent   = errors.New("附件内容不能为空")
	ErrStorageQuotaExceeded     = errors.New("存储配额已满，无法上传")
	ErrFailedToCalculateQuota   = errors.New("计算存储配额失败")
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
	storageManager storage.Manager
	settings       SettingService
}

// NewAttachmentService 创建附件服务实例。
func NewAttachmentService(attachmentRepo repo.AttachmentRepo, userRepo repo.UserRepo, storageManager storage.Manager, settings SettingService) AttachmentService {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		userRepo:       userRepo,
		storageManager: storageManager,
		settings:       settings,
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

	if input.Content == nil || len(input.ContentSample) == 0 {
		return nil, ErrEmptyAttachmentContent
	}

	if err := validateAttachmentUpload(ctx, s.settings, input.OriginalFilename, input.ContentType, input.ContentSample); err != nil {
		return nil, err
	}

	bizType := resolveAttachmentBizType(input.ContentSample)

	if err := validateAttachmentBizType(bizType); err != nil {
		return nil, err
	}

	fileSize := input.FileSize
	if fileSize <= 0 {
		return nil, ErrInvalidFileSize
	}

	defaultQuota, err := s.getDefaultStorageQuota(ctx)
	if err != nil {
		return nil, err
	}

	fileKey, err := s.storageManager.NewObjectKey(storage.CategoryFile, input.UserID, input.OriginalFilename)
	if err != nil {
		log.Printf("生成附件 key 失败: %v", err)
		return nil, ErrInternal
	}

	if err := s.storageManager.UploadStream(ctx, fileKey, input.Content, fileSize, strings.TrimSpace(input.ContentType)); err != nil {
		log.Printf("上传附件失败: %v", err)
		return nil, ErrInternal
	}

	attachment := &model.Attachment{
		UserID:      input.UserID,
		StorageType: s.storageManager.BackendType(),
		FileKey:     fileKey,
		BizType:     bizType,
		FileSize:    fileSize,
	}

	if err := s.attachmentRepo.CreateWithQuotaTx(ctx, attachment, defaultQuota); err != nil {
		log.Printf("创建附件失败: %v", err)
		if deleteErr := s.storageManager.Delete(ctx, fileKey); deleteErr != nil {
			log.Printf("回滚附件对象失败: %v", deleteErr)
		}
		if errors.Is(err, repo.ErrAttachmentUploadQuotaExceeded) {
			return nil, ErrStorageQuotaExceeded
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, ErrInternal
	}

	d, err := s.toAttachmentDTO(ctx, attachment)
	if err != nil {
		log.Printf("生成附件 URL 失败: %v", err)
		return nil, ErrInternal
	}
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

	d, err := s.toAttachmentDTO(ctx, attachment)
	if err != nil {
		log.Printf("生成附件 URL 失败: %v", err)
		return nil, ErrInternal
	}
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

	d, err := s.toAttachmentDTO(ctx, attachment)
	if err != nil {
		log.Printf("生成附件 URL 失败: %v", err)
		return nil, ErrInternal
	}
	return &d, nil
}

// Delete 删除附件记录。
func (s *attachmentService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrAttachmentNotFound
	}

	attachment, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAttachmentNotFound
		}
		log.Printf("查询附件失败: %v", err)
		return ErrInternal
	}

	if err := s.storageManager.Delete(ctx, attachment.FileKey); err != nil {
		log.Printf("删除附件对象失败: %v", err)
		return ErrInternal
	}

	if err := s.attachmentRepo.DeleteWithStorageUsedTx(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAttachmentNotFound
		}
		log.Printf("删除附件失败: %v", err)
		return ErrInternal
	}
	return nil
}

// ListByUser 查询用户附件列表。
func (s *attachmentService) ListByUser(ctx context.Context, input dto.AttachmentListInput) (*dto.AttachmentListResult, error) {
	createdFrom := toTimePtrFromUnix(input.CreatedFrom)
	createdTo := toTimePtrFromUnix(input.CreatedTo)

	attachments, total, err := s.attachmentRepo.List(ctx, repo.AttachmentFilter{
		ID:          input.ID,
		UserID:      input.UserID,
		BizType:     input.BizType,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
	}, repo.ListOptions{
		Limit:  normalizeServiceListLimit(input.Limit),
		Offset: normalizeServiceListOffset(input.Offset),
		Order:  normalizeAttachmentListOrder(input.Order),
	})
	if err != nil {
		log.Printf("查询附件列表失败: %v", err)
		return nil, ErrInternal
	}

	items, err := s.toAttachmentDTOs(ctx, attachments)
	if err != nil {
		log.Printf("生成附件 URL 失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.AttachmentListResult{
		Items: items,
		Total: total,
	}, nil
}

func toTimePtrFromUnix(v *int64) *time.Time {
	if v == nil {
		return nil
	}
	t := time.Unix(*v, 0)
	return &t
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

// resolveAttachmentBizType 基于文件内容探测 MIME 类型并判定附件业务类型。
func resolveAttachmentBizType(content []byte) int {
	mediaType := ""
	if len(content) > 0 {
		mediaType = strings.ToLower(strings.TrimSpace(http.DetectContentType(content)))
		if idx := strings.Index(mediaType, ";"); idx >= 0 {
			mediaType = strings.TrimSpace(mediaType[:idx])
		}
	}

	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return model.AttachmentTypeImage
	case strings.HasPrefix(mediaType, "video/"):
		return model.AttachmentTypeVideo
	case strings.HasPrefix(mediaType, "audio/"):
		return model.AttachmentTypeAudio
	default:
		return model.AttachmentTypeFile
	}
}

// toAttachmentDTO 将附件模型转换为 DTO。
func (s *attachmentService) toAttachmentDTO(ctx context.Context, attachment *model.Attachment) (dto.AttachmentDTO, error) {
	url, err := s.storageManager.URL(ctx, attachment.FileKey)
	if err != nil {
		return dto.AttachmentDTO{}, err
	}

	return dto.AttachmentDTO{
		ID:          attachment.ID,
		UserID:      attachment.UserID,
		StorageType: attachment.StorageType,
		FileKey:     attachment.FileKey,
		URL:         url,
		BizType:     attachment.BizType,
		FileSize:    attachment.FileSize,
		CreatedAt:   attachment.CreatedAt,
	}, nil
}

// toAttachmentDTOs 将附件模型切片转换为 DTO 列表。
func (s *attachmentService) toAttachmentDTOs(ctx context.Context, attachments []model.Attachment) ([]dto.AttachmentDTO, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	items := make([]dto.AttachmentDTO, 0, len(attachments))
	for i := range attachments {
		item, err := s.toAttachmentDTO(ctx, &attachments[i])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// getDefaultStorageQuota 获取全局默认存储配额（字节）。
func (s *attachmentService) getDefaultStorageQuota(ctx context.Context) (int64, error) {
	return getDefaultStorageQuota(ctx, s.settings)
}
