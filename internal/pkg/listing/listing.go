package listing

import "strings"

const (
	defaultPage = 1
	defaultSize = 20
)

// Bounds 定义分页边界。
type Bounds struct {
	DefaultPage int
	DefaultSize int
	MaxSize     int
}

// OrderRules 定义排序白名单与默认值。
type OrderRules struct {
	Default string
	Allowed map[string]string
}

func (b Bounds) NormalizePage(page int) int {
	if page <= 0 {
		return b.defaultPage()
	}
	return page
}

func (b Bounds) NormalizeSize(size int) int {
	defaultSize := b.defaultSize()
	if size <= 0 {
		return defaultSize
	}

	maxSize := b.maxSize(defaultSize)
	if size > maxSize {
		return maxSize
	}

	return size
}

func (b Bounds) NormalizePageSize(page, size int) (int, int) {
	return b.NormalizePage(page), b.NormalizeSize(size)
}

func (b Bounds) NormalizeLimitOffset(limit, offset int) (int, int) {
	limit = b.NormalizeSize(limit)
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// OffsetFromPage 将 page/size 转换为 offset，并避免整数溢出。
func (b Bounds) OffsetFromPage(page, size int) int {
	normalizedPage, normalizedSize := b.NormalizePageSize(page, size)
	deltaPage := int64(normalizedPage - 1)
	size64 := int64(normalizedSize)
	maxInt64 := int64(maxIntValue())

	if deltaPage > 0 && size64 > 0 && deltaPage > maxInt64/size64 {
		return maxIntValue()
	}

	return int(deltaPage * size64)
}

func (r OrderRules) Normalize(order string) string {
	normalizedInput := strings.ToLower(strings.TrimSpace(order))
	if normalizedInput == "" {
		return strings.TrimSpace(r.Default)
	}

	if len(r.Allowed) == 0 {
		return normalizedInput
	}

	if normalized, ok := r.Allowed[normalizedInput]; ok {
		return normalized
	}

	return strings.TrimSpace(r.Default)
}

func (b Bounds) defaultPage() int {
	if b.DefaultPage > 0 {
		return b.DefaultPage
	}
	return defaultPage
}

func (b Bounds) defaultSize() int {
	if b.DefaultSize > 0 {
		return b.DefaultSize
	}
	return defaultSize
}

func (b Bounds) maxSize(fallback int) int {
	if b.MaxSize > 0 {
		return b.MaxSize
	}
	return fallback
}

func maxIntValue() int {
	return int(^uint(0) >> 1)
}
