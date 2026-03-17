package captcha

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewVerifier_ConfigValidation(t *testing.T) {
	t.Parallel()

	_, err := NewVerifier(&AllCaptchaConfig{Enabled: true}, nil)
	if !errors.Is(err, ErrEmptyProvider) {
		t.Fatalf("期望错误 ErrEmptyProvider，实际为: %v", err)
	}

	_, err = NewVerifier(&AllCaptchaConfig{Enabled: true, Provider: "unknown"}, nil)
	if !errors.Is(err, ErrUnsupportedProvider) {
		t.Fatalf("期望错误 ErrUnsupportedProvider，实际为: %v", err)
	}

	_, err = NewVerifier(&AllCaptchaConfig{Enabled: true, Provider: ProviderHCaptcha}, nil)
	if !errors.Is(err, ErrEmptySecretKey) {
		t.Fatalf("期望错误 ErrEmptySecretKey，实际为: %v", err)
	}

	_, err = NewVerifier(&AllCaptchaConfig{Enabled: true, Provider: ProviderHCaptcha, HCaptcha: ProviderConfig{SecretKey: "secret"}}, nil)
	if !errors.Is(err, ErrEmptySiteKey) {
		t.Fatalf("期望错误 ErrEmptySiteKey，实际为: %v", err)
	}
}

func TestVerify_HCaptchaSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("解析请求失败: %v", err)
		}
		if r.Form.Get("secret") != "secret-h" || r.Form.Get("response") != "token-h" {
			t.Fatalf("请求参数不符合预期: %v", r.Form)
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	result, err := verifyHCaptcha(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: "secret-h",
			SiteKey:   "site-h",
			VerifyURL: server.URL,
		},
		HTTPClient: server.Client(),
	}, Payload{Token: "token-h"})
	if err != nil {
		t.Fatalf("校验 hCaptcha 失败: %v", err)
	}
	if !result.Success {
		t.Fatalf("期望校验成功")
	}
}

func TestVerify_RecaptchaFail(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error-codes":["invalid-input-response"]}`))
	}))
	defer server.Close()

	_, err := verifyRecaptcha(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: "secret-r",
			SiteKey:   "site-r",
			VerifyURL: server.URL,
		},
		HTTPClient: server.Client(),
	}, Payload{Token: "bad"})
	if !errors.Is(err, ErrVerifyFailed) {
		t.Fatalf("期望错误 ErrVerifyFailed，实际为: %v", err)
	}
}

func TestVerify_TurnstileSuccessWithScore(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"score":0.92}`))
	}))
	defer server.Close()

	result, err := verifyTurnstile(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: "secret-t",
			SiteKey:   "site-t",
			VerifyURL: server.URL,
		},
		HTTPClient: server.Client(),
	}, Payload{Token: "ok"})
	if err != nil {
		t.Fatalf("校验 Turnstile 失败: %v", err)
	}
	if !result.Success || result.Score <= 0 {
		t.Fatalf("Turnstile 返回结果不符合预期: %+v", result)
	}
}

func TestVerify_GeetestV4Success(t *testing.T) {
	t.Parallel()

	secret := "geetest-secret"
	lotNumber := "lot-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("解析请求失败: %v", err)
		}
		expectSign := geetestSignForTest(lotNumber, secret)
		if r.Form.Get("sign_token") != expectSign {
			t.Fatalf("Geetest sign_token 不符合预期，want=%s got=%s", expectSign, r.Form.Get("sign_token"))
		}
		_, _ = w.Write([]byte(`{"result":"success","reason":"ok"}`))
	}))
	defer server.Close()

	result, err := verifyGeetestV4(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: secret,
			SiteKey:   "captcha-id",
			VerifyURL: server.URL,
		},
		HTTPClient: server.Client(),
	}, Payload{
		LotNumber:     lotNumber,
		CaptchaOutput: "captcha-output",
		PassToken:     "pass-token",
		GenTime:       "1700000000",
	})
	if err != nil {
		t.Fatalf("校验 Geetest v4 失败: %v", err)
	}
	if !result.Success {
		t.Fatalf("期望 Geetest v4 校验成功")
	}
}

func TestVerify_HTTPRequestFailed(t *testing.T) {
	t.Parallel()

	_, err := verifyHCaptcha(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: "secret-h",
			SiteKey:   "site-h",
			VerifyURL: "http://127.0.0.1:1",
		},
		Timeout: 50 * time.Millisecond,
	}, Payload{Token: "token"})
	if err == nil || !strings.Contains(err.Error(), ErrRequestFailed.Error()) {
		t.Fatalf("期望请求失败错误，实际为: %v", err)
	}
}

func TestNormalizeProvider(t *testing.T) {
	t.Parallel()

	if normalizeProvider(" ReCaptcha ") != ProviderRecaptcha {
		t.Fatalf("provider 归一化失败")
	}
}

func geetestSignForTest(lotNumber, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(lotNumber))
	return hex.EncodeToString(h.Sum(nil))
}

func Example_verifyHCaptcha() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer mockServer.Close()

	result, err := verifyHCaptcha(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: "secret",
			SiteKey:   "site",
			VerifyURL: mockServer.URL,
		},
		HTTPClient: mockServer.Client(),
	}, Payload{Token: "token"})

	fmt.Println(err == nil, result.Success)
	// Output: true true
}

func TestVerify_RequestBodyForm(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("解析请求失败: %v", err)
		}
		if r.Form.Get("response") != "token-form" {
			t.Fatalf("response 参数不符合预期")
		}
		if _, err := url.ParseQuery(r.Form.Encode()); err != nil {
			t.Fatalf("表单编码不合法: %v", err)
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	_, err := verifyHCaptcha(context.Background(), VerifyConfig{
		ProviderConfig: ProviderConfig{
			SecretKey: "secret",
			SiteKey:   "site",
			VerifyURL: server.URL,
		},
		HTTPClient: server.Client(),
	}, Payload{Token: "token-form"})
	if err != nil {
		t.Fatalf("验证码校验失败: %v", err)
	}
}
