package repository

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
)

func TestQueueSettingRepositoryEnsureSchemaWithLegacyUniqueIndex(t *testing.T) {
	db := newQueueSettingDBForTest(t, "queue-settings-legacy.db")
	if err := createLegacyQueueSettingsTable(db, true); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}

	repo := NewQueueSettingRepository(db)
	if err := repo.EnsureSchema(); err != nil {
		t.Fatalf("ensure schema with legacy index: %v", err)
	}

	assertQueueKeyUniqueConstraint(t, db)
}

func TestQueueSettingRepositoryEnsureSchemaCreatesMissingQueueKeyUniqueIndex(t *testing.T) {
	db := newQueueSettingDBForTest(t, "queue-settings-missing-index.db")
	if err := createLegacyQueueSettingsTable(db, false); err != nil {
		t.Fatalf("create table without legacy index: %v", err)
	}

	repo := NewQueueSettingRepository(db)
	if err := repo.EnsureSchema(); err != nil {
		t.Fatalf("ensure schema without legacy index: %v", err)
	}

	assertQueueKeyUniqueConstraint(t, db)
}

func TestQueueSettingsLegacyUniqueDropMissingErrorIsIgnored(t *testing.T) {
	err := &mysql.MySQLError{
		Number:  1091,
		Message: "Can't DROP 'uni_queue_settings_queue_key'; check that column/key exists",
	}
	if !isQueueSettingsLegacyUniqueDropMissingError(err) {
		t.Fatalf("expected queue settings legacy drop error to be ignored")
	}

	other := &mysql.MySQLError{
		Number:  1091,
		Message: "Can't DROP 'some_other_index'; check that column/key exists",
	}
	if isQueueSettingsLegacyUniqueDropMissingError(other) {
		t.Fatalf("unexpectedly ignored unrelated drop error")
	}

	wrapped := "migrate queue settings failed: Error 1091 (42000): Can't DROP 'uni_queue_settings_queue_key'; check that column/key exists"
	if !isQueueSettingsLegacyUniqueDropMissingError(fmt.Errorf("%s", wrapped)) {
		t.Fatalf("expected wrapped queue settings legacy drop error to be ignored")
	}
}

func newQueueSettingDBForTest(t *testing.T, name string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sqlite sql db: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close sqlite db: %v", err)
		}
	})

	return db
}

func createLegacyQueueSettingsTable(db *gorm.DB, withLegacyIndex bool) error {
	if err := db.Exec(`
		CREATE TABLE queue_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at datetime,
			updated_at datetime,
			deleted_at datetime,
			queue_key varchar(32) NOT NULL,
			worker_num integer NOT NULL DEFAULT 5,
			max_execution integer NOT NULL DEFAULT 3600,
			backoff_factor integer NOT NULL DEFAULT 2,
			max_backoff integer NOT NULL DEFAULT 60,
			max_retry integer NOT NULL DEFAULT 0,
			retry_delay integer NOT NULL DEFAULT 0
		)`,
	).Error; err != nil {
		return err
	}
	if !withLegacyIndex {
		return nil
	}
	return db.Exec(`CREATE UNIQUE INDEX uni_queue_settings_queue_key ON queue_settings(queue_key)`).Error
}

func assertQueueKeyUniqueConstraint(t *testing.T, db *gorm.DB) {
	t.Helper()

	first := model.QueueSetting{QueueKey: "offline", WorkerNum: 1}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("create first queue setting: %v", err)
	}
	second := model.QueueSetting{QueueKey: "offline", WorkerNum: 2}
	err := db.Create(&second).Error
	if err == nil {
		t.Fatalf("expected duplicate queue_key to fail")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("duplicate queue_key error = %v, want unique constraint", err)
	}
}
