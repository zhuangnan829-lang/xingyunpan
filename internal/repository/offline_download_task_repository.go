package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// OfflineDownloadTaskListFilter filters the user task list.
type OfflineDownloadTaskListFilter struct {
	UserID  uint
	Status  string
	Keyword string
}

// OfflineDownloadTaskRepository persists offline download tasks.
type OfflineDownloadTaskRepository interface {
	EnsureSchema() error
	ListByUser(ctx context.Context, filter OfflineDownloadTaskListFilter) ([]model.OfflineDownloadTask, error)
	GetByIDForUser(ctx context.Context, userID uint, id uint) (*model.OfflineDownloadTask, error)
	GetByID(ctx context.Context, id uint) (*model.OfflineDownloadTask, error)
	Create(ctx context.Context, task *model.OfflineDownloadTask) error
	Save(ctx context.Context, task *model.OfflineDownloadTask) error
	Delete(ctx context.Context, userID uint, id uint) error
	BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error)
}

type offlineDownloadTaskRepository struct {
	db *gorm.DB
}

func NewOfflineDownloadTaskRepository(db *gorm.DB) OfflineDownloadTaskRepository {
	return &offlineDownloadTaskRepository{db: db}
}

func (r *offlineDownloadTaskRepository) EnsureSchema() error {
	if r.db == nil {
		return fmt.Errorf("offline download database is not initialized")
	}

	statement := `CREATE TABLE IF NOT EXISTS offline_download_tasks (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		created_at DATETIME(3) NULL,
		updated_at DATETIME(3) NULL,
		deleted_at DATETIME(3) NULL,
		user_id BIGINT UNSIGNED NOT NULL,
		task_token VARCHAR(64) NOT NULL,
		name VARCHAR(255) NOT NULL,
		source_url TEXT NOT NULL,
		save_path VARCHAR(1000) NOT NULL DEFAULT '/offline-downloads',
		status VARCHAR(20) NOT NULL DEFAULT 'queued',
		progress BIGINT NOT NULL DEFAULT 0,
		speed_text VARCHAR(80) NOT NULL DEFAULT 'waiting',
		size_text VARCHAR(80) NOT NULL DEFAULT 'unknown',
		downloaded_bytes BIGINT NOT NULL DEFAULT 0,
		total_bytes BIGINT NOT NULL DEFAULT 0,
		error_message TEXT NULL,
		queue_job_id BIGINT UNSIGNED NULL,
		node_id BIGINT UNSIGNED NULL,
		downloader VARCHAR(64) NULL,
		rpc_url TEXT NULL,
		rpc_secret TEXT NULL,
		task_options TEXT NULL,
		temp_dir TEXT NULL,
		refresh_interval BIGINT NOT NULL DEFAULT 5,
		wait_for_seeding TINYINT(1) NOT NULL DEFAULT 0,
		remote_task_id VARCHAR(191) NULL,
		dispatch_node_id BIGINT UNSIGNED NULL,
		dispatch_node_name VARCHAR(255) NULL,
		dispatch_node_type VARCHAR(32) NULL,
		saved_file_id BIGINT UNSIGNED NULL,
		saved_folder_id BIGINT UNSIGNED NULL,
		completed_at DATETIME(3) NULL,
		PRIMARY KEY (id),
		UNIQUE KEY idx_offline_download_tasks_task_token (task_token),
		INDEX idx_offline_download_tasks_deleted_at (deleted_at),
		INDEX idx_offline_tasks_user_status (user_id, status),
		INDEX idx_offline_download_tasks_queue_job_id (queue_job_id),
		INDEX idx_offline_download_tasks_node_id (node_id),
		INDEX idx_offline_download_tasks_dispatch_node_id (dispatch_node_id),
		INDEX idx_offline_download_tasks_saved_file_id (saved_file_id),
		INDEX idx_offline_download_tasks_saved_folder_id (saved_folder_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if err := r.db.Exec(statement).Error; err != nil {
		return fmt.Errorf("ensure offline download tasks table failed: %w", err)
	}
	return r.ensureColumns()
}

func (r *offlineDownloadTaskRepository) ensureColumns() error {
	columns := []struct {
		name string
		sql  string
	}{
		{"deleted_at", "ALTER TABLE offline_download_tasks ADD COLUMN deleted_at DATETIME(3) NULL"},
		{"task_token", "ALTER TABLE offline_download_tasks ADD COLUMN task_token VARCHAR(64) NOT NULL"},
		{"save_path", "ALTER TABLE offline_download_tasks ADD COLUMN save_path VARCHAR(1000) NOT NULL DEFAULT '/offline-downloads'"},
		{"progress", "ALTER TABLE offline_download_tasks ADD COLUMN progress BIGINT NOT NULL DEFAULT 0"},
		{"speed_text", "ALTER TABLE offline_download_tasks ADD COLUMN speed_text VARCHAR(80) NOT NULL DEFAULT 'waiting'"},
		{"size_text", "ALTER TABLE offline_download_tasks ADD COLUMN size_text VARCHAR(80) NOT NULL DEFAULT 'unknown'"},
		{"downloaded_bytes", "ALTER TABLE offline_download_tasks ADD COLUMN downloaded_bytes BIGINT NOT NULL DEFAULT 0"},
		{"total_bytes", "ALTER TABLE offline_download_tasks ADD COLUMN total_bytes BIGINT NOT NULL DEFAULT 0"},
		{"error_message", "ALTER TABLE offline_download_tasks ADD COLUMN error_message TEXT NULL"},
		{"queue_job_id", "ALTER TABLE offline_download_tasks ADD COLUMN queue_job_id BIGINT UNSIGNED NULL"},
		{"node_id", "ALTER TABLE offline_download_tasks ADD COLUMN node_id BIGINT UNSIGNED NULL"},
		{"downloader", "ALTER TABLE offline_download_tasks ADD COLUMN downloader VARCHAR(64) NULL"},
		{"rpc_url", "ALTER TABLE offline_download_tasks ADD COLUMN rpc_url TEXT NULL"},
		{"rpc_secret", "ALTER TABLE offline_download_tasks ADD COLUMN rpc_secret TEXT NULL"},
		{"task_options", "ALTER TABLE offline_download_tasks ADD COLUMN task_options TEXT NULL"},
		{"temp_dir", "ALTER TABLE offline_download_tasks ADD COLUMN temp_dir TEXT NULL"},
		{"refresh_interval", "ALTER TABLE offline_download_tasks ADD COLUMN refresh_interval BIGINT NOT NULL DEFAULT 5"},
		{"wait_for_seeding", "ALTER TABLE offline_download_tasks ADD COLUMN wait_for_seeding TINYINT(1) NOT NULL DEFAULT 0"},
		{"remote_task_id", "ALTER TABLE offline_download_tasks ADD COLUMN remote_task_id VARCHAR(191) NULL"},
		{"dispatch_node_id", "ALTER TABLE offline_download_tasks ADD COLUMN dispatch_node_id BIGINT UNSIGNED NULL"},
		{"dispatch_node_name", "ALTER TABLE offline_download_tasks ADD COLUMN dispatch_node_name VARCHAR(255) NULL"},
		{"dispatch_node_type", "ALTER TABLE offline_download_tasks ADD COLUMN dispatch_node_type VARCHAR(32) NULL"},
		{"saved_file_id", "ALTER TABLE offline_download_tasks ADD COLUMN saved_file_id BIGINT UNSIGNED NULL"},
		{"saved_folder_id", "ALTER TABLE offline_download_tasks ADD COLUMN saved_folder_id BIGINT UNSIGNED NULL"},
		{"completed_at", "ALTER TABLE offline_download_tasks ADD COLUMN completed_at DATETIME(3) NULL"},
	}
	for _, column := range columns {
		if r.db.Migrator().HasColumn(&model.OfflineDownloadTask{}, column.name) {
			continue
		}
		if err := r.db.Exec(column.sql).Error; err != nil {
			return fmt.Errorf("add offline download tasks.%s failed: %w", column.name, err)
		}
	}
	return nil
}

func (r *offlineDownloadTaskRepository) ListByUser(ctx context.Context, filter OfflineDownloadTaskListFilter) ([]model.OfflineDownloadTask, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).Where("user_id = ?", filter.UserID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR source_url LIKE ? OR save_path LIKE ? OR status LIKE ?", like, like, like, like)
	}

	var tasks []model.OfflineDownloadTask
	if err := query.Order("updated_at DESC, id DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list offline download tasks failed: %w", err)
	}
	return tasks, nil
}

func (r *offlineDownloadTaskRepository) GetByIDForUser(ctx context.Context, userID uint, id uint) (*model.OfflineDownloadTask, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var task model.OfflineDownloadTask
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("offline download task not found")
		}
		return nil, fmt.Errorf("query offline download task failed: %w", err)
	}
	return &task, nil
}

func (r *offlineDownloadTaskRepository) GetByID(ctx context.Context, id uint) (*model.OfflineDownloadTask, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var task model.OfflineDownloadTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("offline download task not found")
		}
		return nil, fmt.Errorf("query offline download task failed: %w", err)
	}
	return &task, nil
}

func (r *offlineDownloadTaskRepository) Create(ctx context.Context, task *model.OfflineDownloadTask) error {
	if err := r.EnsureSchema(); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
		return fmt.Errorf("create offline download task failed: %w", err)
	}
	return nil
}

func (r *offlineDownloadTaskRepository) Save(ctx context.Context, task *model.OfflineDownloadTask) error {
	if err := r.EnsureSchema(); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Save(task).Error; err != nil {
		return fmt.Errorf("save offline download task failed: %w", err)
	}
	return nil
}

func (r *offlineDownloadTaskRepository) Delete(ctx context.Context, userID uint, id uint) error {
	if err := r.EnsureSchema(); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.OfflineDownloadTask{})
	if result.Error != nil {
		return fmt.Errorf("delete offline download task failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("offline download task not found")
	}
	return nil
}

func (r *offlineDownloadTaskRepository) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).Where("user_id = ? AND id IN ?", userID, ids).Delete(&model.OfflineDownloadTask{})
	if result.Error != nil {
		return 0, fmt.Errorf("batch delete offline download tasks failed: %w", result.Error)
	}
	return result.RowsAffected, nil
}
