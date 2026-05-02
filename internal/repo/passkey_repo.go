package repo

import (
	"context"
	"encoding/json"
	"errors"

	"easydrop/internal/model"

	wa "github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

var ErrPasskeyLimitExceeded = errors.New("达到通行密钥数量上限")

type PasskeyRepo interface {
	Create(ctx context.Context, p *model.PasskeyCredential) error
	CreateWithLimit(ctx context.Context, p *model.PasskeyCredential, max int) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.PasskeyCredential, error)
	FindByCredentialID(ctx context.Context, credentialID string) (*model.PasskeyCredential, error)
	FindByUserID(ctx context.Context, userID uint) ([]model.PasskeyCredential, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	UpdateName(ctx context.Context, id uint, name string) error
	UpdateCredential(ctx context.Context, id uint, credential *wa.Credential) error
	Delete(ctx context.Context, id uint) error
}

type gormPasskeyRepo struct {
	db *gorm.DB
}

func NewPasskeyRepo(db *gorm.DB) PasskeyRepo {
	return &gormPasskeyRepo{db: db}
}

func (r *gormPasskeyRepo) Create(ctx context.Context, p *model.PasskeyCredential) error {
	return r.db.WithContext(withContext(ctx)).Create(p).Error
}

// CreateWithLimit 在事务中原子检查数量上限后创建通行密钥，防止并发突破上限。
// 返回创建前的已有数量，用于自动命名。
func (r *gormPasskeyRepo) CreateWithLimit(ctx context.Context, p *model.PasskeyCredential, max int) (int64, error) {
	var count int64

	err := r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.PasskeyCredential{}).Where("user_id = ?", p.UserID).Count(&count).Error; err != nil {
			return err
		}
		if count >= int64(max) {
			return ErrPasskeyLimitExceeded
		}
		return tx.Create(p).Error
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *gormPasskeyRepo) FindByID(ctx context.Context, id uint) (*model.PasskeyCredential, error) {
	var p model.PasskeyCredential
	err := r.db.WithContext(withContext(ctx)).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *gormPasskeyRepo) FindByCredentialID(ctx context.Context, credentialID string) (*model.PasskeyCredential, error) {
	var p model.PasskeyCredential
	err := r.db.WithContext(withContext(ctx)).Where("credential_id = ?", credentialID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *gormPasskeyRepo) FindByUserID(ctx context.Context, userID uint) ([]model.PasskeyCredential, error) {
	var passkeys []model.PasskeyCredential
	err := r.db.WithContext(withContext(ctx)).Where("user_id = ?", userID).Order("created_at ASC").Find(&passkeys).Error
	return passkeys, err
}

func (r *gormPasskeyRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(withContext(ctx)).Model(&model.PasskeyCredential{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *gormPasskeyRepo) UpdateName(ctx context.Context, id uint, name string) error {
	return r.db.WithContext(withContext(ctx)).Model(&model.PasskeyCredential{}).Where("id = ?", id).Update("name", name).Error
}

func (r *gormPasskeyRepo) UpdateCredential(ctx context.Context, id uint, credential *wa.Credential) error {
	data, err := json.Marshal(credential)
	if err != nil {
		return err
	}
	return r.db.WithContext(withContext(ctx)).Model(&model.PasskeyCredential{}).Where("id = ?", id).Update("credential_json", data).Error
}

func (r *gormPasskeyRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Delete(&model.PasskeyCredential{}, id).Error
}
