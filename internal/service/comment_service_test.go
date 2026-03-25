package service

import (
	"context"
	"testing"

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

