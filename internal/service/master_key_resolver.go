package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xingyunpan-v2/internal/repository"

	"gorm.io/gorm"
)

type MasterKeyResolver interface {
	ResolveMasterKey() ([]byte, *MasterKeyStatusPayload, error)
	Status() *MasterKeyStatusPayload
}

type fileSystemMasterKeyResolver struct {
	db           *gorm.DB
	settingsRepo repository.FileSystemSettingRepository
}

func NewFileSystemMasterKeyResolver(db *gorm.DB, settingsRepo repository.FileSystemSettingRepository) MasterKeyResolver {
	return &fileSystemMasterKeyResolver{db: db, settingsRepo: settingsRepo}
}

func (r *fileSystemMasterKeyResolver) ResolveMasterKey() ([]byte, *MasterKeyStatusPayload, error) {
	mode := r.currentMode()
	key, source, err := r.resolve(mode)
	if err != nil {
		status := &MasterKeyStatusPayload{
			StorageMode: mode,
			Available:   false,
			Source:      source,
			Message:     err.Error(),
		}
		return nil, status, err
	}
	return key, masterKeyStatus(mode, source, key), nil
}

func (r *fileSystemMasterKeyResolver) Status() *MasterKeyStatusPayload {
	key, status, err := r.ResolveMasterKey()
	if err != nil {
		return status
	}
	return masterKeyStatus(status.StorageMode, status.Source, key)
}

func (r *fileSystemMasterKeyResolver) currentMode() string {
	mode := "database"
	if r.settingsRepo != nil {
		if settings, err := r.settingsRepo.Get(); err == nil && settings != nil && strings.TrimSpace(settings.MasterKeyStorage) != "" {
			mode = settings.MasterKeyStorage
		}
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "database", "file", "env":
		return mode
	default:
		return "database"
	}
}

func (r *fileSystemMasterKeyResolver) resolve(mode string) ([]byte, string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "env":
		value := strings.TrimSpace(os.Getenv("XINGYUNPAN_MASTER_KEY"))
		if value == "" {
			return nil, "env:XINGYUNPAN_MASTER_KEY", fmt.Errorf("XINGYUNPAN_MASTER_KEY is empty")
		}
		return []byte(value), "env:XINGYUNPAN_MASTER_KEY", nil
	case "file":
		path := strings.TrimSpace(os.Getenv("XINGYUNPAN_MASTER_KEY_FILE"))
		if path == "" {
			path = filepath.Join("data", "master.key")
		}
		key, err := loadFileMasterKey(path)
		return key, "file:" + path, err
	default:
		key, err := r.loadOrCreateDatabaseKey()
		return key, "database:file_system_secrets.master_key", err
	}
}

func loadFileMasterKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("master key file %s is not available: %w", path, err)
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return nil, fmt.Errorf("master key file %s is empty", path)
	}
	return []byte(value), nil
}

func (r *fileSystemMasterKeyResolver) loadOrCreateDatabaseKey() ([]byte, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is not available")
	}
	if err := r.db.Exec(`CREATE TABLE IF NOT EXISTS file_system_secrets (
		key_name VARCHAR(128) NOT NULL PRIMARY KEY,
		secret_value TEXT NOT NULL,
		updated_at DATETIME NOT NULL
	)`).Error; err != nil {
		return nil, err
	}

	var row fileSystemSecretRow
	err := r.db.Raw(`SELECT key_name, secret_value, updated_at FROM file_system_secrets WHERE key_name = ?`, "master_key").Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(row.SecretValue) != "" {
		return []byte(strings.TrimSpace(row.SecretValue)), nil
	}
	key, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	if err := r.db.Exec(`INSERT INTO file_system_secrets (key_name, secret_value, updated_at) VALUES (?, ?, ?)`, "master_key", key, time.Now()).Error; err != nil {
		return nil, err
	}
	return []byte(key), nil
}

func masterKeyStatus(mode, source string, key []byte) *MasterKeyStatusPayload {
	sum := sha256.Sum256(key)
	return &MasterKeyStatusPayload{
		StorageMode: strings.ToLower(strings.TrimSpace(mode)),
		Available:   true,
		Source:      strings.TrimSpace(source),
		Fingerprint: hex.EncodeToString(sum[:])[:16],
	}
}
