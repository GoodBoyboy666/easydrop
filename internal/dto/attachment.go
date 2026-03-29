package dto

import (
	"io"
	"time"
)

type AttachmentCreateInput struct {
	UserID           uint   `json:"user_id"`
	OriginalFilename string `json:"original_filename"`
	ContentType      string `json:"content_type"`
	FileSize         int64  `json:"file_size"`
	Content          io.Reader
	ContentSample    []byte `json:"-"`
}

type AttachmentIDURIInput struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type AttachmentUpdateInput struct {
	ID      uint `json:"id"`
	BizType *int `json:"biz_type"`
}

type AttachmentListInput struct {
	ID          *uint  `json:"id" form:"id" binding:"omitempty,gt=0"`
	UserID      *uint  `json:"user_id" form:"user_id" binding:"omitempty,gt=0"`
	BizType     *int   `json:"biz_type" form:"biz_type"`
	CreatedFrom *int64 `json:"created_from" form:"created_from" binding:"omitempty,gte=0"`
	CreatedTo   *int64 `json:"created_to" form:"created_to" binding:"omitempty,gte=0"`
	Page        int    `json:"page" form:"page"`
	Size        int    `json:"size" form:"size"`
	Order       string `json:"order" form:"order"`
}

type AttachmentBatchDeleteInput struct {
	IDs []uint `json:"ids"`
}

type AttachmentBatchDeleteFailedItem struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type AttachmentBatchDeleteResult struct {
	SuccessIDs []uint                            `json:"success_ids"`
	Failed     []AttachmentBatchDeleteFailedItem `json:"failed"`
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
