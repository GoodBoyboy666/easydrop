package service

import (
	"context"
	"log"
	"time"

	"easydrop/internal/dto"
	"easydrop/internal/repo"
)

const adminOverviewRecentDays = 7

// AdminOverviewService 定义后台概览聚合服务。
type AdminOverviewService interface {
	// Get 返回后台统计总览与趋势数据。
	Get(ctx context.Context) (*dto.AdminOverviewResult, error)
}

type adminOverviewService struct {
	overviewRepo repo.OverviewRepo
	now          func() time.Time
}

// NewAdminOverviewService 创建后台概览聚合服务。
func NewAdminOverviewService(overviewRepo repo.OverviewRepo) AdminOverviewService {
	return &adminOverviewService{
		overviewRepo: overviewRepo,
		now:          time.Now,
	}
}

// Get 聚合统计总览数据，并构建最近 N 天趋势序列。
func (s *adminOverviewService) Get(ctx context.Context) (*dto.AdminOverviewResult, error) {
	// 计算统计窗口：从最近 N 天的起始日到次日零点。
	today := startOfDay(s.now())
	start := today.AddDate(0, 0, -(adminOverviewRecentDays - 1))
	until := today.AddDate(0, 0, 1)

	// 拉取一次性聚合快照。
	snapshot, err := s.overviewRepo.GetSnapshot(ctx, start, until)
	if err != nil {
		log.Printf("查询后台概览聚合失败: %v", err)
		return nil, ErrInternal
	}
	if snapshot == nil {
		snapshot = &repo.OverviewSnapshot{}
	}

	// 将日统计映射为日期索引，再按固定天数补齐趋势序列。
	postDaily := toOverviewDailyMap(snapshot.PostDaily)
	commentDaily := toOverviewDailyMap(snapshot.CommentDaily)
	recentActivity := make([]dto.AdminOverviewTrendItem, 0, adminOverviewRecentDays)
	for i := 0; i < adminOverviewRecentDays; i++ {
		current := start.AddDate(0, 0, i)
		day := current.Format(time.DateOnly)
		recentActivity = append(recentActivity, dto.AdminOverviewTrendItem{
			Date:     day,
			Posts:    postDaily[day],
			Comments: commentDaily[day],
		})
	}

	return &dto.AdminOverviewResult{
		Totals: dto.AdminOverviewTotals{
			Users:       snapshot.UserTotal,
			Posts:       snapshot.PostTotal,
			Comments:    snapshot.CommentTotal,
			Attachments: snapshot.AttachmentTotal,
		},
		RecentActivity: recentActivity,
	}, nil
}

// startOfDay 将时间归一化到当天零点。
func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

// toOverviewDailyMap 将按天计数列表转为日期到计数的映射。
func toOverviewDailyMap(items []repo.OverviewDailyCount) map[string]int64 {
	if len(items) == 0 {
		return map[string]int64{}
	}

	result := make(map[string]int64, len(items))
	for _, item := range items {
		if item.Day == "" {
			continue
		}
		result[string(item.Day)] = item.Total
	}
	return result
}
