package service

import (
	"context"
	"errors"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

type mockJWTManager struct {
	token string
	err   error
}

func (m *mockJWTManager) IssueAccessToken(userID uint, username string, admin bool) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.token != "" {
		return m.token, nil
	}
	return "jwt-token", nil
}

func (m *mockJWTManager) ParseToken(string) (*jwt.Claims, error) {
	return nil, nil
}

type mockCaptchaVerifier struct {
	enabled bool
	err     error
	payload captcha.Payload
}

type mockAuthSettingService struct {
	valueByKey map[string]string
}

func (m *mockAuthSettingService) GetValue(_ context.Context, key string) (string, bool, error) {
	if m == nil || m.valueByKey == nil {
		return "", false, nil
	}
	value, ok := m.valueByKey[key]
	return value, ok, nil
}

func (m *mockAuthSettingService) ListItems(context.Context, dto.SettingListInput) (*dto.SettingListResult, error) {
	return nil, nil
}

func (m *mockAuthSettingService) UpdateItem(context.Context, dto.SettingUpdateInput) error {
	return nil
}

func (m *mockAuthSettingService) GetPublicItems(context.Context) (*dto.SettingPublicResult, error) {
	return nil, nil
}

func (m *mockCaptchaVerifier) Enabled() bool {
	return m.enabled
}

func (m *mockCaptchaVerifier) Verify(_ context.Context, payload captcha.Payload) (captcha.Result, error) {
	m.payload = payload
	if m.err != nil {
		return captcha.Result{}, m.err
	}
	return captcha.Result{Success: true}, nil
}

func TestAuthServiceRegisterSendsVerifyEmail(t *testing.T) {
	repo := &mockUserRepo{users: map[uint]*model.User{}}
	tokens := &mockTokenManager{nextToken: "verify-token"}
	emailSender := &mockEmailSender{}
	emails := &mockEmailService{sender: emailSender}
	settings := &mockAuthSettingService{valueByKey: map[string]string{"site.allow_register": "true"}}
	svc := NewAuthService(repo, settings, &mockJWTManager{token: "jwt-1"}, nil, tokens, emails)

	result, err := svc.Register(context.Background(), dto.RegisterInput{
		Username: "neo",
		Nickname: "Neo",
		Email:    "neo@example.com",
		Password: "Pass1234",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if result.Message != "注册成功，请先完成邮箱验证后登录" {
		t.Fatalf("unexpected register message: %q", result.Message)
	}
	if len(emailSender.to) != 1 || emailSender.to[0] != "neo@example.com" {
		t.Fatalf("expected verify email sent, got %#v", emailSender.to)
	}
	record, ok := tokens.recordByToken["verify-token"]
	if !ok {
		t.Fatal("expected verify token to be issued")
	}
	if record.Kind != token.KindVerifyEmail {
		t.Fatalf("expected token kind %s, got %s", token.KindVerifyEmail, record.Kind)
	}
	if record.UserID == 0 {
		t.Fatal("expected persisted user id to be set on verify token")
	}
}

func TestAuthServiceConfirmVerifyEmail(t *testing.T) {
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			3: {
				ID:            3,
				Username:      "alice",
				Email:         "alice@example.com",
				EmailVerified: false,
				Status:        1,
			},
		},
	}
	tokens := &mockTokenManager{
		recordByToken: map[string]*token.Record{
			"verify-3": {
				UserID:  3,
				Kind:    token.KindVerifyEmail,
				Token:   "verify-3",
				Payload: "alice@example.com",
			},
		},
	}
	svc := NewAuthService(repo, nil, &mockJWTManager{}, nil, tokens, nil)

	err := svc.ConfirmVerifyEmail(context.Background(), dto.EmailVerifyConfirmInput{
		Token: "verify-3",
	})
	if err != nil {
		t.Fatalf("ConfirmVerifyEmail returned error: %v", err)
	}
	if !repo.users[3].EmailVerified {
		t.Fatal("expected email verified to be updated")
	}
}

func TestAuthServiceRequestPasswordResetDoesNotLeakMissingUser(t *testing.T) {
	captchaVerifier := &mockCaptchaVerifier{enabled: true}
	svc := NewAuthService(&mockUserRepo{users: map[uint]*model.User{}}, nil, &mockJWTManager{}, captchaVerifier, &mockTokenManager{}, &mockEmailService{})

	err := svc.RequestPasswordReset(context.Background(), dto.PasswordResetRequestInput{
		Email: "missing@example.com",
		Captcha: &dto.CaptchaInput{
			Token: "captcha-ok",
		},
	})
	if err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if captchaVerifier.payload.Token != "captcha-ok" {
		t.Fatalf("expected captcha to be verified, got %#v", captchaVerifier.payload)
	}
}

func TestAuthServiceRequestPasswordResetSendsEmail(t *testing.T) {
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			4: {
				ID:       4,
				Username: "morpheus",
				Email:    "morpheus@example.com",
				Status:   1,
			},
		},
	}
	tokens := &mockTokenManager{nextToken: "reset-token"}
	emailSender := &mockEmailSender{}
	emails := &mockEmailService{sender: emailSender}
	svc := NewAuthService(repo, nil, &mockJWTManager{}, nil, tokens, emails)

	err := svc.RequestPasswordReset(context.Background(), dto.PasswordResetRequestInput{
		Email: "morpheus@example.com",
	})
	if err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if len(emailSender.to) != 1 || emailSender.to[0] != "morpheus@example.com" {
		t.Fatalf("expected reset email sent, got %#v", emailSender.to)
	}
	if record, ok := tokens.recordByToken["reset-token"]; !ok || record.Kind != token.KindResetPassword {
		t.Fatalf("expected reset token record, got %#v", record)
	}
}

func TestAuthServiceConfirmPasswordReset(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			5: {
				ID:       5,
				Username: "switch",
				Email:    "switch@example.com",
				Password: string(hash),
				Status:   1,
			},
		},
	}
	tokens := &mockTokenManager{
		recordByToken: map[string]*token.Record{
			"reset-5": {
				UserID:  5,
				Kind:    token.KindResetPassword,
				Token:   "reset-5",
				Payload: "switch@example.com",
			},
		},
	}
	svc := NewAuthService(repo, nil, &mockJWTManager{}, nil, tokens, nil)

	err = svc.ConfirmPasswordReset(context.Background(), dto.PasswordResetConfirmInput{
		Token:       "reset-5",
		NewPassword: "NewPass123",
	})
	if err != nil {
		t.Fatalf("ConfirmPasswordReset returned error: %v", err)
	}
	if compareErr := bcrypt.CompareHashAndPassword([]byte(repo.users[5].Password), []byte("NewPass123")); compareErr != nil {
		t.Fatalf("expected password updated, got %v", compareErr)
	}
	if !repo.users[5].EmailVerified {
		t.Fatal("expected password reset to mark email as verified")
	}
}

func TestAuthServiceConfirmPasswordResetRejectsInvalidToken(t *testing.T) {
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			8: {
				ID:     8,
				Email:  "broken@example.com",
				Status: 1,
			},
		},
	}
	tokens := &mockTokenManager{consumeErr: token.ErrTokenExpired}
	svc := NewAuthService(repo, nil, &mockJWTManager{}, nil, tokens, nil)

	err := svc.ConfirmPasswordReset(context.Background(), dto.PasswordResetConfirmInput{
		Token:       "expired",
		NewPassword: "NewPass123",
	})
	if !errors.Is(err, ErrInvalidPasswordReset) {
		t.Fatalf("expected ErrInvalidPasswordReset, got %v", err)
	}
}

func TestAuthServiceLoginRequiresEmailVerificationWhenEnabled(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Pass1234"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			10: {
				ID:            10,
				Username:      "alice",
				Email:         "alice@example.com",
				Password:      string(hash),
				EmailVerified: false,
				Status:        1,
			},
		},
	}
	settings := &mockAuthSettingService{valueByKey: map[string]string{"auth.require_email_verification": "true"}}
	svc := NewAuthService(repo, settings, &mockJWTManager{token: "jwt-login"}, nil, nil, nil)

	result, err := svc.Login(context.Background(), dto.LoginInput{
		Account:  "alice",
		Password: "Pass1234",
	})
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Fatalf("expected ErrEmailNotVerified, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestAuthServiceLoginAllowsUnverifiedWhenVerificationDisabled(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Pass1234"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			11: {
				ID:            11,
				Username:      "bob",
				Email:         "bob@example.com",
				Password:      string(hash),
				EmailVerified: false,
				Status:        1,
			},
		},
	}
	settings := &mockAuthSettingService{valueByKey: map[string]string{"auth.require_email_verification": "false"}}
	svc := NewAuthService(repo, settings, &mockJWTManager{token: "jwt-login"}, nil, nil, nil)

	result, err := svc.Login(context.Background(), dto.LoginInput{
		Account:  "bob",
		Password: "Pass1234",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if result == nil || result.AccessToken != "jwt-login" {
		t.Fatalf("unexpected login result: %#v", result)
	}
}

func TestAuthServiceLoginRejectsInvalidRequireEmailSetting(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Pass1234"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			12: {
				ID:            12,
				Username:      "charlie",
				Email:         "charlie@example.com",
				Password:      string(hash),
				EmailVerified: false,
				Status:        1,
			},
		},
	}
	settings := &mockAuthSettingService{valueByKey: map[string]string{"auth.require_email_verification": "not-a-bool"}}
	svc := NewAuthService(repo, settings, &mockJWTManager{token: "jwt-login"}, nil, nil, nil)

	result, err := svc.Login(context.Background(), dto.LoginInput{
		Account:  "charlie",
		Password: "Pass1234",
	})
	if !errors.Is(err, ErrInvalidSiteSetting) {
		t.Fatalf("expected ErrInvalidSiteSetting, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}
