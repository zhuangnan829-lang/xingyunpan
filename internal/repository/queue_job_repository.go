package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"xingyunpan-v2/internal/model"
)

// QueueJobStatusCount stores grouped queue job status counts.
type QueueJobStatusCount struct {
	QueueKey string
	Status   string
	Count    int64
}

// QueueJobListFilter filters queue job listing.
type QueueJobListFilter struct {
	QueueKey  string
	Status    string
	NodeID    uint
	Page      int
	PageSize  int
	Cursor    uint
	UseCursor bool
}

// QueueJobClearFilter filters queue jobs for cleanup.
type QueueJobClearFilter struct {
	QueueKey string
	Status   string
}

// QueueJobRepository handles persistence for unified queue jobs.
type QueueJobRepository interface {
	EnsureSchema() error
	EnqueueIfAbsent(job *model.QueueJob) (*model.QueueJob, error)
	GetByID(id uint) (*model.QueueJob, error)
	ClaimDueJobs(queueKey string, limit int) ([]model.QueueJob, error)
	MarkCompleted(id uint, result string) error
	MarkRetry(id uint, attempts int, scheduledAt time.Time, lastError string) error
	MarkFailed(id uint, attempts int, lastError string) error
	Retry(id uint) (*model.QueueJob, error)
	RequeueStaleProcessing(queueKey string, before time.Time, maxRetry int) (int64, error)
	ListStatusCounts() ([]QueueJobStatusCount, error)
	List(filter QueueJobListFilter) ([]model.QueueJob, int64, error)
	Delete(id uint) (int64, error)
	BatchDelete(ids []uint) (int64, error)
	Clear(filter QueueJobClearFilter) (int64, error)
}

type queueJobRepository struct {
	db *gorm.DB
}

// NewQueueJobRepository creates a unified queue job repository.
func NewQueueJobRepository(db *gorm.DB) QueueJobRepository {
	return &queueJobRepository{db: db}
}

// EnsureSchema creates the queue jobs table if it is missing.
func (r *queueJobRepository) EnsureSchema() error {
	if r.db == nil {
		return fmt.Errorf("queue jobs database is not initialized")
	}

	if r.db.Dialector.Name() == "mysql" {
		return r.ensureMySQLSchema()
	}

	if err := r.db.AutoMigrate(&model.QueueJob{}); err != nil {
		return fmt.Errorf("migrate queue jobs failed: %w", err)
	}

	return nil
}

func (r *queueJobRepository) ensureMySQLSchema() error {
	statement := `CREATE TABLE IF NOT EXISTS queue_jobs (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		created_at DATETIME(3) NULL,
		updated_at DATETIME(3) NULL,
		deleted_at DATETIME(3) NULL,
		queue_key VARCHAR(32) NOT NULL,
		job_type VARCHAR(64) NOT NULL,
		resource_type VARCHAR(64) NOT NULL,
		resource_id VARCHAR(128) NOT NULL,
		dedupe_key VARCHAR(191) NOT NULL,
		payload TEXT NOT NULL,
		dispatch_node_id BIGINT UNSIGNED NULL,
		dispatch_node_name VARCHAR(255) NULL,
		dispatch_node_type VARCHAR(32) NULL,
		node_capability VARCHAR(64) NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		attempts INT NOT NULL DEFAULT 0,
		max_attempts INT NOT NULL DEFAULT 0,
		scheduled_at DATETIME(3) NOT NULL,
		started_at DATETIME(3) NULL,
		finished_at DATETIME(3) NULL,
		last_error TEXT NULL,
		result TEXT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY idx_queue_jobs_dedupe_key (dedupe_key),
		INDEX idx_queue_jobs_deleted_at (deleted_at),
		INDEX idx_queue_jobs_lookup (queue_key, status, scheduled_at),
		INDEX idx_queue_jobs_dispatch_node_id (dispatch_node_id),
		INDEX idx_queue_jobs_node_capability (node_capability)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if err := r.db.Exec(statement).Error; err != nil {
		return fmt.Errorf("ensure queue jobs table failed: %w", err)
	}
	if err := r.ensureMySQLColumns(); err != nil {
		return err
	}
	return r.ensureMySQLIndexes()
}

func (r *queueJobRepository) ensureMySQLColumns() error {
	columns := []struct {
		name string
		sql  string
	}{
		{"deleted_at", "ALTER TABLE queue_jobs ADD COLUMN deleted_at DATETIME(3) NULL"},
		{"queue_key", "ALTER TABLE queue_jobs ADD COLUMN queue_key VARCHAR(32) NOT NULL DEFAULT 'io'"},
		{"job_type", "ALTER TABLE queue_jobs ADD COLUMN job_type VARCHAR(64) NOT NULL DEFAULT 'queue.job'"},
		{"resource_type", "ALTER TABLE queue_jobs ADD COLUMN resource_type VARCHAR(64) NOT NULL DEFAULT 'resource'"},
		{"resource_id", "ALTER TABLE queue_jobs ADD COLUMN resource_id VARCHAR(128) NOT NULL DEFAULT ''"},
		{"dedupe_key", "ALTER TABLE queue_jobs ADD COLUMN dedupe_key VARCHAR(191) NOT NULL DEFAULT ''"},
		{"payload", "ALTER TABLE queue_jobs ADD COLUMN payload TEXT NOT NULL"},
		{"dispatch_node_id", "ALTER TABLE queue_jobs ADD COLUMN dispatch_node_id BIGINT UNSIGNED NULL"},
		{"dispatch_node_name", "ALTER TABLE queue_jobs ADD COLUMN dispatch_node_name VARCHAR(255) NULL"},
		{"dispatch_node_type", "ALTER TABLE queue_jobs ADD COLUMN dispatch_node_type VARCHAR(32) NULL"},
		{"node_capability", "ALTER TABLE queue_jobs ADD COLUMN node_capability VARCHAR(64) NULL"},
		{"status", "ALTER TABLE queue_jobs ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending'"},
		{"attempts", "ALTER TABLE queue_jobs ADD COLUMN attempts INT NOT NULL DEFAULT 0"},
		{"max_attempts", "ALTER TABLE queue_jobs ADD COLUMN max_attempts INT NOT NULL DEFAULT 0"},
		{"scheduled_at", "ALTER TABLE queue_jobs ADD COLUMN scheduled_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"},
		{"started_at", "ALTER TABLE queue_jobs ADD COLUMN started_at DATETIME(3) NULL"},
		{"finished_at", "ALTER TABLE queue_jobs ADD COLUMN finished_at DATETIME(3) NULL"},
		{"last_error", "ALTER TABLE queue_jobs ADD COLUMN last_error TEXT NULL"},
		{"result", "ALTER TABLE queue_jobs ADD COLUMN result TEXT NULL"},
	}
	for _, column := range columns {
		if r.db.Migrator().HasColumn(&model.QueueJob{}, column.name) {
			continue
		}
		if err := r.db.Exec(column.sql).Error; err != nil {
			return fmt.Errorf("add queue_jobs.%s failed: %w", column.name, err)
		}
	}
	return nil
}

func (r *queueJobRepository) ensureMySQLIndexes() error {
	indexes := []struct {
		name string
		sql  string
	}{
		{"idx_queue_jobs_deleted_at", "CREATE INDEX idx_queue_jobs_deleted_at ON queue_jobs (deleted_at)"},
		{"idx_queue_jobs_lookup", "CREATE INDEX idx_queue_jobs_lookup ON queue_jobs (queue_key, status, scheduled_at)"},
		{"idx_queue_jobs_dispatch_node_id", "CREATE INDEX idx_queue_jobs_dispatch_node_id ON queue_jobs (dispatch_node_id)"},
		{"idx_queue_jobs_node_capability", "CREATE INDEX idx_queue_jobs_node_capability ON queue_jobs (node_capability)"},
	}
	for _, index := range indexes {
		if ok, err := r.hasMySQLIndex(index.name); err != nil {
			return err
		} else if !ok {
			if err := r.db.Exec(index.sql).Error; err != nil {
				return fmt.Errorf("create queue_jobs index %s failed: %w", index.name, err)
			}
		}
	}

	if ok, err := r.hasMySQLUniqueIndexOnColumn("dedupe_key"); err != nil {
		return err
	} else if !ok {
		if err := r.db.Exec("CREATE UNIQUE INDEX idx_queue_jobs_dedupe_key ON queue_jobs (dedupe_key)").Error; err != nil {
			return fmt.Errorf("create queue_jobs dedupe index failed: %w", err)
		}
	}

	return nil
}

func (r *queueJobRepository) hasMySQLIndex(name string) (bool, error) {
	var count int64
	if err := r.db.Raw(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		"queue_jobs",
		name,
	).Scan(&count).Error; err != nil {
		return false, fmt.Errorf("query queue_jobs index %s failed: %w", name, err)
	}
	return count > 0, nil
}

func (r *queueJobRepository) hasMySQLUniqueIndexOnColumn(column string) (bool, error) {
	var count int64
	if err := r.db.Raw(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ? AND non_unique = 0",
		"queue_jobs",
		column,
	).Scan(&count).Error; err != nil {
		return false, fmt.Errorf("query queue_jobs unique index on %s failed: %w", column, err)
	}
	return count > 0, nil
}

// EnqueueIfAbsent inserts one queue job unless the dedupe key already exists.
func (r *queueJobRepository) EnqueueIfAbsent(job *model.QueueJob) (*model.QueueJob, error) {
	if job == nil {
		return nil, fmt.Errorf("queue job cannot be nil")
	}

	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var existing model.QueueJob
	err := r.db.Where("dedupe_key = ?", job.DedupeKey).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query queue job failed: %w", err)
	}

	if err := r.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("create queue job failed: %w", err)
	}

	return job, nil
}

// GetByID returns one queue job by id.
func (r *queueJobRepository) GetByID(id uint) (*model.QueueJob, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var job model.QueueJob
	if err := r.db.First(&job, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query queue job failed: %w", err)
	}

	return &job, nil
}

// ClaimDueJobs moves due jobs into processing state and returns them.
func (r *queueJobRepository) ClaimDueJobs(queueKey string, limit int) ([]model.QueueJob, error) {
	if limit <= 0 {
		return []model.QueueJob{}, nil
	}

	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	now := time.Now()
	var jobs []model.QueueJob

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var dueJobs []model.QueueJob
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("queue_key = ? AND status = ? AND scheduled_at <= ?", queueKey, model.QueueJobStatusPending, now).
			Order("scheduled_at ASC, id ASC").
			Limit(limit).
			Find(&dueJobs).Error; err != nil {
			return fmt.Errorf("claim queue jobs failed: %w", err)
		}

		if len(dueJobs) == 0 {
			jobs = []model.QueueJob{}
			return nil
		}

		ids := make([]uint, 0, len(dueJobs))
		for _, job := range dueJobs {
			ids = append(ids, job.ID)
		}

		if err := tx.Model(&model.QueueJob{}).
			Where("id IN ? AND status = ?", ids, model.QueueJobStatusPending).
			Updates(map[string]interface{}{
				"status":      model.QueueJobStatusProcessing,
				"started_at":  now,
				"finished_at": nil,
			}).Error; err != nil {
			return fmt.Errorf("update queue jobs to processing failed: %w", err)
		}

		if err := tx.Where("id IN ?", ids).
			Order("scheduled_at ASC, id ASC").
			Find(&jobs).Error; err != nil {
			return fmt.Errorf("reload queue jobs failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// MarkCompleted marks a job as completed.
func (r *queueJobRepository) MarkCompleted(id uint, result string) error {
	now := time.Now()
	return r.db.Model(&model.QueueJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.QueueJobStatusCompleted,
			"finished_at": &now,
			"last_error":  "",
			"result":      result,
		}).Error
}

// MarkRetry schedules a failed job for another attempt.
func (r *queueJobRepository) MarkRetry(id uint, attempts int, scheduledAt time.Time, lastError string) error {
	return r.db.Model(&model.QueueJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.QueueJobStatusPending,
			"attempts":     attempts,
			"scheduled_at": scheduledAt,
			"last_error":   lastError,
			"started_at":   nil,
			"finished_at":  nil,
		}).Error
}

// MarkFailed marks a job as permanently failed.
func (r *queueJobRepository) MarkFailed(id uint, attempts int, lastError string) error {
	now := time.Now()
	return r.db.Model(&model.QueueJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.QueueJobStatusFailed,
			"attempts":    attempts,
			"finished_at": &now,
			"last_error":  lastError,
		}).Error
}

// Retry resets a failed or pending job back to pending for manual execution.
func (r *queueJobRepository) Retry(id uint) (*model.QueueJob, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	now := time.Now()
	var job model.QueueJob
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return gorm.ErrRecordNotFound
			}
			return fmt.Errorf("query queue job failed: %w", err)
		}
		if job.Status != model.QueueJobStatusFailed && job.Status != model.QueueJobStatusPending {
			return fmt.Errorf("only failed or pending queue jobs can be retried")
		}

		updates := map[string]interface{}{
			"status":       model.QueueJobStatusPending,
			"attempts":     0,
			"scheduled_at": now,
			"started_at":   nil,
			"finished_at":  nil,
			"last_error":   "",
			"result":       "",
		}
		if err := tx.Model(&model.QueueJob{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("retry queue job failed: %w", err)
		}

		return tx.First(&job, id).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &job, nil
}

// RequeueStaleProcessing restores processing jobs that outlived their execution window.
func (r *queueJobRepository) RequeueStaleProcessing(queueKey string, before time.Time, maxRetry int) (int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return 0, err
	}

	now := time.Now()
	var changed int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var jobs []model.QueueJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("queue_key = ? AND status = ? AND started_at IS NOT NULL AND started_at < ?", queueKey, model.QueueJobStatusProcessing, before).
			Find(&jobs).Error; err != nil {
			return fmt.Errorf("query stale queue jobs failed: %w", err)
		}

		for _, job := range jobs {
			attempts := job.Attempts + 1
			updates := map[string]interface{}{
				"attempts":    attempts,
				"last_error":  "job was requeued because processing timed out or worker exited",
				"finished_at": nil,
			}
			if attempts > maxRetry {
				updates["status"] = model.QueueJobStatusFailed
				updates["finished_at"] = &now
			} else {
				updates["status"] = model.QueueJobStatusPending
				updates["scheduled_at"] = now
				updates["started_at"] = nil
			}

			result := tx.Model(&model.QueueJob{}).
				Where("id = ? AND status = ?", job.ID, model.QueueJobStatusProcessing).
				Updates(updates)
			if result.Error != nil {
				return fmt.Errorf("requeue stale queue job failed: %w", result.Error)
			}
			changed += result.RowsAffected
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return changed, nil
}

// ListStatusCounts returns grouped job counts by queue and status.
func (r *queueJobRepository) ListStatusCounts() ([]QueueJobStatusCount, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var rows []QueueJobStatusCount
	if err := r.db.Model(&model.QueueJob{}).
		Select("queue_key, status, COUNT(*) AS count").
		Group("queue_key, status").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("query queue job counts failed: %w", err)
	}

	return rows, nil
}

// List returns paged queue jobs with optional queue/status filters.
func (r *queueJobRepository) List(filter QueueJobListFilter) ([]model.QueueJob, int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	query := r.db.Model(&model.QueueJob{})
	if filter.QueueKey != "" {
		query = query.Where("queue_key = ?", filter.QueueKey)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.NodeID > 0 {
		query = query.Where("dispatch_node_id = ?", filter.NodeID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count queue jobs failed: %w", err)
	}

	var jobs []model.QueueJob
	dataQuery := query
	if filter.UseCursor && filter.Cursor > 0 {
		dataQuery = dataQuery.Where("id < ?", filter.Cursor)
	}
	if err := dataQuery.
		Order("created_at DESC, id DESC").
		Offset(queueJobListOffset(filter.UseCursor, page, pageSize)).
		Limit(pageSize).
		Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("list queue jobs failed: %w", err)
	}

	return jobs, total, nil
}

func queueJobListOffset(useCursor bool, page, pageSize int) int {
	if useCursor {
		return 0
	}
	return (page - 1) * pageSize
}

// Delete removes one queue job unless it is currently processing.
func (r *queueJobRepository) Delete(id uint) (int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return 0, err
	}

	result := r.db.
		Where("id = ? AND status <> ?", id, model.QueueJobStatusProcessing).
		Delete(&model.QueueJob{})
	if result.Error != nil {
		return 0, fmt.Errorf("delete queue job failed: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// BatchDelete removes selected queue jobs unless they are currently processing.
func (r *queueJobRepository) BatchDelete(ids []uint) (int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	result := r.db.
		Where("id IN ? AND status <> ?", ids, model.QueueJobStatusProcessing).
		Delete(&model.QueueJob{})
	if result.Error != nil {
		return 0, fmt.Errorf("batch delete queue jobs failed: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// Clear removes terminal queue jobs matching the provided filter.
func (r *queueJobRepository) Clear(filter QueueJobClearFilter) (int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return 0, err
	}

	query := r.db.Model(&model.QueueJob{})
	if filter.QueueKey != "" {
		query = query.Where("queue_key = ?", filter.QueueKey)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		query = query.Where("status IN ?", []string{model.QueueJobStatusCompleted, model.QueueJobStatusFailed})
	}

	result := query.Delete(&model.QueueJob{})
	if result.Error != nil {
		return 0, fmt.Errorf("clear queue jobs failed: %w", result.Error)
	}

	return result.RowsAffected, nil
}
