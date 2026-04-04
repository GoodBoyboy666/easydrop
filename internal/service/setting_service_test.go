package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"easydrop/internal/consts"
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
	}).Create(&model.Setting{Key: consts.SiteNameSettingKey, Value: "EasyDrop", Category: "site", Sensitive: false, Public: true}).Error; err != nil {
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
	}).Create(&model.Setting{Key: consts.StorageQuotaSettingKey, Value: "10737418240", Category: "storage", Public: true}).Error; err != nil {
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
	hasUploadMaxBody := false
	for i := range result.Items {
		if result.Items[i].Key == consts.StorageQuotaSettingKey {
			hasQuota = true
		}
		if result.Items[i].Key == "site.secret" {
			hasSecret = true
		}
		if result.Items[i].Key == consts.AttachmentAllowedExtensionsSettingKey {
			hasAttachmentExtensions = true
		}
		if result.Items[i].Key == consts.UploadMaxRequestBodySettingKey {
			hasUploadMaxBody = true
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
	if hasUploadMaxBody {
		t.Fatal("expected upload max body setting to be excluded")
	}
}

func TestSettingServiceSeedsRequireEmailVerificationSetting(t *testing.T) {
	svc, _ := newTestSettingService(t)

	value, found, err := svc.GetValue(context.Background(), consts.AuthRequireEmailVerificationSettingKey)
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if !found {
		t.Fatal("expected auth.require_email_verification to exist")
	}
	if value != "false" {
		t.Fatalf("expected default false, got %q", value)
	}
}

func TestSettingServiceSeedsPublicSiteFaviconSetting(t *testing.T) {
	svc, db := newTestSettingService(t)

	value, found, err := svc.GetValue(context.Background(), consts.SiteFaviconSettingKey)
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if !found {
		t.Fatal("expected site.favicon to exist")
	}
	if value != "" {
		t.Fatalf("expected default empty favicon, got %q", value)
	}

	var setting model.Setting
	if err := db.Where("key = ?", consts.SiteFaviconSettingKey).First(&setting).Error; err != nil {
		t.Fatalf("load site.favicon failed: %v", err)
	}
	if setting.Category != "site" {
		t.Fatalf("expected category site, got %q", setting.Category)
	}
	if !setting.Public {
		t.Fatal("expected site.favicon to be public")
	}
}

func TestSettingServiceSeedsAttachmentExtensionsFromPresetDefaults(t *testing.T) {
	svc, _ := newTestSettingService(t)

	value, found, err := svc.GetValue(context.Background(), consts.AttachmentAllowedExtensionsSettingKey)
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if !found {
		t.Fatal("expected attachment extension setting to exist")
	}
	if value != consts.DefaultAttachmentAllowedExtensionsSettingValue {
		t.Fatalf("expected preset attachment extensions %q, got %q", consts.DefaultAttachmentAllowedExtensionsSettingValue, value)
	}
}

func TestNewSettingServiceSyncsBlankAttachmentExtensionsValueToPresetDefaults(t *testing.T) {
	db := newSettingServiceTestDB(t)
	if err := db.Create(&model.Setting{
		Key:      consts.AttachmentAllowedExtensionsSettingKey,
		Value:    "",
		Desc:     "旧描述",
		Category: "storage",
		Public:   false,
	}).Error; err != nil {
		t.Fatalf("seed attachment extension setting failed: %v", err)
	}

	kvCache, err := cache.NewCache(nil)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}

	settingRepo := repo.NewSettingRepo(db)
	if _, err := NewSettingService(db, settingRepo, kvCache); err != nil {
		t.Fatalf("NewSettingService returned error: %v", err)
	}

	var setting model.Setting
	if err := db.Where("key = ?", consts.AttachmentAllowedExtensionsSettingKey).First(&setting).Error; err != nil {
		t.Fatalf("load attachment extension setting failed: %v", err)
	}
	if setting.Value != consts.DefaultAttachmentAllowedExtensionsSettingValue {
		t.Fatalf("expected preset attachment extensions %q, got %q", consts.DefaultAttachmentAllowedExtensionsSettingValue, setting.Value)
	}
}

func TestNewSettingServiceDeletesGhostSettings(t *testing.T) {
	db := newSettingServiceTestDB(t)
	if err := db.Create(&model.Setting{
		Key:      "ghost.setting",
		Value:    "ghost",
		Desc:     "幽灵配置",
		Category: "ghost",
		Public:   true,
	}).Error; err != nil {
		t.Fatalf("seed ghost setting failed: %v", err)
	}

	kvCache, err := cache.NewCache(nil)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}

	settingRepo := repo.NewSettingRepo(db)
	if _, err := NewSettingService(db, settingRepo, kvCache); err != nil {
		t.Fatalf("NewSettingService returned error: %v", err)
	}

	var total int64
	if err := db.Model(&model.Setting{}).Where("key = ?", "ghost.setting").Count(&total).Error; err != nil {
		t.Fatalf("count ghost setting failed: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected ghost setting deleted, got %d", total)
	}
}

func TestNewSettingServiceSyncsDefaultSettingMetadataWithoutOverwritingValue(t *testing.T) {
	db := newSettingServiceTestDB(t)
	if err := db.Create(&model.Setting{
		Key:       consts.SiteNameSettingKey,
		Value:     "Custom Site Name",
		Desc:      "旧描述",
		Category:  "legacy",
		Sensitive: true,
		Public:    false,
	}).Error; err != nil {
		t.Fatalf("seed site.name failed: %v", err)
	}

	kvCache, err := cache.NewCache(nil)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}

	settingRepo := repo.NewSettingRepo(db)
	if _, err := NewSettingService(db, settingRepo, kvCache); err != nil {
		t.Fatalf("NewSettingService returned error: %v", err)
	}

	var setting model.Setting
	if err := db.Where("key = ?", consts.SiteNameSettingKey).First(&setting).Error; err != nil {
		t.Fatalf("load site.name failed: %v", err)
	}
	if setting.Value != "Custom Site Name" {
		t.Fatalf("expected custom value preserved, got %q", setting.Value)
	}
	if setting.Desc != "站点名称" {
		t.Fatalf("expected desc synced, got %q", setting.Desc)
	}
	if setting.Category != "site" {
		t.Fatalf("expected category synced, got %q", setting.Category)
	}
	if !setting.Public {
		t.Fatal("expected public synced to true")
	}
	if !setting.Sensitive {
		t.Fatal("expected sensitive flag preserved")
	}
}

func newSettingServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "settings.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return db
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
