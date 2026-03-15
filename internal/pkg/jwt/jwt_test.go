package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	testKeyOnce    sync.Once
	testPrivatePEM string
	testPublicPEM  string
	testKeyErr     error
)

func testRSAKeyPair(t *testing.T) (string, string) {
	t.Helper()

	testKeyOnce.Do(func() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			testKeyErr = err
			return
		}

		privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
		testPrivatePEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateDER}))

		publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			testKeyErr = err
			return
		}
		testPublicPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}))
	})

	if testKeyErr != nil {
		t.Fatalf("生成测试 RSA 密钥对失败: %v", testKeyErr)
	}

	return testPrivatePEM, testPublicPEM
}

func writeKeyFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("写入测试密钥失败: %v", err)
	}
	return path
}

func testKeyPairPaths(t *testing.T) (string, string) {
	t.Helper()

	privateKeyPEM, publicKeyPEM := testRSAKeyPair(t)
	dir := t.TempDir()
	privatePath := writeKeyFile(t, dir, "private.pem", privateKeyPEM)
	publicPath := writeKeyFile(t, dir, "public.pem", publicKeyPEM)
	return privatePath, publicPath
}

func TestNewManager(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)

	_, err := NewManager(nil)
	if !errors.Is(err, ErrNilConfig) {
		t.Fatalf("期望错误 ErrNilConfig，实际为: %v", err)
	}

	_, err = NewManager(&Config{PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Hour})
	if !errors.Is(err, ErrEmptyPrivateKey) {
		t.Fatalf("期望错误 ErrEmptyPrivateKey，实际为: %v", err)
	}

	_, err = NewManager(&Config{PrivateKeyPath: privatePath, Issuer: "easydrop", Expire: time.Hour})
	if !errors.Is(err, ErrEmptyPublicKey) {
		t.Fatalf("期望错误 ErrEmptyPublicKey，实际为: %v", err)
	}

	badDir := t.TempDir()
	badPrivatePath := writeKeyFile(t, badDir, "bad-private.pem", "bad-key")
	_, err = NewManager(&Config{PrivateKeyPath: badPrivatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Hour})
	if !errors.Is(err, ErrInvalidPrivateKey) {
		t.Fatalf("期望错误 ErrInvalidPrivateKey，实际为: %v", err)
	}

	badPublicPath := writeKeyFile(t, badDir, "bad-public.pem", "bad-key")
	_, err = NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: badPublicPath, Issuer: "easydrop", Expire: time.Hour})
	if !errors.Is(err, ErrInvalidPublicKey) {
		t.Fatalf("期望错误 ErrInvalidPublicKey，实际为: %v", err)
	}

	_, err = NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: 0})
	if !errors.Is(err, ErrInvalidExpire) {
		t.Fatalf("期望错误 ErrInvalidExpire，实际为: %v", err)
	}
}

func TestIssueAndParseToken(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)

	manager, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	fixedNow := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return fixedNow }

	token, err := manager.IssueAccessToken(1001, "alice", true)
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	claims, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("解析 token 失败: %v", err)
	}

	if claims.UserID != 1001 || claims.Username != "alice" || !claims.Admin {
		t.Fatalf("claims 不符合预期: %+v", claims)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Unix() != fixedNow.Add(time.Hour).Unix() {
		t.Fatalf("expires_at 不符合预期: %+v", claims.ExpiresAt)
	}
}

func TestParseExpiredToken(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)

	manager, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Minute})
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	issuedAt := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return issuedAt }

	token, err := manager.IssueAccessToken(1001, "alice", false)
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	manager.now = func() time.Time { return issuedAt.Add(2 * time.Minute) }

	_, err = manager.ParseToken(token)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("期望错误 ErrExpiredToken，实际为: %v", err)
	}
}

func TestParseTokenWithWrongPublicKey(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)
	_, otherPublicPEM := testNewRSAKeyPair(t)
	otherPublicPath := writeKeyFile(t, t.TempDir(), "other-public.pem", otherPublicPEM)

	issuer, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建签发方管理器失败: %v", err)
	}
	verifier, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: otherPublicPath, Issuer: "easydrop", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建校验方管理器失败: %v", err)
	}

	fixedNow := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	issuer.now = func() time.Time { return fixedNow }
	verifier.now = func() time.Time { return fixedNow }

	token, err := issuer.IssueAccessToken(1001, "alice", false)
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	_, err = verifier.ParseToken(token)
	if !errors.Is(err, ErrBadTokenSignature) {
		t.Fatalf("期望错误 ErrBadTokenSignature，实际为: %v", err)
	}
}

func TestParseTokenWithWrongIssuer(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)

	issuer, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "issuer-a", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建签发方管理器失败: %v", err)
	}
	verifier, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "issuer-b", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建校验方管理器失败: %v", err)
	}

	fixedNow := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	issuer.now = func() time.Time { return fixedNow }
	verifier.now = func() time.Time { return fixedNow }

	token, err := issuer.IssueAccessToken(1001, "alice", false)
	if err != nil {
		t.Fatalf("签发 token 失败: %v", err)
	}

	_, err = verifier.ParseToken(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("期望错误 ErrInvalidToken，实际为: %v", err)
	}
}

func TestParseMalformedToken(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)

	manager, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	_, err = manager.ParseToken("not-a-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("期望错误 ErrInvalidToken，实际为: %v", err)
	}
}

func TestParseHS256Token(t *testing.T) {
	t.Parallel()
	privatePath, publicPath := testKeyPairPaths(t)

	manager, err := NewManager(&Config{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, Issuer: "easydrop", Expire: time.Hour})
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	fixedNow := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return fixedNow }

	claims := Claims{
		UserID:   1001,
		Username: "alice",
		Admin:    false,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    "easydrop",
			Subject:   "1001",
			IssuedAt:  jwtv5.NewNumericDate(fixedNow),
			NotBefore: jwtv5.NewNumericDate(fixedNow),
			ExpiresAt: jwtv5.NewNumericDate(fixedNow.Add(time.Hour)),
		},
	}

	hsToken := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := hsToken.SignedString([]byte("hs-secret"))
	if err != nil {
		t.Fatalf("签发 HS256 token 失败: %v", err)
	}

	_, err = manager.ParseToken(tokenString)
	if !errors.Is(err, ErrBadTokenSignature) {
		t.Fatalf("期望错误 ErrBadTokenSignature，实际为: %v", err)
	}
}

func testNewRSAKeyPair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成 RSA 密钥失败: %v", err)
	}

	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateDER}))

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("序列化公钥失败: %v", err)
	}
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}))

	return privatePEM, publicPEM
}
