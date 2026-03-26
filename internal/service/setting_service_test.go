package service

import (
	"context"
	"strings"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"
	"easydrop/internal/repo"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestNewSettingServiceRequiresCache(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	settingRepo := repo.NewSettingRepo(db)
	_, err = NewSettingService(db, settingRepo, nil)
	if err == nil {
		t.Fatal("expected error when cache is nil")
	}
	if !strings.Contains(err.Error(), "cache is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSettingServiceUpdateItemAndGetValue(t *testing.T) {
	svc, db := newTestSettingService(t)
	ctx := context.Background()
	v3 := "v3"

	if err := upsertSetting(db, "custom.key", "v1"); err != nil {
		t.Fatalf("upsert setting failed: %v", err)
	}

	value, found, err := svc.GetValue(ctx, "custom.key")
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if !found || value != "v1" {
		t.Fatalf("unexpected setting value: found=%v value=%s", found, value)
	}

	if err := upsertSetting(db, "custom.key", "v2"); err != nil {
		t.Fatalf("update setting failed: %v", err)
	}

	value, found, err = svc.GetValue(ctx, "custom.key")
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if !found || value != "v1" {
		t.Fatalf("expected cached old value, got: found=%v value=%s", found, value)
	}

	if err := svc.UpdateItem(ctx, dto.SettingUpdateInput{Key: "custom.key", Value: &v3}); err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}

	value, found, err = svc.GetValue(ctx, "custom.key")
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if !found || value != "v3" {
		t.Fatalf("expected updated value, got: found=%v value=%s", found, value)
	}
}

func TestSettingServiceGetPublicItems(t *testing.T) {
	svc, db := newTestSettingService(t)
	ctx := context.Background()

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "category", "sensitive", "public"}),
	}).Create(&model.Setting{Key: "site.name", Value: "EasyDrop", Category: "site", Sensitive: false, Public: true}).Error; err != nil {
		t.Fatalf("upsert site.name failed: %v", err)
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "category", "sensitive", "public"}),
	}).Create(&model.Setting{Key: "site.secret", Value: "hidden", Category: "site", Sensitive: true, Public: false}).Error; err != nil {
		t.Fatalf("upsert site.secret failed: %v", err)
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "category", "public"}),
	}).Create(&model.Setting{Key: "storage.quota", Value: "10737418240", Category: "storage", Public: true}).Error; err != nil {
		t.Fatalf("upsert storage.quota failed: %v", err)
	}

	result, err := svc.GetPublicItems(ctx)
	if err != nil {
		t.Fatalf("GetPublicItems returned error: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected public items")
	}
	hasQuota := false
	hasSecret := false
	hasAttachmentExtensions := false
	for i := range result.Items {
		if result.Items[i].Key == "storage.quota" {
			hasQuota = true
		}
		if result.Items[i].Key == "site.secret" {
			hasSecret = true
		}
		if result.Items[i].Key == attachmentAllowedExtensionsSettingKey {
			hasAttachmentExtensions = true
		}
	}
	if !hasQuota {
		t.Fatal("expected storage.quota in public items")
	}
	if hasSecret {
		t.Fatal("expected non-public setting to be excluded")
	}
	if hasAttachmentExtensions {
		t.Fatal("expected attachment extension whitelist setting to be excluded")
	}
}

func newTestSettingService(t *testing.T) (SettingService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	kvCache, err := cache.NewCache(nil)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}

	settingRepo := repo.NewSettingRepo(db)
	svc, err := NewSettingService(db, settingRepo, kvCache)
	if err != nil {
		t.Fatalf("NewSettingService returned error: %v", err)
	}
	return svc, db
}

func upsertSetting(db *gorm.DB, key, value string) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&model.Setting{Key: key, Value: value}).Error
}
