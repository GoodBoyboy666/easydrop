package handler

import "github.com/google/wire"

var HandlerSet = wire.NewSet(
	NewErrorResponder,
	NewAttachmentAdminHandler,
	NewAttachmentHandler,
	NewAuthHandler,
	NewCaptchaHandler,
	NewCommentAdminHandler,
	NewCommentHandler,
	NewFeedHandler,
	NewInitHandler,
	NewOverviewAdminHandler,
	NewPasskeyHandler,
	NewPostAdminHandler,
	NewPostHandler,
	NewSettingAdminHandler,
	NewTagHandler,
	NewUserAdminHandler,
	NewUserHandler,
)
