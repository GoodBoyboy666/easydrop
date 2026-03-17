package dto

import "time"

type AttachmentCreateInput struct {
	UserID      uint   `json:"user_id"`
	StorageType string `json:"storage_type"`
	FileKey     string `json:"file_key"`
	BizType     int    `json:"biz_type"`
	FileSize    int64  `json:"file_size"`
}

type AttachmentUpdateInput struct {
	ID      uint `json:"id"`
	BizType *int `json:"biz_type"`
}

type AttachmentListInput struct {
	UserID  *uint  `json:"user_id"`
	BizType *int   `json:"biz_type"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	Order   string `json:"order"`
}

type AttachmentDTO struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	StorageType string    `json:"storage_type"`
	FileKey     string    `json:"file_key"`
	BizType     int       `json:"biz_type"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
}

type AttachmentListResult struct {
	Items []AttachmentDTO `json:"items"`
	Total int64           `json:"total"`
}
