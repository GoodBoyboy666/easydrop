package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/storage"
)

// toUserDTO 将用户模型转换为对外返回的 DTO。
func toUserDTO(ctx context.Context, user *model.User, storageManager storage.Manager) (dto.UserDTO, error) {
	avatar, err := resolveUserAvatar(ctx, user.Avatar, user.Email, storageManager)
	if err != nil {
		return dto.UserDTO{}, err
	}

	return dto.UserDTO{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Admin:         user.Admin,
		Status:        user.Status,
		Avatar:        avatar,
		EmailVerified: user.EmailVerified,
		StorageQuota:  user.StorageQuota,
		StorageUsed:   user.StorageUsed,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// toUserDTOs 将用户模型切片转换为 DTO 列表。
func toUserDTOs(ctx context.Context, users []model.User, storageManager storage.Manager) ([]dto.UserDTO, error) {
	if len(users) == 0 {
		return nil, nil
	}

	items := make([]dto.UserDTO, 0, len(users))
	for i := range users {
		item, err := toUserDTO(ctx, &users[i], storageManager)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func resolveUserAvatar(ctx context.Context, avatar *string, email string, storageManager storage.Manager) (*string, error) {
	if avatar == nil {
		return buildGravatarURL(email), nil
	}

	trimmed := strings.TrimSpace(*avatar)
	if trimmed == "" {
		return nil, nil
	}
	if !isManagedAvatarKey(trimmed) || storageManager == nil {
		return &trimmed, nil
	}

	url, err := storageManager.URL(ctx, trimmed)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func buildGravatarURL(email string) *string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil
	}

	hash := md5.Sum([]byte(normalized))
	url := "https://gravatar.furwolf.com/avatar/" + hex.EncodeToString(hash[:])
	return &url
}

func isManagedAvatarKey(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), storage.CategoryAvatar+"/")
}
