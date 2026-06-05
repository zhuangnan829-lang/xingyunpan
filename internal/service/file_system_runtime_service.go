package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/storage"

	"gorm.io/gorm"
)

type MapRuntimeConfig struct {
	Provider      string `json:"provider"`
	Engine        string `json:"engine"`
	TileURL       string `json:"tile_url"`
	StyleURL      string `json:"style_url,omitempty"`
	Attribution   string `json:"attribution"`
	RequiresToken bool   `json:"requires_token"`
}

type MasterKeyStatusPayload struct {
	StorageMode string `json:"storage_mode"`
	Available   bool   `json:"available"`
	Source      string `json:"source"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Message     string `json:"message,omitempty"`
}

type PackageDownloadSessionPayload struct {
	SessionID    string               `json:"session_id"`
	FileIDs      []uint               `json:"file_ids"`
	NodeID       uint                 `json:"node_id,omitempty"`
	NodeName     string               `json:"node_name,omitempty"`
	DispatchNode *SelectedNodePayload `json:"dispatch_node,omitempty"`
	ExpiresAt    time.Time            `json:"expires_at"`
	TTLSeconds   int                  `json:"ttl_seconds"`
	DownloadURL  string               `json:"download_url"`
}

type ExtractArchivePayload struct {
	TaskID         string     `json:"task_id"`
	QueueJobID     *uint      `json:"queue_job_id,omitempty"`
	NodeID         uint       `json:"node_id"`
	NodeName       string     `json:"node_name"`
	SourceFileID   uint       `json:"source_file_id"`
	TargetFolderID *uint      `json:"target_folder_id"`
	Status         string     `json:"status"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	ExtractedFiles []uint     `json:"extracted_files"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type WOPISessionPayload struct {
	SessionID   string    `json:"session_id"`
	FileID      uint      `json:"file_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	TTLSeconds  int       `json:"ttl_seconds"`
	AccessToken string    `json:"access_token"`
}

type SlaveSignaturePayload struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	BodySHA256 string `json:"body_sha256"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
	TTLSeconds int    `json:"ttl_seconds"`
}

type SlaveSignatureVerifyPayload struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	BodySHA256 string `json:"body_sha256"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
	Valid      bool   `json:"valid"`
	Reason     string `json:"reason,omitempty"`
}

type OAuthRefreshStatusPayload struct {
	Interval          string                         `json:"interval"`
	IntervalSeconds   int                            `json:"interval_seconds"`
	EnabledApps       int64                          `json:"enabled_apps"`
	TotalApps         int64                          `json:"total_apps"`
	ActiveCredentials int64                          `json:"active_credentials"`
	TotalCredentials  int64                          `json:"total_credentials"`
	DueCredentials    int64                          `json:"due_credentials"`
	LastRunAt         *time.Time                     `json:"last_run_at,omitempty"`
	NextRunAt         *time.Time                     `json:"next_run_at,omitempty"`
	Message           string                         `json:"message"`
	RefreshSummary    *OAuthCredentialRefreshSummary `json:"refresh_summary,omitempty"`
}

type FileSystemRuntimeService interface {
	GetMapRuntimeConfig() (*MapRuntimeConfig, error)
	GetMasterKeyStatus() *MasterKeyStatusPayload
	CreatePackageDownloadSession(ctx context.Context, userID uint, fileIDs []uint) (*PackageDownloadSessionPayload, error)
	WritePackageDownload(ctx context.Context, userID uint, sessionID string, writer io.Writer) (*PackageDownloadSessionPayload, error)
	ExtractArchive(ctx context.Context, userID uint, sourceFileID uint, targetFolderID *uint) (*ExtractArchivePayload, error)
	ExecuteArchiveCreate(ctx context.Context, userID uint, sessionID string, fileIDs []uint) (string, error)
	ExecuteArchiveExtract(ctx context.Context, userID uint, taskID string) (string, error)
	CreateWOPISession(ctx context.Context, userID uint, fileID uint) (*WOPISessionPayload, error)
	GetWOPISession(ctx context.Context, userID uint, sessionID string) (*WOPISessionPayload, error)
	SignSlaveRequest(method, path, bodySHA256 string) (*SlaveSignaturePayload, error)
	VerifySlaveRequest(method, path, bodySHA256 string, timestamp int64, signature string) (*SlaveSignatureVerifyPayload, error)
	GetOAuthRefreshStatus() (*OAuthRefreshStatusPayload, error)
	RunOAuthRefresh() (*OAuthRefreshStatusPayload, error)
	OAuthRefreshInterval() (time.Duration, error)
}

type fileSystemRuntimeService struct {
	db           *gorm.DB
	settingsRepo repository.FileSystemSettingRepository
	userFiles    repository.UserFileRepository
	storage      storage.Storage
	baseURL      string
	nodeDispatch NodeDispatchService
	oauthTokens  OAuthCredentialService
	masterKeys   MasterKeyResolver

	mu              sync.Mutex
	packageSessions map[string]*packageDownloadSession
	extractTasks    map[string]*ExtractArchivePayload
	wopiSessions    map[string]*wopiSession
	oauthLastRun    *time.Time
}

type packageDownloadSession struct {
	UserID       uint
	FileIDs      []uint
	NodeID       uint
	NodeName     string
	DispatchNode *SelectedNodePayload
	QueueJobID   *uint
	ExpiresAt    time.Time
}

type wopiSession struct {
	UserID      uint
	FileID      uint
	ExpiresAt   time.Time
	AccessToken string
}

type fileSystemSecretRow struct {
	KeyName     string
	SecretValue string
	UpdatedAt   time.Time
}

func NewFileSystemRuntimeService(db *gorm.DB, settingsRepo repository.FileSystemSettingRepository, userFiles repository.UserFileRepository, store storage.Storage, baseURL string, nodeDispatch NodeDispatchService, oauthTokens ...OAuthCredentialService) FileSystemRuntimeService {
	var tokenService OAuthCredentialService
	if len(oauthTokens) > 0 {
		tokenService = oauthTokens[0]
	}
	if tokenService == nil && db != nil {
		tokenService = NewOAuthCredentialService(db)
	}
	return &fileSystemRuntimeService{
		db:              db,
		settingsRepo:    settingsRepo,
		userFiles:       userFiles,
		storage:         store,
		baseURL:         strings.TrimRight(baseURL, "/"),
		nodeDispatch:    nodeDispatch,
		oauthTokens:     tokenService,
		masterKeys:      NewFileSystemMasterKeyResolver(db, settingsRepo),
		packageSessions: map[string]*packageDownloadSession{},
		extractTasks:    map[string]*ExtractArchivePayload{},
		wopiSessions:    map[string]*wopiSession{},
	}
}

func (s *fileSystemRuntimeService) GetMapRuntimeConfig() (*MapRuntimeConfig, error) {
	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}

	switch settings.MapProvider {
	case "google-leaflet":
		return &MapRuntimeConfig{
			Provider:      settings.MapProvider,
			Engine:        "leaflet",
			TileURL:       "https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}",
			Attribution:   "Google Maps",
			RequiresToken: false,
		}, nil
	case "osm-mapbox":
		return &MapRuntimeConfig{
			Provider:      settings.MapProvider,
			Engine:        "mapbox",
			StyleURL:      "mapbox://styles/mapbox/streets-v12",
			Attribution:   "OpenStreetMap contributors, Mapbox",
			RequiresToken: true,
		}, nil
	default:
		return &MapRuntimeConfig{
			Provider:      "osm-leaflet",
			Engine:        "leaflet",
			TileURL:       "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
			Attribution:   "OpenStreetMap contributors",
			RequiresToken: false,
		}, nil
	}
}

func (s *fileSystemRuntimeService) GetMasterKeyStatus() *MasterKeyStatusPayload {
	if s.masterKeys == nil {
		s.masterKeys = NewFileSystemMasterKeyResolver(s.db, s.settingsRepo)
	}
	return s.masterKeys.Status()
}

func (s *fileSystemRuntimeService) CreatePackageDownloadSession(ctx context.Context, userID uint, fileIDs []uint) (*PackageDownloadSessionPayload, error) {
	if len(fileIDs) == 0 {
		s.recordArchiveQueueJob(ctx, userID, "package-missing-files", fileIDs, NodeCapabilityCreateArchive, nil, model.QueueJobStatusFailed, "parameter error: file_ids cannot be empty")
		return nil, fmt.Errorf("file_ids cannot be empty")
	}
	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}
	var dispatchNode *SelectedNodePayload
	if s.nodeDispatch != nil {
		dispatchNode, err = s.nodeDispatch.SelectCreateArchiveNode()
		if err != nil {
			s.recordArchiveQueueJob(ctx, userID, "package-node-selection-failed", fileIDs, NodeCapabilityCreateArchive, nil, model.QueueJobStatusFailed, "node selection failed: "+err.Error())
			return nil, err
		}
	} else {
		s.recordArchiveQueueJob(ctx, userID, "package-node-dispatch-missing", fileIDs, NodeCapabilityCreateArchive, nil, model.QueueJobStatusFailed, "node selection failed: node dispatch service is not initialized")
		return nil, fmt.Errorf("node dispatch service is not initialized")
	}
	ttl := time.Duration(settings.ServerSideDownloadSessionTTL) * time.Second
	for _, fileID := range fileIDs {
		file, err := s.userFiles.GetByIDWithPhysicalFile(fileID)
		if err != nil {
			s.recordArchiveQueueJob(ctx, userID, fmt.Sprintf("package-file-%d", fileID), fileIDs, NodeCapabilityCreateArchive, dispatchNode, model.QueueJobStatusFailed, fmt.Sprintf("parameter error: file %d not found: %v", fileID, err))
			return nil, fmt.Errorf("file %d not found: %w", fileID, err)
		}
		if file.UserID != userID {
			s.recordArchiveQueueJob(ctx, userID, fmt.Sprintf("package-file-%d", fileID), fileIDs, NodeCapabilityCreateArchive, dispatchNode, model.QueueJobStatusFailed, fmt.Sprintf("permission error: file %d is not owned by current user", fileID))
			return nil, fmt.Errorf("file %d is not owned by current user", fileID)
		}
		if file.IsFolder {
			s.recordArchiveQueueJob(ctx, userID, fmt.Sprintf("package-file-%d", fileID), fileIDs, NodeCapabilityCreateArchive, dispatchNode, model.QueueJobStatusFailed, fmt.Sprintf("parameter error: folder %d is not supported in package download sessions yet", fileID))
			return nil, fmt.Errorf("folder %d is not supported in package download sessions yet", fileID)
		}
		if file.PhysicalFile == nil || file.PhysicalFile.StoragePath == "" {
			s.recordArchiveQueueJob(ctx, userID, fmt.Sprintf("package-file-%d", fileID), fileIDs, NodeCapabilityCreateArchive, dispatchNode, model.QueueJobStatusFailed, fmt.Sprintf("parameter error: file %d has no physical storage", fileID))
			return nil, fmt.Errorf("file %d has no physical storage", fileID)
		}
	}

	sessionID, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(ttl)
	queueJob := s.recordArchiveQueueJob(ctx, userID, sessionID, fileIDs, NodeCapabilityCreateArchive, dispatchNode, model.QueueJobStatusPending, "")
	var queueJobID *uint
	if queueJob != nil {
		queueJobID = &queueJob.ID
	}
	s.mu.Lock()
	s.packageSessions[sessionID] = &packageDownloadSession{UserID: userID, FileIDs: append([]uint{}, fileIDs...), NodeID: dispatchNode.ID, NodeName: dispatchNode.Name, DispatchNode: dispatchNode, QueueJobID: queueJobID, ExpiresAt: expiresAt}
	s.cleanupLocked(time.Now())
	s.mu.Unlock()

	return s.packagePayload(sessionID, fileIDs, dispatchNode, expiresAt, settings.ServerSideDownloadSessionTTL), nil
}

func (s *fileSystemRuntimeService) WritePackageDownload(ctx context.Context, userID uint, sessionID string, writer io.Writer) (*PackageDownloadSessionPayload, error) {
	session, err := s.getPackageSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if s.storage != nil {
		archivePath := archivePackageStoragePath(sessionID)
		if s.storage.Exists(archivePath) {
			reader, readErr := s.storage.Read(archivePath)
			if readErr != nil {
				s.markArchiveQueueFailed(ctx, session.QueueJobID, "read generated archive failed: "+readErr.Error())
				return nil, readErr
			}
			_, copyErr := io.Copy(writer, reader)
			closeErr := reader.Close()
			if copyErr != nil {
				return nil, copyErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
			settings, _ := s.currentSettings()
			return s.packagePayload(sessionID, session.FileIDs, session.DispatchNode, session.ExpiresAt, settings.ServerSideDownloadSessionTTL), nil
		}
	}
	if session.DispatchNode == nil {
		s.markArchiveQueueFailed(ctx, session.QueueJobID, "node selection failed: archive dispatch node is missing")
		return nil, fmt.Errorf("archive dispatch node is missing")
	}
	if session.DispatchNode.Type != "master" {
		s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("node selection failed: remote archive execution is not supported yet for node %s", session.DispatchNode.Name))
		return nil, fmt.Errorf("remote archive execution is not supported yet for node %s", session.DispatchNode.Name)
	}
	s.markArchiveQueueProcessing(ctx, session.QueueJobID)

	zipWriter := zip.NewWriter(writer)
	for _, fileID := range session.FileIDs {
		file, err := s.userFiles.GetByIDWithPhysicalFile(fileID)
		if err != nil {
			_ = zipWriter.Close()
			s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("parameter error: %v", err))
			return nil, err
		}
		if file.UserID != userID || file.IsFolder || file.PhysicalFile == nil {
			_ = zipWriter.Close()
			s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("permission error: file %d is not downloadable", fileID))
			return nil, fmt.Errorf("file %d is not downloadable", fileID)
		}
		reader, err := readPhysicalBlob(s.storage, file.PhysicalFile, s.masterKeys)
		if err != nil {
			_ = zipWriter.Close()
			s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("permission error: %v", err))
			return nil, err
		}
		entry, err := zipWriter.Create(s.uniqueZipName(file.FileName, fileID))
		if err != nil {
			_ = reader.Close()
			_ = zipWriter.Close()
			s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("parameter error: %v", err))
			return nil, err
		}
		if _, err := io.Copy(entry, reader); err != nil {
			_ = reader.Close()
			_ = zipWriter.Close()
			s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("permission error: %v", err))
			return nil, err
		}
		_ = reader.Close()
	}
	if err := zipWriter.Close(); err != nil {
		s.markArchiveQueueFailed(ctx, session.QueueJobID, fmt.Sprintf("parameter error: %v", err))
		return nil, err
	}
	s.markArchiveQueueCompleted(ctx, session.QueueJobID, "archive package generated")

	settings, _ := s.currentSettings()
	return s.packagePayload(sessionID, session.FileIDs, session.DispatchNode, session.ExpiresAt, settings.ServerSideDownloadSessionTTL), nil
}

func (s *fileSystemRuntimeService) ExecuteArchiveCreate(ctx context.Context, userID uint, sessionID string, fileIDs []uint) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if userID == 0 || sessionID == "" || len(fileIDs) == 0 {
		return "", fmt.Errorf("archive create payload is incomplete")
	}
	if s.storage == nil {
		return "", fmt.Errorf("storage is not initialized")
	}

	var dispatchNode *SelectedNodePayload
	if session, err := s.getPackageSession(userID, sessionID); err == nil {
		dispatchNode = session.DispatchNode
	} else if s.nodeDispatch != nil {
		node, err := s.nodeDispatch.SelectCreateArchiveNode()
		if err != nil {
			return "", err
		}
		dispatchNode = node
	} else {
		return "", fmt.Errorf("node dispatch service is not initialized")
	}
	if dispatchNode == nil {
		return "", fmt.Errorf("archive dispatch node is missing")
	}
	if dispatchNode.Type != "master" {
		return "", fmt.Errorf("remote archive execution is not supported yet for node %s", dispatchNode.Name)
	}

	var archive bytes.Buffer
	if err := s.writePackageZip(ctx, userID, fileIDs, &archive); err != nil {
		return "", err
	}
	archivePath := archivePackageStoragePath(sessionID)
	if err := s.storage.Save(bytes.NewReader(archive.Bytes()), archivePath); err != nil {
		return "", fmt.Errorf("save generated archive failed: %w", err)
	}

	result, err := json.Marshal(map[string]interface{}{
		"status":         "completed",
		"execution_mode": "unified_runner",
		"archive_path":   archivePath,
		"session_id":     sessionID,
		"file_ids":       fileIDs,
		"node_id":        dispatchNode.ID,
		"node_name":      dispatchNode.Name,
	})
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (s *fileSystemRuntimeService) ExtractArchive(ctx context.Context, userID uint, sourceFileID uint, targetFolderID *uint) (*ExtractArchivePayload, error) {
	if sourceFileID == 0 {
		s.recordArchiveQueueJob(ctx, userID, "extract-missing-source", []uint{sourceFileID}, NodeCapabilityExtractArchive, nil, model.QueueJobStatusFailed, "parameter error: source_file_id is required")
		return nil, fmt.Errorf("source_file_id is required")
	}
	if s.nodeDispatch == nil {
		s.recordArchiveQueueJob(ctx, userID, fmt.Sprintf("extract-source-%d", sourceFileID), []uint{sourceFileID}, NodeCapabilityExtractArchive, nil, model.QueueJobStatusFailed, "node selection failed: node dispatch service is not initialized")
		return nil, fmt.Errorf("node dispatch service is not initialized")
	}
	node, err := s.nodeDispatch.SelectExtractArchiveNode()
	if err != nil {
		s.recordArchiveQueueJob(ctx, userID, fmt.Sprintf("extract-source-%d", sourceFileID), []uint{sourceFileID}, NodeCapabilityExtractArchive, nil, model.QueueJobStatusFailed, "node selection failed: "+err.Error())
		return nil, err
	}
	taskID, err := randomToken(18)
	if err != nil {
		return nil, err
	}
	task := &ExtractArchivePayload{
		TaskID:         taskID,
		NodeID:         node.ID,
		NodeName:       node.Name,
		SourceFileID:   sourceFileID,
		TargetFolderID: copyUintPtr(targetFolderID),
		Status:         "processing",
		ExtractedFiles: []uint{},
		CreatedAt:      time.Now(),
	}
	queueJob := s.recordArchiveQueueJob(ctx, userID, taskID, []uint{sourceFileID}, NodeCapabilityExtractArchive, node, model.QueueJobStatusPending, "")
	if queueJob != nil {
		task.QueueJobID = &queueJob.ID
	}
	s.rememberExtractTask(task)
	if err := s.persistExtractTask(ctx, task); err != nil {
		return nil, err
	}

	return cloneExtractTask(task), nil
}

func (s *fileSystemRuntimeService) ExecuteArchiveExtract(ctx context.Context, userID uint, taskID string) (string, error) {
	task, err := s.getExtractTask(ctx, strings.TrimSpace(taskID))
	if err != nil {
		return "", err
	}
	if task == nil {
		return "", fmt.Errorf("archive extract task not found")
	}
	if userID == 0 || task.SourceFileID == 0 {
		return "", fmt.Errorf("archive extract payload is incomplete")
	}
	if task.Status == "completed" {
		result, _ := json.Marshal(map[string]interface{}{
			"status":          "completed",
			"execution_mode":  "unified_runner",
			"task_id":         task.TaskID,
			"extracted_files": task.ExtractedFiles,
		})
		return string(result), nil
	}

	task.Status = "processing"
	task.ErrorMessage = ""
	task.CompletedAt = nil
	s.rememberExtractTask(task)
	_ = s.persistExtractTask(ctx, task)

	node := &SelectedNodePayload{ID: task.NodeID, Name: task.NodeName, Type: "master"}
	if s.nodeDispatch != nil {
		selected, err := s.nodeDispatch.SelectExtractArchiveNode()
		if err != nil {
			return "", err
		}
		node = selected
		task.NodeID = selected.ID
		task.NodeName = selected.Name
	}
	if node.Type != "master" {
		failed := s.failExtractTask(task, fmt.Sprintf("node selection failed: remote archive extraction is not supported yet for node %s", node.Name))
		return "", errors.New(failed.ErrorMessage)
	}

	source, err := s.userFiles.GetByIDWithPhysicalFile(task.SourceFileID)
	if err != nil {
		failed := s.failExtractTask(task, fmt.Sprintf("parameter error: source file %d not found: %v", task.SourceFileID, err))
		return "", errors.New(failed.ErrorMessage)
	}
	if source.UserID != userID || source.IsFolder || source.PhysicalFile == nil {
		failed := s.failExtractTask(task, "permission error: source file is not extractable")
		return "", errors.New(failed.ErrorMessage)
	}
	if !strings.EqualFold(filepath.Ext(source.FileName), ".zip") {
		failed := s.failExtractTask(task, "parameter error: only zip archives are supported for extraction")
		return "", errors.New(failed.ErrorMessage)
	}
	if task.TargetFolderID != nil {
		target, err := s.userFiles.GetByID(*task.TargetFolderID)
		if err != nil || target.UserID != userID || !target.IsFolder {
			failed := s.failExtractTask(task, "permission error: target folder not found")
			return "", errors.New(failed.ErrorMessage)
		}
	}
	if s.storage == nil {
		failed := s.failExtractTask(task, "permission error: storage is not initialized")
		return "", errors.New(failed.ErrorMessage)
	}

	reader, err := readPhysicalBlob(s.storage, source.PhysicalFile, s.masterKeys)
	if err != nil {
		failed := s.failExtractTask(task, "permission error: "+err.Error())
		return "", errors.New(failed.ErrorMessage)
	}
	archiveBytes, err := io.ReadAll(reader)
	closeErr := reader.Close()
	if err != nil {
		failed := s.failExtractTask(task, "permission error: "+err.Error())
		return "", errors.New(failed.ErrorMessage)
	}
	if closeErr != nil {
		failed := s.failExtractTask(task, "permission error: "+closeErr.Error())
		return "", errors.New(failed.ErrorMessage)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		failed := s.failExtractTask(task, fmt.Sprintf("parameter error: open zip archive failed: %v", err))
		return "", errors.New(failed.ErrorMessage)
	}

	task.ExtractedFiles = []uint{}
	for _, entry := range zipReader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		fileName := sanitizeExtractedFileName(entry.Name)
		if fileName == "" {
			continue
		}
		extractedID, err := s.extractZipEntry(ctx, userID, task.TargetFolderID, fileName, entry)
		if err != nil {
			failed := s.failExtractTask(task, "permission error: "+err.Error())
			return "", errors.New(failed.ErrorMessage)
		}
		task.ExtractedFiles = append(task.ExtractedFiles, extractedID)
	}

	completedAt := time.Now()
	task.Status = "completed"
	task.ErrorMessage = ""
	task.CompletedAt = &completedAt
	s.rememberExtractTask(task)
	_ = s.persistExtractTask(context.Background(), task)
	result, err := json.Marshal(map[string]interface{}{
		"status":          "completed",
		"execution_mode":  "unified_runner",
		"task_id":         task.TaskID,
		"extracted_files": task.ExtractedFiles,
		"message":         fmt.Sprintf("extracted %d files", len(task.ExtractedFiles)),
	})
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (s *fileSystemRuntimeService) extractZipEntry(ctx context.Context, userID uint, targetFolderID *uint, fileName string, entry *zip.File) (uint, error) {
	entryReader, err := entry.Open()
	if err != nil {
		return 0, fmt.Errorf("open zip entry %s failed: %w", entry.Name, err)
	}
	data, err := io.ReadAll(entryReader)
	closeErr := entryReader.Close()
	if err != nil {
		return 0, fmt.Errorf("read zip entry %s failed: %w", entry.Name, err)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("close zip entry %s failed: %w", entry.Name, closeErr)
	}

	sum := sha256.Sum256(data)
	fileHash := hex.EncodeToString(sum[:])
	storagePath := fmt.Sprintf("files/%s", fileHash)
	if err := s.storage.Save(bytes.NewReader(data), storagePath); err != nil {
		return 0, err
	}

	var physicalID *uint
	if s.db != nil {
		physical := model.PhysicalFile{
			FileHash:    fileHash,
			FileSize:    int64(len(data)),
			StoragePath: storagePath,
			StorageType: "local",
			RefCount:    1,
		}
		if err := s.db.WithContext(ctx).Create(&physical).Error; err != nil {
			return 0, fmt.Errorf("create extracted physical file failed: %w", err)
		}
		physicalID = &physical.ID
	}

	userFile := &model.UserFile{
		UserID:         userID,
		ParentID:       copyUintPtr(targetFolderID),
		FileName:       fileName,
		PhysicalFileID: physicalID,
		IsFolder:       false,
		FileSize:       int64(len(data)),
		FilePath:       fileName,
	}
	if err := s.userFiles.Create(userFile); err != nil {
		return 0, fmt.Errorf("create extracted user file failed: %w", err)
	}
	return userFile.ID, nil
}

func (s *fileSystemRuntimeService) writePackageZip(ctx context.Context, userID uint, fileIDs []uint, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	for _, fileID := range fileIDs {
		select {
		case <-ctx.Done():
			_ = zipWriter.Close()
			return ctx.Err()
		default:
		}

		file, err := s.userFiles.GetByIDWithPhysicalFile(fileID)
		if err != nil {
			_ = zipWriter.Close()
			return err
		}
		if file.UserID != userID || file.IsFolder || file.PhysicalFile == nil {
			_ = zipWriter.Close()
			return fmt.Errorf("file %d is not downloadable", fileID)
		}
		reader, err := readPhysicalBlob(s.storage, file.PhysicalFile, s.masterKeys)
		if err != nil {
			_ = zipWriter.Close()
			return err
		}
		entry, err := zipWriter.Create(s.uniqueZipName(file.FileName, fileID))
		if err != nil {
			_ = reader.Close()
			_ = zipWriter.Close()
			return err
		}
		if _, err := io.Copy(entry, reader); err != nil {
			_ = reader.Close()
			_ = zipWriter.Close()
			return err
		}
		_ = reader.Close()
	}
	return zipWriter.Close()
}

func archivePackageStoragePath(sessionID string) string {
	return fmt.Sprintf("archives/%s.zip", strings.TrimSpace(sessionID))
}

func (s *fileSystemRuntimeService) CreateWOPISession(ctx context.Context, userID uint, fileID uint) (*WOPISessionPayload, error) {
	file, err := s.userFiles.GetByIDWithPhysicalFile(fileID)
	if err != nil {
		return nil, err
	}
	if file.UserID != userID {
		return nil, fmt.Errorf("file is not owned by current user")
	}
	if file.IsFolder {
		return nil, fmt.Errorf("folders cannot create WOPI sessions")
	}

	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}
	sessionID, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	accessToken, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(time.Duration(settings.WOPISessionTTL) * time.Second)
	s.mu.Lock()
	s.wopiSessions[sessionID] = &wopiSession{UserID: userID, FileID: fileID, ExpiresAt: expiresAt, AccessToken: accessToken}
	s.cleanupLocked(time.Now())
	s.mu.Unlock()

	return &WOPISessionPayload{SessionID: sessionID, FileID: fileID, ExpiresAt: expiresAt, TTLSeconds: settings.WOPISessionTTL, AccessToken: accessToken}, nil
}

func (s *fileSystemRuntimeService) GetWOPISession(ctx context.Context, userID uint, sessionID string) (*WOPISessionPayload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(time.Now())
	session, ok := s.wopiSessions[sessionID]
	if !ok || session.UserID != userID {
		return nil, fmt.Errorf("wopi session not found")
	}
	settings, _ := s.currentSettings()
	return &WOPISessionPayload{SessionID: sessionID, FileID: session.FileID, ExpiresAt: session.ExpiresAt, TTLSeconds: settings.WOPISessionTTL, AccessToken: session.AccessToken}, nil
}

func (s *fileSystemRuntimeService) SignSlaveRequest(method, path, bodySHA256 string) (*SlaveSignaturePayload, error) {
	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}
	key, _, err := s.masterKeys.ResolveMasterKey()
	if err != nil {
		return nil, err
	}
	timestamp := time.Now().Unix()
	signature := signSlavePayload(key, method, path, bodySHA256, timestamp)
	return &SlaveSignaturePayload{
		Method:     strings.ToUpper(strings.TrimSpace(method)),
		Path:       strings.TrimSpace(path),
		BodySHA256: strings.TrimSpace(bodySHA256),
		Timestamp:  timestamp,
		Signature:  signature,
		TTLSeconds: settings.SlaveAPISignTTL,
	}, nil
}

func (s *fileSystemRuntimeService) VerifySlaveRequest(method, path, bodySHA256 string, timestamp int64, signature string) (*SlaveSignatureVerifyPayload, error) {
	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}
	result := &SlaveSignatureVerifyPayload{
		Method:     strings.ToUpper(strings.TrimSpace(method)),
		Path:       strings.TrimSpace(path),
		BodySHA256: strings.TrimSpace(bodySHA256),
		Timestamp:  timestamp,
		Signature:  strings.TrimSpace(signature),
	}
	now := time.Now().Unix()
	if timestamp <= 0 || absInt64(now-timestamp) > int64(settings.SlaveAPISignTTL) {
		result.Valid = false
		result.Reason = "signature timestamp expired"
		return result, nil
	}
	key, _, err := s.masterKeys.ResolveMasterKey()
	if err != nil {
		return nil, err
	}
	expected := signSlavePayload(key, method, path, bodySHA256, timestamp)
	result.Valid = hmac.Equal([]byte(expected), []byte(result.Signature))
	if !result.Valid {
		result.Reason = "signature mismatch"
	}
	return result, nil
}

func (s *fileSystemRuntimeService) GetOAuthRefreshStatus() (*OAuthRefreshStatusPayload, error) {
	return s.oauthStatus(false, nil)
}

func (s *fileSystemRuntimeService) RunOAuthRefresh() (*OAuthRefreshStatusPayload, error) {
	now := time.Now()
	s.mu.Lock()
	s.oauthLastRun = &now
	s.mu.Unlock()
	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}
	interval, err := parseEveryDuration(settings.OAuthRefreshInterval)
	if err != nil {
		return nil, err
	}
	var summary *OAuthCredentialRefreshSummary
	if s.oauthTokens != nil {
		summary, err = s.oauthTokens.RefreshDue(context.Background(), interval)
		if err != nil {
			return nil, err
		}
	}
	return s.oauthStatus(true, summary)
}

func (s *fileSystemRuntimeService) OAuthRefreshInterval() (time.Duration, error) {
	settings, err := s.currentSettings()
	if err != nil {
		return 0, err
	}
	return parseEveryDuration(settings.OAuthRefreshInterval)
}

func (s *fileSystemRuntimeService) currentSettings() (*FileSystemSettingPayload, error) {
	setting, err := s.settingsRepo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return defaultFileSystemSettingPayload(), nil
	}
	payload := toFileSystemSettingPayload(setting)
	if payload.MapProvider == "" || payload.MasterKeyStorage == "" || payload.OAuthRefreshInterval == "" {
		defaults := defaultFileSystemSettingPayload()
		if payload.MapProvider == "" {
			payload.MapProvider = defaults.MapProvider
		}
		if payload.MasterKeyStorage == "" {
			payload.MasterKeyStorage = defaults.MasterKeyStorage
		}
		if payload.OAuthRefreshInterval == "" {
			payload.OAuthRefreshInterval = defaults.OAuthRefreshInterval
		}
	}
	return payload, nil
}

func (s *fileSystemRuntimeService) getPackageSession(userID uint, sessionID string) (*packageDownloadSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(time.Now())
	session, ok := s.packageSessions[sessionID]
	if !ok || session.UserID != userID {
		return nil, fmt.Errorf("package download session not found")
	}
	return &packageDownloadSession{UserID: session.UserID, FileIDs: append([]uint{}, session.FileIDs...), NodeID: session.NodeID, NodeName: session.NodeName, DispatchNode: session.DispatchNode, QueueJobID: copyUintPtr(session.QueueJobID), ExpiresAt: session.ExpiresAt}, nil
}

func (s *fileSystemRuntimeService) recordArchiveQueueJob(ctx context.Context, userID uint, resourceID string, fileIDs []uint, capability NodeCapability, node *SelectedNodePayload, status string, lastError string) *model.QueueJob {
	if s.db == nil {
		return nil
	}
	payload, err := json.Marshal(map[string]interface{}{
		"user_id":         userID,
		"file_ids":        fileIDs,
		"session_id":      resourceID,
		"task_id":         resourceID,
		"node_capability": string(capability),
		"execution_mode":  "unified_runner",
		"execution_note":  "archive task is executed by the unified queue runner and controlled by IO queue settings",
	})
	if err != nil {
		payload = []byte("{}")
	}
	now := time.Now()
	jobType := queue.JobTypeArchiveCreate
	resourceType := queue.ResourceTypeArchiveSession
	if capability == NodeCapabilityExtractArchive {
		jobType = queue.JobTypeArchiveExtract
		resourceType = queue.ResourceTypeArchiveExtract
	}
	job := &model.QueueJob{
		QueueKey:       "io",
		JobType:        jobType,
		ResourceType:   resourceType,
		ResourceID:     strings.TrimSpace(resourceID),
		DedupeKey:      fmt.Sprintf("io:%s:%s:%d", resourceType, strings.TrimSpace(resourceID), now.UnixNano()),
		Payload:        string(payload),
		NodeCapability: string(capability),
		Status:         status,
		MaxAttempts:    0,
		ScheduledAt:    now,
		LastError:      strings.TrimSpace(lastError),
	}
	if node != nil {
		job.DispatchNodeID = &node.ID
		job.DispatchNodeName = node.Name
		job.DispatchNodeType = node.Type
	}
	if status == model.QueueJobStatusProcessing {
		job.StartedAt = &now
	}
	if status == model.QueueJobStatusCompleted || status == model.QueueJobStatusFailed {
		job.FinishedAt = &now
	}
	repo := repository.NewQueueJobRepository(s.db)
	created, err := repo.EnqueueIfAbsent(job)
	if err != nil {
		return nil
	}
	return created
}

func (s *fileSystemRuntimeService) markArchiveQueueProcessing(ctx context.Context, jobID *uint) {
	if s.db == nil || jobID == nil {
		return
	}
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *jobID).Updates(map[string]interface{}{
		"status":      model.QueueJobStatusProcessing,
		"started_at":  &now,
		"finished_at": nil,
		"last_error":  "",
	}).Error
}

func (s *fileSystemRuntimeService) markArchiveQueueCompleted(ctx context.Context, jobID *uint, result string) {
	if s.db == nil || jobID == nil {
		return
	}
	now := time.Now()
	resultJSON, err := json.Marshal(map[string]interface{}{
		"status":         "completed",
		"execution_mode": "runtime_record",
		"message":        strings.TrimSpace(result),
	})
	if err != nil {
		resultJSON = []byte(strings.TrimSpace(result))
	}
	_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *jobID).Updates(map[string]interface{}{
		"status":      model.QueueJobStatusCompleted,
		"finished_at": &now,
		"last_error":  "",
		"result":      string(resultJSON),
	}).Error
}

func (s *fileSystemRuntimeService) markArchiveQueueFailed(ctx context.Context, jobID *uint, message string) {
	if s.db == nil || jobID == nil {
		return
	}
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *jobID).Updates(map[string]interface{}{
		"status":      model.QueueJobStatusFailed,
		"finished_at": &now,
		"last_error":  strings.TrimSpace(message),
	}).Error
}

func (s *fileSystemRuntimeService) rememberExtractTask(task *ExtractArchivePayload) {
	if task == nil {
		return
	}
	s.mu.Lock()
	s.extractTasks[task.TaskID] = cloneExtractTask(task)
	s.mu.Unlock()
}

func (s *fileSystemRuntimeService) getExtractTask(ctx context.Context, taskID string) (*ExtractArchivePayload, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("archive extract task id is required")
	}

	s.mu.Lock()
	if task, ok := s.extractTasks[taskID]; ok {
		clone := cloneExtractTask(task)
		s.mu.Unlock()
		return clone, nil
	}
	s.mu.Unlock()

	if s.db == nil {
		return nil, nil
	}
	var row model.ArchiveExtractTask
	if err := s.db.WithContext(ctx).Where("task_id = ?", taskID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	var extracted []uint
	if row.ExtractedFiles != "" {
		_ = json.Unmarshal([]byte(row.ExtractedFiles), &extracted)
	}
	task := &ExtractArchivePayload{
		TaskID:         row.TaskID,
		QueueJobID:     copyUintPtr(row.QueueJobID),
		NodeID:         row.NodeID,
		NodeName:       row.NodeName,
		SourceFileID:   row.SourceFileID,
		TargetFolderID: copyUintPtr(row.TargetFolderID),
		Status:         row.Status,
		ErrorMessage:   row.ErrorMessage,
		ExtractedFiles: extracted,
		CreatedAt:      row.CreatedAt,
		CompletedAt:    row.CompletedAt,
	}
	s.rememberExtractTask(task)
	return cloneExtractTask(task), nil
}

func (s *fileSystemRuntimeService) failExtractTask(task *ExtractArchivePayload, message string) *ExtractArchivePayload {
	if task == nil {
		return nil
	}
	completedAt := time.Now()
	task.Status = "failed"
	task.ErrorMessage = strings.TrimSpace(message)
	task.CompletedAt = &completedAt
	s.rememberExtractTask(task)
	_ = s.persistExtractTask(context.Background(), task)
	s.markArchiveQueueFailed(context.Background(), task.QueueJobID, task.ErrorMessage)
	return cloneExtractTask(task)
}

func (s *fileSystemRuntimeService) persistExtractTask(ctx context.Context, task *ExtractArchivePayload) error {
	if s.db == nil || task == nil {
		return nil
	}
	if err := s.db.WithContext(ctx).AutoMigrate(&model.ArchiveExtractTask{}); err != nil {
		return err
	}
	rawExtracted, err := json.Marshal(task.ExtractedFiles)
	if err != nil {
		return err
	}
	var existing model.ArchiveExtractTask
	err = s.db.WithContext(ctx).Where("task_id = ?", task.TaskID).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	row := model.ArchiveExtractTask{
		TaskID:         task.TaskID,
		QueueJobID:     copyUintPtr(task.QueueJobID),
		NodeCapability: string(NodeCapabilityExtractArchive),
		NodeID:         task.NodeID,
		NodeName:       task.NodeName,
		SourceFileID:   task.SourceFileID,
		TargetFolderID: copyUintPtr(task.TargetFolderID),
		Status:         task.Status,
		ErrorMessage:   task.ErrorMessage,
		ExtractedFiles: string(rawExtracted),
		CompletedAt:    task.CompletedAt,
	}
	if err == gorm.ErrRecordNotFound {
		return s.db.WithContext(ctx).Create(&row).Error
	}
	row.ID = existing.ID
	return s.db.WithContext(ctx).Model(&model.ArchiveExtractTask{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
		"node_id":          row.NodeID,
		"node_name":        row.NodeName,
		"queue_job_id":     row.QueueJobID,
		"node_capability":  row.NodeCapability,
		"source_file_id":   row.SourceFileID,
		"target_folder_id": row.TargetFolderID,
		"status":           row.Status,
		"error_message":    row.ErrorMessage,
		"extracted_files":  row.ExtractedFiles,
		"completed_at":     row.CompletedAt,
	}).Error
}

func cloneExtractTask(task *ExtractArchivePayload) *ExtractArchivePayload {
	if task == nil {
		return nil
	}
	clone := *task
	clone.TargetFolderID = copyUintPtr(task.TargetFolderID)
	clone.ExtractedFiles = append([]uint{}, task.ExtractedFiles...)
	return &clone
}

func copyUintPtr(value *uint) *uint {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func (s *fileSystemRuntimeService) packagePayload(sessionID string, fileIDs []uint, dispatchNode *SelectedNodePayload, expiresAt time.Time, ttl int) *PackageDownloadSessionPayload {
	downloadURL := fmt.Sprintf("/api/v1/file/package-downloads/%s/download", sessionID)
	if s.baseURL != "" {
		downloadURL = s.baseURL + downloadURL
	}
	var nodeID uint
	var nodeName string
	if dispatchNode != nil {
		nodeID = dispatchNode.ID
		nodeName = dispatchNode.Name
	}
	return &PackageDownloadSessionPayload{SessionID: sessionID, FileIDs: append([]uint{}, fileIDs...), NodeID: nodeID, NodeName: nodeName, DispatchNode: dispatchNode, ExpiresAt: expiresAt, TTLSeconds: ttl, DownloadURL: downloadURL}
}

func (s *fileSystemRuntimeService) uniqueZipName(filename string, id uint) string {
	clean := filepath.Base(strings.TrimSpace(filename))
	if clean == "." || clean == "" {
		clean = fmt.Sprintf("file-%d", id)
	}
	return fmt.Sprintf("%d-%s", id, clean)
}

func sanitizeExtractedFileName(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Trim(name, "/")
	if name == "" {
		return ""
	}
	base := filepath.Base(name)
	base = strings.TrimSpace(base)
	if base == "." || base == ".." || base == "" {
		return ""
	}
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	base = replacer.Replace(base)
	if len([]rune(base)) > 255 {
		runes := []rune(base)
		base = string(runes[:255])
	}
	return base
}

func (s *fileSystemRuntimeService) cleanupLocked(now time.Time) {
	for id, session := range s.packageSessions {
		if now.After(session.ExpiresAt) {
			delete(s.packageSessions, id)
		}
	}
	for id, session := range s.wopiSessions {
		if now.After(session.ExpiresAt) {
			delete(s.wopiSessions, id)
		}
	}
}

func (s *fileSystemRuntimeService) oauthStatus(ran bool, summary *OAuthCredentialRefreshSummary) (*OAuthRefreshStatusPayload, error) {
	settings, err := s.currentSettings()
	if err != nil {
		return nil, err
	}
	interval, err := parseEveryDuration(settings.OAuthRefreshInterval)
	if err != nil {
		return nil, err
	}
	var total int64
	var enabled int64
	var credentialTotal int64
	var credentialActive int64
	var credentialDue int64
	if s.db != nil {
		if err := s.db.Model(&model.OAuthApp{}).Count(&total).Error; err != nil && !strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil, err
		}
		if err := s.db.Model(&model.OAuthApp{}).Where("enabled = ?", true).Count(&enabled).Error; err != nil && !strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil, err
		}
	}
	if s.oauthTokens != nil {
		credentialTotal, credentialActive, credentialDue, err = s.oauthTokens.Stats(context.Background(), interval)
		if err != nil {
			return nil, err
		}
	}
	s.mu.Lock()
	lastRun := s.oauthLastRun
	s.mu.Unlock()
	var nextRun *time.Time
	if lastRun != nil {
		next := lastRun.Add(interval)
		nextRun = &next
	}
	message := "oauth refresh scheduler is ready"
	if ran {
		message = "oauth credential refresh pass completed"
	}
	return &OAuthRefreshStatusPayload{
		Interval:          settings.OAuthRefreshInterval,
		IntervalSeconds:   int(interval.Seconds()),
		EnabledApps:       enabled,
		TotalApps:         total,
		ActiveCredentials: credentialActive,
		TotalCredentials:  credentialTotal,
		DueCredentials:    credentialDue,
		LastRunAt:         lastRun,
		NextRunAt:         nextRun,
		Message:           message,
		RefreshSummary:    summary,
	}, nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func signSlavePayload(key []byte, method, path, bodySHA256 string, timestamp int64) string {
	parts := []string{
		strings.ToUpper(strings.TrimSpace(method)),
		strings.TrimSpace(path),
		strings.TrimSpace(bodySHA256),
		strconv.FormatInt(timestamp, 10),
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(mac.Sum(nil))
}

func parseEveryDuration(value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "@every ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "@every "))
	}
	if trimmed == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}
	duration, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return duration, nil
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
