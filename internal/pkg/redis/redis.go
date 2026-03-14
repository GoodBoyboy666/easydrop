package redis

import (
	"errors"
	"strings"
	"time"

	"github.com/google/wire"
	red "github.com/redis/go-redis/v9"
)

const (
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 3 * time.Second
	defaultWriteTimeout = 3 * time.Second
)

var (
	ErrNilConfig = errors.New("redis 配置不能为空")
	ErrEmptyAddr = errors.New("redis 地址不能为空")
)

var ProviderSet = wire.NewSet(NewClient)

type Config struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
}

func NewClient(cfg *Config) (*red.Client, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		return nil, ErrEmptyAddr
	}

	opts := &red.Options{
		Addr:         addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  defaultOr(cfg.DialTimeout, defaultDialTimeout),
		ReadTimeout:  defaultOr(cfg.ReadTimeout, defaultReadTimeout),
		WriteTimeout: defaultOr(cfg.WriteTimeout, defaultWriteTimeout),
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	}

	return red.NewClient(opts), nil
}

func defaultOr(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}

