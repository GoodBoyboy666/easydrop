package service

import "easydrop/internal/pkg/listing"

var serviceListBounds = listing.Bounds{
	DefaultPage: 1,
	DefaultSize: 20,
	MaxSize:     100,
}

var postListOrderRules = listing.OrderRules{
	Default: "pin desc, created_at desc",
	Allowed: map[string]string{
		"created_at_asc":  "pin desc, created_at asc",
		"created_at_desc": "pin desc, created_at desc",
	},
}

var tagListOrderRules = listing.OrderRules{
	Default: "created_at desc",
	Allowed: map[string]string{
		"hot":             "hot_desc",
		"hot_desc":        "hot_desc",
		"hot_asc":         "hot_asc",
		"name_asc":        "name asc",
		"name_desc":       "name desc",
		"created_at_asc":  "created_at asc",
		"created_at_desc": "created_at desc",
	},
}

var userListOrderRules = listing.OrderRules{
	Default: "created_at desc",
	Allowed: map[string]string{
		"created_at_asc":  "created_at asc",
		"created_at_desc": "created_at desc",
		"username_asc":    "username asc",
		"username_desc":   "username desc",
		"status_asc":      "status asc",
		"status_desc":     "status desc",
	},
}

var commentListOrderRules = listing.OrderRules{
	Default: "created_at desc",
	Allowed: map[string]string{
		"created_at_asc":  "created_at asc",
		"created_at_desc": "created_at desc",
	},
}

var attachmentListOrderRules = listing.OrderRules{
	Default: "created_at desc",
	Allowed: map[string]string{
		"created_at_asc":  "created_at asc",
		"created_at_desc": "created_at desc",
	},
}

var settingListOrderRules = listing.OrderRules{
	Default: "key asc",
	Allowed: map[string]string{
		"key_asc":  "key asc",
		"key_desc": "key desc",
	},
}

// normalizeServiceListPage 规范化 service 层分页页码，页码从 1 开始。
func normalizeServiceListPage(page int) int {
	return serviceListBounds.NormalizePage(page)
}

// normalizeServiceListSize 规范化 service 层分页大小，并限制最大值。
func normalizeServiceListSize(size int) int {
	return serviceListBounds.NormalizeSize(size)
}

// normalizeServiceListPageSize 同时规范化分页页码和每页大小。
func normalizeServiceListPageSize(page, size int) (int, int) {
	return serviceListBounds.NormalizePageSize(page, size)
}

// pageSizeToOffset 将 page/size 转换为 offset，并避免整数溢出。
func pageSizeToOffset(page, size int) int {
	return serviceListBounds.OffsetFromPage(page, size)
}

// normalizePostListOrder 将说说列表排序参数映射为允许的排序表达式。
func normalizePostListOrder(order string) string {
	return postListOrderRules.Normalize(order)
}

// normalizeTagListOrder 将标签列表排序参数映射为允许的排序表达式。
func normalizeTagListOrder(order string) string {
	return tagListOrderRules.Normalize(order)
}

// normalizeUserListOrder 将用户列表排序参数映射为允许的排序表达式。
func normalizeUserListOrder(order string) string {
	return userListOrderRules.Normalize(order)
}

// normalizeCommentListOrder 将评论列表排序参数映射为允许的排序表达式。
func normalizeCommentListOrder(order string) string {
	return commentListOrderRules.Normalize(order)
}

// normalizeAttachmentListOrder 将附件列表排序参数映射为允许的排序表达式。
func normalizeAttachmentListOrder(order string) string {
	return attachmentListOrderRules.Normalize(order)
}

// normalizeSettingListOrder 将配置列表排序参数映射为允许的排序表达式。
func normalizeSettingListOrder(order string) string {
	return settingListOrderRules.Normalize(order)
}
