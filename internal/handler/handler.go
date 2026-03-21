package handler

import "github.com/google/wire"

var HandlerSet = wire.NewSet(
	NewAttachmentAdminHandler,
	NewAttachmentHandler,
	NewAuthHandler,
	NewCommentAdminHandler,
	NewCommentHandler,
	NewPostAdminHandler,
	NewSettingAdminHandler,
	NewUserAdminHandler,
	NewUserHandler,
)
