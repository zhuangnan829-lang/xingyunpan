package repository

import (
	"path/filepath"
	"testing"
	"time"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

func TestQueueJobRepositoryStoresNodeCapabilityAndFiltersByNode(t *testing.T) {
	repo := newQueueJobRepositoryForTest(t)
	now := time.Now()
	nodeOne := uint(11)
	nodeTwo := uint(22)

	_, err := repo.EnqueueIfAbsent(&model.QueueJob{
		QueueKey:         "offline",
		JobType:          "offline.download",
		ResourceType:     "offline_download_task",
		ResourceID:       "task-a",
		DedupeKey:        "offline:task-a",
		DispatchNodeID:   &nodeOne,
		DispatchNodeName: "主节点",
		DispatchNodeType: "local",
		NodeCapability:   "offline_download",
		Status:           model.QueueJobStatusFailed,
		ScheduledAt:      now,
		LastError:        "rpc failed: connection refused",
	})
	if err != nil {
		t.Fatalf("enqueue first queue job: %v", err)
	}
	_, err = repo.EnqueueIfAbsent(&model.QueueJob{
		QueueKey:         "io",
		JobType:          "archive.create",
		ResourceType:     "archive_package_session",
		ResourceID:       "pkg-b",
		DedupeKey:        "archive:pkg-b",
		DispatchNodeID:   &nodeTwo,
		DispatchNodeName: "备用节点",
		DispatchNodeType: "remote",
		NodeCapability:   "create_archive",
		Status:           model.QueueJobStatusPending,
		ScheduledAt:      now,
	})
	if err != nil {
		t.Fatalf("enqueue second queue job: %v", err)
	}

	rows, total, err := repo.List(QueueJobListFilter{NodeID: nodeOne, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list by node: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected exactly one node-filtered job, got total=%d len=%d", total, len(rows))
	}
	if rows[0].DispatchNodeID == nil || *rows[0].DispatchNodeID != nodeOne {
		t.Fatalf("expected dispatch node %d, got %#v", nodeOne, rows[0].DispatchNodeID)
	}
	if rows[0].DispatchNodeName != "主节点" {
		t.Fatalf("expected node name to be stored, got %q", rows[0].DispatchNodeName)
	}
	if rows[0].NodeCapability != "offline_download" {
		t.Fatalf("expected node capability to be stored, got %q", rows[0].NodeCapability)
	}
	if rows[0].LastError != "rpc failed: connection refused" {
		t.Fatalf("expected concrete failure reason, got %q", rows[0].LastError)
	}
}

func newQueueJobRepositoryForTest(t *testing.T) QueueJobRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "queue-jobs.db")), &gorm.Config{})
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
	return NewQueueJobRepository(db)
}
