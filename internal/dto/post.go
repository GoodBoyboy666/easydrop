package dto

import "time"

type PostCreateInput struct {
	UserID  uint   `json:"user_id"`
	Content string `json:"content"`
}

type PostUpdateInput struct {
	ID      uint    `json:"id"`
	Content *string `json:"content"`
}

type PostListInput struct {
	UserID *uint  `json:"user_id"`
	TagID  *uint  `json:"tag_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Order  string `json:"order"`
}

type PostDTO struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	Tags      []TagDTO  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostListResult struct {
	Items []PostDTO `json:"items"`
	Total int64     `json:"total"`
}
