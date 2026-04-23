package service

import "easydrop/internal/pkg/listing"

const (
	defaultCreatedAtDescOrder = "created_at_desc"
	defaultCreatedAtAscOrder  = "created_at_asc"
	defaultKeyAscOrder        = "key_asc"
	tagHotOrderAscSQL         = "COUNT(post_tags.post_id) ASC"
	tagHotOrderDescSQL        = "COUNT(post_tags.post_id) DESC"
)

var postListOrderMap = listing.SQLOrderMap{
	"created_at_asc":  "pin desc, created_at asc",
	"created_at_desc": "pin desc, created_at desc",
}

var tagListOrderMap = listing.SQLOrderMap{
	"hot":             tagHotOrderDescSQL,
	"hot_asc":         tagHotOrderAscSQL,
	"hot_desc":        tagHotOrderDescSQL,
	"name_asc":        "name asc",
	"name_desc":       "name desc",
	"created_at_asc":  "created_at asc",
	"created_at_desc": "created_at desc",
}

var userListOrderMap = listing.SQLOrderMap{
	"created_at_asc":  "created_at asc",
	"created_at_desc": "created_at desc",
	"username_asc":    "username asc",
	"username_desc":   "username desc",
	"status_asc":      "status asc",
	"status_desc":     "status desc",
}

var commentListOrderMap = listing.SQLOrderMap{
	"created_at_asc":  "created_at asc",
	"created_at_desc": "created_at desc",
}

var publicCommentListOrderMap = listing.SQLOrderMap{
	"created_at_asc":  "comments.created_at asc",
	"created_at_desc": "comments.created_at desc",
}

var attachmentListOrderMap = listing.SQLOrderMap{
	"created_at_asc":  "created_at asc",
	"created_at_desc": "created_at desc",
}

var settingListOrderMap = listing.SQLOrderMap{
	"key_asc":  "key asc",
	"key_desc": "key desc",
}

func resolvePostListOrder(order string) string {
	return postListOrderMap.Resolve(order, defaultCreatedAtDescOrder)
}

func resolveTagListOrder(order string) string {
	return tagListOrderMap.Resolve(order, defaultCreatedAtDescOrder)
}

func resolveUserListOrder(order string) string {
	return userListOrderMap.Resolve(order, defaultCreatedAtDescOrder)
}

func resolveCommentListOrder(order string) string {
	return commentListOrderMap.Resolve(order, defaultCreatedAtDescOrder)
}

func resolvePublicCommentListOrder(order string) string {
	return publicCommentListOrderMap.Resolve(order, defaultCreatedAtDescOrder)
}

func resolveAttachmentListOrder(order string) string {
	return attachmentListOrderMap.Resolve(order, defaultCreatedAtDescOrder)
}

func resolveSettingListOrder(order string) string {
	return settingListOrderMap.Resolve(order, defaultKeyAscOrder)
}
