package dto

// CaptchaConfigResult 表示前端可读取的验证码配置。
type CaptchaConfigResult struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	SiteKey  string `json:"site_key"`
}
