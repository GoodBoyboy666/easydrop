package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNilConfig            = errors.New("storage 配置不能为空")
	ErrEmptyStorageBackend  = errors.New("storage backend 不能为空")
	ErrUnsupportedBackend   = errors.New("不支持的 storage backend")
	ErrEmptyObjectKey       = errors.New("object key 不能为空")
	ErrEmptyObjectData      = errors.New("object 数据不能为空")
	ErrInvalidLocalBasePath = errors.New("local base path 不能为空")
	ErrInvalidS3Bucket      = errors.New("s3 bucket 不能为空")
	ErrInvalidS3Region      = errors.New("s3 region 不能为空")
	ErrInvalidS3AccessKey   = errors.New("s3 access key 不能为空")
	ErrInvalidS3SecretKey   = errors.New("s3 secret key 不能为空")
)

// Backend 定义底层存储实现能力。
type Backend interface {
	Upload(ctx context.Context, objectKey string, data []byte, contentType string) error
	Download(ctx context.Context, objectKey string) ([]byte, error)
	GetSize(ctx context.Context, objectKey string) (int64, error)
	Delete(ctx context.Context, objectKey string) error
	URL(ctx context.Context, objectKey string) (string, error)
}

// Manager 是上层统一入口，负责路由到 local/s3 并生成 object key。
type Manager struct {
	backend string
	svc     Backend
	now     func() time.Time
}

// NewManager 根据配置创建 storage manager。
func NewManager(cfg *Config) (*Manager, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	backend := strings.TrimSpace(cfg.Backend)
	if backend == "" {
		return nil, ErrEmptyStorageBackend
	}

	var svc Backend
	var err error

	switch backend {
	case BackendLocal:
		svc, err = NewLocalStorage(cfg.Local)
	case BackendS3:
		svc, err = NewS3Storage(cfg.S3)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedBackend, backend)
	}
	if err != nil {
		return nil, err
	}

	return &Manager{
		backend: backend,
		svc:     svc,
		now:     time.Now,
	}, nil
}

// BackendType 返回当前使用的存储后端类型。
func (m *Manager) BackendType() string {
	return m.backend
}

// NewObjectKey 按分类规则生成对象 key，文件名为随机 UUID，并保留原始文件名扩展名。
func (m *Manager) NewObjectKey(category string, userID uint, originalFilename string) (string, error) {
	return buildObjectKey(category, userID, originalFilename, m.now())
}

// Upload 上传对象内容。
func (m *Manager) Upload(ctx context.Context, objectKey string, data []byte, contentType string) error {
	if strings.TrimSpace(objectKey) == "" {
		return ErrEmptyObjectKey
	}
	if len(data) == 0 {
		return ErrEmptyObjectData
	}
	return m.svc.Upload(ctx, objectKey, data, contentType)
}

// Download 下载对象内容。
func (m *Manager) Download(ctx context.Context, objectKey string) ([]byte, error) {
	if strings.TrimSpace(objectKey) == "" {
		return nil, ErrEmptyObjectKey
	}
	return m.svc.Download(ctx, objectKey)
}

// GetSize 返回对象大小（字节）。
func (m *Manager) GetSize(ctx context.Context, objectKey string) (int64, error) {
	if strings.TrimSpace(objectKey) == "" {
		return 0, ErrEmptyObjectKey
	}
	return m.svc.GetSize(ctx, objectKey)
}

// Delete 删除对象。
func (m *Manager) Delete(ctx context.Context, objectKey string) error {
	if strings.TrimSpace(objectKey) == "" {
		return ErrEmptyObjectKey
	}
	return m.svc.Delete(ctx, objectKey)
}

// URL 返回对象访问地址。
func (m *Manager) URL(ctx context.Context, objectKey string) (string, error) {
	if strings.TrimSpace(objectKey) == "" {
		return "", ErrEmptyObjectKey
	}
	return m.svc.URL(ctx, objectKey)
}
