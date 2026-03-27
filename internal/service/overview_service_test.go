package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"easydrop/internal/repo"
)

type mockOverviewRepo struct {
	getSnapshotFn func(ctx context.Context, since time.Time, until time.Time) (*repo.OverviewSnapshot, error)
}

func (m *mockOverviewRepo) GetSnapshot(ctx context.Context, since time.Time, until time.Time) (*repo.OverviewSnapshot, error) {
	if m.getSnapshotFn == nil {
		return nil, nil
	}
	return m.getSnapshotFn(ctx, since, until)
}

func TestAdminOverviewServiceGetBuildsRecentActivity(t *testing.T) {
	fixedNow := time.Date(2026, 3, 27, 15, 4, 5, 0, time.UTC)
	var gotSince time.Time
	var gotUntil time.Time

	svc := &adminOverviewService{
		overviewRepo: &mockOverviewRepo{
			getSnapshotFn: func(_ context.Context, since time.Time, until time.Time) (*repo.OverviewSnapshot, error) {
				gotSince = since
				gotUntil = until
				return &repo.OverviewSnapshot{
					UserTotal:       12,
					PostTotal:       34,
					CommentTotal:    56,
					AttachmentTotal: 78,
					PostDaily: []repo.OverviewDailyCount{
						{Day: "2026-03-21", Total: 2},
						{Day: "2026-03-24", Total: 5},
						{Day: "2026-03-27", Total: 1},
					},
					CommentDaily: []repo.OverviewDailyCount{
						{Day: "2026-03-22", Total: 3},
						{Day: "2026-03-24", Total: 7},
					},
				}, nil
			},
		},
		now: func() time.Time { return fixedNow },
	}

	result, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if gotSince.Format(time.DateOnly) != "2026-03-21" {
		t.Fatalf("expected since 2026-03-21, got %s", gotSince.Format(time.DateOnly))
	}
	if gotUntil.Format(time.DateOnly) != "2026-03-28" {
		t.Fatalf("expected until 2026-03-28, got %s", gotUntil.Format(time.DateOnly))
	}
	if result.Totals.Users != 12 || result.Totals.Posts != 34 || result.Totals.Comments != 56 || result.Totals.Attachments != 78 {
		t.Fatalf("unexpected totals: %+v", result.Totals)
	}
	if len(result.RecentActivity) != adminOverviewRecentDays {
		t.Fatalf("expected %d recent activity items, got %d", adminOverviewRecentDays, len(result.RecentActivity))
	}
	if result.RecentActivity[0].Date != "2026-03-21" || result.RecentActivity[0].Posts != 2 || result.RecentActivity[0].Comments != 0 {
		t.Fatalf("unexpected first recent activity item: %+v", result.RecentActivity[0])
	}
	if result.RecentActivity[3].Date != "2026-03-24" || result.RecentActivity[3].Posts != 5 || result.RecentActivity[3].Comments != 7 {
		t.Fatalf("unexpected middle recent activity item: %+v", result.RecentActivity[3])
	}
	if result.RecentActivity[6].Date != "2026-03-27" || result.RecentActivity[6].Posts != 1 || result.RecentActivity[6].Comments != 0 {
		t.Fatalf("unexpected last recent activity item: %+v", result.RecentActivity[6])
	}
}

func TestAdminOverviewServiceGetReturnsInternalOnRepoError(t *testing.T) {
	svc := &adminOverviewService{
		overviewRepo: &mockOverviewRepo{
			getSnapshotFn: func(_ context.Context, _ time.Time, _ time.Time) (*repo.OverviewSnapshot, error) {
				return nil, errors.New("db down")
			},
		},
		now: func() time.Time { return time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC) },
	}

	_, err := svc.Get(context.Background())
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
}
