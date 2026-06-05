package queue

import (
	"fmt"

	"xingyunpan-v2/internal/repository"
)

// ResolveSetting loads one queue setting from DB and falls back to defaults.
func ResolveSetting(repo repository.QueueSettingRepository, queueKey string) (Setting, error) {
	if repo == nil {
		return Setting{}, fmt.Errorf("queue setting repository is nil")
	}

	if err := repo.EnsureSchema(); err != nil {
		return Setting{}, err
	}

	setting, err := repo.GetByQueueKey(queueKey)
	if err != nil {
		return Setting{}, err
	}
	if setting != nil {
		return Setting{
			QueueKey:      setting.QueueKey,
			WorkerNum:     setting.WorkerNum,
			MaxExecution:  setting.MaxExecution,
			BackoffFactor: setting.BackoffFactor,
			MaxBackoff:    setting.MaxBackoff,
			MaxRetry:      setting.MaxRetry,
			RetryDelay:    setting.RetryDelay,
		}, nil
	}

	defaultSetting, ok := DefaultSettingByKey(queueKey)
	if !ok {
		return Setting{}, fmt.Errorf("unknown queue key: %s", queueKey)
	}

	return defaultSetting, nil
}
