package captcha

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

type Provider string

const (
	ProviderGeetestV4 Provider = "geetest_v4"
	ProviderHCaptcha  Provider = "hcaptcha"
	ProviderRecaptcha Provider = "recaptcha"
	ProviderTurnstile Provider = "turnstile"

	defaultVerifyTimeout = 5 * time.Second
)

var (
	ErrEmptyProvider      = errors.New("验证码类型不能为空")
	ErrUnsupportedProvider = errors.New("不支持的验证码类型")
	ErrEmptySecretKey     = errors.New("验证码密钥不能为空")
	ErrEmptyToken         = errors.New("验证码 token 不能为空")
	ErrEmptyVerifyURL     = errors.New("验证码校验地址不能为空")
	ErrRequestFailed      = errors.New("验证码请求失败")
	ErrDecodeResponse     = errors.New("验证码响应解析失败")
	ErrVerifyFailed       = errors.New("验证码校验失败")

	ErrEmptyGeetestCaptchaID = errors.New("Geetest captcha_id 不能为空")
	ErrEmptyLotNumber        = errors.New("Geetest lot_number 不能为空")
	ErrEmptyCaptchaOutput    = errors.New("Geetest captcha_output 不能为空")
	ErrEmptyPassToken        = errors.New("Geetest pass_token 不能为空")
	ErrEmptyGenTime          = errors.New("Geetest gen_time 不能为空")
)

type Config struct {
	Provider Provider

	SecretKey string
	SiteKey   string
	VerifyURL string
	RemoteIP  string

	HTTPClient *http.Client
	Timeout    time.Duration
}

type Payload struct {
	Token string

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

func Verify(ctx context.Context, cfg Config, payload Payload) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	switch normalizeProvider(cfg.Provider) {
	case ProviderHCaptcha:
		return verifyHCaptcha(ctx, cfg, payload)
	case ProviderRecaptcha:
		return verifyRecaptcha(ctx, cfg, payload)
	case ProviderTurnstile:
		return verifyTurnstile(ctx, cfg, payload)
	case ProviderGeetestV4:
		return verifyGeetestV4(ctx, cfg, payload)
	case "":
		return Result{}, ErrEmptyProvider
	default:
		return Result{}, ErrUnsupportedProvider
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


