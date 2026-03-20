package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"easydrop/internal/config"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type captureEmailSender struct {
	to      []string
	subject string
	body    string
	err     error
}

func (s *captureEmailSender) SendHTML(_ context.Context, to []string, subject, htmlBody string) error {
	if s.err != nil {
		return s.err
	}
	s.to = append([]string{}, to...)
	s.subject = subject
	s.body = htmlBody
	return nil
}

func TestEmailServiceUsesDefaultTemplateWhenFileMissing(t *testing.T) {
	oldConfigDir := config.GlobalConfigDir
	config.GlobalConfigDir = ""
	t.Cleanup(func() {
		config.GlobalConfigDir = oldConfigDir
	})

	sender := &captureEmailSender{}
	svc := NewEmailService(sender, nil)

	err := svc.SendPasswordResetEmail(context.Background(), "u@example.com", "token-abc", 30*time.Minute)
	if err != nil {
		t.Fatalf("SendPasswordResetEmail returned error: %v", err)
	}

	if sender.subject != passwordResetSubject {
		t.Fatalf("unexpected subject: %s", sender.subject)
	}
	if len(sender.to) != 1 || sender.to[0] != "u@example.com" {
		t.Fatalf("unexpected recipients: %#v", sender.to)
	}
	if !strings.Contains(sender.body, "/reset-password?token=token-abc") {
		t.Fatalf("expected reset action url in body, got: %s", sender.body)
	}
	if strings.Contains(sender.body, "令牌：") {
		t.Fatalf("expected no plaintext token label in body, got: %s", sender.body)
	}
}

func TestEmailServiceUsesTemplateFileFromConfigDir(t *testing.T) {
	configDir := t.TempDir()
	templateDir := filepath.Join(configDir, "templates", "email")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("create template dir failed: %v", err)
	}

	customTemplate := "custom {{.ActionURL}}"
	if err := os.WriteFile(filepath.Join(templateDir, passwordResetTemplateFile), []byte(customTemplate), 0o644); err != nil {
		t.Fatalf("write template file failed: %v", err)
	}

	oldConfigDir := config.GlobalConfigDir
	config.GlobalConfigDir = configDir
	t.Cleanup(func() {
		config.GlobalConfigDir = oldConfigDir
	})

	sender := &captureEmailSender{}
	svc := NewEmailService(sender, nil)

	err := svc.SendPasswordResetEmail(context.Background(), "u@example.com", "abc", time.Hour)
	if err != nil {
		t.Fatalf("SendPasswordResetEmail returned error: %v", err)
	}

	if !strings.Contains(sender.body, "custom /reset-password?token=abc") {
		t.Fatalf("expected custom template content, got: %s", sender.body)
	}
}

func TestEmailServiceBuildsAbsoluteActionURLFromSiteURL(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	dbConfig, err := config.NewDBConfig(db)
	if err != nil {
		t.Fatalf("create db config failed: %v", err)
	}

	if err := dbConfig.Set(context.Background(), "site.url", "https://example.com"); err != nil {
		t.Fatalf("set site.url failed: %v", err)
	}

	oldConfigDir := config.GlobalConfigDir
	config.GlobalConfigDir = ""
	t.Cleanup(func() {
		config.GlobalConfigDir = oldConfigDir
	})

	sender := &captureEmailSender{}
	svc := NewEmailService(sender, dbConfig)

	err = svc.SendVerifyEmail(context.Background(), "u@example.com", "verify-1", 10*time.Minute)
	if err != nil {
		t.Fatalf("SendVerifyEmail returned error: %v", err)
	}

	if !strings.Contains(sender.body, "https://example.com/verify-email?token=verify-1") {
		t.Fatalf("expected absolute action url in body, got: %s", sender.body)
	}
}

func TestEmailServicePropagatesSendError(t *testing.T) {
	wantErr := errors.New("send failed")
	sender := &captureEmailSender{err: wantErr}
	svc := NewEmailService(sender, nil)

	err := svc.SendChangeEmailEmail(context.Background(), "u@example.com", "new@example.com", "change-1", 5*time.Minute)
	if err == nil {
		t.Fatal("expected send error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped send error, got: %v", err)
	}
}

func TestEmailServiceChangeEmailTemplateContainsNewEmail(t *testing.T) {
	configDir := t.TempDir()
	templateDir := filepath.Join(configDir, "templates", "email")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("create template dir failed: %v", err)
	}

	customTemplate := "change to {{.NewEmail}} {{.ActionURL}}"
	if err := os.WriteFile(filepath.Join(templateDir, emailChangeTemplateFile), []byte(customTemplate), 0o644); err != nil {
		t.Fatalf("write template file failed: %v", err)
	}

	oldConfigDir := config.GlobalConfigDir
	config.GlobalConfigDir = configDir
	t.Cleanup(func() {
		config.GlobalConfigDir = oldConfigDir
	})

	sender := &captureEmailSender{}
	svc := NewEmailService(sender, nil)

	err := svc.SendChangeEmailEmail(context.Background(), "u@example.com", "new@example.com", "abc", time.Hour)
	if err != nil {
		t.Fatalf("SendChangeEmailEmail returned error: %v", err)
	}

	if !strings.Contains(sender.body, "change to new@example.com /change-email?token=abc") {
		t.Fatalf("expected new email in template body, got: %s", sender.body)
	}
}
