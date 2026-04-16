package initsecret

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	red "github.com/redis/go-redis/v9"
)

const storageKey = "easydrop:init:secret"

var (
	ErrRequired = errors.New("init secret 不能为空")
	ErrInvalid  = errors.New("init secret 无效")
	ErrNotReady = errors.New("init secret 未就绪")
)

// NewGuard 创建初始化 secret 守卫。Redis 客户端为空时回退到进程内存存储。
func NewGuard(client *red.Client) Guard {
	if client == nil {
		return newGuardWithStore(newMemoryStore())
	}
	return newGuardWithStore(&redisStore{client: client})
}

// Guard 管理首次初始化阶段使用的 secret。
type Guard interface {
	EnsureSecret(ctx context.Context) (string, error)
	Validate(ctx context.Context, secret string) error
}

type store interface {
	Get(ctx context.Context, key string) (string, bool, error)
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
}

type guard struct {
	store store
}

func newGuardWithStore(store store) Guard {
	return &guard{store: store}
}

func (g *guard) EnsureSecret(ctx context.Context) (string, error) {
	if g == nil || g.store == nil {
		return "", ErrNotReady
	}

	currentSecret, ok, err := g.store.Get(ctx, storageKey)
	if err != nil {
		return "", err
	}
	if ok && currentSecret != "" {
		return currentSecret, nil
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	generated := hex.EncodeToString(buf)

	stored, err := g.store.SetNX(ctx, storageKey, generated, 0)
	if err != nil {
		return "", err
	}
	if stored {
		return generated, nil
	}

	currentSecret, ok, err = g.store.Get(ctx, storageKey)
	if err != nil {
		return "", err
	}
	if !ok || currentSecret == "" {
		return "", ErrNotReady
	}
	return currentSecret, nil
}

func (g *guard) Validate(ctx context.Context, secret string) error {
	cleanSecret := strings.TrimSpace(secret)
	if cleanSecret == "" {
		return ErrRequired
	}
	if g == nil || g.store == nil {
		return ErrNotReady
	}

	currentSecret, ok, err := g.store.Get(ctx, storageKey)
	if err != nil {
		return err
	}
	if !ok || currentSecret == "" {
		return ErrNotReady
	}

	if len(cleanSecret) != len(currentSecret) || subtle.ConstantTimeCompare([]byte(cleanSecret), []byte(currentSecret)) != 1 {
		return ErrInvalid
	}

	return nil
}

var _ Guard = (*guard)(nil)
