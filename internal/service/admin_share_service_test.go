package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type adminShareTestEnv struct {
	db      *gorm.DB
	service AdminShareService
	user    model.User
}

func newAdminShareTestEnv(t *testing.T) *adminShareTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-shares.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserFile{}, &model.Share{}, &model.ShareFile{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	user := model.User{Username: "Sincerity", Email: "sincerity@example.com", Password: "x", Role: "admin", Enabled: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return &adminShareTestEnv{
		db:      db,
		user:    user,
		service: NewAdminShareService(db, repository.NewShareRepository(db), "https://pan.example.com"),
	}
}

func (e *adminShareTestEnv) createShare(t *testing.T, token, fileName string, expiresAt *time.Time, downloads int, access int, maxDownloads *int, accessHash string) model.Share {
	t.Helper()
	file := model.UserFile{UserID: e.user.ID, FileName: fileName}
	if err := e.db.Create(&file).Error; err != nil {
		t.Fatalf("create user file: %v", err)
	}
	share := model.Share{
		UserID:         e.user.ID,
		ShareToken:     token,
		AccessCodeHash: accessHash,
		ExpiresAt:      expiresAt,
		MaxDownloads:   maxDownloads,
		DownloadCount:  downloads,
		AccessCount:    access,
	}
	if err := e.db.Create(&share).Error; err != nil {
		t.Fatalf("create share: %v", err)
	}
	if err := e.db.Create(&model.ShareFile{ShareID: share.ID, FileID: file.ID}).Error; err != nil {
		t.Fatalf("create share file: %v", err)
	}
	return share
}

func TestAdminShareListFiltersAndUnavailableStatus(t *testing.T) {
	env := newAdminShareTestEnv(t)
	now := time.Now()
	soon := now.Add(48 * time.Hour)
	later := now.Add(10 * 24 * time.Hour)
	past := now.Add(-time.Hour)
	limit := 2

	active := env.createShare(t, "active-token", "active-report.pdf", &later, 1, 2, nil, "")
	env.createShare(t, "protected-token", "private.txt", &later, 0, 1, nil, "hashed")
	env.createShare(t, "soon-token", "soon.jpg", &soon, 3, 4, nil, "")
	expired := env.createShare(t, "expired-token", "old.jpg", &past, 4, 5, nil, "")
	limitReached := env.createShare(t, "limit-token", "limited.zip", &later, 2, 6, &limit, "")

	items, _, err := env.service.List(context.Background(), &AdminShareListQuery{Keyword: "report"})
	if err != nil {
		t.Fatalf("keyword list: %v", err)
	}
	if len(items) != 1 || items[0].ShareID != active.ID {
		t.Fatalf("keyword results = %#v, want active report only", items)
	}

	minDownloads := 3
	items, _, err = env.service.List(context.Background(), &AdminShareListQuery{MinDownloads: &minDownloads})
	if err != nil {
		t.Fatalf("min downloads list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("min downloads result count = %d, want 2", len(items))
	}

	days := 3
	items, _, err = env.service.List(context.Background(), &AdminShareListQuery{ExpiringWithinDays: &days})
	if err != nil {
		t.Fatalf("expiring list: %v", err)
	}
	if len(items) != 1 || items[0].ShareToken != "soon-token" {
		t.Fatalf("expiring results = %#v, want soon-token only", items)
	}

	items, _, err = env.service.List(context.Background(), &AdminShareListQuery{Status: "active"})
	if err != nil {
		t.Fatalf("active list: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("active result count = %d, want 3", len(items))
	}

	items, _, err = env.service.List(context.Background(), &AdminShareListQuery{Status: "unavailable"})
	if err != nil {
		t.Fatalf("unavailable list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unavailable result count = %d, want 2", len(items))
	}

	items, _, err = env.service.List(context.Background(), &AdminShareListQuery{Status: "download_limit_reached"})
	if err != nil {
		t.Fatalf("download limit list: %v", err)
	}
	if len(items) != 1 || items[0].ShareID != limitReached.ID || items[0].StatusReason != "download_limit_reached" || !items[0].IsUnavailable {
		t.Fatalf("limit result = %#v, want reached/unavailable", items)
	}

	items, _, err = env.service.List(context.Background(), &AdminShareListQuery{Status: "expired"})
	if err != nil {
		t.Fatalf("expired list: %v", err)
	}
	if len(items) != 1 || items[0].ShareID != expired.ID || items[0].StatusReason != "time_expired" {
		t.Fatalf("expired result = %#v, want time expired", items)
	}
}

func TestAdminShareMetrics(t *testing.T) {
	env := newAdminShareTestEnv(t)
	now := time.Now()
	soon := now.Add(48 * time.Hour)
	later := now.Add(10 * 24 * time.Hour)
	past := now.Add(-time.Hour)
	limit := 2

	env.createShare(t, "active-token", "active.pdf", &later, 1, 2, nil, "")
	env.createShare(t, "protected-token", "private.txt", &later, 0, 1, nil, "hashed")
	env.createShare(t, "soon-token", "soon.jpg", &soon, 3, 4, nil, "")
	env.createShare(t, "expired-token", "old.jpg", &past, 4, 5, nil, "")
	env.createShare(t, "limit-token", "limited.zip", &later, 2, 6, &limit, "")

	metrics, err := env.service.Metrics(context.Background(), 3)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if metrics.TotalShares != 5 || metrics.ActiveShares != 3 || metrics.ExpiredShares != 1 || metrics.ProtectedShares != 1 {
		t.Fatalf("metrics counts = %#v", metrics)
	}
	if metrics.TotalAccessCount != 18 || metrics.TotalDownloadCount != 10 {
		t.Fatalf("metrics sums = access %d download %d, want 18/10", metrics.TotalAccessCount, metrics.TotalDownloadCount)
	}
	if metrics.ExpiringSoonCount != 1 || metrics.DownloadLimitReachedCount != 1 {
		t.Fatalf("metrics expiry/limit = %#v", metrics)
	}
}

func TestAdminShareDeleteAndBatchDelete(t *testing.T) {
	env := newAdminShareTestEnv(t)
	first := env.createShare(t, "first-token", "first.txt", nil, 0, 0, nil, "")
	second := env.createShare(t, "second-token", "second.txt", nil, 0, 0, nil, "")
	third := env.createShare(t, "third-token", "third.txt", nil, 0, 0, nil, "")

	if err := env.service.Delete(context.Background(), first.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := env.service.BatchDelete(context.Background(), []uint{second.ID, third.ID}); err != nil {
		t.Fatalf("batch delete: %v", err)
	}

	var shareCount int64
	if err := env.db.Model(&model.Share{}).Count(&shareCount).Error; err != nil {
		t.Fatalf("count shares: %v", err)
	}
	var linkCount int64
	if err := env.db.Model(&model.ShareFile{}).Count(&linkCount).Error; err != nil {
		t.Fatalf("count share files: %v", err)
	}
	if shareCount != 0 || linkCount != 0 {
		t.Fatalf("remaining shares=%d links=%d, want 0/0", shareCount, linkCount)
	}
}
