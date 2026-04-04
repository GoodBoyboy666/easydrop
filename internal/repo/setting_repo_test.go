package repo

import (
	"context"
	"testing"

	"easydrop/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSettingRepoSyncDefaultsDeletesGhostAndSyncsMetadata(t *testing.T) {
	db := newSettingRepoTestDB(t)
	repo := NewSettingRepo(db)

	if err := db.Create(&model.Setting{
		Key:      "ghost.setting",
		Value:    "ghost",
		Desc:     "幽灵配置",
		Category: "ghost",
		Public:   true,
	}).Error; err != nil {
		t.Fatalf("seed ghost setting failed: %v", err)
	}
	if err := db.Create(&model.Setting{
		Key:       "site.name",
		Value:     "Custom Site Name",
		Desc:      "旧描述",
		Category:  "legacy",
		Sensitive: true,
		Public:    false,
	}).Error; err != nil {
		t.Fatalf("seed site.name failed: %v", err)
	}

	defaults := []model.Setting{
		{
			Key:      "site.name",
			Value:    "EasyDrop",
			Desc:     "站点名称",
			Category: "site",
			Public:   true,
		},
		{
			Key:      "system.initialized",
			Value:    "false",
			Desc:     "系统已初始化",
			Category: "system",
			Public:   false,
		},
	}

	if err := repo.SyncDefaults(context.Background(), defaults); err != nil {
		t.Fatalf("SyncDefaults returned error: %v", err)
	}

	var ghostCount int64
	if err := db.Model(&model.Setting{}).Where("key = ?", "ghost.setting").Count(&ghostCount).Error; err != nil {
		t.Fatalf("count ghost setting failed: %v", err)
	}
	if ghostCount != 0 {
		t.Fatalf("expected ghost setting deleted, got %d", ghostCount)
	}

	var siteName model.Setting
	if err := db.Where("key = ?", "site.name").First(&siteName).Error; err != nil {
		t.Fatalf("load site.name failed: %v", err)
	}
	if siteName.Value != "Custom Site Name" {
		t.Fatalf("expected value preserved, got %q", siteName.Value)
	}
	if siteName.Desc != "站点名称" {
		t.Fatalf("expected desc synced, got %q", siteName.Desc)
	}
	if siteName.Category != "site" {
		t.Fatalf("expected category synced, got %q", siteName.Category)
	}
	if !siteName.Public {
		t.Fatal("expected public synced to true")
	}
	if !siteName.Sensitive {
		t.Fatal("expected sensitive preserved")
	}

	var initialized model.Setting
	if err := db.Where("key = ?", "system.initialized").First(&initialized).Error; err != nil {
		t.Fatalf("load system.initialized failed: %v", err)
	}
	if initialized.Value != "false" {
		t.Fatalf("expected missing default inserted, got %q", initialized.Value)
	}
	if initialized.Desc != "系统已初始化" {
		t.Fatalf("expected desc inserted, got %q", initialized.Desc)
	}
}

func newSettingRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
}
