package service

import (
	"context"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/repo"

	"gorm.io/gorm"
)

type mockCommentRepoForDelete struct {
	comments            map[uint]*model.Comment
	deleteIDs           []uint
	deleteRootWithChild []uint
}

func (m *mockCommentRepoForDelete) Create(_ context.Context, _ *model.Comment) error {
	return nil
}

func (m *mockCommentRepoForDelete) GetByID(_ context.Context, id uint) (*model.Comment, error) {
	comment, ok := m.comments[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *comment
	return &copy, nil
}

func (m *mockCommentRepoForDelete) Update(_ context.Context, _ *model.Comment) error {
	return nil
}

func (m *mockCommentRepoForDelete) Delete(_ context.Context, id uint) error {
	m.deleteIDs = append(m.deleteIDs, id)
	return nil
}

func (m *mockCommentRepoForDelete) DeleteByPostID(_ context.Context, _ uint) error {
	return nil
}

func (m *mockCommentRepoForDelete) DeleteRootWithChildren(_ context.Context, rootID uint) error {
	m.deleteRootWithChild = append(m.deleteRootWithChild, rootID)
	return nil
}

func (m *mockCommentRepoForDelete) List(_ context.Context, _ repo.CommentFilter, _ repo.ListOptions) ([]model.Comment, int64, error) {
	return nil, 0, nil
}

func (m *mockCommentRepoForDelete) ListPublic(_ context.Context, _ bool, _ repo.ListOptions) ([]model.Comment, int64, error) {
	return nil, 0, nil
}

type mockCommentRepoForVisibility struct {
	createCalled            bool
	createdComment          *model.Comment
	getByIDComments         map[uint]*model.Comment
	listFilter              repo.CommentFilter
	listCalled              bool
	listPublicCalled        bool
	listPublicIncludeHidden bool
}

func (m *mockCommentRepoForVisibility) Create(_ context.Context, comment *model.Comment) error {
	m.createCalled = true
	comment.ID = 99
	clone := *comment
	m.createdComment = &clone
	return nil
}

func (m *mockCommentRepoForVisibility) GetByID(_ context.Context, id uint) (*model.Comment, error) {
	comment, ok := m.getByIDComments[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *comment
	return &clone, nil
}

func (m *mockCommentRepoForVisibility) Update(_ context.Context, _ *model.Comment) error {
	return nil
}

func (m *mockCommentRepoForVisibility) Delete(_ context.Context, _ uint) error {
	return nil
}

func (m *mockCommentRepoForVisibility) DeleteByPostID(_ context.Context, _ uint) error {
	return nil
}

func (m *mockCommentRepoForVisibility) DeleteRootWithChildren(_ context.Context, _ uint) error {
	return nil
}

func (m *mockCommentRepoForVisibility) List(_ context.Context, filter repo.CommentFilter, _ repo.ListOptions) ([]model.Comment, int64, error) {
	m.listCalled = true
	m.listFilter = filter
	return []model.Comment{
		{
			ID:     1,
			PostID: *filter.PostID,
			User:   model.User{ID: 7, Nickname: "tester"},
		},
	}, 1, nil
}

func (m *mockCommentRepoForVisibility) ListPublic(_ context.Context, includeHidden bool, _ repo.ListOptions) ([]model.Comment, int64, error) {
	m.listPublicCalled = true
	m.listPublicIncludeHidden = includeHidden
	return []model.Comment{
		{
			ID:     1,
			PostID: 5,
			User:   model.User{ID: 7, Nickname: "tester"},
		},
	}, 1, nil
}

type mockPostRepoForVisibility struct {
	posts map[uint]*model.Post
}

func (m *mockPostRepoForVisibility) Create(_ context.Context, _ *model.Post) error {
	return nil
}

func (m *mockPostRepoForVisibility) GetByID(_ context.Context, id uint) (*model.Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *post
	return &clone, nil
}

func (m *mockPostRepoForVisibility) Update(_ context.Context, _ *model.Post) error {
	return nil
}

func (m *mockPostRepoForVisibility) Delete(_ context.Context, _ uint) error {
	return nil
}

func (m *mockPostRepoForVisibility) List(_ context.Context, _ repo.PostFilter, _ repo.ListOptions) ([]model.Post, int64, error) {
	return nil, 0, nil
}

type mockUserRepoForVisibility struct {
	users map[uint]*model.User
}

func (m *mockUserRepoForVisibility) Create(_ context.Context, _ *model.User) error {
	return nil
}

func (m *mockUserRepoForVisibility) GetByID(_ context.Context, id uint) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *user
	return &clone, nil
}

func (m *mockUserRepoForVisibility) GetByUsername(_ context.Context, _ string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepoForVisibility) GetByEmail(_ context.Context, _ string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepoForVisibility) GetByUsernameOrEmail(_ context.Context, _ string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepoForVisibility) UpdateAvatarWithStorageUsedTx(_ context.Context, _ uint, _ *string, _ int64, _ int64) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepoForVisibility) Update(_ context.Context, _ *model.User) error {
	return nil
}

func (m *mockUserRepoForVisibility) Delete(_ context.Context, _ uint) error {
	return nil
}

func (m *mockUserRepoForVisibility) List(_ context.Context, _ repo.UserFilter, _ repo.ListOptions) ([]model.User, int64, error) {
	return nil, 0, nil
}

func TestCommentServiceDeleteRootCommentDoesNotCascade(t *testing.T) {
	repo := &mockCommentRepoForDelete{
		comments: map[uint]*model.Comment{
			1: {ID: 1, PostID: 10, UserID: 100},
		},
	}
	svc := &commentService{commentRepo: repo}

	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if len(repo.deleteIDs) != 1 || repo.deleteIDs[0] != 1 {
		t.Fatalf("expected Delete to be called once with id=1, got %v", repo.deleteIDs)
	}
	if len(repo.deleteRootWithChild) != 0 {
		t.Fatalf("expected DeleteRootWithChildren not to be called, got %v", repo.deleteRootWithChild)
	}
}

func TestCommentServiceDeleteChildComment(t *testing.T) {
	parentID := uint(1)
	repo := &mockCommentRepoForDelete{
		comments: map[uint]*model.Comment{
			2: {ID: 2, PostID: 10, UserID: 101, ParentID: &parentID, RootID: &parentID},
		},
	}
	svc := &commentService{commentRepo: repo}

	if err := svc.Delete(context.Background(), 2); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if len(repo.deleteIDs) != 1 || repo.deleteIDs[0] != 2 {
		t.Fatalf("expected Delete to be called once with id=2, got %v", repo.deleteIDs)
	}
	if len(repo.deleteRootWithChild) != 0 {
		t.Fatalf("expected DeleteRootWithChildren not to be called, got %v", repo.deleteRootWithChild)
	}
}

func TestCommentServiceCreateRejectsHiddenPostForNormalUser(t *testing.T) {
	commentRepo := &mockCommentRepoForVisibility{}
	postRepo := &mockPostRepoForVisibility{
		posts: map[uint]*model.Post{
			9: {ID: 9, Hide: true},
		},
	}
	userRepo := &mockUserRepoForVisibility{
		users: map[uint]*model.User{
			7: {ID: 7},
		},
	}
	svc := &commentService{commentRepo: commentRepo, postRepo: postRepo, userRepo: userRepo}

	_, err := svc.Create(context.Background(), dto.CommentCreateInput{
		PostID:  9,
		UserID:  7,
		Content: "hello",
	})

	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
	if commentRepo.createCalled {
		t.Fatal("expected create to be blocked for hidden post")
	}
}

func TestCommentServiceCreateAllowsHiddenPostForAdmin(t *testing.T) {
	commentRepo := &mockCommentRepoForVisibility{
		getByIDComments: map[uint]*model.Comment{
			99: {
				ID:      99,
				PostID:  9,
				UserID:  7,
				Content: "hello",
				User:    model.User{ID: 7, Nickname: "tester"},
			},
		},
	}
	postRepo := &mockPostRepoForVisibility{
		posts: map[uint]*model.Post{
			9: {ID: 9, Hide: true},
		},
	}
	userRepo := &mockUserRepoForVisibility{
		users: map[uint]*model.User{
			7: {ID: 7},
		},
	}
	svc := &commentService{commentRepo: commentRepo, postRepo: postRepo, userRepo: userRepo}

	result, err := svc.Create(context.Background(), dto.CommentCreateInput{
		PostID:        9,
		UserID:        7,
		CanViewHidden: true,
		Content:       "hello",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || result.ID != 99 {
		t.Fatalf("expected created comment id 99, got %+v", result)
	}
}

func TestCommentServiceListByPostRejectsHiddenPostForNormalUser(t *testing.T) {
	commentRepo := &mockCommentRepoForVisibility{}
	postRepo := &mockPostRepoForVisibility{
		posts: map[uint]*model.Post{
			9: {ID: 9, Hide: true},
		},
	}
	svc := &commentService{commentRepo: commentRepo, postRepo: postRepo}

	_, err := svc.ListByPost(context.Background(), dto.CommentListInput{PostID: 9})

	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
	if commentRepo.listCalled {
		t.Fatal("expected hidden post comments to be blocked")
	}
}

func TestCommentServiceListByPostAllowsHiddenPostForAdmin(t *testing.T) {
	commentRepo := &mockCommentRepoForVisibility{}
	postRepo := &mockPostRepoForVisibility{
		posts: map[uint]*model.Post{
			9: {ID: 9, Hide: true},
		},
	}
	svc := &commentService{commentRepo: commentRepo, postRepo: postRepo}

	result, err := svc.ListByPost(context.Background(), dto.CommentListInput{
		PostID:        9,
		CanViewHidden: true,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || result.Total != 1 {
		t.Fatalf("expected one visible comment, got %+v", result)
	}
	if !commentRepo.listCalled {
		t.Fatal("expected comment list to be queried")
	}
}

func TestCommentServiceListPublicFiltersHiddenPostsForNormalUser(t *testing.T) {
	commentRepo := &mockCommentRepoForVisibility{}
	svc := &commentService{commentRepo: commentRepo}

	result, err := svc.ListPublic(context.Background(), dto.CommentPublicListInput{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || result.Total != 1 {
		t.Fatalf("expected one public comment, got %+v", result)
	}
	if !commentRepo.listPublicCalled {
		t.Fatal("expected public comment repo query to run")
	}
	if commentRepo.listPublicIncludeHidden {
		t.Fatal("expected normal public list to exclude hidden posts")
	}
}

func TestCommentServiceListPublicAllowsHiddenPostsForAdmin(t *testing.T) {
	commentRepo := &mockCommentRepoForVisibility{}
	svc := &commentService{commentRepo: commentRepo}

	_, err := svc.ListPublic(context.Background(), dto.CommentPublicListInput{CanViewHidden: true})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !commentRepo.listPublicIncludeHidden {
		t.Fatal("expected admin public list to include hidden posts")
	}
}
