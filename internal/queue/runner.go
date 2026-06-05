package queue

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/logger"

	"go.uber.org/zap"
)

// Runner executes unified queue jobs for all implemented queues.
type Runner struct {
	settingsRepo repository.QueueSettingRepository
	fileSettings repository.FileSystemSettingRepository
	jobsRepo     repository.QueueJobRepository
	executor     *Executor
	pollInterval time.Duration
	stopChan     chan struct{}
}

// NewRunner creates a unified queue runner.
func NewRunner(
	settingsRepo repository.QueueSettingRepository,
	jobsRepo repository.QueueJobRepository,
	executor *Executor,
	pollInterval time.Duration,
	fileSettings ...repository.FileSystemSettingRepository,
) *Runner {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	var fileSettingsRepo repository.FileSystemSettingRepository
	if len(fileSettings) > 0 {
		fileSettingsRepo = fileSettings[0]
	}

	return &Runner{
		settingsRepo: settingsRepo,
		fileSettings: fileSettingsRepo,
		jobsRepo:     jobsRepo,
		executor:     executor,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
	}
}

// Start begins polling all implemented queues.
func (r *Runner) Start() {
	logger.Info("unified queue runner started", zap.Duration("poll_interval", r.pollInterval))

	var wg sync.WaitGroup
	for _, definition := range Definitions() {
		if !definition.Implemented {
			continue
		}

		wg.Add(1)
		go func(queueKey string) {
			defer wg.Done()
			r.loop(queueKey)
		}(string(definition.Key))
	}

	<-r.stopChan
	wg.Wait()
	logger.Info("unified queue runner stopped")
}

// Stop stops the unified queue runner.
func (r *Runner) Stop() {
	close(r.stopChan)
}

func (r *Runner) loop(queueKey string) {
	currentInterval := r.pollIntervalForQueue(queueKey)
	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	r.processQueue(queueKey)

	for {
		select {
		case <-ticker.C:
			nextInterval := r.pollIntervalForQueue(queueKey)
			if nextInterval > 0 && nextInterval != currentInterval {
				ticker.Reset(nextInterval)
				currentInterval = nextInterval
			}
			r.processQueue(queueKey)
		case <-r.stopChan:
			return
		}
	}
}

func (r *Runner) pollIntervalForQueue(queueKey string) time.Duration {
	if r.fileSettings == nil {
		return r.pollInterval
	}

	setting, err := r.fileSettings.Get()
	if err != nil || setting == nil {
		return r.pollInterval
	}

	var raw string
	switch queueKey {
	case string(KeyBlob):
		raw = setting.BlobRecycleInterval
	case string(KeyOffline):
		raw = setting.RecycleScanInterval
	default:
		return r.pollInterval
	}

	duration := parseEveryDuration(raw)
	if duration <= 0 {
		return r.pollInterval
	}
	return duration
}

func parseEveryDuration(raw string) time.Duration {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "@every")
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return duration
}

func (r *Runner) processQueue(queueKey string) {
	setting, err := ResolveSetting(r.settingsRepo, queueKey)
	if err != nil {
		logger.Error("load queue setting failed", zap.String("queue_key", queueKey), zap.Error(err))
		return
	}

	if setting.WorkerNum <= 0 {
		return
	}

	timeoutSeconds := setting.MaxExecution
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}
	recovered, err := r.jobsRepo.RequeueStaleProcessing(queueKey, time.Now().Add(-time.Duration(timeoutSeconds)*time.Second), setting.MaxRetry)
	if err != nil {
		logger.Error("recover stale queue jobs failed", zap.String("queue_key", queueKey), zap.Error(err))
		return
	}
	if recovered > 0 {
		logger.Warn("stale queue jobs recovered", zap.String("queue_key", queueKey), zap.Int64("count", recovered))
	}

	jobs, err := r.jobsRepo.ClaimDueJobs(queueKey, setting.WorkerNum)
	if err != nil {
		logger.Error("claim queue jobs failed", zap.String("queue_key", queueKey), zap.Error(err))
		return
	}
	if len(jobs) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(item model.QueueJob) {
			defer wg.Done()
			r.executeJob(setting, item)
		}(job)
	}
	wg.Wait()
}

func (r *Runner) executeJob(setting Setting, job model.QueueJob) {
	timeoutSeconds := setting.MaxExecution
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	result, err := r.executor.Execute(ctx, job)
	if err == nil {
		if markErr := r.jobsRepo.MarkCompleted(job.ID, result); markErr != nil {
			logger.Error("mark queue job completed failed", zap.Uint("job_id", job.ID), zap.Error(markErr))
		}
		return
	}

	attempts := job.Attempts + 1
	if attempts > setting.MaxRetry {
		if markErr := r.jobsRepo.MarkFailed(job.ID, attempts, err.Error()); markErr != nil {
			logger.Error("mark queue job failed failed", zap.Uint("job_id", job.ID), zap.Error(markErr))
		}
		return
	}

	retryAt := time.Now().Add(backoffDuration(setting, attempts))
	if markErr := r.jobsRepo.MarkRetry(job.ID, attempts, retryAt, err.Error()); markErr != nil {
		logger.Error("mark queue job retry failed", zap.Uint("job_id", job.ID), zap.Error(markErr))
	}
}

func backoffDuration(setting Setting, attempts int) time.Duration {
	baseSeconds := setting.RetryDelay
	if baseSeconds <= 0 {
		baseSeconds = 5
	}

	factor := setting.BackoffFactor
	if factor < 1 {
		factor = 1
	}

	delay := baseSeconds
	for i := 1; i < attempts; i++ {
		delay *= factor
		if setting.MaxBackoff > 0 && delay >= setting.MaxBackoff {
			delay = setting.MaxBackoff
			break
		}
	}

	if delay <= 0 {
		delay = 5
	}

	return time.Duration(delay) * time.Second
}

// DescribeImplementedQueues returns the enabled queue keys for logging.
func DescribeImplementedQueues() []string {
	result := make([]string, 0, len(Definitions()))
	for _, definition := range Definitions() {
		if definition.Implemented {
			result = append(result, fmt.Sprintf("%s", definition.Key))
		}
	}

	return result
}
