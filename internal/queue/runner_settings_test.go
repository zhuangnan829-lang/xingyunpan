package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

func TestIORunnerUsesQueueSettingWorkerNumForClaimLimit(t *testing.T) {
	settings := &fakeRunnerSettingsRepo{setting: &model.QueueSetting{
		QueueKey:      string(KeyIO),
		WorkerNum:     7,
		MaxExecution:  300,
		BackoffFactor: 2,
		MaxBackoff:    60,
		MaxRetry:      3,
		RetryDelay:    5,
	}}
	jobs := &fakeRunnerJobsRepo{}
	runner := NewRunner(settings, jobs, nil, time.Second)

	runner.processQueue(string(KeyIO))

	if jobs.claimQueueKey != string(KeyIO) {
		t.Fatalf("claim queue key = %q, want io", jobs.claimQueueKey)
	}
	if jobs.claimLimit != 7 {
		t.Fatalf("claim limit = %d, want worker_num 7", jobs.claimLimit)
	}
}

func TestRunnerWorkerNumLimitsClaimedJobsPerRound(t *testing.T) {
	settings := &fakeRunnerSettingsRepo{setting: &model.QueueSetting{
		QueueKey:      string(KeyIO),
		WorkerNum:     2,
		MaxExecution:  300,
		BackoffFactor: 2,
		MaxBackoff:    60,
		MaxRetry:      0,
		RetryDelay:    5,
	}}
	jobs := &fakeRunnerJobsRepo{pending: []model.QueueJob{
		failingQueueJob(1, 0),
		failingQueueJob(2, 0),
		failingQueueJob(3, 0),
		failingQueueJob(4, 0),
	}}
	runner := NewRunner(settings, jobs, NewExecutor(nil, nil, nil, nil, nil, nil), time.Second)

	runner.processQueue(string(KeyIO))

	if jobs.claimLimit != 2 {
		t.Fatalf("claim limit = %d, want worker_num 2", jobs.claimLimit)
	}
	if got := len(jobs.failedIDs()); got != 2 {
		t.Fatalf("executed jobs = %d, want 2 claimed jobs in one round", got)
	}
}

func TestIORunnerWorkerNumZeroDoesNotClaimJobs(t *testing.T) {
	settings := &fakeRunnerSettingsRepo{setting: &model.QueueSetting{
		QueueKey:      string(KeyIO),
		WorkerNum:     0,
		MaxExecution:  300,
		BackoffFactor: 2,
		MaxBackoff:    60,
		MaxRetry:      3,
		RetryDelay:    5,
	}}
	jobs := &fakeRunnerJobsRepo{}
	runner := NewRunner(settings, jobs, nil, time.Second)

	runner.processQueue(string(KeyIO))

	if jobs.claimCalls != 0 {
		t.Fatalf("claim calls = %d, want 0 when worker_num is disabled", jobs.claimCalls)
	}
	if jobs.requeueCalls != 0 {
		t.Fatalf("requeue calls = %d, want 0 when worker_num is disabled", jobs.requeueCalls)
	}
}

func TestRunnerMaxExecutionControlsStaleProcessingRecovery(t *testing.T) {
	settings := &fakeRunnerSettingsRepo{setting: &model.QueueSetting{
		QueueKey:      string(KeyIO),
		WorkerNum:     1,
		MaxExecution:  42,
		BackoffFactor: 2,
		MaxBackoff:    60,
		MaxRetry:      3,
		RetryDelay:    5,
	}}
	jobs := &fakeRunnerJobsRepo{}
	runner := NewRunner(settings, jobs, NewExecutor(nil, nil, nil, nil, nil, nil), time.Second)

	start := time.Now()
	runner.processQueue(string(KeyIO))
	end := time.Now()

	if jobs.requeueCalls != 1 {
		t.Fatalf("requeue calls = %d, want 1", jobs.requeueCalls)
	}
	if jobs.requeueQueueKey != string(KeyIO) {
		t.Fatalf("requeue queue key = %q, want io", jobs.requeueQueueKey)
	}
	if jobs.requeueMaxRetry != 3 {
		t.Fatalf("requeue max retry = %d, want 3", jobs.requeueMaxRetry)
	}
	minBefore := start.Add(-42*time.Second - time.Second)
	maxBefore := end.Add(-42*time.Second + time.Second)
	if jobs.requeueBefore.Before(minBefore) || jobs.requeueBefore.After(maxBefore) {
		t.Fatalf("requeue before = %s, want around now - max_execution", jobs.requeueBefore.Format(time.RFC3339Nano))
	}
}

func TestRunnerMaxRetryControlsRetryThenFinalFailure(t *testing.T) {
	setting := Setting{
		QueueKey:      string(KeyIO),
		WorkerNum:     1,
		MaxExecution:  300,
		BackoffFactor: 2,
		MaxBackoff:    60,
		MaxRetry:      1,
		RetryDelay:    5,
	}
	jobs := &fakeRunnerJobsRepo{}
	runner := NewRunner(&fakeRunnerSettingsRepo{}, jobs, NewExecutor(nil, nil, nil, nil, nil, nil), time.Second)

	runner.executeJob(setting, failingQueueJob(11, 0))
	if got := jobs.retryByID(11); got == nil || got.attempts != 1 {
		t.Fatalf("first failure should retry with attempts=1, got %#v", got)
	}

	runner.executeJob(setting, failingQueueJob(12, 1))
	if got := jobs.failedByID(12); got == nil || got.attempts != 2 {
		t.Fatalf("second failure should be final failed with attempts=2, got %#v", got)
	}
}

func TestRunnerRetryBackoffControlsNextSchedule(t *testing.T) {
	setting := Setting{
		QueueKey:      string(KeyIO),
		WorkerNum:     1,
		MaxExecution:  300,
		BackoffFactor: 3,
		MaxBackoff:    20,
		MaxRetry:      5,
		RetryDelay:    10,
	}
	jobs := &fakeRunnerJobsRepo{}
	runner := NewRunner(&fakeRunnerSettingsRepo{}, jobs, NewExecutor(nil, nil, nil, nil, nil, nil), time.Second)

	start := time.Now()
	runner.executeJob(setting, failingQueueJob(21, 2))
	end := time.Now()

	retry := jobs.retryByID(21)
	if retry == nil {
		t.Fatalf("expected retry to be scheduled")
	}
	if retry.attempts != 3 {
		t.Fatalf("retry attempts = %d, want 3", retry.attempts)
	}
	minScheduled := start.Add(20*time.Second - time.Second)
	maxScheduled := end.Add(20*time.Second + time.Second)
	if retry.scheduledAt.Before(minScheduled) || retry.scheduledAt.After(maxScheduled) {
		t.Fatalf("scheduled_at = %s, want retry_delay/backoff capped to about 20s", retry.scheduledAt.Format(time.RFC3339Nano))
	}
}

func TestResolveSettingUsesDatabaseThenDefault(t *testing.T) {
	dbSetting := &model.QueueSetting{
		QueueKey:      string(KeyIO),
		WorkerNum:     4,
		MaxExecution:  123,
		BackoffFactor: 7,
		MaxBackoff:    89,
		MaxRetry:      6,
		RetryDelay:    11,
	}
	settings := &fakeRunnerSettingsRepo{setting: dbSetting}

	resolved, err := ResolveSetting(settings, string(KeyIO))
	if err != nil {
		t.Fatalf("resolve db setting: %v", err)
	}
	if resolved.WorkerNum != dbSetting.WorkerNum ||
		resolved.MaxExecution != dbSetting.MaxExecution ||
		resolved.BackoffFactor != dbSetting.BackoffFactor ||
		resolved.MaxBackoff != dbSetting.MaxBackoff ||
		resolved.MaxRetry != dbSetting.MaxRetry ||
		resolved.RetryDelay != dbSetting.RetryDelay {
		t.Fatalf("resolved setting = %#v, want db setting %#v", resolved, dbSetting)
	}

	settings.setting = nil
	fallback, err := ResolveSetting(settings, string(KeyThumbnail))
	if err != nil {
		t.Fatalf("resolve default setting: %v", err)
	}
	want, ok := DefaultSettingByKey(string(KeyThumbnail))
	if !ok {
		t.Fatalf("missing thumbnail default setting")
	}
	if fallback != want {
		t.Fatalf("fallback setting = %#v, want default %#v", fallback, want)
	}
}

func TestExecutorRunsArchiveJobsThroughArchiveExecutor(t *testing.T) {
	executor := NewExecutor(nil, nil, nil, nil, nil, nil)
	archive := &fakeArchiveExecutor{}
	executor.SetArchiveExecutor(archive)

	createPayload, err := EncodePayload(ArchiveCreatePayload{UserID: 42, SessionID: "session-a", FileIDs: []uint{1, 2}})
	if err != nil {
		t.Fatalf("encode create payload: %v", err)
	}
	if _, err := executor.Execute(context.Background(), model.QueueJob{JobType: JobTypeArchiveCreate, ResourceID: "session-a", Payload: createPayload}); err != nil {
		t.Fatalf("archive.create should be handled by unified queue executor: %v", err)
	}
	if archive.createCalls != 1 {
		t.Fatalf("archive create calls = %d, want 1", archive.createCalls)
	}

	extractPayload, err := EncodePayload(ArchiveExtractPayload{UserID: 42, TaskID: "task-a"})
	if err != nil {
		t.Fatalf("encode extract payload: %v", err)
	}
	if _, err := executor.Execute(context.Background(), model.QueueJob{JobType: JobTypeArchiveExtract, ResourceID: "task-a", Payload: extractPayload}); err != nil {
		t.Fatalf("archive.extract should be handled by unified queue executor: %v", err)
	}
	if archive.extractCalls != 1 {
		t.Fatalf("archive extract calls = %d, want 1", archive.extractCalls)
	}
}

type fakeRunnerSettingsRepo struct {
	setting *model.QueueSetting
}

type fakeArchiveExecutor struct {
	createCalls  int
	extractCalls int
}

func (e *fakeArchiveExecutor) ExecuteArchiveCreate(ctx context.Context, userID uint, sessionID string, fileIDs []uint) (string, error) {
	e.createCalls++
	return `{"status":"completed"}`, nil
}

func (e *fakeArchiveExecutor) ExecuteArchiveExtract(ctx context.Context, userID uint, taskID string) (string, error) {
	e.extractCalls++
	return `{"status":"completed"}`, nil
}

func failingQueueJob(id uint, attempts int) model.QueueJob {
	return model.QueueJob{
		BaseModel: model.BaseModel{ID: id},
		QueueKey:  string(KeyIO),
		JobType:   "test.unsupported",
		Attempts:  attempts,
	}
}

func (r *fakeRunnerSettingsRepo) EnsureSchema() error { return nil }
func (r *fakeRunnerSettingsRepo) List() ([]model.QueueSetting, error) {
	if r.setting == nil {
		return nil, nil
	}
	return []model.QueueSetting{*r.setting}, nil
}
func (r *fakeRunnerSettingsRepo) GetByQueueKey(queueKey string) (*model.QueueSetting, error) {
	if r.setting != nil && r.setting.QueueKey == queueKey {
		return r.setting, nil
	}
	return nil, nil
}
func (r *fakeRunnerSettingsRepo) Save(setting *model.QueueSetting) error {
	r.setting = setting
	return nil
}

type fakeRunnerJobsRepo struct {
	mu sync.Mutex

	pending []model.QueueJob

	claimCalls    int
	claimQueueKey string
	claimLimit    int

	requeueCalls    int
	requeueQueueKey string
	requeueBefore   time.Time
	requeueMaxRetry int

	retries []fakeRetryRecord
	failed  []fakeFailureRecord
}

type fakeRetryRecord struct {
	id          uint
	attempts    int
	scheduledAt time.Time
	lastError   string
}

type fakeFailureRecord struct {
	id        uint
	attempts  int
	lastError string
}

func (r *fakeRunnerJobsRepo) EnsureSchema() error { return nil }
func (r *fakeRunnerJobsRepo) EnqueueIfAbsent(job *model.QueueJob) (*model.QueueJob, error) {
	return job, nil
}
func (r *fakeRunnerJobsRepo) GetByID(id uint) (*model.QueueJob, error) { return nil, nil }
func (r *fakeRunnerJobsRepo) ClaimDueJobs(queueKey string, limit int) ([]model.QueueJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.claimCalls++
	r.claimQueueKey = queueKey
	r.claimLimit = limit
	if limit <= 0 || len(r.pending) == 0 {
		return []model.QueueJob{}, nil
	}
	count := limit
	if count > len(r.pending) {
		count = len(r.pending)
	}
	claimed := make([]model.QueueJob, count)
	copy(claimed, r.pending[:count])
	r.pending = r.pending[count:]
	return claimed, nil
}
func (r *fakeRunnerJobsRepo) MarkCompleted(id uint, result string) error { return nil }
func (r *fakeRunnerJobsRepo) MarkRetry(id uint, attempts int, scheduledAt time.Time, lastError string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.retries = append(r.retries, fakeRetryRecord{id: id, attempts: attempts, scheduledAt: scheduledAt, lastError: lastError})
	return nil
}
func (r *fakeRunnerJobsRepo) MarkFailed(id uint, attempts int, lastError string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failed = append(r.failed, fakeFailureRecord{id: id, attempts: attempts, lastError: lastError})
	return nil
}
func (r *fakeRunnerJobsRepo) Retry(id uint) (*model.QueueJob, error) { return nil, nil }
func (r *fakeRunnerJobsRepo) RequeueStaleProcessing(queueKey string, before time.Time, maxRetry int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requeueCalls++
	r.requeueQueueKey = queueKey
	r.requeueBefore = before
	r.requeueMaxRetry = maxRetry
	return 0, nil
}
func (r *fakeRunnerJobsRepo) ListStatusCounts() ([]repository.QueueJobStatusCount, error) {
	return nil, nil
}
func (r *fakeRunnerJobsRepo) List(filter repository.QueueJobListFilter) ([]model.QueueJob, int64, error) {
	return nil, 0, nil
}
func (r *fakeRunnerJobsRepo) Delete(id uint) (int64, error)         { return 0, nil }
func (r *fakeRunnerJobsRepo) BatchDelete(ids []uint) (int64, error) { return 0, nil }
func (r *fakeRunnerJobsRepo) Clear(filter repository.QueueJobClearFilter) (int64, error) {
	return 0, nil
}

func (r *fakeRunnerJobsRepo) retryByID(id uint) *fakeRetryRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.retries {
		if item.id == id {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *fakeRunnerJobsRepo) failedByID(id uint) *fakeFailureRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.failed {
		if item.id == id {
			copy := item
			return &copy
		}
	}
	return nil
}

func (r *fakeRunnerJobsRepo) failedIDs() []uint {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]uint, 0, len(r.failed))
	for _, item := range r.failed {
		ids = append(ids, item.id)
	}
	return ids
}
