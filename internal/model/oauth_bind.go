package model

import "time"

type OAuthBind struct {
	ID             uint `gorm:"primaryKey;autoIncrement"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UserID         uint   `gorm:"not null;uniqueIndex:idx_user_provider"`
	Provider       string `gorm:"not null;size:50;uniqueIndex:idx_user_provider;uniqueIndex:idx_provider_uid"`
	ProviderUserID string `gorm:"not null;size:255;uniqueIndex:idx_provider_uid"`
	ProviderEmail  string `gorm:"size:255"`
}
