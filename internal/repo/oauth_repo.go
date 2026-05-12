package repo

import (
	"context"

	"easydrop/internal/model"

	"gorm.io/gorm"
)

// OAuthBindRepo 定义社交登录绑定的仓储接口。
type OAuthBindRepo interface {
	// Create 创建一条社交账号绑定记录。
	Create(ctx context.Context, bind *model.OAuthBind) error
	// FindByProviderAndUID 根据社交平台和平台用户ID查找绑定记录。
	FindByProviderAndUID(ctx context.Context, provider, providerUserID string) (*model.OAuthBind, error)
	// FindByUserIDAndProvider 根据本地用户ID和社交平台查找绑定记录。
	FindByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*model.OAuthBind, error)
	// FindByUserID 查找指定用户的所有社交账号绑定。
	FindByUserID(ctx context.Context, userID uint) ([]model.OAuthBind, error)
	// Delete 根据主键ID删除一条绑定记录。
	Delete(ctx context.Context, id uint) error
}

// GormOAuthBindRepo 基于 Gorm 的社交登录绑定仓储实现。
type GormOAuthBindRepo struct {
	db *gorm.DB
}

// NewOAuthBindRepo 创建社交登录绑定仓储实例。
func NewOAuthBindRepo(db *gorm.DB) OAuthBindRepo {
	return &GormOAuthBindRepo{db: db}
}

// Create 创建一条社交账号绑定记录。
func (r *GormOAuthBindRepo) Create(ctx context.Context, bind *model.OAuthBind) error {
	return r.db.WithContext(withContext(ctx)).Create(bind).Error
}

// FindByProviderAndUID 根据社交平台和平台用户ID查找绑定记录。
// 返回 gorm.ErrRecordNotFound 表示不存在。
func (r *GormOAuthBindRepo) FindByProviderAndUID(ctx context.Context, provider, providerUserID string) (*model.OAuthBind, error) {
	var bind model.OAuthBind
	err := r.db.WithContext(withContext(ctx)).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&bind).Error
	if err != nil {
		return nil, err
	}
	return &bind, nil
}

// FindByUserIDAndProvider 根据本地用户ID和社交平台查找绑定记录。
// 返回 gorm.ErrRecordNotFound 表示不存在。
func (r *GormOAuthBindRepo) FindByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*model.OAuthBind, error) {
	var bind model.OAuthBind
	err := r.db.WithContext(withContext(ctx)).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&bind).Error
	if err != nil {
		return nil, err
	}
	return &bind, nil
}

// FindByUserID 查找指定用户的所有社交账号绑定。
func (r *GormOAuthBindRepo) FindByUserID(ctx context.Context, userID uint) ([]model.OAuthBind, error) {
	var binds []model.OAuthBind
	err := r.db.WithContext(withContext(ctx)).
		Where("user_id = ?", userID).
		Find(&binds).Error
	if err != nil {
		return nil, err
	}
	return binds, nil
}

// Delete 根据主键ID删除一条绑定记录。
func (r *GormOAuthBindRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Delete(&model.OAuthBind{}, id).Error
}
