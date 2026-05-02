package repo

import "github.com/google/wire"

var RepositorySet = wire.NewSet(
	NewAttachmentRepo,
	NewCommentRepo,
	NewInitRepo,
	NewOverviewRepo,
	NewPasskeyRepo,
	NewPostRepo,
	NewSettingRepo,
	NewTagRepo,
	NewUserRepo,
)
