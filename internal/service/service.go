package service

import "github.com/google/wire"

// ServiceSet 汇总 service 层依赖提供器。
var ServiceSet = wire.NewSet(NewSettingService, NewEmailService, NewAuthService, NewAttachmentService, NewCommentService, NewPostService, NewTagService, NewUserService, NewCaptchaConfigService, NewInitService, NewAdminOverviewService)
