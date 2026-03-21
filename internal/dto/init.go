package dto

// InitStatusResult 表示系统初始化状态。
type InitStatusResult struct {
	Initialized bool `json:"initialized"`
}

// InitInput 表示系统初始化请求。
type InitInput struct {
	Username         string `json:"username"`
	Nickname         string `json:"nickname"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	SiteName         string `json:"site_name"`
	SiteURL          string `json:"site_url"`
	SiteAnnouncement string `json:"site_announcement"`
	AllowRegister    *bool  `json:"allow_register"`
}
