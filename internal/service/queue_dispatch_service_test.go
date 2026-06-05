package service

import (
	"path/filepath"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
)

func TestQueueDispatchSkipsUnsupportedThumbnailFormats(t *testing.T) {
	db := newQueueDispatchTestDB(t)
	dispatch := NewQueueDispatchService(
		repository.NewQueueSettingRepository(db),
		repository.NewQueueJobRepository(db),
	)

	if err := dispatch.EnqueueFilePostProcess(1, 2, "vector.svg", "files/vector.svg", "local"); err != nil {
		t.Fatalf("enqueue svg post process: %v", err)
	}

	assertQueueJobCount(t, db, queue.JobTypeFileMetadata, 1)
	assertQueueJobCount(t, db, queue.JobTypeFileThumbnail, 0)

	if err := dispatch.EnqueueFilePostProcess(5, 6, "modern.webp", "files/modern.webp", "local"); err != nil {
		t.Fatalf("enqueue webp post process: %v", err)
	}

	assertQueueJobCount(t, db, queue.JobTypeFileMetadata, 2)
	assertQueueJobCount(t, db, queue.JobTypeFileThumbnail, 0)

	if err := dispatch.EnqueueFilePostProcess(3, 4, "photo.jpg", "files/photo.jpg", "local"); err != nil {
		t.Fatalf("enqueue jpg post process: %v", err)
	}

	assertQueueJobCount(t, db, queue.JobTypeFileMetadata, 3)
	assertQueueJobCount(t, db, queue.JobTypeFileThumbnail, 1)
}

func newQueueDispatchTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "queue-dispatch.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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

func assertQueueJobCount(t *testing.T, db *gorm.DB, jobType string, want int64) {
	t.Helper()

	var got int64
	if err := db.Model(&model.QueueJob{}).Where("job_type = ?", jobType).Count(&got).Error; err != nil {
		t.Fatalf("count %s jobs: %v", jobType, err)
	}
	if got != want {
		t.Fatalf("%s job count = %d, want %d", jobType, got, want)
	}
}
