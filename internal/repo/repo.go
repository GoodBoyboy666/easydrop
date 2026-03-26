package repo

import "github.com/google/wire"

var RepositorySet = wire.NewSet(
	NewAttachmentRepo,
	NewCommentRepo,
	NewInitRepo,
	NewPostRepo,
	NewSettingRepo,
	NewTagRepo,
	NewUserRepo,
)
