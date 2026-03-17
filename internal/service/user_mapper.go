package service

import (
	"easydrop/internal/dto"
	"easydrop/internal/model"
)

// toUserDTO 将用户模型转换为对外返回的 DTO。
func toUserDTO(user *model.User) dto.UserDTO {
	return dto.UserDTO{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Admin:         user.Admin,
		Status:        user.Status,
		Avatar:        user.Avatar,
		EmailVerified: user.EmailVerified,
		StorageQuota:  user.StorageQuota,
		StorageUsed:   user.StorageUsed,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

// toUserDTOs 将用户模型切片转换为 DTO 列表。
func toUserDTOs(users []model.User) []dto.UserDTO {
	if len(users) == 0 {
		return nil
	}

	items := make([]dto.UserDTO, 0, len(users))
	for i := range users {
		items = append(items, toUserDTO(&users[i]))
	}
	return items
}
