package model

import "time"

type Tag struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex:idx_user_tag" json:"name"`
	Posts     []Post    `gorm:"many2many:post_tags;" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
