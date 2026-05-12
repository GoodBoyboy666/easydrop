package repo

import "github.com/google/wire"

var RepositorySet = wire.NewSet(
	NewAttachmentRepo,
	NewCommentRepo,
	NewInitRepo,
	NewOAuthBindRepo,
	NewOverviewRepo,
	NewPasskeyRepo,
	NewPostRepo,
	NewSettingRepo,
	NewTagRepo,
	NewUserRepo,
)
