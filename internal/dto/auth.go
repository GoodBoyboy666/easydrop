package dto

type CaptchaInput struct {
	Provider      string `json:"provider"`
	Token         string `json:"token"`
	RemoteIP      string `json:"-"`
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

type AuthResult struct {
	AccessToken string `json:"access_token"`
}
