package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/wire"
	mail "github.com/xhit/go-simple-mail/v2"
)

const (
	TLSModeNoTLS    = "notls"
	TLSModeSSL      = "ssl"
	TLSModeStartTLS = "starttls"

	defaultConnectTimeout = 10 * time.Second
	defaultSendTimeout    = 10 * time.Second
)

var (
	ErrNilConfig       = errors.New("邮件配置不能为空")
	ErrEmptyHost       = errors.New("邮件服务器地址不能为空")
	ErrInvalidPort     = errors.New("邮件服务器端口必须大于 0")
	ErrEmptyUsername   = errors.New("邮件用户名不能为空")
	ErrEmptyFromEmail  = errors.New("发件人邮箱不能为空")
	ErrInvalidTLSMode  = errors.New("不支持的邮件 TLS 模式")
	ErrEmptyRecipient  = errors.New("收件人不能为空")
	ErrEmptySubject    = errors.New("邮件主题不能为空")
	ErrEmptyHTMLBody   = errors.New("邮件 HTML 内容不能为空")
)

var ProviderSet = wire.NewSet(NewClient)

type Config struct {
	Host           string
	Port           int
	Username       string
	Password       string
	FromEmail      string
	TLSMode        string
	ConnectTimeout time.Duration
	SendTimeout    time.Duration
}

type Client struct {
	host           string
	port           int
	username       string
	password       string
	fromEmail      string
	encryption     mail.Encryption
	connectTimeout time.Duration
	sendTimeout    time.Duration

	connectFn func(*mail.SMTPServer) (*mail.SMTPClient, error)
	sendFn    func(*mail.Email, *mail.SMTPClient) error
	closeFn   func(*mail.SMTPClient) error
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return nil, ErrEmptyHost
	}
	if cfg.Port <= 0 {
		return nil, ErrInvalidPort
	}
	username := strings.TrimSpace(cfg.Username)
	if username == "" {
		return nil, ErrEmptyUsername
	}
	fromEmail := strings.TrimSpace(cfg.FromEmail)
	if fromEmail == "" {
		return nil, ErrEmptyFromEmail
	}

	encryption, err := parseTLSMode(cfg.TLSMode)
	if err != nil {
		return nil, err
	}

	client := &Client{
		host:           host,
		port:           cfg.Port,
		username:       username,
		password:       cfg.Password,
		fromEmail:      fromEmail,
		encryption:     encryption,
		connectTimeout: defaultOr(cfg.ConnectTimeout, defaultConnectTimeout),
		sendTimeout:    defaultOr(cfg.SendTimeout, defaultSendTimeout),
	}

	client.connectFn = func(server *mail.SMTPServer) (*mail.SMTPClient, error) {
		return server.Connect()
	}
	client.sendFn = func(message *mail.Email, smtpClient *mail.SMTPClient) error {
		return message.Send(smtpClient)
	}
	client.closeFn = func(smtpClient *mail.SMTPClient) error {
		if smtpClient == nil {
			return nil
		}
		return smtpClient.Close()
	}

	return client, nil
}

func (c *Client) SendHTML(ctx context.Context, to []string, subject, htmlBody string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	recipients := normalizeRecipients(to)
	if len(recipients) == 0 {
		return ErrEmptyRecipient
	}
	if strings.TrimSpace(subject) == "" {
		return ErrEmptySubject
	}
	if strings.TrimSpace(htmlBody) == "" {
		return ErrEmptyHTMLBody
	}

	server := mail.NewSMTPClient()
	server.Host = c.host
	server.Port = c.port
	server.Username = c.username
	server.Password = c.password
	server.Encryption = c.encryption
	server.ConnectTimeout = c.connectTimeout
	server.SendTimeout = c.sendTimeout

	smtpClient, err := c.connectFn(server)
	if err != nil {
		return fmt.Errorf("连接邮件服务器失败: %w", err)
	}
	defer func() { _ = c.closeFn(smtpClient) }()

	message := mail.NewMSG()
	message.SetFrom(c.fromEmail)
	message.AddTo(recipients...)
	message.SetSubject(subject)
	message.SetBody(mail.TextHTML, htmlBody)

	if message.Error != nil {
		return fmt.Errorf("构建邮件内容失败: %w", message.Error)
	}

	if err := c.sendFn(message, smtpClient); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

func parseTLSMode(mode string) (mail.Encryption, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", TLSModeStartTLS:
		return mail.EncryptionSTARTTLS, nil
	case TLSModeNoTLS:
		return mail.EncryptionNone, nil
	case TLSModeSSL:
		return mail.EncryptionSSL, nil
	default:
		return mail.EncryptionNone, ErrInvalidTLSMode
	}
}

func normalizeRecipients(to []string) []string {
	if len(to) == 0 {
		return nil
	}

	uniq := make(map[string]struct{}, len(to))
	result := make([]string, 0, len(to))
	for _, raw := range to {
		addr := strings.TrimSpace(raw)
		if addr == "" {
			continue
		}
		if _, exists := uniq[addr]; exists {
			continue
		}
		uniq[addr] = struct{}{}
		result = append(result, addr)
	}

	return result
}

func defaultOr(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}

