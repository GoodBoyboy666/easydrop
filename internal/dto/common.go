package dto

// ErrorResponse 表示通用错误响应。
type ErrorResponse struct {
	Message string `json:"message"`
}

// MessageResponse 表示通用消息响应。
type MessageResponse struct {
	Message string `json:"message"`
}
