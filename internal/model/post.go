package model

import (
	"time"
)

type Post struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Content   string    `gorm:"type:text" json:"content"`
	Hide      bool      `gorm:"not null;default:false" json:"hide"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Tags      []Tag     `gorm:"many2many:post_tags;" json:"tags"`
	Pin       *uint     `gorm:"" json:"pin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
}
