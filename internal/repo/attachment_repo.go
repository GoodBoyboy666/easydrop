package repo

import (
	"context"

	"easydrop/internal/model"

	"gorm.io/gorm"
)

// AttachmentRepo 定义附件仓储接口。
type AttachmentRepo interface {
	Create(ctx context.Context, attachment *model.Attachment) error
	GetByID(ctx context.Context, id uint) (*model.Attachment, error)
	GetByFileKey(ctx context.Context, fileKey string) (*model.Attachment, error)
	Update(ctx context.Context, attachment *model.Attachment) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter AttachmentFilter, opts ListOptions) ([]model.Attachment, int64, error)
}

// AttachmentFilter 附件查询过滤条件。
type AttachmentFilter struct {
	UserID  *uint
	BizType *int
}

// GormAttachmentRepo 基于 Gorm 的附件仓储实现。
type GormAttachmentRepo struct {
	db *gorm.DB
}

// NewAttachmentRepo 创建附件仓储实例。
func NewAttachmentRepo(db *gorm.DB) *GormAttachmentRepo {
	return &GormAttachmentRepo{db: db}
}

func (r *GormAttachmentRepo) Create(ctx context.Context, attachment *model.Attachment) error {
	return r.db.WithContext(withContext(ctx)).Create(attachment).Error
}

func (r *GormAttachmentRepo) GetByID(ctx context.Context, id uint) (*model.Attachment, error) {
	var attachment model.Attachment
	err := r.db.WithContext(withContext(ctx)).First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *GormAttachmentRepo) GetByFileKey(ctx context.Context, fileKey string) (*model.Attachment, error) {
	var attachment model.Attachment
	err := r.db.WithContext(withContext(ctx)).Where("file_key = ?", fileKey).First(&attachment).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *GormAttachmentRepo) Update(ctx context.Context, attachment *model.Attachment) error {
	return r.db.WithContext(withContext(ctx)).Save(attachment).Error
}

func (r *GormAttachmentRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Delete(&model.Attachment{}, id).Error
}

func (r *GormAttachmentRepo) List(ctx context.Context, filter AttachmentFilter, opts ListOptions) ([]model.Attachment, int64, error) {
	db := r.db.WithContext(withContext(ctx)).Model(&model.Attachment{})

	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.BizType != nil {
		db = db.Where("biz_type = ?", *filter.BizType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var attachments []model.Attachment
	db = applyListOptions(db, opts, "created_at desc")
	if err := db.Find(&attachments).Error; err != nil {
		return nil, 0, err
	}
	return attachments, total, nil
}
