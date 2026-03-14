package model

import "time"

type User struct {
	ID            uint `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Username      string  `json:"username" gorm:"unique;not null"`
	Nickname      string  `json:"nickname" gorm:"size:100;not null"`
	Password      string  `json:"-" gorm:"not null"`
	Admin         bool    `json:"admin" gorm:"not null;default:false;"`
	Status        int     `json:"status" gorm:"default:1"` // 1: 正常, 2: 封禁, 3: 软删除(停用)
	Avatar        *string `json:"avatar"`
	Email         string  `json:"email" gorm:"unique;index;size:255"`
	EmailVerified bool    `json:"email_verified" gorm:"default:false"`
	StorageQuota  *int64  `json:"storage_quota"`
	StorageUsed   int64   `json:"storage_used" gorm:"default:0"` // 已用存储空间 (Bytes)
}
