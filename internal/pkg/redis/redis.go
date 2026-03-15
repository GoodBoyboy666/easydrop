package redis

import (
	"errors"
	"strings"
	"time"

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

type Config struct {
	Addr         string        `mapstructure:"addr" yaml:"addr"`
	Username     string        `mapstructure:"username" yaml:"username"`
	Password     string        `mapstructure:"password" yaml:"password"`
	DB           int           `mapstructure:"db" yaml:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries" yaml:"max_retries"`
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
