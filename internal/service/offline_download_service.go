package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/token"

	"gorm.io/gorm"
)

type OfflineDownloadService interface {
	List(ctx context.Context, userID uint, status string, keyword string) ([]OfflineDownloadTaskPayload, error)
	Create(ctx context.Context, userID uint, req OfflineDownloadCreateRequest) (*OfflineDownloadTaskPayload, error)
	Refresh(ctx context.Context, userID uint) ([]OfflineDownloadTaskPayload, error)
	Pause(ctx context.Context, userID uint, id uint) (*OfflineDownloadTaskPayload, error)
	Resume(ctx context.Context, userID uint, id uint) (*OfflineDownloadTaskPayload, error)
	Retry(ctx context.Context, userID uint, id uint) (*OfflineDownloadTaskPayload, error)
	ExecuteOfflineDownload(ctx context.Context, taskID uint) (string, error)
	Delete(ctx context.Context, userID uint, id uint) error
	BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error)
}

type offlineDownloadService struct {
	repo         repository.OfflineDownloadTaskRepository
	files        FileService
	db           *gorm.DB
	shares       ShareService
	nodeDispatch NodeDispatchService
	settings     FileSystemSettingService
}

var activeOfflineDownloads sync.Map

func NewOfflineDownloadService(repo repository.OfflineDownloadTaskRepository, files FileService, db *gorm.DB, shares ShareService, nodeDispatch ...NodeDispatchService) OfflineDownloadService {
	var dispatch NodeDispatchService
	if len(nodeDispatch) > 0 {
		dispatch = nodeDispatch[0]
	}
	return &offlineDownloadService{repo: repo, files: files, db: db, shares: shares, nodeDispatch: dispatch}
}

func NewOfflineDownloadServiceWithSettings(repo repository.OfflineDownloadTaskRepository, files FileService, db *gorm.DB, shares ShareService, settings FileSystemSettingService, nodeDispatch ...NodeDispatchService) OfflineDownloadService {
	svc := NewOfflineDownloadService(repo, files, db, shares, nodeDispatch...)
	if concrete, ok := svc.(*offlineDownloadService); ok {
		concrete.settings = settings
	}
	return svc
}

type OfflineDownloadCreateRequest struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	SavePath string `json:"save_path"`
}

type OfflineDownloadBatchDeleteRequest struct {
	IDs []uint `json:"ids"`
}

type OfflineDownloadTaskPayload struct {
	ID              uint                 `json:"id"`
	TaskToken       string               `json:"task_token"`
	Name            string               `json:"name"`
	URL             string               `json:"url"`
	SavePath        string               `json:"save_path"`
	Status          string               `json:"status"`
	Progress        int                  `json:"progress"`
	Speed           string               `json:"speed"`
	Size            string               `json:"size"`
	DownloadedBytes int64                `json:"downloaded_bytes"`
	TotalBytes      int64                `json:"total_bytes"`
	ErrorMessage    string               `json:"error_message"`
	QueueJobID      *uint                `json:"queue_job_id"`
	DispatchNode    *SelectedNodePayload `json:"dispatch_node,omitempty"`
	Downloader      string               `json:"downloader,omitempty"`
	RPCURL          string               `json:"rpc_url,omitempty"`
	TaskOptions     string               `json:"task_options,omitempty"`
	TempDir         string               `json:"temp_dir,omitempty"`
	RefreshInterval int                  `json:"refresh_interval,omitempty"`
	WaitForSeeding  bool                 `json:"wait_for_seeding,omitempty"`
	RemoteTaskID    string               `json:"remote_task_id,omitempty"`
	SavedFileID     *uint                `json:"saved_file_id"`
	SavedFolderID   *uint                `json:"saved_folder_id"`
	ExpiresAt       *time.Time           `json:"expires_at"`
	TTLSeconds      int                  `json:"ttl_seconds"`
	Expired         bool                 `json:"expired"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	CompletedAt     *time.Time           `json:"completed_at"`
}

func (s *offlineDownloadService) List(ctx context.Context, userID uint, status string, keyword string) ([]OfflineDownloadTaskPayload, error) {
	status = strings.TrimSpace(status)
	if status != "" && !isOfflineTaskStatus(status) {
		return nil, fmt.Errorf("invalid offline download status")
	}
	tasks, err := s.repo.ListByUser(ctx, repository.OfflineDownloadTaskListFilter{
		UserID:  userID,
		Status:  status,
		Keyword: strings.TrimSpace(keyword),
	})
	if err != nil {
		return nil, err
	}
	s.expireOfflineTasks(ctx, tasks)
	for i := range tasks {
		if tasks[i].Status == model.OfflineTaskStatusQueued || tasks[i].Status == model.OfflineTaskStatusDownloading {
			if tasks[i].QueueJobID == nil {
				go s.runDownload(context.Background(), tasks[i].ID)
			}
			continue
		}
		if tasks[i].Status == model.OfflineTaskStatusCompleted && tasks[i].SavedFileID == nil {
			s.resolveSavedFile(ctx, &tasks[i])
		}
	}
	return s.offlineTaskPayloads(tasks), nil
}

func (s *offlineDownloadService) Create(ctx context.Context, userID uint, req OfflineDownloadCreateRequest) (*OfflineDownloadTaskPayload, error) {
	normalized, err := normalizeOfflineDownloadCreate(req)
	if err != nil {
		return nil, err
	}

	taskToken, err := token.GenerateShareTokenWithLength(18)
	if err != nil {
		return nil, err
	}
	taskToken = strings.TrimRight(taskToken, "=")

	var dispatchNode *SelectedNodePayload
	if s.nodeDispatch != nil {
		dispatchNode, err = s.nodeDispatch.SelectOfflineDownloadNode()
		if err != nil {
			s.recordOfflineQueueJob(ctx, userID, taskToken, normalized, nil, model.QueueJobStatusFailed, "node selection failed: "+err.Error())
			return nil, err
		}
	} else {
		s.recordOfflineQueueJob(ctx, userID, taskToken, normalized, nil, model.QueueJobStatusFailed, "node selection failed: no enabled node supports offline_download")
		return nil, fmt.Errorf("no enabled node supports offline_download")
	}
	if !isSupportedOfflineDownloader(dispatchNode.Downloader) {
		s.recordOfflineQueueJob(ctx, userID, taskToken, normalized, dispatchNode, model.QueueJobStatusFailed, fmt.Sprintf("parameter error: downloader %s does not support submitting offline download tasks yet", dispatchNode.Downloader))
		return nil, fmt.Errorf("%s does not support submitting offline download tasks yet", dispatchNode.Downloader)
	}

	task := &model.OfflineDownloadTask{
		UserID:    userID,
		TaskToken: taskToken,
		Name:      normalized.Name,
		SourceURL: normalized.URL,
		SavePath:  normalized.SavePath,
		Status:    model.OfflineTaskStatusQueued,
		Progress:  0,
		SpeedText: "waiting for scheduler",
		SizeText:  "unknown",
	}
	assignOfflineTaskDispatchNode(task, dispatchNode)
	assignOfflineTaskRuntimeSnapshot(task, dispatchNode)
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	if job := s.recordOfflineQueueJob(ctx, userID, fmt.Sprintf("%d", task.ID), normalized, dispatchNode, model.QueueJobStatusPending, ""); job != nil {
		task.QueueJobID = &job.ID
		_ = s.repo.Save(ctx, task)
	}
	if task.QueueJobID == nil {
		go s.runDownload(context.Background(), task.ID)
	}
	payload := s.offlineTaskPayload(task)
	return &payload, nil
}

func (s *offlineDownloadService) Refresh(ctx context.Context, userID uint) ([]OfflineDownloadTaskPayload, error) {
	tasks, err := s.repo.ListByUser(ctx, repository.OfflineDownloadTaskListFilter{UserID: userID})
	if err != nil {
		return nil, err
	}
	s.expireOfflineTasks(ctx, tasks)
	for i := range tasks {
		if tasks[i].Status == model.OfflineTaskStatusCompleted && tasks[i].SavedFileID == nil {
			s.resolveSavedFile(ctx, &tasks[i])
		}
	}

	return s.offlineTaskPayloads(tasks), nil
}

func (s *offlineDownloadService) Pause(ctx context.Context, userID uint, id uint) (*OfflineDownloadTaskPayload, error) {
	task, err := s.repo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if s.expireOfflineTask(ctx, task, time.Now()) {
		payload := s.offlineTaskPayload(task)
		return &payload, nil
	}
	if task.Status == model.OfflineTaskStatusCompleted {
		return nil, fmt.Errorf("completed task cannot be paused")
	}
	if task.Status == model.OfflineTaskStatusFailed {
		return nil, fmt.Errorf("failed task should be retried")
	}
	if task.RemoteTaskID != "" {
		if err := newOfflineDownloader(task.Downloader).Pause(ctx, offlineRuntimeFromTask(task)); err != nil {
			return nil, err
		}
	}
	task.Status = model.OfflineTaskStatusPaused
	task.SpeedText = "paused"
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	payload := s.offlineTaskPayload(task)
	return &payload, nil
}

func (s *offlineDownloadService) Resume(ctx context.Context, userID uint, id uint) (*OfflineDownloadTaskPayload, error) {
	task, err := s.repo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if s.expireOfflineTask(ctx, task, time.Now()) {
		payload := s.offlineTaskPayload(task)
		return &payload, nil
	}
	if task.Status == model.OfflineTaskStatusCompleted {
		return nil, fmt.Errorf("completed task cannot be resumed")
	}
	if task.RemoteTaskID != "" {
		if err := newOfflineDownloader(task.Downloader).Resume(ctx, offlineRuntimeFromTask(task)); err != nil {
			return nil, err
		}
	}
	task.Status = model.OfflineTaskStatusDownloading
	task.SpeedText = "waiting for scheduler"
	task.ErrorMessage = ""
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	s.requeueOfflineQueue(ctx, task)
	if task.QueueJobID == nil {
		go s.runDownload(context.Background(), task.ID)
	}
	payload := s.offlineTaskPayload(task)
	return &payload, nil
}

func (s *offlineDownloadService) Retry(ctx context.Context, userID uint, id uint) (*OfflineDownloadTaskPayload, error) {
	task, err := s.repo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if s.expireOfflineTask(ctx, task, time.Now()) {
		payload := s.offlineTaskPayload(task)
		return &payload, nil
	}
	task.Status = model.OfflineTaskStatusQueued
	if task.Progress >= 100 {
		task.Progress = 0
	}
	task.SpeedText = "waiting for scheduler"
	task.ErrorMessage = ""
	task.CompletedAt = nil
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	s.requeueOfflineQueue(ctx, task)
	if task.QueueJobID == nil {
		go s.runDownload(context.Background(), task.ID)
	}
	payload := s.offlineTaskPayload(task)
	return &payload, nil
}

var errOfflineDownloadPaused = errors.New("offline download paused")

type progressReadCloser struct {
	source       io.ReadCloser
	taskID       uint
	userID       uint
	repo         repository.OfflineDownloadTaskRepository
	total        int64
	downloaded   int64
	startedAt    time.Time
	lastUpdateAt time.Time
	mu           sync.Mutex
}

func (r *progressReadCloser) Read(p []byte) (int, error) {
	if time.Since(r.lastUpdateAt) >= time.Second {
		if err := r.checkPaused(); err != nil {
			return 0, err
		}
	}
	n, err := r.source.Read(p)
	if n > 0 {
		r.downloaded += int64(n)
		r.maybeUpdate(false)
	}
	if err == io.EOF {
		r.maybeUpdate(true)
	}
	return n, err
}

func (r *progressReadCloser) Close() error {
	return r.source.Close()
}

func (r *progressReadCloser) maybeUpdate(force bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if !force && now.Sub(r.lastUpdateAt) < time.Second {
		return
	}
	r.lastUpdateAt = now

	task, err := r.repo.GetByIDForUser(context.Background(), r.userID, r.taskID)
	if err != nil {
		return
	}
	if task.Status == model.OfflineTaskStatusPaused {
		return
	}
	task.DownloadedBytes = r.downloaded
	if r.total > 0 {
		task.TotalBytes = r.total
		task.Progress = int((r.downloaded * 100) / r.total)
		if task.Progress > 99 && !force {
			task.Progress = 99
		}
	}
	elapsed := now.Sub(r.startedAt).Seconds()
	if elapsed > 0 {
		task.SpeedText = formatBytes(int64(float64(r.downloaded)/elapsed)) + "/s"
	}
	task.SizeText = humanSize("", task.TotalBytes)
	_ = r.repo.Save(context.Background(), task)
}

func (r *progressReadCloser) checkPaused() error {
	task, err := r.repo.GetByIDForUser(context.Background(), r.userID, r.taskID)
	if err != nil {
		return err
	}
	if task.Status == model.OfflineTaskStatusPaused {
		return errOfflineDownloadPaused
	}
	return nil
}

func (s *offlineDownloadService) runDownload(ctx context.Context, taskID uint) {
	_, _ = s.ExecuteOfflineDownload(ctx, taskID)
}

// ExecuteOfflineDownload runs one offline download task for the unified queue worker.
func (s *offlineDownloadService) ExecuteOfflineDownload(ctx context.Context, taskID uint) (string, error) {
	if _, loaded := activeOfflineDownloads.LoadOrStore(taskID, struct{}{}); loaded {
		return "", fmt.Errorf("offline download task is already running")
	}
	defer activeOfflineDownloads.Delete(taskID)

	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return "", err
	}
	if task.Status == model.OfflineTaskStatusPaused || task.Status == model.OfflineTaskStatusCompleted {
		return fmt.Sprintf(`{"status":"skipped","task_status":"%s"}`, task.Status), nil
	}
	if s.expireOfflineTask(ctx, task, time.Now()) {
		s.markOfflineQueueFailed(ctx, task, "offline download task expired")
		return "", fmt.Errorf("offline download task expired")
	}

	task.Status = model.OfflineTaskStatusDownloading
	task.SpeedText = "submitting"
	task.ErrorMessage = ""
	_ = s.repo.Save(ctx, task)
	s.markOfflineQueueProcessing(ctx, task)

	downloader := newOfflineDownloader(task.Downloader)
	runtime := offlineRuntimeFromTask(task)
	if task.RemoteTaskID == "" {
		remoteTaskID, err := downloader.Submit(ctx, runtime)
		if err != nil {
			s.failTask(ctx, task, "rpc failed: "+err.Error())
			return "", err
		}
		task.RemoteTaskID = remoteTaskID
		task.SpeedText = "waiting for scheduler"
		if err := s.repo.Save(ctx, task); err != nil {
			return "", err
		}
		runtime.RemoteTaskID = remoteTaskID
	}

	return s.pollRemoteDownload(ctx, task.ID)
}

func (s *offlineDownloadService) pollRemoteDownload(ctx context.Context, taskID uint) (string, error) {
	for {
		task, err := s.repo.GetByID(ctx, taskID)
		if err != nil {
			return "", err
		}
		if task.Status == model.OfflineTaskStatusPaused || task.Status == model.OfflineTaskStatusCompleted || task.Status == model.OfflineTaskStatusFailed || task.Status == model.OfflineTaskStatusExpired {
			return fmt.Sprintf(`{"status":"skipped","task_status":"%s"}`, task.Status), nil
		}
		if s.expireOfflineTask(ctx, task, time.Now()) {
			s.markOfflineQueueFailed(ctx, task, "offline download task expired")
			return "", fmt.Errorf("offline download task expired")
		}
		status, err := newOfflineDownloader(task.Downloader).Status(ctx, offlineRuntimeFromTask(task))
		if err != nil {
			s.failTask(ctx, task, "rpc failed: "+err.Error())
			return "", err
		}

		task.DownloadedBytes = status.DownloadedBytes
		task.TotalBytes = status.TotalBytes
		if task.TotalBytes > 0 {
			task.Progress = int((task.DownloadedBytes * 100) / task.TotalBytes)
			if task.Progress > 100 {
				task.Progress = 100
			}
		}
		task.SizeText = humanSize("", task.TotalBytes)
		task.SpeedText = status.SpeedText
		if status.WaitingSeeding {
			task.SpeedText = "seeding"
		}
		if status.Completed {
			completedAt := time.Now()
			task.Status = model.OfflineTaskStatusCompleted
			task.Progress = 100
			task.SpeedText = "completed"
			task.ErrorMessage = ""
			task.CompletedAt = &completedAt
			_ = s.repo.Save(ctx, task)
			s.markOfflineQueueCompleted(ctx, task, "offline download completed")
			return fmt.Sprintf(`{"status":"completed","remote_task_id":%q,"downloaded_bytes":%d,"total_bytes":%d}`, task.RemoteTaskID, task.DownloadedBytes, task.TotalBytes), nil
		}
		if err := s.repo.Save(ctx, task); err != nil {
			return "", err
		}

		interval := time.Duration(clampInt(task.RefreshInterval, 1, 3600)) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (s *offlineDownloadService) runLegacyHTTPDownload(ctx context.Context, task *model.OfflineDownloadTask) {
	s.failTask(ctx, task, "local HTTP offline download is disabled; configure an enabled offline_download node")
}

func (s *offlineDownloadService) runShareDownload(ctx context.Context, task *model.OfflineDownloadTask, shareToken string) {
	if shareToken == "" {
		return
	}
	reader, fileName, _, err := s.shares.DownloadShare(ctx, shareToken)
	if err != nil {
		s.failTask(ctx, task, err.Error())
		return
	}
	defer reader.Close()

	fileName = sanitizeOfflineFileName(fileName)
	if fileName == "" {
		fileName = sanitizeOfflineFileName(task.Name)
	}
	if fileName == "" {
		fileName = inferOfflineTaskName(task.SourceURL)
	}

	task.Name = fileName
	task.TotalBytes = 0
	task.SizeText = "unknown"
	_ = s.repo.Save(ctx, task)

	parentID, err := s.ensureSaveFolder(ctx, task.UserID, task.SavePath)
	if err != nil {
		s.failTask(ctx, task, err.Error())
		return
	}

	progressReader := &progressReadCloser{
		source:       reader,
		taskID:       task.ID,
		userID:       task.UserID,
		repo:         s.repo,
		total:        0,
		startedAt:    time.Now(),
		lastUpdateAt: time.Now().Add(-time.Second),
	}
	if err := progressReader.checkPaused(); err != nil {
		return
	}
	savedFile, err := s.files.Upload(task.UserID, fileName, 0, progressReader, parentID)
	if err != nil {
		current, currentErr := s.repo.GetByID(ctx, task.ID)
		if currentErr == nil && current.Status == model.OfflineTaskStatusPaused {
			return
		}
		s.failTask(ctx, task, err.Error())
		return
	}

	completedAt := time.Now()
	current, err := s.repo.GetByID(ctx, task.ID)
	if err != nil {
		return
	}
	current.Name = fileName
	current.Status = model.OfflineTaskStatusCompleted
	current.Progress = 100
	current.SpeedText = "completed"
	current.DownloadedBytes = progressReader.downloaded
	if savedFile != nil {
		current.SavedFileID = &savedFile.ID
		current.SavedFolderID = savedFile.ParentID
	}
	current.TotalBytes = progressReader.downloaded
	current.SizeText = humanSize("", current.TotalBytes)
	current.ErrorMessage = ""
	current.CompletedAt = &completedAt
	_ = s.repo.Save(ctx, current)
}

func (s *offlineDownloadService) failTask(ctx context.Context, task *model.OfflineDownloadTask, message string) {
	current, err := s.repo.GetByID(ctx, task.ID)
	if err != nil {
		current = task
	}
	if current.Status == model.OfflineTaskStatusPaused {
		return
	}
	current.Status = model.OfflineTaskStatusFailed
	current.SpeedText = "failed"
	current.ErrorMessage = strings.TrimSpace(message)
	_ = s.repo.Save(ctx, current)
	s.markOfflineQueueFailed(ctx, current, current.ErrorMessage)
}

func (s *offlineDownloadService) recordOfflineQueueJob(ctx context.Context, userID uint, resourceID string, req OfflineDownloadCreateRequest, node *SelectedNodePayload, status string, lastError string) *model.QueueJob {
	if s.db == nil {
		return nil
	}
	now := time.Now()
	var job *model.QueueJob

	parsedTaskID, parseErr := strconv.ParseUint(strings.TrimSpace(resourceID), 10, 32)
	if parseErr == nil && status == model.QueueJobStatusPending {
		setting, err := queue.ResolveSetting(repository.NewQueueSettingRepository(s.db), string(queue.KeyOffline))
		if err != nil {
			if fallback, ok := queue.DefaultSettingByKey(string(queue.KeyOffline)); ok {
				setting = fallback
			}
		}
		built, err := queue.BuildOfflineDownloadJob(queue.OfflineDownloadPayload{TaskID: uint(parsedTaskID)}, setting)
		if err == nil {
			job = built
		}
	}

	if job == nil {
		payload, err := json.Marshal(map[string]interface{}{
			"user_id":   userID,
			"url":       req.URL,
			"name":      req.Name,
			"save_path": req.SavePath,
		})
		if err != nil {
			payload = []byte("{}")
		}
		job = &model.QueueJob{
			QueueKey:       string(queue.KeyOffline),
			JobType:        queue.JobTypeOfflineDownload,
			ResourceType:   queue.ResourceTypeOfflineTask,
			ResourceID:     strings.TrimSpace(resourceID),
			DedupeKey:      fmt.Sprintf("%s:%s:%s:%d", queue.KeyOffline, queue.ResourceTypeOfflineTask, strings.TrimSpace(resourceID), now.UnixNano()),
			Payload:        string(payload),
			NodeCapability: string(NodeCapabilityOfflineDownload),
			Status:         status,
			MaxAttempts:    0,
			ScheduledAt:    now,
			LastError:      strings.TrimSpace(lastError),
		}
	}
	job.Status = status
	job.LastError = strings.TrimSpace(lastError)
	if status == model.QueueJobStatusFailed || status == model.QueueJobStatusCompleted {
		job.FinishedAt = &now
	}
	if status == model.QueueJobStatusProcessing {
		job.StartedAt = &now
	}
	if node != nil {
		job.DispatchNodeID = &node.ID
		job.DispatchNodeName = node.Name
		job.DispatchNodeType = node.Type
	}
	repo := repository.NewQueueJobRepository(s.db)
	created, err := repo.EnqueueIfAbsent(job)
	if err != nil {
		return nil
	}
	return created
}

func (s *offlineDownloadService) requeueOfflineQueue(ctx context.Context, task *model.OfflineDownloadTask) {
	if s.db == nil || task == nil {
		return
	}
	now := time.Now()
	if task.QueueJobID != nil {
		_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *task.QueueJobID).Updates(map[string]interface{}{
			"status":       model.QueueJobStatusPending,
			"attempts":     0,
			"scheduled_at": now,
			"started_at":   nil,
			"finished_at":  nil,
			"last_error":   "",
			"result":       "",
		}).Error
		return
	}
	req := OfflineDownloadCreateRequest{URL: task.SourceURL, Name: task.Name, SavePath: task.SavePath}
	if job := s.recordOfflineQueueJob(ctx, task.UserID, fmt.Sprintf("%d", task.ID), req, offlineTaskDispatchNode(task), model.QueueJobStatusPending, ""); job != nil {
		task.QueueJobID = &job.ID
		_ = s.repo.Save(ctx, task)
	}
}

func (s *offlineDownloadService) markOfflineQueueProcessing(ctx context.Context, task *model.OfflineDownloadTask) {
	if s.db == nil || task == nil || task.QueueJobID == nil {
		return
	}
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *task.QueueJobID).Updates(map[string]interface{}{
		"status":      model.QueueJobStatusProcessing,
		"started_at":  &now,
		"last_error":  "",
		"finished_at": nil,
	}).Error
}

func (s *offlineDownloadService) markOfflineQueueCompleted(ctx context.Context, task *model.OfflineDownloadTask, result string) {
	if s.db == nil || task == nil || task.QueueJobID == nil {
		return
	}
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *task.QueueJobID).Updates(map[string]interface{}{
		"status":      model.QueueJobStatusCompleted,
		"finished_at": &now,
		"last_error":  "",
		"result":      result,
	}).Error
}

func (s *offlineDownloadService) markOfflineQueueFailed(ctx context.Context, task *model.OfflineDownloadTask, message string) {
	if s.db == nil || task == nil || task.QueueJobID == nil {
		return
	}
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&model.QueueJob{}).Where("id = ?", *task.QueueJobID).Updates(map[string]interface{}{
		"status":      model.QueueJobStatusFailed,
		"finished_at": &now,
		"last_error":  strings.TrimSpace(message),
	}).Error
}

func (s *offlineDownloadService) ensureSaveFolder(ctx context.Context, userID uint, savePath string) (*uint, error) {
	if s.db == nil {
		return nil, nil
	}
	segments := splitSavePath(savePath)
	var parentID *uint
	for _, segment := range segments {
		var folder model.UserFile
		query := s.db.WithContext(ctx).Where("user_id = ? AND file_name = ? AND is_folder = ?", userID, segment, true)
		if parentID == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}
		err := query.First(&folder).Error
		if err == nil {
			id := folder.ID
			parentID = &id
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		folder = model.UserFile{
			UserID:   userID,
			FileName: segment,
			IsFolder: true,
			ParentID: parentID,
		}
		if err := s.db.WithContext(ctx).Create(&folder).Error; err != nil {
			return nil, err
		}
		id := folder.ID
		parentID = &id
	}
	return parentID, nil
}

func (s *offlineDownloadService) findSaveFolder(ctx context.Context, userID uint, savePath string) (*uint, error) {
	if s.db == nil {
		return nil, nil
	}
	segments := splitSavePath(savePath)
	var parentID *uint
	for _, segment := range segments {
		var folder model.UserFile
		query := s.db.WithContext(ctx).Where("user_id = ? AND file_name = ? AND is_folder = ?", userID, segment, true)
		if parentID == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}
		if err := query.First(&folder).Error; err != nil {
			return nil, err
		}
		id := folder.ID
		parentID = &id
	}
	return parentID, nil
}

func (s *offlineDownloadService) resolveSavedFile(ctx context.Context, task *model.OfflineDownloadTask) {
	if s.db == nil || task == nil || task.Name == "" {
		return
	}
	parentID, err := s.findSaveFolder(ctx, task.UserID, task.SavePath)
	if err != nil {
		return
	}
	var file model.UserFile
	query := s.db.WithContext(ctx).
		Where("user_id = ? AND file_name = ? AND is_folder = ?", task.UserID, task.Name, false)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if err := query.Order("updated_at DESC, id DESC").First(&file).Error; err != nil {
		return
	}
	task.SavedFileID = &file.ID
	task.SavedFolderID = file.ParentID
	_ = s.repo.Save(ctx, task)
}

func (s *offlineDownloadService) Delete(ctx context.Context, userID uint, id uint) error {
	task, err := s.repo.GetByIDForUser(ctx, userID, id)
	if err == nil && task.RemoteTaskID != "" {
		if deleteErr := newOfflineDownloader(task.Downloader).Delete(ctx, offlineRuntimeFromTask(task)); deleteErr != nil {
			return deleteErr
		}
	}
	return s.repo.Delete(ctx, userID, id)
}

func (s *offlineDownloadService) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	return s.repo.BatchDelete(ctx, userID, ids)
}

func normalizeOfflineDownloadCreate(req OfflineDownloadCreateRequest) (OfflineDownloadCreateRequest, error) {
	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		return req, fmt.Errorf("download url cannot be empty")
	}
	parsed, err := url.Parse(req.URL)
	if err != nil || parsed.Scheme == "" {
		return req, fmt.Errorf("invalid download url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "magnet" {
		return req, fmt.Errorf("only http, https and magnet urls are supported")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = inferOfflineTaskName(req.URL)
	}
	if len([]rune(req.Name)) > 255 {
		return req, fmt.Errorf("task name cannot exceed 255 characters")
	}

	req.SavePath = strings.TrimSpace(req.SavePath)
	if req.SavePath == "" {
		req.SavePath = "/offline-downloads"
	}
	req.SavePath = strings.ReplaceAll(req.SavePath, "\\", "/")
	if !strings.HasPrefix(req.SavePath, "/") {
		req.SavePath = "/" + req.SavePath
	}
	req.SavePath = path.Clean(req.SavePath)
	if req.SavePath == "." {
		req.SavePath = "/offline-downloads"
	}
	return req, nil
}

func splitSavePath(savePath string) []string {
	savePath = strings.TrimSpace(strings.ReplaceAll(savePath, "\\", "/"))
	savePath = strings.Trim(savePath, "/")
	if savePath == "" {
		return nil
	}
	rawSegments := strings.Split(savePath, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		segment = sanitizeOfflineFileName(segment)
		if segment == "" || segment == "." || segment == ".." {
			continue
		}
		segments = append(segments, segment)
	}
	return segments
}

func sanitizeOfflineFileName(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)
	if len([]rune(name)) > 255 {
		runes := []rune(name)
		name = string(runes[:255])
	}
	return name
}

func inferOfflineTaskName(rawURL string) string {
	if strings.HasPrefix(strings.ToLower(rawURL), "magnet:") {
		return "magnet task"
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "offline download task"
	}
	name := path.Base(parsed.Path)
	if name == "." || name == "/" || strings.TrimSpace(name) == "" {
		return "offline download task"
	}
	if decoded, err := url.PathUnescape(name); err == nil && strings.TrimSpace(decoded) != "" {
		return decoded
	}
	return name
}

func extractInternalShareToken(rawURL string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", false
	}
	cleanPath := path.Clean("/" + strings.TrimSpace(parsed.Path))
	segments := strings.Split(strings.Trim(cleanPath, "/"), "/")
	if len(segments) >= 2 && segments[0] == "s" {
		token := strings.TrimSpace(segments[1])
		return token, token != ""
	}
	if len(segments) >= 5 && segments[0] == "api" && segments[1] == "v1" && segments[2] == "shares" && segments[4] == "download" {
		token := strings.TrimSpace(segments[3])
		return token, token != ""
	}
	return "", false
}

func isOfflineTaskStatus(status string) bool {
	switch status {
	case model.OfflineTaskStatusQueued, model.OfflineTaskStatusDownloading, model.OfflineTaskStatusPaused, model.OfflineTaskStatusCompleted, model.OfflineTaskStatusFailed, model.OfflineTaskStatusExpired:
		return true
	default:
		return false
	}
}

func normalizeSizeText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "unknown" {
		return "unknown"
	}
	return value
}

func (s *offlineDownloadService) offlineTaskPayloads(tasks []model.OfflineDownloadTask) []OfflineDownloadTaskPayload {
	payloads := make([]OfflineDownloadTaskPayload, 0, len(tasks))
	for i := range tasks {
		payloads = append(payloads, s.offlineTaskPayload(&tasks[i]))
	}
	return payloads
}

func (s *offlineDownloadService) offlineTaskPayload(task *model.OfflineDownloadTask) OfflineDownloadTaskPayload {
	ttlSeconds := s.offlineTTLSeconds()
	expiresAt := offlineTaskExpiresAt(task, ttlSeconds)
	expired := offlineTaskExpiredAt(task, ttlSeconds, time.Now())
	return OfflineDownloadTaskPayload{
		ID:              task.ID,
		TaskToken:       task.TaskToken,
		Name:            task.Name,
		URL:             task.SourceURL,
		SavePath:        task.SavePath,
		Status:          task.Status,
		Progress:        task.Progress,
		Speed:           humanSpeed(task.SpeedText),
		Size:            humanSize(task.SizeText, task.TotalBytes),
		DownloadedBytes: task.DownloadedBytes,
		TotalBytes:      task.TotalBytes,
		ErrorMessage:    task.ErrorMessage,
		QueueJobID:      task.QueueJobID,
		DispatchNode:    offlineTaskDispatchNode(task),
		Downloader:      task.Downloader,
		RPCURL:          task.RPCURL,
		TaskOptions:     task.TaskOptions,
		TempDir:         task.TempDir,
		RefreshInterval: task.RefreshInterval,
		WaitForSeeding:  task.WaitForSeeding,
		RemoteTaskID:    task.RemoteTaskID,
		SavedFileID:     task.SavedFileID,
		SavedFolderID:   task.SavedFolderID,
		ExpiresAt:       expiresAt,
		TTLSeconds:      ttlSeconds,
		Expired:         expired,
		CreatedAt:       task.CreatedAt,
		UpdatedAt:       task.UpdatedAt,
		CompletedAt:     task.CompletedAt,
	}
}

func (s *offlineDownloadService) offlineTTLSeconds() int {
	if s.settings == nil {
		return defaultFileSystemSettingPayload().OfflineTTL
	}
	settings, err := s.settings.Get()
	if err != nil || settings == nil {
		return defaultFileSystemSettingPayload().OfflineTTL
	}
	return clampInt(settings.OfflineTTL, 0, 31536000)
}

func (s *offlineDownloadService) expireOfflineTasks(ctx context.Context, tasks []model.OfflineDownloadTask) {
	now := time.Now()
	for i := range tasks {
		if s.expireOfflineTask(ctx, &tasks[i], now) {
			tasks[i].Status = model.OfflineTaskStatusExpired
		}
	}
}

func (s *offlineDownloadService) expireOfflineTask(ctx context.Context, task *model.OfflineDownloadTask, now time.Time) bool {
	if task == nil || task.Status == model.OfflineTaskStatusExpired {
		return false
	}
	ttlSeconds := s.offlineTTLSeconds()
	if !offlineTaskExpiredAt(task, ttlSeconds, now) {
		return false
	}
	task.Status = model.OfflineTaskStatusExpired
	task.SpeedText = "expired"
	if strings.TrimSpace(task.ErrorMessage) == "" {
		task.ErrorMessage = "offline download task expired"
	}
	_ = s.repo.Save(ctx, task)
	return true
}

func offlineTaskExpiresAt(task *model.OfflineDownloadTask, ttlSeconds int) *time.Time {
	if task == nil || ttlSeconds <= 0 || task.CreatedAt.IsZero() {
		return nil
	}
	expiresAt := task.CreatedAt.Add(time.Duration(ttlSeconds) * time.Second)
	return &expiresAt
}

func offlineTaskExpiredAt(task *model.OfflineDownloadTask, ttlSeconds int, now time.Time) bool {
	expiresAt := offlineTaskExpiresAt(task, ttlSeconds)
	return expiresAt != nil && !now.Before(*expiresAt)
}

func assignOfflineTaskDispatchNode(task *model.OfflineDownloadTask, node *SelectedNodePayload) {
	if task == nil || node == nil {
		return
	}
	task.DispatchNodeID = &node.ID
	task.DispatchNodeName = node.Name
	task.DispatchNodeType = node.Type
}

func assignOfflineTaskRuntimeSnapshot(task *model.OfflineDownloadTask, node *SelectedNodePayload) {
	if task == nil || node == nil {
		return
	}
	task.NodeID = &node.ID
	task.Downloader = node.Downloader
	task.RPCURL = node.RPCURL
	task.RPCSecret = node.rpcSecret
	task.TaskOptions = node.TaskOptions
	task.TempDir = node.TempDir
	task.RefreshInterval = clampInt(node.RefreshInterval, 1, 3600)
	task.WaitForSeeding = node.WaitForSeeding
}

func offlineRuntimeFromTask(task *model.OfflineDownloadTask) *modelOfflineRuntime {
	return &modelOfflineRuntime{
		URL:            task.SourceURL,
		Downloader:     task.Downloader,
		RPCURL:         task.RPCURL,
		RPCSecret:      task.RPCSecret,
		TaskOptions:    task.TaskOptions,
		TempDir:        task.TempDir,
		RemoteTaskID:   task.RemoteTaskID,
		WaitForSeeding: task.WaitForSeeding,
	}
}

func isSupportedOfflineDownloader(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "Aria2") || strings.TrimSpace(name) == ""
}

func offlineTaskDispatchNode(task *model.OfflineDownloadTask) *SelectedNodePayload {
	if task == nil || task.DispatchNodeID == nil {
		return nil
	}
	return &SelectedNodePayload{
		ID:              *task.DispatchNodeID,
		Name:            task.DispatchNodeName,
		Type:            task.DispatchNodeType,
		Downloader:      task.Downloader,
		RPCURL:          task.RPCURL,
		TaskOptions:     task.TaskOptions,
		TempDir:         task.TempDir,
		RefreshInterval: task.RefreshInterval,
		WaitForSeeding:  task.WaitForSeeding,
	}
}

func humanSpeed(value string) string {
	switch strings.TrimSpace(value) {
	case "", "waiting":
		return "等待调度"
	case "waiting for scheduler":
		return "等待调度"
	case "connecting":
		return "连接中"
	case "paused":
		return "已暂停"
	case "completed":
		return "已完成"
	case "merging":
		return "等待合并"
	case "failed":
		return "失败"
	default:
		return value
	}
}

func humanSize(value string, totalBytes int64) string {
	if totalBytes > 0 {
		return formatBytes(totalBytes)
	}
	switch strings.TrimSpace(value) {
	case "", "unknown":
		return "未知"
	default:
		return value
	}
}
func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(bytes)
	unit := "B"
	for _, item := range units {
		value = value / 1024
		unit = item
		if value < 1024 {
			break
		}
	}
	return fmt.Sprintf("%.1f %s", value, unit)
}

func ParseOfflineDownloadTaskID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid offline download task id")
	}
	return uint(parsed), nil
}
