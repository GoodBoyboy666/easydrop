package repo

import (
	"context"

	"easydrop/internal/pkg/listing"

	"gorm.io/gorm"
)

var repoListBounds = listing.Bounds{
	DefaultPage: 1,
	DefaultSize: 20,
	MaxSize:     100,
}

// ListOptions 列表分页参数。
type ListOptions struct {
	Limit  int
	Offset int
	Order  string
}

func normalizeListOptions(opts ListOptions, defaultOrder string) ListOptions {
	opts.Limit, opts.Offset = repoListBounds.NormalizeLimitOffset(opts.Limit, opts.Offset)
	opts.Order = listing.OrderRules{Default: defaultOrder}.Normalize(opts.Order)
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
