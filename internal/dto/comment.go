package dto

import "time"

type CommentCreateInput struct {
	PostID   uint   `json:"post_id"`
	UserID   uint   `json:"user_id"`
	Content  string `json:"content"`
	ParentID *uint  `json:"parent_id"`
}

type CommentUpdateInput struct {
	ID      uint    `json:"id"`
	Content *string `json:"content"`
}

type CommentListInput struct {
	PostID uint   `json:"post_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Order  string `json:"order"`
}

type CommentUserListInput struct {
	UserID uint   `json:"user_id"`
	PostID *uint  `json:"post_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Order  string `json:"order"`
}

type CommentAdminListInput struct {
	PostID *uint  `json:"post_id"`
	UserID *uint  `json:"user_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Order  string `json:"order"`
}

type CommentDTO struct {
	ID            uint      `json:"id"`
	PostID        uint      `json:"post_id"`
	UserID        uint      `json:"user_id"`
	Content       string    `json:"content"`
	ParentID      *uint     `json:"parent_id"`
	RootID        *uint     `json:"root_id"`
	ReplyToUserID *uint     `json:"reply_to_user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CommentListResult struct {
	Items []CommentDTO `json:"items"`
	Total int64        `json:"total"`
}
