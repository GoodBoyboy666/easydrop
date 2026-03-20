package storage

import (
	"errors"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	CategoryFile   = "file"
	CategoryAvatar = "avatar"
)

var (
	ErrInvalidCategory = errors.New("存储分类不合法")
	ErrInvalidUserID   = errors.New("avatar 分类需要有效 user id")
)

func buildObjectKey(category string, userID uint, originalFilename string, now time.Time) (string, error) {
	filename := buildObjectFilename(originalFilename)

	switch strings.TrimSpace(category) {
	case CategoryFile:
		return path.Join(
			CategoryFile,
			now.Format("2006"),
			now.Format("01"),
			now.Format("02"),
			filename,
		), nil
	case CategoryAvatar:
		if userID == 0 {
			return "", ErrInvalidUserID
		}
		return path.Join(CategoryAvatar, strconv.FormatUint(uint64(userID), 10), filename), nil
	default:
		return "", ErrInvalidCategory
	}
}

func buildObjectFilename(originalFilename string) string {
	id := uuid.NewString()
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filepath.Base(strings.TrimSpace(originalFilename)))))
	if ext == "." {
		ext = ""
	}
	return id + ext
}
