package service

import (
	"fmt"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
)

// QueueJobListItemPayload is the API-facing shape for one queue job row.
type QueueJobListItemPayload struct {
	ID             uint                 `json:"id"`
	QueueKey       string               `json:"queue_key"`
	JobType        string               `json:"job_type"`
	ResourceType   string               `json:"resource_type"`
	ResourceID     string               `json:"resource_id"`
	DispatchNode   *SelectedNodePayload `json:"dispatch_node,omitempty"`
	NodeCapability string               `json:"node_capability"`
	ExecutionMode  string               `json:"execution_mode"`
	ExecutionNote  string               `json:"execution_note"`
	Status         string               `json:"status"`
	Attempts       int                  `json:"attempts"`
	MaxAttempts    int                  `json:"max_attempts"`
	ScheduledAt    time.Time            `json:"scheduled_at"`
	StartedAt      *time.Time           `json:"started_at"`
	FinishedAt     *time.Time           `json:"finished_at"`
	LastError      string               `json:"last_error"`
	Result         string               `json:"result"`
	Payload        string               `json:"payload"`
	CreatedAt      time.Time            `json:"created_at"`
}

// QueueJobListPayload is the paged queue job list response.
type QueueJobListPayload struct {
	List           []QueueJobListItemPayload `json:"list"`
	Total          int64                     `json:"total"`
	Page           int                       `json:"page"`
	PageSize       int                       `json:"page_size"`
	PaginationMode string                    `json:"pagination_mode"`
	NextCursor     string                    `json:"next_cursor"`
	MaxPageSize    int                       `json:"max_page_size"`
}

// QueueJobBatchDeletePayload is the request body for batch deletion.
type QueueJobBatchDeletePayload struct {
	JobIDs []uint `json:"job_ids"`
}

// QueueJobClearPayload is the request body for cleanup by filter.
type QueueJobClearPayload struct {
	QueueKey string `json:"queue_key"`
	Status   string `json:"status"`
}

// QueueJobMutationPayload is returned by delete/cleanup operations.
type QueueJobMutationPayload struct {
	Deleted int64 `json:"deleted"`
}

// QueueJobStaleRecoveryPayload is returned by stale processing recovery.
type QueueJobStaleRecoveryPayload struct {
	Recovered int64 `json:"recovered"`
}

// QueueJobService provides admin access to queue job details.
type QueueJobService interface {
	List(queueKey, status string, nodeID uint, page, pageSize int, cursor uint, useCursor bool) (*QueueJobListPayload, error)
	Get(id uint) (*QueueJobListItemPayload, error)
	Retry(id uint) (*QueueJobListItemPayload, error)
	RecoverStale(queueKey string) (*QueueJobStaleRecoveryPayload, error)
	Delete(id uint) (*QueueJobMutationPayload, error)
	BatchDelete(payload *QueueJobBatchDeletePayload) (*QueueJobMutationPayload, error)
	Clear(payload *QueueJobClearPayload) (*QueueJobMutationPayload, error)
}

type queueJobService struct {
	repo     repository.QueueJobRepository
	settings repository.QueueSettingRepository
}

// NewQueueJobService creates a queue job service.
func NewQueueJobService(repo repository.QueueJobRepository, settings repository.QueueSettingRepository) QueueJobService {
	return &queueJobService{repo: repo, settings: settings}
}

// List returns paged queue jobs.
func (s *queueJobService) List(queueKey, status string, nodeID uint, page, pageSize int, cursor uint, useCursor bool) (*QueueJobListPayload, error) {
	if queueKey != "" && !isValidQueueKey(queueKey) {
		return nil, fmt.Errorf("invalid queue key")
	}
	if status != "" && !isValidQueueStatus(status) {
		return nil, fmt.Errorf("invalid queue job status")
	}

	rows, total, err := s.repo.List(repository.QueueJobListFilter{
		QueueKey:  queueKey,
		Status:    status,
		NodeID:    nodeID,
		Page:      page,
		PageSize:  pageSize,
		Cursor:    cursor,
		UseCursor: useCursor,
	})
	if err != nil {
		return nil, err
	}

	items := make([]QueueJobListItemPayload, 0, len(rows))
	for _, row := range rows {
		items = append(items, toQueueJobPayload(row))
	}

	return &QueueJobListPayload{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Get returns one queue job detail.
func (s *queueJobService) Get(id uint) (*QueueJobListItemPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid queue job id")
	}

	job, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("queue job not found")
	}

	payload := toQueueJobPayload(*job)
	return &payload, nil
}

// Retry resets a failed or pending job back to pending.
func (s *queueJobService) Retry(id uint) (*QueueJobListItemPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid queue job id")
	}

	job, err := s.repo.Retry(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("queue job not found")
	}

	payload := toQueueJobPayload(*job)
	return &payload, nil
}

// RecoverStale restores stale processing jobs for one queue or all queues.
func (s *queueJobService) RecoverStale(queueKey string) (*QueueJobStaleRecoveryPayload, error) {
	if queueKey != "" && !isValidQueueKey(queueKey) {
		return nil, fmt.Errorf("invalid queue key")
	}

	total := int64(0)
	for _, definition := range queue.Definitions() {
		key := string(definition.Key)
		if queueKey != "" && key != queueKey {
			continue
		}

		setting, err := queue.ResolveSetting(s.settings, key)
		if err != nil {
			defaultSetting, ok := queue.DefaultSettingByKey(key)
			if !ok {
				continue
			}
			setting = defaultSetting
		}
		timeoutSeconds := setting.MaxExecution
		if timeoutSeconds <= 0 {
			timeoutSeconds = 300
		}

		recovered, recoverErr := s.repo.RequeueStaleProcessing(key, time.Now().Add(-time.Duration(timeoutSeconds)*time.Second), setting.MaxRetry)
		if recoverErr != nil {
			return nil, recoverErr
		}
		total += recovered
	}

	return &QueueJobStaleRecoveryPayload{Recovered: total}, nil
}

// Delete removes one queue job. Processing jobs are protected from deletion.
func (s *queueJobService) Delete(id uint) (*QueueJobMutationPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid queue job id")
	}

	deleted, err := s.repo.Delete(id)
	if err != nil {
		return nil, err
	}
	if deleted == 0 {
		return nil, fmt.Errorf("queue job not found or is processing")
	}

	return &QueueJobMutationPayload{Deleted: deleted}, nil
}

// BatchDelete removes selected queue jobs. Processing jobs are skipped.
func (s *queueJobService) BatchDelete(payload *QueueJobBatchDeletePayload) (*QueueJobMutationPayload, error) {
	if payload == nil || len(payload.JobIDs) == 0 {
		return nil, fmt.Errorf("queue job ids cannot be empty")
	}

	ids := make([]uint, 0, len(payload.JobIDs))
	seen := map[uint]struct{}{}
	for _, id := range payload.JobIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("queue job ids cannot be empty")
	}

	deleted, err := s.repo.BatchDelete(ids)
	if err != nil {
		return nil, err
	}

	return &QueueJobMutationPayload{Deleted: deleted}, nil
}

// Clear removes terminal queue jobs matching the filter.
func (s *queueJobService) Clear(payload *QueueJobClearPayload) (*QueueJobMutationPayload, error) {
	if payload == nil {
		payload = &QueueJobClearPayload{}
	}
	if payload.QueueKey != "" && !isValidQueueKey(payload.QueueKey) {
		return nil, fmt.Errorf("invalid queue key")
	}
	if payload.Status != "" && !isValidQueueStatus(payload.Status) {
		return nil, fmt.Errorf("invalid queue job status")
	}
	if payload.Status == model.QueueJobStatusPending || payload.Status == model.QueueJobStatusProcessing {
		return nil, fmt.Errorf("only completed or failed jobs can be cleared by filter")
	}

	deleted, err := s.repo.Clear(repository.QueueJobClearFilter{
		QueueKey: payload.QueueKey,
		Status:   payload.Status,
	})
	if err != nil {
		return nil, err
	}

	return &QueueJobMutationPayload{Deleted: deleted}, nil
}

func isValidQueueKey(value string) bool {
	for _, definition := range queue.Definitions() {
		if string(definition.Key) == value {
			return true
		}
	}
	return false
}

func isValidQueueStatus(value string) bool {
	switch value {
	case model.QueueJobStatusPending, model.QueueJobStatusProcessing, model.QueueJobStatusCompleted, model.QueueJobStatusFailed:
		return true
	default:
		return false
	}
}

func toQueueJobPayload(row model.QueueJob) QueueJobListItemPayload {
	return QueueJobListItemPayload{
		ID:             row.ID,
		QueueKey:       row.QueueKey,
		JobType:        row.JobType,
		ResourceType:   row.ResourceType,
		ResourceID:     row.ResourceID,
		DispatchNode:   queueJobDispatchNode(row),
		NodeCapability: row.NodeCapability,
		ExecutionMode:  queueJobExecutionMode(row.JobType),
		ExecutionNote:  queueJobExecutionNote(row.JobType),
		Status:         row.Status,
		Attempts:       row.Attempts,
		MaxAttempts:    row.MaxAttempts,
		ScheduledAt:    row.ScheduledAt,
		StartedAt:      row.StartedAt,
		FinishedAt:     row.FinishedAt,
		LastError:      row.LastError,
		Result:         row.Result,
		Payload:        row.Payload,
		CreatedAt:      row.CreatedAt,
	}
}

func queueJobExecutionMode(jobType string) string {
	switch jobType {
	case "archive.create", "archive.extract":
		return "unified_runner"
	default:
		return "unified_runner"
	}
}

func queueJobExecutionNote(jobType string) string {
	switch jobType {
	case "archive.create", "archive.extract":
		return "统一队列 runner 任务：归档创建/解压由 IO 队列 runner 执行，受 worker_num、max_execution、max_retry 和退避参数控制。"
	default:
		return "统一队列 runner 任务：由对应 queue_settings 控制并发、超时和重试。"
	}
}

func queueJobDispatchNode(row model.QueueJob) *SelectedNodePayload {
	if row.DispatchNodeID == nil {
		return nil
	}
	return &SelectedNodePayload{
		ID:   *row.DispatchNodeID,
		Name: row.DispatchNodeName,
		Type: row.DispatchNodeType,
	}
}
