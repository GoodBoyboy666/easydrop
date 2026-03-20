package dto

import "time"

type AttachmentCreateInput struct {
	UserID      uint   `json:"user_id"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}

type AttachmentUpdateInput struct {
	ID      uint `json:"id"`
	BizType *int `json:"biz_type"`
}

type AttachmentListInput struct {
	ID          *uint  `json:"id"`
	UserID      *uint  `json:"user_id"`
	BizType     *int   `json:"biz_type"`
	CreatedFrom *int64 `json:"created_from"`
	CreatedTo   *int64 `json:"created_to"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	Order       string `json:"order"`
}

type AttachmentDTO struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	StorageType string    `json:"storage_type"`
	FileKey     string    `json:"file_key"`
	URL         string    `json:"url"`
	BizType     int       `json:"biz_type"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
}

type AttachmentListResult struct {
	Items []AttachmentDTO `json:"items"`
	Total int64           `json:"total"`
}
