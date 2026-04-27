package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"easydrop/internal/consts"
	"easydrop/internal/model"
	"easydrop/internal/repo"

	"github.com/gorilla/feeds"
)

const feedItemLimit = 20

var (
	ErrFeedGenerateFailed = errors.New("生成Feed失败")
)

type FeedService interface {
	GetRSS(ctx context.Context) (string, error)
	GetAtom(ctx context.Context) (string, error)
}

type feedService struct {
	postRepo       repo.PostRepo
	settingService SettingService
}

func NewFeedService(postRepo repo.PostRepo, settingService SettingService) FeedService {
	return &feedService{
		postRepo:       postRepo,
		settingService: settingService,
	}
}

func (s *feedService) GetRSS(ctx context.Context) (string, error) {
	feed, err := s.buildFeed(ctx)
	if err != nil {
		return "", err
	}
	rss, err := feed.ToRss()
	if err != nil {
		return "", ErrFeedGenerateFailed
	}
	return rss, nil
}

func (s *feedService) GetAtom(ctx context.Context) (string, error) {
	feed, err := s.buildFeed(ctx)
	if err != nil {
		return "", err
	}
	atom, err := feed.ToAtom()
	if err != nil {
		return "", ErrFeedGenerateFailed
	}
	return atom, nil
}

func (s *feedService) buildFeed(ctx context.Context) (*feeds.Feed, error) {
	siteName := s.getSetting(ctx, consts.SiteNameSettingKey, "EasyDrop")
	siteDesc := s.getSetting(ctx, consts.SiteDescriptionSettingKey, "")
	siteURL := s.getSetting(ctx, consts.SiteURLSettingKey, "http://localhost:8080")

	hideFalse := false
	posts, _, err := s.postRepo.List(ctx, repo.PostFilter{
		Hide: &hideFalse,
	}, repo.ListOptions{
		Limit:  feedItemLimit,
		Offset: 0,
		Order:  "pin desc, created_at desc",
	})
	if err != nil {
		return nil, ErrInternal
	}

	feed := &feeds.Feed{
		Title:       siteName,
		Link:        &feeds.Link{Href: siteURL},
		Description: siteDesc,
	}

	for i := range posts {
		item := s.postToFeedItem(&posts[i], siteURL)
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}

func (s *feedService) getSetting(ctx context.Context, key, fallback string) string {
	if s.settingService == nil {
		return fallback
	}
	value, found, err := s.settingService.GetValue(ctx, key)
	if err != nil || !found {
		return fallback
	}
	return value
}

func (s *feedService) postToFeedItem(post *model.Post, siteURL string) *feeds.Item {
	title := firstLine(post.Content)
	link := fmt.Sprintf("%s/posts/%d", strings.TrimRight(siteURL, "/"), post.ID)

	return &feeds.Item{
		Title:       title,
		Link:        &feeds.Link{Href: link},
		Description: post.Content,
		Author:      &feeds.Author{Name: post.User.Nickname},
		Created:     post.CreatedAt,
		Id:          link,
	}
}

func firstLine(content string) string {
	idx := strings.IndexByte(content, '\n')
	if idx < 0 {
		return content
	}
	return content[:idx]
}
