package repo

import (
	"context"
	"errors"

	"easydrop/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrAttachmentUploadQuotaExceeded = errors.New("attachment upload quota exceeded")

// AttachmentRepo 定义附件仓储接口。
type AttachmentRepo interface {
	Create(ctx context.Context, attachment *model.Attachment) error
	CreateWithQuotaTx(ctx context.Context, attachment *model.Attachment, defaultQuota int64) error
	DeleteWithStorageUsedTx(ctx context.Context, id uint) error
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
func NewAttachmentRepo(db *gorm.DB) AttachmentRepo {
	return &GormAttachmentRepo{db: db}
}

func (r *GormAttachmentRepo) Create(ctx context.Context, attachment *model.Attachment) error {
	return r.db.WithContext(withContext(ctx)).Create(attachment).Error
}

// CreateWithQuotaTx 在一个事务中完成配额校验、附件创建与用户已用空间递增。
func (r *GormAttachmentRepo) CreateWithQuotaTx(ctx context.Context, attachment *model.Attachment, defaultQuota int64) error {
	return r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, attachment.UserID).Error; err != nil {
			return err
		}

		effectiveQuota := defaultQuota
		if user.StorageQuota != nil && *user.StorageQuota > 0 {
			effectiveQuota = *user.StorageQuota
		}

		if user.StorageUsed+attachment.FileSize > effectiveQuota {
			return ErrAttachmentUploadQuotaExceeded
		}

		if err := tx.Create(attachment).Error; err != nil {
			return err
		}

		return tx.Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("storage_used", gorm.Expr("storage_used + ?", attachment.FileSize)).Error
	})
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

// DeleteWithStorageUsedTx 在一个事务中删除附件并扣减用户已用存储。
func (r *GormAttachmentRepo) DeleteWithStorageUsedTx(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		var attachment model.Attachment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&attachment, id).Error; err != nil {
			return err
		}

		var user model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, attachment.UserID).Error; err != nil {
			return err
		}

		newUsed := user.StorageUsed - attachment.FileSize
		if newUsed < 0 {
			newUsed = 0
		}

		if err := tx.Delete(&model.Attachment{}, attachment.ID).Error; err != nil {
			return err
		}

		return tx.Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("storage_used", newUsed).Error
	})
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
