package handler

import "github.com/google/wire"

var HandlerSet = wire.NewSet(
	NewAttachmentAdminHandler,
	NewAttachmentHandler,
	NewAuthHandler,
	NewCaptchaHandler,
	NewCommentAdminHandler,
	NewCommentHandler,
	NewInitHandler,
	NewPostAdminHandler,
	NewPostHandler,
	NewSettingAdminHandler,
	NewUserAdminHandler,
	NewUserHandler,
)
