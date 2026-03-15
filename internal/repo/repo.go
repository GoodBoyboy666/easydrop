package repo

import "github.com/google/wire"

var RepositorySet = wire.NewSet(NewAttachmentRepo, NewPostRepo, NewSettingRepo, NewTagRepo, NewUserRepo)
