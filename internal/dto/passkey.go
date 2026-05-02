package dto

import "time"

type PasskeyLoginBeginResponse struct {
	Options   map[string]any `json:"options"`
	SessionID string         `json:"session_id"`
}

type PasskeyLoginFinishRequest struct {
	SessionID string `json:"session_id"`
}

type PasskeyRegisterBeginResponse struct {
	Options   map[string]any `json:"options"`
	SessionID string         `json:"session_id"`
}

type PasskeyRegisterFinishRequest struct {
	SessionID string `json:"session_id"`
}

type PasskeyRenameRequest struct {
	Name string `json:"name" binding:"required,min=1,max=15"`
}

type PasskeyItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
