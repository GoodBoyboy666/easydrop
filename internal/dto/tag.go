package dto

import "time"

type TagDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TagListInput struct {
	Keyword string `json:"keyword" form:"keyword"`
	Limit   int    `json:"limit" form:"limit"`
	Offset  int    `json:"offset" form:"offset"`
	Order   string `json:"order" form:"order"`
}

type TagListResult struct {
	Items []TagDTO `json:"items"`
	Total int64    `json:"total"`
}
