package repo

import (
	"context"
	"encoding/json"

	"easydrop/internal/model"

	wa "github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type PasskeyRepo interface {
	Create(ctx context.Context, p *model.PasskeyCredential) error
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
