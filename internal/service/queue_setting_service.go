package service

import (
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
)

// QueueSettingItemPayload is the API-facing shape for one queue config row.
type QueueSettingItemPayload struct {
	QueueKey      string `json:"queue_key"`
	WorkerNum     int    `json:"worker_num"`
	MaxExecution  int    `json:"max_execution"`
	BackoffFactor int    `json:"backoff_factor"`
	MaxBackoff    int    `json:"max_backoff"`
	MaxRetry      int    `json:"max_retry"`
	RetryDelay    int    `json:"retry_delay"`
}

// QueueSettingService provides admin access to queue settings.
type QueueSettingService interface {
	GetAll() ([]QueueSettingItemPayload, error)
	UpdateAll(payload []QueueSettingItemPayload) ([]QueueSettingItemPayload, error)
}

type queueSettingService struct {
	repo repository.QueueSettingRepository
}

// NewQueueSettingService creates a queue settings service.
func NewQueueSettingService(repo repository.QueueSettingRepository) QueueSettingService {
	return &queueSettingService{repo: repo}
}

// GetAll returns queue settings, seeding defaults when no rows exist.
func (s *queueSettingService) GetAll() ([]QueueSettingItemPayload, error) {
	if err := s.repo.EnsureSchema(); err != nil {
		return nil, err
	}

	settings, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	defaultQueueSettings := defaultQueueSettingPayloads()

	if len(settings) == 0 {
		for _, item := range defaultQueueSettings {
			record := toQueueSettingModel(item)
			if err := s.repo.Save(&record); err != nil {
				return nil, err
			}
		}
		return cloneDefaultQueueSettings(), nil
	}

	merged := make([]QueueSettingItemPayload, 0, len(defaultQueueSettings))
	for _, def := range defaultQueueSettings {
		match := findQueueSetting(settings, def.QueueKey)
		if match == nil {
			record := toQueueSettingModel(def)
			if err := s.repo.Save(&record); err != nil {
				return nil, err
			}
			merged = append(merged, def)
			continue
		}

		merged = append(merged, toQueueSettingPayload(match))
	}

	return merged, nil
}

// UpdateAll persists all queue settings rows.
func (s *queueSettingService) UpdateAll(payload []QueueSettingItemPayload) ([]QueueSettingItemPayload, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("queue settings cannot be empty")
	}

	if err := s.repo.EnsureSchema(); err != nil {
		return nil, err
	}

	normalized, err := normalizeQueueSettings(payload)
	if err != nil {
		return nil, err
	}

	for _, item := range normalized {
		existing, err := s.repo.GetByQueueKey(item.QueueKey)
		if err != nil {
			return nil, err
		}

		if existing == nil {
			record := toQueueSettingModel(item)
			if err := s.repo.Save(&record); err != nil {
				return nil, err
			}
			continue
		}

		existing.WorkerNum = item.WorkerNum
		existing.MaxExecution = item.MaxExecution
		existing.BackoffFactor = item.BackoffFactor
		existing.MaxBackoff = item.MaxBackoff
		existing.MaxRetry = item.MaxRetry
		existing.RetryDelay = item.RetryDelay
		if err := s.repo.Save(existing); err != nil {
			return nil, err
		}
	}

	return s.GetAll()
}

func normalizeQueueSettings(payload []QueueSettingItemPayload) ([]QueueSettingItemPayload, error) {
	defaultQueueSettings := defaultQueueSettingPayloads()
	expected := make(map[string]struct{}, len(defaultQueueSettings))
	for _, item := range defaultQueueSettings {
		expected[item.QueueKey] = struct{}{}
	}

	normalized := make([]QueueSettingItemPayload, 0, len(defaultQueueSettings))
	seen := make(map[string]struct{}, len(payload))

	for _, item := range payload {
		item.QueueKey = strings.TrimSpace(strings.ToLower(item.QueueKey))
		if _, ok := expected[item.QueueKey]; !ok {
			return nil, fmt.Errorf("unsupported queue key: %s", item.QueueKey)
		}
		if _, exists := seen[item.QueueKey]; exists {
			return nil, fmt.Errorf("duplicate queue key: %s", item.QueueKey)
		}
		seen[item.QueueKey] = struct{}{}

		if item.WorkerNum < 0 {
			item.WorkerNum = 0
		}
		if item.MaxExecution < 0 {
			item.MaxExecution = 0
		}
		if item.BackoffFactor < 0 {
			item.BackoffFactor = 0
		}
		if item.MaxBackoff < 0 {
			item.MaxBackoff = 0
		}
		if item.MaxRetry < 0 {
			item.MaxRetry = 0
		}
		if item.RetryDelay < 0 {
			item.RetryDelay = 0
		}

		normalized = append(normalized, item)
	}

	if len(normalized) != len(defaultQueueSettings) {
		return nil, fmt.Errorf("incomplete queue settings payload")
	}

	return normalized, nil
}

func toQueueSettingPayload(setting *model.QueueSetting) QueueSettingItemPayload {
	return QueueSettingItemPayload{
		QueueKey:      setting.QueueKey,
		WorkerNum:     setting.WorkerNum,
		MaxExecution:  setting.MaxExecution,
		BackoffFactor: setting.BackoffFactor,
		MaxBackoff:    setting.MaxBackoff,
		MaxRetry:      setting.MaxRetry,
		RetryDelay:    setting.RetryDelay,
	}
}

func toQueueSettingModel(item QueueSettingItemPayload) model.QueueSetting {
	return model.QueueSetting{
		QueueKey:      item.QueueKey,
		WorkerNum:     item.WorkerNum,
		MaxExecution:  item.MaxExecution,
		BackoffFactor: item.BackoffFactor,
		MaxBackoff:    item.MaxBackoff,
		MaxRetry:      item.MaxRetry,
		RetryDelay:    item.RetryDelay,
	}
}

func cloneDefaultQueueSettings() []QueueSettingItemPayload {
	defaultQueueSettings := defaultQueueSettingPayloads()
	items := make([]QueueSettingItemPayload, len(defaultQueueSettings))
	copy(items, defaultQueueSettings)
	return items
}

func findQueueSetting(settings []model.QueueSetting, queueKey string) *model.QueueSetting {
	for i := range settings {
		if settings[i].QueueKey == queueKey {
			return &settings[i]
		}
	}
	return nil
}

func defaultQueueSettingPayloads() []QueueSettingItemPayload {
	items := queue.DefaultSettings()
	payloads := make([]QueueSettingItemPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, QueueSettingItemPayload{
			QueueKey:      item.QueueKey,
			WorkerNum:     item.WorkerNum,
			MaxExecution:  item.MaxExecution,
			BackoffFactor: item.BackoffFactor,
			MaxBackoff:    item.MaxBackoff,
			MaxRetry:      item.MaxRetry,
			RetryDelay:    item.RetryDelay,
		})
	}
	return payloads
}
