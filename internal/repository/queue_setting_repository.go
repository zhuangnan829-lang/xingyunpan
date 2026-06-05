package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// QueueSettingRepository handles persistence for queue settings.
type QueueSettingRepository interface {
	EnsureSchema() error
	List() ([]model.QueueSetting, error)
	GetByQueueKey(queueKey string) (*model.QueueSetting, error)
	Save(setting *model.QueueSetting) error
}

type queueSettingRepository struct {
	db *gorm.DB
}

// NewQueueSettingRepository creates a queue settings repository.
func NewQueueSettingRepository(db *gorm.DB) QueueSettingRepository {
	return &queueSettingRepository{db: db}
}

// EnsureSchema creates the queue settings table if it is missing.
func (r *queueSettingRepository) EnsureSchema() error {
	if r.db == nil {
		return fmt.Errorf("queue settings database is not initialized")
	}

	if r.db.Dialector.Name() == "mysql" {
		return r.ensureMySQLSchema()
	}

	if err := r.db.AutoMigrate(&model.QueueSetting{}); err != nil {
		if !isQueueSettingsLegacyUniqueDropMissingError(err) {
			return fmt.Errorf("migrate queue settings failed: %w", err)
		}
	}

	if err := r.ensureColumns(); err != nil {
		return fmt.Errorf("ensure queue settings columns failed: %w", err)
	}
	if err := r.ensureQueueKeyUniqueIndex(); err != nil {
		return fmt.Errorf("ensure queue settings queue_key unique index failed: %w", err)
	}

	return nil
}

func (r *queueSettingRepository) ensureMySQLSchema() error {
	statement := `CREATE TABLE IF NOT EXISTS queue_settings (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		created_at DATETIME(3) NULL,
		updated_at DATETIME(3) NULL,
		deleted_at DATETIME(3) NULL,
		queue_key VARCHAR(32) NOT NULL,
		worker_num INT NOT NULL DEFAULT 5,
		max_execution INT NOT NULL DEFAULT 3600,
		backoff_factor INT NOT NULL DEFAULT 2,
		max_backoff INT NOT NULL DEFAULT 60,
		max_retry INT NOT NULL DEFAULT 0,
		retry_delay INT NOT NULL DEFAULT 0,
		PRIMARY KEY (id),
		INDEX idx_queue_settings_deleted_at (deleted_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	if err := r.db.Exec(statement).Error; err != nil {
		return fmt.Errorf("ensure queue settings table failed: %w", err)
	}
	if err := r.ensureColumns(); err != nil {
		return fmt.Errorf("ensure queue settings columns failed: %w", err)
	}
	if err := r.ensureQueueKeyUniqueIndex(); err != nil {
		return fmt.Errorf("ensure queue settings queue_key unique index failed: %w", err)
	}
	return nil
}

func (r *queueSettingRepository) ensureColumns() error {
	if !r.db.Migrator().HasTable(&model.QueueSetting{}) {
		return r.db.AutoMigrate(&model.QueueSetting{})
	}

	columns := []string{
		"CreatedAt",
		"UpdatedAt",
		"DeletedAt",
		"QueueKey",
		"WorkerNum",
		"MaxExecution",
		"BackoffFactor",
		"MaxBackoff",
		"MaxRetry",
		"RetryDelay",
	}
	for _, column := range columns {
		if r.db.Migrator().HasColumn(&model.QueueSetting{}, column) {
			continue
		}
		if err := r.db.Migrator().AddColumn(&model.QueueSetting{}, column); err != nil {
			return err
		}
	}

	return nil
}

// List returns all queue settings ordered by id.
func (r *queueSettingRepository) List() ([]model.QueueSetting, error) {
	var settings []model.QueueSetting
	if err := r.db.Order("id asc").Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("query queue settings failed: %w", err)
	}

	return settings, nil
}

// GetByQueueKey returns a queue setting by key.
func (r *queueSettingRepository) GetByQueueKey(queueKey string) (*model.QueueSetting, error) {
	var setting model.QueueSetting
	if err := r.db.Where("queue_key = ?", queueKey).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query queue setting failed: %w", err)
	}

	return &setting, nil
}

// Save creates or updates a queue setting row.
func (r *queueSettingRepository) Save(setting *model.QueueSetting) error {
	if setting == nil {
		return fmt.Errorf("queue setting cannot be nil")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("create queue setting failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("update queue setting failed: %w", err)
	}

	return nil
}

func (r *queueSettingRepository) ensureQueueKeyUniqueIndex() error {
	if !r.db.Migrator().HasTable(&model.QueueSetting{}) {
		return r.db.AutoMigrate(&model.QueueSetting{})
	}

	hasUnique, err := queueSettingsHasQueueKeyUniqueIndex(r.db)
	if err != nil {
		return err
	}
	if hasUnique {
		return nil
	}

	if r.db.Dialector.Name() == "mysql" {
		return r.createMySQLQueueKeyUniqueIndex()
	}

	return r.db.Migrator().CreateIndex(&model.QueueSetting{}, "QueueKey")
}

func (r *queueSettingRepository) createMySQLQueueKeyUniqueIndex() error {
	indexNames := []string{
		"idx_queue_settings_queue_key",
		"uni_queue_settings_queue_key_current",
		"uni_queue_settings_queue_key_v2",
	}
	for _, indexName := range indexNames {
		exists, err := mysqlIndexExists(r.db, indexName)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		return r.db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX %s ON queue_settings (queue_key)", indexName)).Error
	}

	return fmt.Errorf("queue_settings queue_key unique index cannot be created: all compatible index names already exist")
}

func queueSettingsHasQueueKeyUniqueIndex(db *gorm.DB) (bool, error) {
	switch db.Dialector.Name() {
	case "mysql":
		var count int64
		err := db.Raw(`
			SELECT COUNT(*) FROM (
				SELECT index_name
				FROM information_schema.statistics
				WHERE table_schema = DATABASE()
					AND table_name = 'queue_settings'
					AND non_unique = 0
				GROUP BY index_name
				HAVING COUNT(*) = 1 AND SUM(column_name = 'queue_key') = 1
			) AS queue_key_unique_indexes`,
		).Scan(&count).Error
		return count > 0, err
	case "sqlite":
		return sqliteHasQueueKeyUniqueIndex(db)
	default:
		if db.Migrator().HasIndex(&model.QueueSetting{}, "idx_queue_settings_queue_key") ||
			db.Migrator().HasIndex(&model.QueueSetting{}, "uni_queue_settings_queue_key") {
			return true, nil
		}
		return false, nil
	}
}

func mysqlIndexExists(db *gorm.DB, indexName string) (bool, error) {
	var count int64
	err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
			AND table_name = 'queue_settings'
			AND index_name = ?`,
		indexName,
	).Scan(&count).Error
	return count > 0, err
}

func sqliteHasQueueKeyUniqueIndex(db *gorm.DB) (bool, error) {
	type sqliteIndex struct {
		Name   string `gorm:"column:name"`
		Unique int    `gorm:"column:unique"`
	}
	var indexes []sqliteIndex
	if err := db.Raw(`PRAGMA index_list('queue_settings')`).Scan(&indexes).Error; err != nil {
		return false, err
	}

	type sqliteIndexColumn struct {
		Name string `gorm:"column:name"`
	}
	for _, index := range indexes {
		if index.Unique != 1 {
			continue
		}

		var columns []sqliteIndexColumn
		if err := db.Raw(fmt.Sprintf("PRAGMA index_info(%s)", quoteSQLiteString(index.Name))).Scan(&columns).Error; err != nil {
			return false, err
		}
		if len(columns) == 1 && columns[0].Name == "queue_key" {
			return true, nil
		}
	}

	return false, nil
}

func quoteSQLiteString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func isQueueSettingsLegacyUniqueDropMissingError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1091 {
		message := strings.ToLower(mysqlErr.Message)
		return isQueueSettingsLegacyDropMessage(message)
	}

	return isQueueSettingsLegacyDropMessage(strings.ToLower(err.Error()))
}

func isQueueSettingsLegacyDropMessage(message string) bool {
	return strings.Contains(message, "uni_queue_settings_queue_key") &&
		strings.Contains(message, "drop") &&
		(strings.Contains(message, "check that column/key exists") ||
			strings.Contains(message, "can't drop") ||
			strings.Contains(message, "error 1091") ||
			strings.Contains(message, "1091"))
}
