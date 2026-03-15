package repo

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

const defaultLimit = 20

// ListOptions 列表分页参数。
type ListOptions struct {
	Limit  int
	Offset int
	Order  string
}

func normalizeListOptions(opts ListOptions, defaultOrder string) ListOptions {
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}
	opts.Order = strings.TrimSpace(opts.Order)
	if opts.Order == "" {
		opts.Order = defaultOrder
	}
	return opts
}

func applyListOptions(db *gorm.DB, opts ListOptions, defaultOrder string) *gorm.DB {
	opts = normalizeListOptions(opts, defaultOrder)
	return db.Order(opts.Order).Limit(opts.Limit).Offset(opts.Offset)
}

func withContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
