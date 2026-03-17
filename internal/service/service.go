package service

import "github.com/google/wire"

// ServiceSet 汇总 service 层依赖提供器。
var ServiceSet = wire.NewSet(NewAuthService, NewCommentService, NewPostService, NewTagService, NewUserService)
