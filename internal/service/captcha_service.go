package service

import (
	"context"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/pkg/captcha"
)

// CaptchaConfigService 提供验证码公开配置读取能力。
type CaptchaConfigService interface {
	GetConfig(ctx context.Context) *dto.CaptchaConfigResult
}

type captchaConfigService struct {
	cfg *captcha.AllCaptchaConfig
}

// NewCaptchaConfigService 创建验证码公开配置服务。
func NewCaptchaConfigService(cfg *captcha.AllCaptchaConfig) CaptchaConfigService {
	return &captchaConfigService{cfg: cfg}
}

func (s *captchaConfigService) GetConfig(context.Context) *dto.CaptchaConfigResult {
	if s == nil || s.cfg == nil {
		return &dto.CaptchaConfigResult{}
	}

	provider := strings.TrimSpace(string(s.cfg.Provider))
	result := &dto.CaptchaConfigResult{
		Enabled:  s.cfg.Enabled,
		Provider: provider,
		SiteKey:  "",
	}

	if !s.cfg.Enabled {
		return result
	}

	var siteKey string
	switch captcha.Provider(strings.ToLower(provider)) {
	case captcha.ProviderTurnstile:
		siteKey = s.cfg.Turnstile.SiteKey
	case captcha.ProviderRecaptcha:
		siteKey = s.cfg.Recaptcha.SiteKey
	case captcha.ProviderHCaptcha:
		siteKey = s.cfg.HCaptcha.SiteKey
	case captcha.ProviderGeetestV4:
		siteKey = s.cfg.GeetestV4.SiteKey
	default:
		siteKey = ""
	}

	siteKey = strings.TrimSpace(siteKey)
	result.SiteKey = siteKey
	return result
}

var _ CaptchaConfigService = (*captchaConfigService)(nil)
