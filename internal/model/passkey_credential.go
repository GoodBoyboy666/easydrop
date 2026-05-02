package model

import "time"

type PasskeyCredential struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Name           string    `json:"name" gorm:"not null;size:15"`
	UserID         uint      `json:"user_id" gorm:"not null;index"`
	CredentialID   string    `json:"credential_id" gorm:"uniqueIndex;size:1023"`
	CredentialJSON []byte    `json:"credential_json" gorm:"not null"`
}
