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
