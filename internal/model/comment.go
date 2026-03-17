package model

import "time"

type Comment struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID        uint      `gorm:"index;not null" json:"post_id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	ParentID      *uint     `gorm:"index" json:"parent_id"`
	RootID        *uint     `gorm:"index" json:"root_id"`
	ReplyToUserID *uint     `gorm:"index" json:"reply_to_user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
