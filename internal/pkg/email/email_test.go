package email

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

func mustNewConcreteClient(t *testing.T, cfg *Config) *client {
	t.Helper()

	sender, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	concrete, ok := sender.(*client)
	if !ok {
		t.Fatal("unexpected client implementation")
	}

	return concrete
}

func TestNewClient_ConfigValidation(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil)
	if !errors.Is(err, ErrNilConfig) {
		t.Fatalf("期望错误 ErrNilConfig，实际为: %v", err)
	}

	_, err = NewClient(&Config{Enable: true, Port: 25, Username: "user", FromEmail: "noreply@example.com"})
	if !errors.Is(err, ErrEmptyHost) {
		t.Fatalf("期望错误 ErrEmptyHost，实际为: %v", err)
	}

	_, err = NewClient(&Config{Enable: true, Host: "smtp.example.com", Port: 0, Username: "user", FromEmail: "noreply@example.com"})
	if !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("期望错误 ErrInvalidPort，实际为: %v", err)
	}

	_, err = NewClient(&Config{Enable: true, Host: "smtp.example.com", Port: 25, FromEmail: "noreply@example.com"})
	if !errors.Is(err, ErrEmptyUsername) {
		t.Fatalf("期望错误 ErrEmptyUsername，实际为: %v", err)
	}

	_, err = NewClient(&Config{Enable: true, Host: "smtp.example.com", Port: 25, Username: "user"})
	if !errors.Is(err, ErrEmptyFromEmail) {
		t.Fatalf("期望错误 ErrEmptyFromEmail，实际为: %v", err)
	}

	_, err = NewClient(&Config{Enable: true, Host: "smtp.example.com", Port: 25, Username: "user", FromEmail: "noreply@example.com", TLSMode: "unknown"})
	if !errors.Is(err, ErrInvalidTLSMode) {
		t.Fatalf("期望错误 ErrInvalidTLSMode，实际为: %v", err)
	}
}

func TestNewClient_DisabledReturnsNil(t *testing.T) {
	t.Parallel()

	got, err := NewClient(&Config{Enable: false})
	if err != nil {
		t.Fatalf("期望禁用邮件时不返回错误，实际为: %v", err)
	}
	if got != nil {
		t.Fatalf("期望禁用邮件时返回 nil 客户端，实际为: %#v", got)
	}
}

func TestNewClient_TLSModeMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		mode    string
		wantEnc mail.Encryption
	}{
		{name: "默认starttls", mode: "", wantEnc: mail.EncryptionSTARTTLS},
		{name: "notls", mode: TLSModeNoTLS, wantEnc: mail.EncryptionNone},
		{name: "ssl", mode: TLSModeSSL, wantEnc: mail.EncryptionSSL},
		{name: "starttls", mode: TLSModeStartTLS, wantEnc: mail.EncryptionSTARTTLS},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := mustNewConcreteClient(t, &Config{
				Enable:    true,
				Host:      "smtp.example.com",
				Port:      25,
				Username:  "user",
				FromEmail: "noreply@example.com",
				TLSMode:   tc.mode,
			})
			if client.encryption != tc.wantEnc {
				t.Fatalf("TLS 模式映射不正确，want=%v，got=%v", tc.wantEnc, client.encryption)
			}
		})
	}
}

func TestSendHTML_ValidateInput(t *testing.T) {
	t.Parallel()

	client := mustNewConcreteClient(t, &Config{
		Enable:    true,
		Host:      "smtp.example.com",
		Port:      25,
		Username:  "user",
		FromEmail: "noreply@example.com",
		TLSMode:   TLSModeNoTLS,
	})

	client.connectFn = func(*mail.SMTPServer) (*mail.SMTPClient, error) {
		return &mail.SMTPClient{}, nil
	}
	client.sendFn = func(*mail.Email, *mail.SMTPClient) error { return nil }
	client.closeFn = func(*mail.SMTPClient) error { return nil }

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.SendHTML(canceledCtx, []string{"a@example.com"}, "主题", "<h1>内容</h1>"); !errors.Is(err, context.Canceled) {
		t.Fatalf("期望 context canceled，实际为: %v", err)
	}

	if err := client.SendHTML(context.Background(), nil, "主题", "<h1>内容</h1>"); !errors.Is(err, ErrEmptyRecipient) {
		t.Fatalf("期望错误 ErrEmptyRecipient，实际为: %v", err)
	}

	if err := client.SendHTML(context.Background(), []string{"a@example.com"}, "", "<h1>内容</h1>"); !errors.Is(err, ErrEmptySubject) {
		t.Fatalf("期望错误 ErrEmptySubject，实际为: %v", err)
	}

	if err := client.SendHTML(context.Background(), []string{"a@example.com"}, "主题", ""); !errors.Is(err, ErrEmptyHTMLBody) {
		t.Fatalf("期望错误 ErrEmptyHTMLBody，实际为: %v", err)
	}
}

func TestSendHTML_Success(t *testing.T) {
	t.Parallel()

	client := mustNewConcreteClient(t, &Config{
		Enable:         true,
		Host:           "smtp.example.com",
		Port:           465,
		Username:       "user",
		Password:       "pass",
		FromEmail:      "noreply@example.com",
		TLSMode:        TLSModeSSL,
		ConnectTimeout: 3 * time.Second,
		SendTimeout:    4 * time.Second,
	})

	calledConnect := false
	calledSend := false
	calledClose := false

	client.connectFn = func(server *mail.SMTPServer) (*mail.SMTPClient, error) {
		calledConnect = true
		if server.Host != "smtp.example.com" || server.Port != 465 {
			t.Fatalf("SMTP 服务器参数不符合预期")
		}
		if server.ConnectTimeout != 3*time.Second || server.SendTimeout != 4*time.Second {
			t.Fatalf("超时参数不符合预期")
		}
		return &mail.SMTPClient{}, nil
	}
	client.sendFn = func(message *mail.Email, _ *mail.SMTPClient) error {
		calledSend = true
		if message.Error != nil {
			t.Fatalf("邮件构建出现错误: %v", message.Error)
		}
		return nil
	}
	client.closeFn = func(*mail.SMTPClient) error {
		calledClose = true
		return nil
	}

	err := client.SendHTML(context.Background(), []string{"a@example.com", "", "a@example.com", "b@example.com"}, "主题", "<p>正文</p>")
	if err != nil {
		t.Fatalf("发送 HTML 邮件失败: %v", err)
	}
	if !calledConnect || !calledSend || !calledClose {
		t.Fatalf("发送流程未完整执行，connect=%v send=%v close=%v", calledConnect, calledSend, calledClose)
	}
}

func TestSendHTML_SendFail(t *testing.T) {
	t.Parallel()

	client := mustNewConcreteClient(t, &Config{
		Enable:    true,
		Host:      "smtp.example.com",
		Port:      25,
		Username:  "user",
		FromEmail: "noreply@example.com",
		TLSMode:   TLSModeNoTLS,
	})

	client.connectFn = func(*mail.SMTPServer) (*mail.SMTPClient, error) {
		return &mail.SMTPClient{}, nil
	}
	client.sendFn = func(*mail.Email, *mail.SMTPClient) error {
		return errors.New("mock send fail")
	}
	client.closeFn = func(*mail.SMTPClient) error { return nil }

	err := client.SendHTML(context.Background(), []string{"a@example.com"}, "主题", "<p>正文</p>")
	if err == nil || !strings.Contains(err.Error(), "发送邮件失败") {
		t.Fatalf("期望发送失败包装错误，实际为: %v", err)
	}
}
