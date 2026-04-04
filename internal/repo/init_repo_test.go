package repo

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"easydrop/internal/consts"
	"easydrop/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestInitRepoInitializeSuccess(t *testing.T) {
	db := newInitRepoTestDB(t)
	repo := NewInitRepo(db)

	err := repo.Initialize(context.Background(), SystemInitInput{
		AdminUser: model.User{
			Username:      "admin",
			Nickname:      "管理员",
			Email:         "admin@example.com",
			Password:      "hashed",
			Admin:         true,
			Status:        1,
			EmailVerified: true,
		},
		Settings: []SettingValueInput{
			{Key: consts.SiteNameSettingKey, Value: "EasyDrop"},
			{Key: consts.SiteURLSettingKey, Value: "http://localhost:8080"},
			{Key: consts.SiteAnnouncementSettingKey, Value: "hello"},
			{Key: consts.SiteAllowRegisterSettingKey, Value: "false"},
			{Key: consts.SystemInitializedSettingKey, Value: "true"},
		},
	})
	if err != nil {
		t.Fatalf("Initialize error: %v", err)
	}

	var user model.User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		t.Fatalf("expected admin user created: %v", err)
	}
	if !user.Admin || !user.EmailVerified {
		t.Fatalf("unexpected user flags: %+v", user)
	}

	assertSettingValue(t, db, consts.SiteNameSettingKey, "EasyDrop")
	assertSettingValue(t, db, consts.SiteAllowRegisterSettingKey, "false")
	assertSettingValue(t, db, consts.SystemInitializedSettingKey, "true")
}

func TestInitRepoInitializeRollbackOnUserCreateFailure(t *testing.T) {
	db := newInitRepoTestDB(t)
	if err := db.Create(&model.User{
		Username: "admin",
		Nickname: "old",
		Email:    "admin@example.com",
		Password: "hashed",
		Status:   1,
	}).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	repo := NewInitRepo(db)
	err := repo.Initialize(context.Background(), SystemInitInput{
		AdminUser: model.User{
			Username:      "admin",
			Nickname:      "管理员",
			Email:         "admin@example.com",
			Password:      "hashed",
			Admin:         true,
			Status:        1,
			EmailVerified: true,
		},
		Settings: []SettingValueInput{
			{Key: consts.SiteNameSettingKey, Value: "Changed"},
			{Key: consts.SiteURLSettingKey, Value: "http://changed.example.com"},
			{Key: consts.SiteAnnouncementSettingKey, Value: "changed"},
			{Key: consts.SiteAllowRegisterSettingKey, Value: "false"},
			{Key: consts.SystemInitializedSettingKey, Value: "true"},
		},
	})
	if err == nil {
		t.Fatal("expected Initialize to fail on duplicate user")
	}

	assertSettingValue(t, db, consts.SiteNameSettingKey, "EasyDrop")
	assertSettingValue(t, db, consts.SystemInitializedSettingKey, "false")

	var total int64
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Count(&total).Error; err != nil {
		t.Fatalf("count users failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected only seeded user to remain, got %d", total)
	}
}

func TestInitRepoInitializeRejectsAlreadyInitialized(t *testing.T) {
	db := newInitRepoTestDB(t)
	if err := db.Model(&model.Setting{}).Where("key = ?", consts.SystemInitializedSettingKey).Update("value", "true").Error; err != nil {
		t.Fatalf("seed init flag failed: %v", err)
	}

	repo := NewInitRepo(db)
	err := repo.Initialize(context.Background(), SystemInitInput{
		AdminUser: model.User{
			Username: "admin",
			Nickname: "管理员",
			Email:    "admin@example.com",
			Password: "hashed",
			Admin:    true,
			Status:   1,
		},
	})
	if !errors.Is(err, ErrInitAlreadyInitialized) {
		t.Fatalf("expected ErrInitAlreadyInitialized, got %v", err)
	}
}

func newInitRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "init.db")
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
	if err := db.AutoMigrate(&model.User{}, &model.Setting{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	for _, setting := range defaultInitSettings() {
		if err := db.Create(&setting).Error; err != nil {
			t.Fatalf("seed setting %s failed: %v", setting.Key, err)
		}
	}

	return db
}

func defaultInitSettings() []model.Setting {
	return []model.Setting{
		{Key: consts.SiteNameSettingKey, Value: "EasyDrop", Category: "site", Public: true},
		{Key: consts.SiteURLSettingKey, Value: "http://localhost:8080", Category: "site", Public: true},
		{Key: consts.SiteAnnouncementSettingKey, Value: "", Category: "site", Public: true},
		{Key: consts.SiteAllowRegisterSettingKey, Value: "true", Category: "site", Public: true},
		{Key: consts.SystemInitializedSettingKey, Value: "false", Category: "system", Public: false},
	}
}

func assertSettingValue(t *testing.T, db *gorm.DB, key, expected string) {
	t.Helper()

	var setting model.Setting
	if err := db.Where("key = ?", key).First(&setting).Error; err != nil {
		t.Fatalf("load setting %s failed: %v", key, err)
	}
	if setting.Value != expected {
		t.Fatalf("expected setting %s=%q, got %q", key, expected, setting.Value)
	}
}
