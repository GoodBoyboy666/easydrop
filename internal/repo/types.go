package repo

import (
	"context"
)

type ListOptions struct {
	Limit  int
	Offset int
	Order  string
}

func withContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
