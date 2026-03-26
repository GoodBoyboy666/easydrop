package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidLocalObjectKey = errors.New("local object key 非法")

type localStorage struct {
	basePath  string
	urlPrefix string
}

// NewLocalStorage 创建本地存储实现。
func NewLocalStorage(cfg LocalConfig) (Backend, error) {
	basePath := strings.TrimSpace(cfg.BasePath)
	if basePath == "" {
		return nil, ErrInvalidLocalBasePath
	}

	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("创建 local base path 失败: %w", err)
	}

	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("解析 local base path 失败: %w", err)
	}

	return &localStorage{
		basePath:  absBasePath,
		urlPrefix: normalizeLocalURLPrefix(cfg.URLPrefix),
	}, nil
}

func (s *localStorage) Upload(_ context.Context, objectKey string, data []byte, contentType string) error {
	return s.UploadStream(context.Background(), objectKey, bytes.NewReader(data), int64(len(data)), contentType)
}

func (s *localStorage) UploadStream(_ context.Context, objectKey string, reader io.Reader, _ int64, _ string) error {
	fullPath, err := s.resolvePath(objectKey)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func (s *localStorage) Download(_ context.Context, objectKey string) ([]byte, error) {
	fullPath, err := s.resolvePath(objectKey)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return data, nil
}

func (s *localStorage) GetSize(_ context.Context, objectKey string) (int64, error) {
	fullPath, err := s.resolvePath(objectKey)
	if err != nil {
		return 0, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("读取文件信息失败: %w", err)
	}

	return info.Size(), nil
}

func (s *localStorage) Delete(_ context.Context, objectKey string) error {
	fullPath, err := s.resolvePath(objectKey)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}

	s.cleanupEmptyDirs(filepath.Dir(fullPath))
	return nil
}

func (s *localStorage) URL(_ context.Context, objectKey string) (string, error) {
	return strings.TrimRight(s.urlPrefix, "/") + "/" + strings.TrimLeft(objectKey, "/"), nil
}

func normalizeLocalURLPrefix(value string) string {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return "/api"
	}
	return "/api/" + trimmed
}

func (s *localStorage) resolvePath(objectKey string) (string, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return "", ErrEmptyObjectKey
	}

	cleanKey := filepath.Clean(filepath.FromSlash(key))
	if cleanKey == "." || cleanKey == ".." || filepath.IsAbs(cleanKey) {
		return "", ErrInvalidLocalObjectKey
	}
	if strings.HasPrefix(cleanKey, ".."+string(filepath.Separator)) {
		return "", ErrInvalidLocalObjectKey
	}

	fullPath := filepath.Join(s.basePath, cleanKey)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	if absPath != s.basePath && !strings.HasPrefix(absPath, s.basePath+string(filepath.Separator)) {
		return "", ErrInvalidLocalObjectKey
	}

	return absPath, nil
}

func (s *localStorage) cleanupEmptyDirs(start string) {
	current := start
	for {
		if current == s.basePath || current == string(filepath.Separator) {
			return
		}

		entries, err := os.ReadDir(current)
		if err != nil || len(entries) > 0 {
			return
		}

		_ = os.Remove(current)
		next := filepath.Dir(current)
		if next == current {
			return
		}
		current = next
	}
}
