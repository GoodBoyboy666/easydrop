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
	NewOverviewAdminHandler,
	NewPostAdminHandler,
	NewPostHandler,
	NewSettingAdminHandler,
	NewTagHandler,
	NewUserAdminHandler,
	NewUserHandler,
)
