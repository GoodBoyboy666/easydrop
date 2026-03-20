package dto

import "time"

type UserCreateInput struct {
	Username      string  `json:"username"`
	Nickname      string  `json:"nickname"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	Admin         bool    `json:"admin"`
	Status        *int    `json:"status"`
	Avatar        *string `json:"avatar"`
	EmailVerified bool    `json:"email_verified"`
	StorageQuota  *int64  `json:"storage_quota"`
}

type UserUpdateInput struct {
	ID            uint    `json:"id"`
	Username      *string `json:"username"`
	Nickname      *string `json:"nickname"`
	Email         *string `json:"email"`
	Password      *string `json:"password"`
	Admin         *bool   `json:"admin"`
	Status        *int    `json:"status"`
	Avatar        *string `json:"avatar"`
	EmailVerified *bool   `json:"email_verified"`
	StorageQuota  *int64  `json:"storage_quota"`
}

type UserProfileUpdateInput struct {
	UserID   uint    `json:"user_id"`
	Nickname *string `json:"nickname"`
}

type UserChangePasswordInput struct {
	UserID      uint   `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserChangeEmailRequestInput struct {
	UserID          uint   `json:"user_id"`
	CurrentPassword string `json:"current_password"`
	NewEmail        string `json:"new_email"`
}

type UserChangeEmailConfirmInput struct {
	UserID            uint   `json:"user_id"`
	VerificationToken string `json:"verification_token"`
}

type UserAvatarUploadInput struct {
	UserID           uint   `json:"user_id"`
	OriginalFilename string `json:"original_filename"`
	ContentType      string `json:"content_type"`
	Content          []byte `json:"content"`
}

type UserListInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   *int   `json:"status"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Order    string `json:"order"`
}

type UserDTO struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname"`
	Email         string    `json:"email"`
	Admin         bool      `json:"admin"`
	Status        int       `json:"status"`
	Avatar        *string   `json:"avatar"`
	EmailVerified bool      `json:"email_verified"`
	StorageQuota  *int64    `json:"storage_quota"`
	StorageUsed   int64     `json:"storage_used"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserListResult struct {
	Items []UserDTO `json:"items"`
	Total int64     `json:"total"`
}
