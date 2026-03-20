package repo

import (
	"context"
	"errors"
	"strings"

	"easydrop/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUserAvatarQuotaExceeded = errors.New("user avatar quota exceeded")

// UserRepo 定义用户仓储接口。
type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsernameOrEmail(ctx context.Context, value string) (*model.User, error)
	UpdateAvatarWithStorageUsedTx(ctx context.Context, userID uint, avatar *string, sizeDelta int64, defaultQuota int64) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter UserFilter, opts ListOptions) ([]model.User, int64, error)
}

// UserFilter 用户查询过滤条件。
type UserFilter struct {
	Username string
	Email    string
	Status   *int
}

// GormUserRepo 基于 Gorm 的用户仓储实现。
type GormUserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储实例。
func NewUserRepo(db *gorm.DB) UserRepo {
	return &GormUserRepo{db: db}
}

func (r *GormUserRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(withContext(ctx)).Create(user).Error
}

func (r *GormUserRepo) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(withContext(ctx)).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(withContext(ctx)).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(withContext(ctx)).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) GetByUsernameOrEmail(ctx context.Context, value string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(withContext(ctx)).Where("username = ? OR email = ?", value, value).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) UpdateAvatarWithStorageUsedTx(ctx context.Context, userID uint, avatar *string, sizeDelta int64, defaultQuota int64) (*model.User, error) {
	var updated model.User

	err := r.db.WithContext(withContext(ctx)).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		newUsed := user.StorageUsed + sizeDelta
		if newUsed < 0 {
			newUsed = 0
		}

		effectiveQuota := defaultQuota
		if user.StorageQuota != nil && *user.StorageQuota > 0 {
			effectiveQuota = *user.StorageQuota
		}
		if sizeDelta > 0 && effectiveQuota > 0 && newUsed > effectiveQuota {
			return ErrUserAvatarQuotaExceeded
		}

		user.Avatar = avatar
		user.StorageUsed = newUsed
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		updated = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *GormUserRepo) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(withContext(ctx)).Save(user).Error
}

func (r *GormUserRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(withContext(ctx)).Delete(&model.User{}, id).Error
}

func (r *GormUserRepo) List(ctx context.Context, filter UserFilter, opts ListOptions) ([]model.User, int64, error) {
	db := r.db.WithContext(withContext(ctx)).Model(&model.User{})

	if username := strings.TrimSpace(filter.Username); username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if email := strings.TrimSpace(filter.Email); email != "" {
		db = db.Where("email LIKE ?", "%"+email+"%")
	}
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	db = applyListOptions(db, opts, "created_at asc")
	if err := db.Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
