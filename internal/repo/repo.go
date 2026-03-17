package repo

import "github.com/google/wire"

var RepositorySet = wire.NewSet(
	NewAttachmentRepo,
	NewCommentRepo,
	NewPostRepo,
	NewSettingRepo,
	NewTagRepo,
	NewUserRepo,
)
