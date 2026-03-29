package dto

import "time"

type TagDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TagListInput struct {
	Keyword string `json:"keyword" form:"keyword"`
	Page    int    `json:"page" form:"page"`
	Size    int    `json:"size" form:"size"`
	Order   string `json:"order" form:"order"`
}

type TagListResult struct {
	Items []TagDTO `json:"items"`
	Total int64    `json:"total"`
}
