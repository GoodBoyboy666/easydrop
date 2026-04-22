package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	avatarpkg "easydrop/internal/pkg/avatar"
	"easydrop/internal/pkg/storage"
)

// toUserDTO 将用户模型转换为对外返回的 DTO。
func toUserDTO(ctx context.Context, user *model.User, storageManager storage.Manager, gravatarBaseURL string) (dto.UserDTO, error) {
	avatar, err := resolveUserAvatar(ctx, user.Avatar, user.Email, storageManager, gravatarBaseURL)
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
func toUserDTOs(ctx context.Context, users []model.User, storageManager storage.Manager, gravatarBaseURL string) ([]dto.UserDTO, error) {
	if len(users) == 0 {
		return nil, nil
	}

	items := make([]dto.UserDTO, 0, len(users))
	for i := range users {
		item, err := toUserDTO(ctx, &users[i], storageManager, gravatarBaseURL)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// resolveUserAvatar 解析用户头像地址，优先用户头像，其次 Gravatar 回退。
func resolveUserAvatar(ctx context.Context, avatar *string, email string, storageManager storage.Manager, gravatarBaseURL string) (*string, error) {
	// 未设置头像时回退到 Gravatar。
	if avatar == nil {
		return buildGravatarURL(email, gravatarBaseURL), nil
	}

	trimmed := strings.TrimSpace(*avatar)
	if trimmed == "" {
		return nil, nil
	}
	if !isManagedAvatarKey(trimmed) || storageManager == nil {
		return &trimmed, nil
	}

	// 托管头像需要转换为可访问 URL。
	url, err := storageManager.URL(ctx, trimmed)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// buildGravatarURL 基于邮箱计算 Gravatar 头像地址。
func buildGravatarURL(email string, gravatarBaseURL string) *string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil
	}

	hash := md5.Sum([]byte(normalized))
	url := normalizeGravatarBaseURL(gravatarBaseURL) + hex.EncodeToString(hash[:])
	return &url
}

// normalizeGravatarBaseURL 规范化 Gravatar 基础地址并保证末尾分隔符。
func normalizeGravatarBaseURL(gravatarBaseURL string) string {
	base := strings.TrimSpace(gravatarBaseURL)
	if base == "" {
		return avatarpkg.DefaultGravatarBaseURL
	}
	return strings.TrimRight(base, "/") + "/"
}

// isManagedAvatarKey 判断头像值是否为系统托管存储对象键。
func isManagedAvatarKey(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), storage.CategoryAvatar+"/")
}
