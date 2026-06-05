package queue

import "time"

// BlobRuntimeConfig is the subset of queue settings used by the file deletion worker.
type BlobRuntimeConfig struct {
	Interval   time.Duration
	BatchSize  int
	MaxRetries int
}

// IORuntimeConfig is the subset of queue settings used by the multipart cleanup worker.
type IORuntimeConfig struct {
	Interval    time.Duration
	ExpireHours int
}

// OfflineRuntimeConfig is the subset of queue settings used by the recycle cleanup worker.
type OfflineRuntimeConfig struct {
	CronHour  int
	BatchSize int
}

// ResolveBlobRuntime converts queue settings into runtime config for the blob worker.
func ResolveBlobRuntime(setting Setting) BlobRuntimeConfig {
	intervalSeconds := setting.RetryDelay
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}

	batchSize := setting.WorkerNum
	if batchSize <= 0 {
		batchSize = 100
	}

	maxRetries := setting.MaxRetry
	if maxRetries < 0 {
		maxRetries = 0
	}

	return BlobRuntimeConfig{
		Interval:   time.Duration(intervalSeconds) * time.Second,
		BatchSize:  batchSize,
		MaxRetries: maxRetries,
	}
}

// ResolveIORuntime converts queue settings into runtime config for the multipart cleanup worker.
func ResolveIORuntime(setting Setting) IORuntimeConfig {
	intervalSeconds := setting.RetryDelay
	if intervalSeconds <= 0 {
		intervalSeconds = 3600
	}

	expireHours := setting.MaxExecution / 3600
	if expireHours <= 0 {
		expireHours = 24
	}

	return IORuntimeConfig{
		Interval:    time.Duration(intervalSeconds) * time.Second,
		ExpireHours: expireHours,
	}
}

// ResolveOfflineRuntime converts queue settings into runtime config for the recycle cleanup worker.
func ResolveOfflineRuntime(setting Setting) OfflineRuntimeConfig {
	batchSize := setting.WorkerNum
	if batchSize <= 0 {
		batchSize = 100
	}

	cronHour := setting.RetryDelay
	if cronHour < 0 || cronHour > 23 {
		cronHour = 2
	}

	return OfflineRuntimeConfig{
		CronHour:  cronHour,
		BatchSize: batchSize,
	}
}
