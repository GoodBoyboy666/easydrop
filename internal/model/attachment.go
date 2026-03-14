package model

import "time"

const (
	AttachmentTypeImage = 1 // 图片
	AttachmentTypeVideo = 2 // 视频
	AttachmentTypeAudio = 3 // 音频
	AttachmentTypeFile  = 4 // 其他普通文件（如 PDF, ZIP）
)

type Attachment struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	StorageType string    `gorm:"size:20;not null" json:"storage_type"`
	FileKey     string    `gorm:"size:500;not null;uniqueIndex" json:"file_key"`
	BizType     int       `gorm:"type:tinyint;not null" json:"biz_type"`
	FileSize    int64     `gorm:"not null" json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
}
