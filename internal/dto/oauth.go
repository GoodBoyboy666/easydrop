package dto

type OAuthProviderItem struct {
	Provider string `json:"provider"`
	AuthURL  string `json:"auth_url"`
}

type OAuthBindDTO struct {
	ID            uint   `json:"id"`
	Provider      string `json:"provider"`
	ProviderEmail string `json:"provider_email"`
}

type OAuthBindManualInput struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// OAuthCallbackInput 是 OAuth 回调 POST 请求体。
// 前端从 OAuth 提供方回跳 URL 中提取 code 和 state 后提交。
type OAuthCallbackInput struct {
	Code  string `json:"code"`
	State string `json:"state"`
}
