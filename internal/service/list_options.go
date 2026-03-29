package service

import "strings"

const (
	defaultServiceListPage = 1
	defaultServiceListSize = 20
	maxServiceListSize     = 100
)

// normalizeServiceListPage 规范化 service 层分页页码，页码从 1 开始。
func normalizeServiceListPage(page int) int {
	if page <= 0 {
		return defaultServiceListPage
	}
	return page
}

// normalizeServiceListSize 规范化 service 层分页大小，并限制最大值。
func normalizeServiceListSize(size int) int {
	if size <= 0 {
		return defaultServiceListSize
	}
	if size > maxServiceListSize {
		return maxServiceListSize
	}
	return size
}

func normalizeServiceListPageSize(page, size int) (int, int) {
	return normalizeServiceListPage(page), normalizeServiceListSize(size)
}

func maxIntValue() int {
	return int(^uint(0) >> 1)
}

// pageSizeToOffset 将 page/size 转换为 offset，并避免整数溢出。
func pageSizeToOffset(page, size int) int {
	normalizedPage, normalizedSize := normalizeServiceListPageSize(page, size)
	deltaPage := int64(normalizedPage - 1)
	size64 := int64(normalizedSize)
	maxInt64 := int64(maxIntValue())

	if deltaPage > 0 && size64 > 0 && deltaPage > maxInt64/size64 {
		return maxIntValue()
	}

	return int(deltaPage * size64)
}

// normalizePostListOrder 将说说列表排序参数映射为允许的排序表达式。
func normalizePostListOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "created_at_asc":
		return "pin desc, created_at asc"
	case "created_at_desc", "":
		return "pin desc, created_at desc"
	default:
		return "pin desc, created_at desc"
	}
}

// normalizeTagListOrder 将标签列表排序参数映射为允许的排序表达式。
func normalizeTagListOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "hot_desc", "hot":
		return "hot_desc"
	case "hot_asc":
		return "hot_asc"
	case "name_asc":
		return "name asc"
	case "name_desc":
		return "name desc"
	case "created_at_asc":
		return "created_at asc"
	case "created_at_desc", "":
		return "created_at desc"
	default:
		return "created_at desc"
	}
}

// normalizeUserListOrder 将用户列表排序参数映射为允许的排序表达式。
func normalizeUserListOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "created_at_asc":
		return "created_at asc"
	case "username_asc":
		return "username asc"
	case "username_desc":
		return "username desc"
	case "status_asc":
		return "status asc"
	case "status_desc":
		return "status desc"
	case "created_at_desc", "":
		return "created_at desc"
	default:
		return "created_at desc"
	}
}

// normalizeCommentListOrder 将评论列表排序参数映射为允许的排序表达式。
func normalizeCommentListOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "created_at_desc":
		return "created_at desc"
	case "created_at_asc":
		return "created_at asc"
	case "":
		return "created_at desc"
	default:
		return "created_at desc"
	}
}

// normalizeAttachmentListOrder 将附件列表排序参数映射为允许的排序表达式。
func normalizeAttachmentListOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "created_at_asc":
		return "created_at asc"
	case "created_at_desc", "":
		return "created_at desc"
	default:
		return "created_at desc"
	}
}
