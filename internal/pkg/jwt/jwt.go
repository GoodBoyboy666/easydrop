package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	ErrNilConfig         = errors.New("jwt 配置不能为空")
	ErrEmptyPrivateKey   = errors.New("jwt 私钥路径不能为空")
	ErrEmptyPublicKey    = errors.New("jwt 公钥路径不能为空")
	ErrInvalidPrivateKey = errors.New("jwt 私钥格式不合法")
	ErrInvalidPublicKey  = errors.New("jwt 公钥格式不合法")
	ErrInvalidExpire     = errors.New("jwt 过期时间必须大于 0")
	ErrInvalidToken      = errors.New("jwt token 无效")
	ErrExpiredToken      = errors.New("jwt token 已过期")
	ErrBadTokenSignature = errors.New("jwt token 签名无效")
)

type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username,omitempty"`
	Admin    bool   `json:"admin"`
	jwtv5.RegisteredClaims
}

type Config struct {
	PrivateKeyPath string        `mapstructure:"private_key_path" yaml:"private_key_path"`
	PublicKeyPath  string        `mapstructure:"public_key_path" yaml:"public_key_path"`
	Issuer         string        `mapstructure:"issuer" yaml:"issuer"`
	Expire         time.Duration `mapstructure:"expire" yaml:"expire"`
}

type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expire     time.Duration
	now        func() time.Time
}

func NewManager(cfg *Config) (*Manager, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	if cfg.PrivateKeyPath == "" {
		return nil, ErrEmptyPrivateKey
	}
	if cfg.PublicKeyPath == "" {
		return nil, ErrEmptyPublicKey
	}
	if cfg.Expire <= 0 {
		return nil, ErrInvalidExpire
	}

	privatePEM, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取 jwt 私钥失败: %w", err)
	}

	publicPEM, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取 jwt 公钥失败: %w", err)
	}

	privateKey, err := parseRSAPrivateKey(privatePEM)
	if err != nil {
		return nil, err
	}

	publicKey, err := parseRSAPublicKey(publicPEM)
	if err != nil {
		return nil, err
	}

	return &Manager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     cfg.Issuer,
		expire:     cfg.Expire,
		now:        time.Now,
	}, nil
}

func (m *Manager) IssueAccessToken(userID uint, username string, admin bool) (string, error) {
	now := m.now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Admin:    admin,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(m.expire)),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", fmt.Errorf("签发 jwt token 失败: %w", err)
	}

	return signed, nil
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	options := []jwtv5.ParserOption{jwtv5.WithTimeFunc(m.now)}
	if m.issuer != "" {
		options = append(options, jwtv5.WithIssuer(m.issuer))
	}

	token, err := jwtv5.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwtv5.Token) (any, error) {
			if token.Method.Alg() != jwtv5.SigningMethodRS256.Alg() {
				return nil, ErrBadTokenSignature
			}
			return m.publicKey, nil
		},
		options...,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrBadTokenSignature):
			return nil, ErrBadTokenSignature
		case errors.Is(err, jwtv5.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwtv5.ErrTokenSignatureInvalid):
			return nil, ErrBadTokenSignature
		default:
			return nil, ErrInvalidToken
		}
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func parseRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, ErrInvalidPrivateKey
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return privateKey, nil
	}

	parsedPKCS8, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
	if parseErr != nil {
		return nil, ErrInvalidPrivateKey
	}

	rsaPrivateKey, ok := parsedPKCS8.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidPrivateKey
	}

	return rsaPrivateKey, nil
}

func parseRSAPublicKey(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, ErrInvalidPublicKey
	}

	if publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return publicKey, nil
	}

	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	rsaPublicKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	return rsaPublicKey, nil
}
