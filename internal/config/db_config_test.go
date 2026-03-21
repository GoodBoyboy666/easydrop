package config

import (
	"context"
	"strings"
	"testing"

	"easydrop/internal/model"
	"easydrop/internal/pkg/cache"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestNewDBConfigRequiresCache(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	_, err = NewDBConfig(db, nil)
	if err == nil {
		t.Fatal("expected error when cache is nil")
	}
	if !strings.Contains(err.Error(), "cache is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDBConfigClearCacheKey(t *testing.T) {
	dbCfg, db := newTestDBConfig(t)
	ctx := context.Background()

	if err := upsertSetting(db, "custom.key", "v1"); err != nil {
		t.Fatalf("upsert setting failed: %v", err)
	}

	setting, err := dbCfg.Get(ctx, "custom.key")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if setting.Value != "v1" {
		t.Fatalf("unexpected setting value: %s", setting.Value)
	}

	if err := upsertSetting(db, "custom.key", "v2"); err != nil {
		t.Fatalf("update setting failed: %v", err)
	}

	setting, err = dbCfg.Get(ctx, "custom.key")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if setting.Value != "v1" {
		t.Fatalf("expected cached old value, got: %s", setting.Value)
	}

	if err := dbCfg.ClearCacheKey(ctx, "custom.key"); err != nil {
		t.Fatalf("ClearCacheKey returned error: %v", err)
	}

	setting, err = dbCfg.Get(ctx, "custom.key")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if setting.Value != "v2" {
		t.Fatalf("expected refreshed value, got: %s", setting.Value)
	}
}

func TestDBConfigClearCache(t *testing.T) {
	dbCfg, db := newTestDBConfig(t)
	ctx := context.Background()

	if err := upsertSetting(db, "custom.a", "a1"); err != nil {
		t.Fatalf("upsert custom.a failed: %v", err)
	}
	if err := upsertSetting(db, "custom.b", "b1"); err != nil {
		t.Fatalf("upsert custom.b failed: %v", err)
	}

	if _, err := dbCfg.Get(ctx, "custom.a"); err != nil {
		t.Fatalf("Get custom.a returned error: %v", err)
	}
	if _, err := dbCfg.Get(ctx, "custom.b"); err != nil {
		t.Fatalf("Get custom.b returned error: %v", err)
	}

	if err := upsertSetting(db, "custom.a", "a2"); err != nil {
		t.Fatalf("update custom.a failed: %v", err)
	}
	if err := upsertSetting(db, "custom.b", "b2"); err != nil {
		t.Fatalf("update custom.b failed: %v", err)
	}

	if err := dbCfg.ClearCache(ctx); err != nil {
		t.Fatalf("ClearCache returned error: %v", err)
	}

	settingA, err := dbCfg.Get(ctx, "custom.a")
	if err != nil {
		t.Fatalf("Get custom.a returned error: %v", err)
	}
	if settingA.Value != "a2" {
		t.Fatalf("unexpected custom.a value: %s", settingA.Value)
	}

	settingB, err := dbCfg.Get(ctx, "custom.b")
	if err != nil {
		t.Fatalf("Get custom.b returned error: %v", err)
	}
	if settingB.Value != "b2" {
		t.Fatalf("unexpected custom.b value: %s", settingB.Value)
	}
}

func newTestDBConfig(t *testing.T) (DBConfig, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	kvCache, err := cache.NewCache(nil)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}

	dbCfg, err := NewDBConfig(db, kvCache)
	if err != nil {
		t.Fatalf("NewDBConfig returned error: %v", err)
	}
	return dbCfg, db
}

func upsertSetting(db *gorm.DB, key, value string) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&model.Setting{Key: key, Value: value}).Error
}
