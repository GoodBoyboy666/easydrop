package storage

import (
	"errors"
	"path"
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

func buildObjectKey(category string, userID uint, now time.Time) (string, error) {
	id := uuid.NewString()

	switch strings.TrimSpace(category) {
	case CategoryFile:
		return path.Join(
			CategoryFile,
			now.Format("2006"),
			now.Format("01"),
			now.Format("02"),
			id,
		), nil
	case CategoryAvatar:
		if userID == 0 {
			return "", ErrInvalidUserID
		}
		return path.Join(CategoryAvatar, strconv.FormatUint(uint64(userID), 10), id), nil
	default:
		return "", ErrInvalidCategory
	}
}
