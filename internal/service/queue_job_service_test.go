package service

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
)

func TestQueueJobPayloadMarksArchiveJobsAsUnifiedRunnerJobs(t *testing.T) {
	db := newQueueJobServiceTestDB(t)
	repo := repository.NewQueueJobRepository(db)
	created, err := repo.EnqueueIfAbsent(&model.QueueJob{
		QueueKey:       string(queue.KeyIO),
		JobType:        "archive.create",
		ResourceType:   "archive_package_session",
		ResourceID:     "session-a",
		DedupeKey:      "io:archive:session-a",
		Payload:        `{"execution_mode":"unified_runner"}`,
		NodeCapability: string(NodeCapabilityCreateArchive),
		Status:         model.QueueJobStatusPending,
		MaxAttempts:    0,
		ScheduledAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("create archive queue job: %v", err)
	}

	service := NewQueueJobService(repo, nil)
	payload, err := service.Get(created.ID)
	if err != nil {
		t.Fatalf("get queue job: %v", err)
	}
	if payload.ExecutionMode != "unified_runner" {
		t.Fatalf("execution_mode = %q, want unified_runner", payload.ExecutionMode)
	}
	if !strings.Contains(payload.ExecutionNote, "IO 队列 runner") {
		t.Fatalf("execution_note = %q, want unified archive runner note", payload.ExecutionNote)
	}
}

func TestQueueJobPayloadMarksUnifiedRunnerJobs(t *testing.T) {
	db := newQueueJobServiceTestDB(t)
	repo := repository.NewQueueJobRepository(db)
	created, err := repo.EnqueueIfAbsent(&model.QueueJob{
		QueueKey:     string(queue.KeyIO),
		JobType:      queue.JobTypeMultipartCleanup,
		ResourceType: queue.ResourceTypeMultipart,
		ResourceID:   "upload-a",
		DedupeKey:    "io:multipart:upload-a",
		Payload:      `{"upload_id":"upload-a"}`,
		Status:       model.QueueJobStatusPending,
		MaxAttempts:  3,
		ScheduledAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("create multipart queue job: %v", err)
	}

	service := NewQueueJobService(repo, nil)
	payload, err := service.Get(created.ID)
	if err != nil {
		t.Fatalf("get queue job: %v", err)
	}
	if payload.ExecutionMode != "unified_runner" {
		t.Fatalf("execution_mode = %q, want unified_runner", payload.ExecutionMode)
	}
}

func newQueueJobServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "queue-job-service.db")), &gorm.Config{
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
