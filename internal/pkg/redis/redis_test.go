package redis

import (
	"errors"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil)
	if !errors.Is(err, ErrNilConfig) {
		t.Fatalf("期望错误 ErrNilConfig，实际为: %v", err)
	}

	_, err = NewClient(&Config{Addr: "  "})
	if !errors.Is(err, ErrEmptyAddr) {
		t.Fatalf("期望错误 ErrEmptyAddr，实际为: %v", err)
	}
}

func TestNewClient_Defaults(t *testing.T) {
	t.Parallel()

	client, err := NewClient(&Config{Addr: "localhost:6379"})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	opts := client.Options()
	if opts.Addr != "localhost:6379" {
		t.Fatalf("地址不符合预期: %s", opts.Addr)
	}
	if opts.DialTimeout != defaultDialTimeout {
		t.Fatalf("连接超时不符合预期: %s", opts.DialTimeout)
	}
	if opts.ReadTimeout != defaultReadTimeout {
		t.Fatalf("读取超时不符合预期: %s", opts.ReadTimeout)
	}
	if opts.WriteTimeout != defaultWriteTimeout {
		t.Fatalf("写入超时不符合预期: %s", opts.WriteTimeout)
	}
}

func TestNewClient_MapsConfig(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Addr:         "127.0.0.1:6380",
		Username:     "user-a",
		Password:     "pass-a",
		DB:           5,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 6 * time.Second,
		PoolSize:     40,
		MinIdleConns: 8,
		MaxRetries:   1,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	opts := client.Options()
	if opts.Addr != cfg.Addr {
		t.Fatalf("地址不符合预期: %s", opts.Addr)
	}
	if opts.Username != cfg.Username {
		t.Fatalf("用户名不符合预期: %s", opts.Username)
	}
	if opts.Password != cfg.Password {
		t.Fatalf("密码不符合预期")
	}
	if opts.DB != cfg.DB {
		t.Fatalf("数据库编号不符合预期: %d", opts.DB)
	}
	if opts.DialTimeout != cfg.DialTimeout {
		t.Fatalf("连接超时不符合预期: %s", opts.DialTimeout)
	}
	if opts.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("读取超时不符合预期: %s", opts.ReadTimeout)
	}
	if opts.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("写入超时不符合预期: %s", opts.WriteTimeout)
	}
	if opts.PoolSize != cfg.PoolSize {
		t.Fatalf("连接池大小不符合预期: %d", opts.PoolSize)
	}
	if opts.MinIdleConns != cfg.MinIdleConns {
		t.Fatalf("最小空闲连接数不符合预期: %d", opts.MinIdleConns)
	}
	if opts.MaxRetries != cfg.MaxRetries {
		t.Fatalf("最大重试次数不符合预期: %d", opts.MaxRetries)
	}
}

