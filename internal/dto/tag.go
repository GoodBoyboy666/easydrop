package dto

import "time"

type TagDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TagListInput struct {
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	Order   string `json:"order"`
}

type TagListResult struct {
	Items []TagDTO `json:"items"`
	Total int64    `json:"total"`
}
