package initsecret

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
)

var (
	ErrRequired = errors.New("init secret 不能为空")
	ErrInvalid  = errors.New("init secret 无效")
	ErrNotReady = errors.New("init secret 未就绪")
)

// Guard 管理首次初始化阶段使用的进程内 secret。
type Guard interface {
	EnsureSecret() (string, error)
	Validate(secret string) error
}

type guard struct {
	mu     sync.RWMutex
	secret string
}

// NewGuard 创建初始化 secret 守卫。
func NewGuard() Guard {
	return &guard{}
}

func (g *guard) EnsureSecret() (string, error) {
	g.mu.RLock()
	if g.secret != "" {
		secret := g.secret
		g.mu.RUnlock()
		return secret, nil
	}
	g.mu.RUnlock()

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(buf)

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.secret == "" {
		g.secret = secret
	}
	return g.secret, nil
}

func (g *guard) Validate(secret string) error {
	cleanSecret := strings.TrimSpace(secret)
	if cleanSecret == "" {
		return ErrRequired
	}

	g.mu.RLock()
	currentSecret := g.secret
	g.mu.RUnlock()
	if currentSecret == "" {
		return ErrNotReady
	}

	if len(cleanSecret) != len(currentSecret) || subtle.ConstantTimeCompare([]byte(cleanSecret), []byte(currentSecret)) != 1 {
		return ErrInvalid
	}

	return nil
}

var _ Guard = (*guard)(nil)
