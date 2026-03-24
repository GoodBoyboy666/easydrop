package service

import "strings"

const (
	defaultServiceListLimit = 20
	maxServiceListLimit     = 100
)

// normalizeServiceListLimit 规范化 service 层列表分页大小，并限制最大值。
func normalizeServiceListLimit(limit int) int {
	if limit <= 0 {
		return defaultServiceListLimit
	}
	if limit > maxServiceListLimit {
		return maxServiceListLimit
	}
	return limit
}

// normalizeServiceListOffset 规范化 service 层列表偏移量，避免负数。
func normalizeServiceListOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
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
