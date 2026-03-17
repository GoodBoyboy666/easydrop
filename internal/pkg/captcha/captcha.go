package captcha

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/wire"
)

type Provider string

var CaptchaSet = wire.NewSet(
	NewHttpClient, NewVerifier)

const (
	ProviderGeetestV4 Provider = "geetest_v4"
	ProviderHCaptcha  Provider = "hcaptcha"
	ProviderRecaptcha Provider = "recaptcha"
	ProviderTurnstile Provider = "turnstile"

	defaultVerifyTimeout = 5 * time.Second
)

var (
	ErrEmptyProvider       = errors.New("验证码类型不能为空")
	ErrUnsupportedProvider = errors.New("不支持的验证码类型")
	ErrEmptySecretKey      = errors.New("验证码密钥不能为空")
	ErrEmptySiteKey        = errors.New("验证码站点密钥不能为空")
	ErrEmptyToken          = errors.New("验证码 token 不能为空")
	ErrEmptyVerifyURL      = errors.New("验证码校验地址不能为空")
	ErrRequestFailed       = errors.New("验证码请求失败")
	ErrDecodeResponse      = errors.New("验证码响应解析失败")
	ErrVerifyFailed        = errors.New("验证码校验失败")

	ErrEmptyGeetestCaptchaID = errors.New("Geetest captcha_id 不能为空")
	ErrEmptyLotNumber        = errors.New("Geetest lot_number 不能为空")
	ErrEmptyCaptchaOutput    = errors.New("Geetest captcha_output 不能为空")
	ErrEmptyPassToken        = errors.New("Geetest pass_token 不能为空")
	ErrEmptyGenTime          = errors.New("Geetest gen_time 不能为空")
)

type VerifyConfig struct {
	ProviderConfig ProviderConfig

	HTTPClient *http.Client
	Timeout    time.Duration
}

type Payload struct {
	Token    string
	RemoteIP string

	LotNumber     string
	CaptchaOutput string
	PassToken     string
	GenTime       string
}

type Result struct {
	Provider Provider
	Success  bool
	Score    float64
	Message  string
	Raw      map[string]any
}

type Verifier interface {
	Enabled() bool
	Verify(ctx context.Context, payload Payload) (Result, error)
}

type ProviderConfig struct {
	SecretKey string `mapstructure:"secret_key" yaml:"secret_key"`
	SiteKey   string `mapstructure:"site_key" yaml:"site_key"`
	VerifyURL string `mapstructure:"verify_url" yaml:"verify_url"`
}

type AllCaptchaConfig struct {
	Enabled   bool           `mapstructure:"enabled" yaml:"enabled"`
	Provider  Provider       `mapstructure:"provider" yaml:"provider"`
	Timeout   time.Duration  `mapstructure:"timeout" yaml:"timeout"`
	Turnstile ProviderConfig `mapstructure:"turnstile" yaml:"turnstile"`
	Recaptcha ProviderConfig `mapstructure:"recaptcha" yaml:"recaptcha"`
	HCaptcha  ProviderConfig `mapstructure:"hcaptcha" yaml:"hcaptcha"`
	GeetestV4 ProviderConfig `mapstructure:"geetest_v4" yaml:"geetest_v4"`
}

type disabledVerifier struct{}
type turnstileVerifier struct{ cfg VerifyConfig }
type recaptchaVerifier struct{ cfg VerifyConfig }
type hcaptchaVerifier struct{ cfg VerifyConfig }
type geetestVerifier struct{ cfg VerifyConfig }

func (v *disabledVerifier) Enabled() bool {
	return false
}
func (v *turnstileVerifier) Enabled() bool { return true }
func (v *recaptchaVerifier) Enabled() bool { return true }
func (v *hcaptchaVerifier) Enabled() bool  { return true }
func (v *geetestVerifier) Enabled() bool   { return true }

func (v *disabledVerifier) Verify(context.Context, Payload) (Result, error) {
	return Result{Success: true}, nil
}

func (v *turnstileVerifier) Verify(ctx context.Context, payload Payload) (Result, error) {
	payload.RemoteIP = strings.TrimSpace(payload.RemoteIP)
	return verifyTurnstile(ctx, v.cfg, payload)
}

func (v *recaptchaVerifier) Verify(ctx context.Context, payload Payload) (Result, error) {
	payload.RemoteIP = strings.TrimSpace(payload.RemoteIP)
	return verifyRecaptcha(ctx, v.cfg, payload)
}

func (v *hcaptchaVerifier) Verify(ctx context.Context, payload Payload) (Result, error) {
	payload.RemoteIP = strings.TrimSpace(payload.RemoteIP)
	return verifyHCaptcha(ctx, v.cfg, payload)
}

func (v *geetestVerifier) Verify(ctx context.Context, payload Payload) (Result, error) {
	payload.RemoteIP = strings.TrimSpace(payload.RemoteIP)
	return verifyGeetestV4(ctx, v.cfg, payload)
}

func NewVerifier(cfg *AllCaptchaConfig, client *http.Client) (Verifier, error) {
	if cfg == nil || !cfg.Enabled {
		return &disabledVerifier{}, nil
	}

	provider := normalizeProvider(cfg.Provider)
	if provider == "" {
		return nil, ErrEmptyProvider
	}
	switch provider {
	case ProviderTurnstile, ProviderRecaptcha, ProviderHCaptcha, ProviderGeetestV4:
		// ok
	default:
		return nil, ErrUnsupportedProvider
	}

	options := cfg.providerConfig(provider)
	secretKey := strings.TrimSpace(options.SecretKey)
	siteKey := strings.TrimSpace(options.SiteKey)
	if secretKey == "" {
		return nil, ErrEmptySecretKey
	}
	if siteKey == "" {
		return nil, ErrEmptySiteKey
	}

	providerConfig := ProviderConfig{
		SecretKey: secretKey,
		SiteKey:   siteKey,
		VerifyURL: strings.TrimSpace(options.VerifyURL),
	}
	vCfg := VerifyConfig{
		ProviderConfig: providerConfig,

		HTTPClient: client,
		Timeout:    cfg.Timeout,
	}

	switch provider {
	case ProviderTurnstile:
		return &turnstileVerifier{cfg: vCfg}, nil
	case ProviderRecaptcha:
		return &recaptchaVerifier{cfg: vCfg}, nil
	case ProviderHCaptcha:
		return &hcaptchaVerifier{cfg: vCfg}, nil
	case ProviderGeetestV4:
		return &geetestVerifier{cfg: vCfg}, nil
	default:
		return nil, ErrUnsupportedProvider
	}
}

func (cfg *AllCaptchaConfig) providerConfig(provider Provider) ProviderConfig {
	if cfg == nil {
		return ProviderConfig{}
	}

	switch normalizeProvider(provider) {
	case ProviderTurnstile:
		return cfg.Turnstile
	case ProviderRecaptcha:
		return cfg.Recaptcha
	case ProviderHCaptcha:
		return cfg.HCaptcha
	case ProviderGeetestV4:
		return cfg.GeetestV4
	default:
		return ProviderConfig{}
	}
}

func normalizeProvider(p Provider) Provider {
	return Provider(strings.ToLower(strings.TrimSpace(string(p))))
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func defaultOr(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}
