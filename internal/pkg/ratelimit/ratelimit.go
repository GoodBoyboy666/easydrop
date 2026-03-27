package ratelimit

import (
	"context"
	"errors"
	"strings"
	"time"

	red "github.com/redis/go-redis/v9"
)

const (
	BackendMemory = "memory"
	BackendRedis  = "redis"

	ModeCooldown = "cooldown"
	ModeWindow   = "window"

	defaultKeyPrefix  = "ratelimit"
	redisProbeTimeout = 500 * time.Millisecond
)

var (
	ErrEmptyKey      = errors.New("限流 key 不能为空")
	ErrInvalidMode   = errors.New("限流模式无效")
	ErrInvalidTTL    = errors.New("限流时间间隔必须大于 0")
	ErrInvalidLimit  = errors.New("窗口限流次数必须大于 0")
	ErrNilRedisProbe = errors.New("redis 限流客户端不能为空")
)

type Config struct {
	Enabled   bool                  `mapstructure:"enabled" yaml:"enabled"`
	KeyPrefix string                `mapstructure:"key_prefix" yaml:"key_prefix"`
	Rules     map[string]RuleConfig `mapstructure:"rules" yaml:"rules"`
}

type RuleConfig struct {
	Interval time.Duration `mapstructure:"interval" yaml:"interval"`
	Limit    int           `mapstructure:"limit" yaml:"limit"`
}

type Rule struct {
	Name     string
	Mode     string
	Interval time.Duration
	Limit    int
}

type Decision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
	ResetAt    time.Time
}

type Limiter interface {
	Backend() string
	Allow(ctx context.Context, key string, rule Rule) (*Decision, error)
}

type redisProbe interface {
	Ping(ctx context.Context) error
}

type redisStore interface {
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

type goRedisClient struct {
	client *red.Client
}

func NewLimiter(cfg *Config, redisClient *red.Client) (Limiter, error) {
	if redisClient == nil {
		return newMemoryLimiter(cfg), nil
	}

	return newLimiterWithRedis(cfg, &goRedisClient{client: redisClient}, &goRedisClient{client: redisClient})
}

func newLimiterWithRedis(cfg *Config, probe redisProbe, store redisStore) (Limiter, error) {
	if probe == nil || store == nil {
		return newMemoryLimiter(cfg), nil
	}

	probeCtx, cancel := context.WithTimeout(context.Background(), redisProbeTimeout)
	defer cancel()
	if err := probe.Ping(probeCtx); err != nil {
		return newMemoryLimiter(cfg), nil
	}

	return newRedisLimiter(cfg, store)
}

func ValidateRule(rule Rule) error {
	_, err := normalizeRule(rule)
	return err
}

func (c *goRedisClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *goRedisClient) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, ttl).Result()
}

func (c *goRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *goRedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *goRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

func normalizeRule(rule Rule) (Rule, error) {
	normalized := Rule{
		Name:     normalizeRuleName(rule.Name),
		Mode:     strings.ToLower(strings.TrimSpace(rule.Mode)),
		Interval: rule.Interval,
		Limit:    rule.Limit,
	}

	if normalized.Mode != ModeCooldown && normalized.Mode != ModeWindow {
		return Rule{}, ErrInvalidMode
	}
	if normalized.Interval <= 0 {
		return Rule{}, ErrInvalidTTL
	}
	if normalized.Mode == ModeWindow && normalized.Limit <= 0 {
		return Rule{}, ErrInvalidLimit
	}
	if normalized.Mode == ModeCooldown && normalized.Limit <= 0 {
		normalized.Limit = 1
	}

	return normalized, nil
}

func normalizeRuleName(name string) string {
	trimmed := strings.Trim(strings.TrimSpace(name), ":")
	if trimmed == "" {
		return "default"
	}
	return trimmed
}

func normalizeKeyPrefix(prefix string) string {
	trimmed := strings.Trim(strings.TrimSpace(prefix), ":")
	if trimmed == "" {
		return defaultKeyPrefix
	}
	return trimmed
}

func normalizeKey(key string) (string, error) {
	trimmed := strings.Trim(strings.TrimSpace(key), ":")
	if trimmed == "" {
		return "", ErrEmptyKey
	}
	return trimmed, nil
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
