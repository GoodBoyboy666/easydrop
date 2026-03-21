package token

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	red "github.com/redis/go-redis/v9"
)

const (
	BackendMemory = "memory"
	BackendRedis  = "redis"

	KindResetPassword = "reset_password"
	KindVerifyEmail   = "verify_email"
	KindChangeEmail   = "change_email"

	defaultKeyPrefix = "token"
	randomTokenBytes = 32
)

var (
	ErrInvalidUserID = errors.New("user id 必须大于 0")
	ErrEmptyKind     = errors.New("token 类型不能为空")
	ErrInvalidTTL    = errors.New("token ttl 必须大于 0")
	ErrEmptyToken    = errors.New("token 不能为空")
	ErrTokenNotFound = errors.New("token 不存在")
	ErrTokenExpired  = errors.New("token 已过期")
	ErrTokenMismatch = errors.New("token 不匹配")
)

type Config struct {
	KeyPrefix string `mapstructure:"key_prefix" yaml:"key_prefix"`
}

type Record struct {
	UserID    uint      `json:"user_id"`
	Kind      string    `json:"kind"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Payload   string    `json:"payload"`
}

type store interface {
	Save(ctx context.Context, record *Record) error
	Consume(ctx context.Context, userID uint, kind, token string, now time.Time) (*Record, error)
}

type Manager interface {
	Backend() string
	Issue(ctx context.Context, userID uint, kind string, ttl time.Duration, payload string) (string, error)
	Consume(ctx context.Context, userID uint, kind, tokenValue string) (*Record, error)
}

type manager struct {
	store   store
	now     func() time.Time
	random  io.Reader
	backend string
}

func NewManager(cfg *Config, redisClient *red.Client) (Manager, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	var backend store
	backendName := BackendMemory
	if redisClient != nil {
		backend = newRedisStore(newGoRedisTxnClient(redisClient), normalizeKeyPrefix(cfg.KeyPrefix))
		backendName = BackendRedis
	} else {
		backend = newMemoryStore()
	}

	return &manager{
		store:   backend,
		now:     time.Now,
		random:  cryptorand.Reader,
		backend: backendName,
	}, nil
}

func (m *manager) Backend() string {
	return m.backend
}

func (m *manager) Issue(ctx context.Context, userID uint, kind string, ttl time.Duration, payload string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedKind, err := validateIssueInput(userID, kind, ttl)
	if err != nil {
		return "", err
	}

	tokenValue, err := m.generateToken()
	if err != nil {
		return "", err
	}

	now := m.now().UTC()
	record := &Record{
		UserID:    userID,
		Kind:      normalizedKind,
		Token:     tokenValue,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Payload:   payload,
	}

	if err := m.store.Save(ctx, record); err != nil {
		return "", err
	}

	return tokenValue, nil
}

func (m *manager) Consume(ctx context.Context, userID uint, kind, tokenValue string) (*Record, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedKind, normalizedToken, err := validateConsumeInput(userID, kind, tokenValue)
	if err != nil {
		return nil, err
	}

	record, err := m.store.Consume(ctx, userID, normalizedKind, normalizedToken, m.now().UTC())
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (m *manager) generateToken() (string, error) {
	reader := m.random
	if reader == nil {
		reader = cryptorand.Reader
	}

	buf := make([]byte, randomTokenBytes)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

var _ Manager = (*manager)(nil)

func validateIssueInput(userID uint, kind string, ttl time.Duration) (string, error) {
	if userID == 0 {
		return "", ErrInvalidUserID
	}

	normalizedKind := normalizeKind(kind)
	if normalizedKind == "" {
		return "", ErrEmptyKind
	}
	if ttl <= 0 {
		return "", ErrInvalidTTL
	}

	return normalizedKind, nil
}

func validateConsumeInput(userID uint, kind, tokenValue string) (string, string, error) {
	if userID == 0 {
		return "", "", ErrInvalidUserID
	}

	normalizedKind := normalizeKind(kind)
	if normalizedKind == "" {
		return "", "", ErrEmptyKind
	}

	normalizedToken := strings.TrimSpace(tokenValue)
	if normalizedToken == "" {
		return "", "", ErrEmptyToken
	}

	return normalizedKind, normalizedToken, nil
}

func normalizeKind(kind string) string {
	return strings.TrimSpace(kind)
}

func normalizeKeyPrefix(prefix string) string {
	normalized := strings.Trim(strings.TrimSpace(prefix), ":")
	if normalized == "" {
		return defaultKeyPrefix
	}
	return normalized
}

func cloneRecord(record *Record) (*Record, error) {
	if record == nil {
		return nil, nil
	}

	cloned := *record
	return &cloned, nil
}
