package handler

import "github.com/google/wire"

var HandlerSet = wire.NewSet(
	NewAttachmentAdminHandler,
	NewAttachmentHandler,
	NewAuthHandler,
	NewCaptchaHandler,
	NewCommentAdminHandler,
	NewCommentHandler,
	NewPostAdminHandler,
	NewSettingAdminHandler,
	NewUserAdminHandler,
	NewUserHandler,
)
