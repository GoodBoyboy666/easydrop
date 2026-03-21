package service

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"easydrop/internal/config"
	"easydrop/internal/pkg/email"
)

const (
	passwordResetTemplateFile = "password_reset.html"
	emailVerifyTemplateFile   = "email_verify.html"
	emailChangeTemplateFile   = "email_change.html"

	passwordResetPath = "reset-password"
	emailVerifyPath   = "verify-email"
	emailChangePath   = "change-email"

	passwordResetSubject = "重置密码确认"
	emailVerifySubject   = "邮箱验证"
	emailChangeSubject   = "邮箱修改确认"
)

var (
	ErrEmailServiceUnavailable = errors.New("邮件服务未初始化")
	ErrEmailRecipientRequired  = errors.New("收件人不能为空")
	ErrEmailTokenRequired      = errors.New("邮件 token 不能为空")
	ErrEmailTTLInvalid         = errors.New("邮件 token 有效期必须大于 0")
)

const passwordResetFallbackTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>重置密码</title>
</head>
<body>
	<p>你正在执行密码重置，请点击下方链接继续操作。</p>
	{{if .ActionURL}}
	<p><a href="{{.ActionURL}}">立即重置密码</a></p>
	<p>如果无法点击，请复制链接：{{.ActionURL}}</p>
	{{end}}
	<p>有效期：{{.TTL}}</p>
	{{if .SiteURL}}
	<p>来源站点：{{.SiteURL}}</p>
	{{end}}
</body>
</html>`

const emailVerifyFallbackTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>邮箱验证</title>
</head>
<body>
	<p>请验证你的邮箱地址，点击下方链接完成验证。</p>
	{{if .ActionURL}}
	<p><a href="{{.ActionURL}}">立即验证邮箱</a></p>
	<p>如果无法点击，请复制链接：{{.ActionURL}}</p>
	{{end}}
	<p>有效期：{{.TTL}}</p>
	{{if .SiteURL}}
	<p>来源站点：{{.SiteURL}}</p>
	{{end}}
</body>
</html>`

const emailChangeFallbackTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>邮箱修改确认</title>
</head>
<body>
	{{if .NewEmail}}
	<p>你正在将账号邮箱修改为：{{.NewEmail}}</p>
	{{else}}
	<p>你正在执行邮箱修改操作，请确认该请求。</p>
	{{end}}
	{{if .ActionURL}}
	<p><a href="{{.ActionURL}}">确认修改邮箱</a></p>
	<p>如果无法点击，请复制链接：{{.ActionURL}}</p>
	{{end}}
	<p>有效期：{{.TTL}}</p>
	{{if .SiteURL}}
	<p>来源站点：{{.SiteURL}}</p>
	{{end}}
</body>
</html>`

// EmailService 负责渲染邮件模板并发送业务邮件。
type EmailService interface {
	SendPasswordResetEmail(ctx context.Context, to, tokenValue string, ttl time.Duration) error
	SendVerifyEmail(ctx context.Context, to, tokenValue string, ttl time.Duration) error
	SendChangeEmailEmail(ctx context.Context, to, newEmail, tokenValue string, ttl time.Duration) error
}

type emailService struct {
	sender   email.Client
	dbConfig config.DBConfig
}

type emailTemplateData struct {
	ActionURL string
	TTL       string
	SiteURL   string
	NewEmail  string
}

// NewEmailService 创建邮件服务。
func NewEmailService(sender email.Client, dbConfig config.DBConfig) EmailService {
	return &emailService{sender: sender, dbConfig: dbConfig}
}

func (s *emailService) SendPasswordResetEmail(ctx context.Context, to, tokenValue string, ttl time.Duration) error {
	data, err := s.newTemplateData(ctx, tokenValue, ttl, passwordResetPath)
	if err != nil {
		return err
	}

	body, err := s.renderTemplate(passwordResetTemplateFile, data)
	if err != nil {
		return err
	}

	return s.send(ctx, to, passwordResetSubject, body)
}

func (s *emailService) SendVerifyEmail(ctx context.Context, to, tokenValue string, ttl time.Duration) error {
	data, err := s.newTemplateData(ctx, tokenValue, ttl, emailVerifyPath)
	if err != nil {
		return err
	}

	body, err := s.renderTemplate(emailVerifyTemplateFile, data)
	if err != nil {
		return err
	}

	return s.send(ctx, to, emailVerifySubject, body)
}

func (s *emailService) SendChangeEmailEmail(ctx context.Context, to, newEmail, tokenValue string, ttl time.Duration) error {
	data, err := s.newTemplateData(ctx, tokenValue, ttl, emailChangePath)
	if err != nil {
		return err
	}
	data.NewEmail = strings.TrimSpace(newEmail)

	body, err := s.renderTemplate(emailChangeTemplateFile, data)
	if err != nil {
		return err
	}

	return s.send(ctx, to, emailChangeSubject, body)
}

func (s *emailService) send(ctx context.Context, to, subject, body string) error {
	if s.sender == nil {
		return ErrEmailServiceUnavailable
	}

	recipient := strings.TrimSpace(to)
	if recipient == "" {
		return ErrEmailRecipientRequired
	}

	if err := s.sender.SendHTML(ctx, []string{recipient}, subject, body); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}

func (s *emailService) newTemplateData(ctx context.Context, tokenValue string, ttl time.Duration, actionPath string) (*emailTemplateData, error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, ErrEmailTokenRequired
	}
	if ttl <= 0 {
		return nil, ErrEmailTTLInvalid
	}

	siteURL := s.getSiteURL(ctx)
	actionURL := buildActionURL(siteURL, actionPath, tokenValue)

	return &emailTemplateData{
		TTL:       ttl.String(),
		SiteURL:   siteURL,
		ActionURL: actionURL,
	}, nil
}

func (s *emailService) getSiteURL(ctx context.Context) string {
	if s.dbConfig == nil {
		return ""
	}

	value, ok, err := s.dbConfig.GetValue(ctx, "site.url")
	if err != nil {
		log.Printf("读取站点地址配置失败: %v", err)
		return ""
	}
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}

func (s *emailService) renderTemplate(templateFile string, data *emailTemplateData) (string, error) {
	content, found, err := loadTemplateContent(templateFile)
	if err != nil {
		return "", err
	}
	if !found {
		content = fallbackTemplateFor(templateFile)
	}

	tpl, err := template.New(templateFile).Parse(content)
	if err != nil {
		return "", fmt.Errorf("解析邮件模板失败: %w", err)
	}

	var builder strings.Builder
	if err := tpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	return builder.String(), nil
}

func fallbackTemplateFor(templateFile string) string {
	switch templateFile {
	case passwordResetTemplateFile:
		return passwordResetFallbackTemplate
	case emailVerifyTemplateFile:
		return emailVerifyFallbackTemplate
	case emailChangeTemplateFile:
		return emailChangeFallbackTemplate
	default:
		return passwordResetFallbackTemplate
	}
}

func loadTemplateContent(templateFile string) (string, bool, error) {
	configDir := strings.TrimSpace(config.GlobalConfigDir)
	if configDir == "" {
		return "", false, nil
	}

	paths := []string{
		filepath.Join(configDir, "templates", "email", templateFile),
		filepath.Join(configDir, "templates", templateFile),
	}

	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", false, fmt.Errorf("读取邮件模板失败: %w", err)
		}
		return string(content), true, nil
	}

	return "", false, nil
}

func buildActionURL(siteURL, actionPath, tokenValue string) string {
	actionPath = strings.Trim(strings.TrimSpace(actionPath), "/")
	if actionPath == "" {
		actionPath = "action"
	}

	query := url.Values{}
	query.Set("token", tokenValue)
	trimmedSiteURL := strings.TrimSpace(siteURL)
	if trimmedSiteURL == "" {
		return "/" + actionPath + "?" + query.Encode()
	}

	base, err := url.Parse(trimmedSiteURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "/" + actionPath + "?" + query.Encode()
	}

	relative := &url.URL{Path: "/" + actionPath, RawQuery: query.Encode()}
	return base.ResolveReference(relative).String()
}
