package dto

import (
	"io"
	"time"
)

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

type UserIDURIInput struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type UserUpdateInput struct {
	ID                     uint    `json:"-"`
	Username               *string `json:"username"`
	Nickname               *string `json:"nickname"`
	Email                  *string `json:"email"`
	Password               *string `json:"password"`
	Admin                  *bool   `json:"admin"`
	Status                 *int    `json:"status"`
	Avatar                 *string `json:"avatar"`
	EmailVerified          *bool   `json:"email_verified"`
	StorageQuota           *int64  `json:"storage_quota"`
	UseDefaultStorageQuota *bool   `json:"use_default_storage_quota"`
}

type UserProfileUpdateInput struct {
	UserID   uint    `json:"-"`
	Nickname *string `json:"nickname"`
}

type UserChangePasswordInput struct {
	UserID      uint   `json:"-"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserChangeEmailInput struct {
	UserID          uint   `json:"-"`
	CurrentPassword string `json:"current_password"`
	NewEmail        string `json:"new_email"`
}

type UserChangeEmailConfirmInput struct {
	VerificationToken string `json:"token"`
}

type UserAvatarUploadInput struct {
	UserID           uint   `json:"-"`
	OriginalFilename string `json:"original_filename"`
	ContentType      string `json:"content_type"`
	FileSize         int64  `json:"file_size"`
	Content          io.Reader
	ContentSample    []byte `json:"-"`
}

type UserListInput struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
	Status   *int   `json:"status" form:"status"`
	Limit    int    `json:"limit" form:"limit"`
	Offset   int    `json:"offset" form:"offset"`
	Order    string `json:"order" form:"order"`
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
