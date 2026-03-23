package dto

import "time"

type PostCreateInput struct {
	UserID  uint   `json:"user_id"`
	Content string `json:"content"`
	Hide    bool   `json:"hide"`
}

type PostIDURIInput struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type PostUpdateInput struct {
	ID      uint    `json:"-"`
	Content *string `json:"content"`
	Hide    *bool   `json:"hide"`
}

type PostListInput struct {
	UserID *uint  `json:"user_id" form:"user_id" binding:"omitempty,gt=0"`
	TagID  *uint  `json:"tag_id" form:"tag_id" binding:"omitempty,gt=0"`
	Hide   *bool  `json:"hide" form:"hide"`
	Limit  int    `json:"limit" form:"limit"`
	Offset int    `json:"offset" form:"offset"`
	Order  string `json:"order" form:"order"`
}

type PostDTO struct {
	ID        uint          `json:"id"`
	Content   string        `json:"content"`
	Hide      bool          `json:"hide"`
	Author    PostAuthorDTO `json:"author"`
	Tags      []TagDTO      `json:"tags"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type PostAuthorDTO struct {
	ID       uint    `json:"id"`
	Nickname string  `json:"nickname"`
	Avatar   *string `json:"avatar"`
}

type PostListResult struct {
	Items []PostDTO `json:"items"`
	Total int64     `json:"total"`
}
