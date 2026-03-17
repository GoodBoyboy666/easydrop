package dto

import "time"

type CaptchaInput struct {
	Provider      string `json:"provider"`
	Token         string `json:"token"`
	RemoteIP      string `json:"remote_ip"`
	LotNumber     string `json:"lot_number"`
	CaptchaOutput string `json:"captcha_output"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
}

type RegisterInput struct {
	Username string        `json:"username"`
	Nickname string        `json:"nickname"`
	Email    string        `json:"email"`
	Password string        `json:"password"`
	Captcha  *CaptchaInput `json:"captcha"`
}

type LoginInput struct {
	Account  string        `json:"account"`
	Password string        `json:"password"`
	Captcha  *CaptchaInput `json:"captcha"`
}

type UserDTO struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname"`
	Email         string    `json:"email"`
	Admin         bool      `json:"admin"`
	Status        int       `json:"status"`
	Avatar        *string   `json:"avatar"`
	EmailVerified bool      `json:"email_verified"`
	StorageQuota  *int64    `json:"storage_quota"`
	StorageUsed   int64     `json:"storage_used"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AuthResult struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
}
