package service

import (
	"encoding/json"
	"fmt"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type EventSettingPayload struct {
	Categories []EventSettingCategory `json:"categories"`
	Events     map[string]bool        `json:"events"`
}

type EventTogglePayload struct {
	Enabled bool `json:"enabled"`
}

// EventSettingService provides admin access to event recording settings.
type EventSettingService interface {
	Get() (*EventSettingPayload, error)
	Update(payload *EventSettingPayload) (*EventSettingPayload, error)
	Reset() (*EventSettingPayload, error)
	ToggleAll(enabled bool) (*EventSettingPayload, error)
	ToggleCategory(categoryKey string, enabled bool) (*EventSettingPayload, error)
	ToggleEvent(eventKey string, enabled bool) (*EventSettingPayload, error)
	IsEventEnabled(eventKey string) (bool, error)
}

type eventSettingService struct {
	repo repository.EventSettingRepository
}

// NewEventSettingService creates an event settings service.
func NewEventSettingService(repo repository.EventSettingRepository) EventSettingService {
	return &eventSettingService{repo: repo}
}

func (s *eventSettingService) Get() (*EventSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return defaultEventSettingPayload(), nil
	}
	return toEventSettingPayload(setting)
}

func (s *eventSettingService) Update(payload *EventSettingPayload) (*EventSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("事件设置不能为空")
	}

	normalized, err := normalizeEventMap(payload.Events)
	if err != nil {
		return nil, err
	}

	return s.saveEvents(normalized)
}

func (s *eventSettingService) Reset() (*EventSettingPayload, error) {
	return s.saveEvents(defaultEventMap())
}

func (s *eventSettingService) ToggleAll(enabled bool) (*EventSettingPayload, error) {
	events := defaultEventMap()
	for key := range events {
		events[key] = enabled
	}
	return s.saveEvents(events)
}

func (s *eventSettingService) ToggleCategory(categoryKey string, enabled bool) (*EventSettingPayload, error) {
	category, ok := eventCategoryByKey(categoryKey)
	if !ok {
		return nil, fmt.Errorf("未知事件分类")
	}

	current, err := s.Get()
	if err != nil {
		return nil, err
	}
	events := cloneEventMap(current.Events)
	for _, item := range category.Items {
		events[item.Key] = enabled
	}
	return s.saveEvents(events)
}

func (s *eventSettingService) ToggleEvent(eventKey string, enabled bool) (*EventSettingPayload, error) {
	if !isKnownEventKey(eventKey) {
		return nil, fmt.Errorf("未知事件")
	}

	current, err := s.Get()
	if err != nil {
		return nil, err
	}
	events := cloneEventMap(current.Events)
	events[eventKey] = enabled
	return s.saveEvents(events)
}

func (s *eventSettingService) IsEventEnabled(eventKey string) (bool, error) {
	if !isKnownEventKey(eventKey) {
		return false, fmt.Errorf("未知事件")
	}

	current, err := s.Get()
	if err != nil {
		return false, err
	}
	return current.Events[eventKey], nil
}

func (s *eventSettingService) saveEvents(events map[string]bool) (*EventSettingPayload, error) {
	normalized, err := normalizeEventMap(events)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(enabledEventKeys(normalized))
	if err != nil {
		return nil, fmt.Errorf("序列化事件设置失败: %w", err)
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.EventSetting{}
	}
	setting.EnabledEventsJSON = string(data)

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return toEventSettingPayload(setting)
}

func toEventSettingPayload(setting *model.EventSetting) (*EventSettingPayload, error) {
	events := defaultEventMap()
	if setting != nil && setting.EnabledEventsJSON != "" {
		var enabledKeys []string
		if err := json.Unmarshal([]byte(setting.EnabledEventsJSON), &enabledKeys); err != nil {
			return nil, fmt.Errorf("解析事件设置失败: %w", err)
		}
		for _, key := range enabledKeys {
			if isKnownEventKey(key) {
				events[key] = true
			}
		}
	}

	return &EventSettingPayload{
		Categories: defaultEventCategories(),
		Events:     events,
	}, nil
}

func defaultEventSettingPayload() *EventSettingPayload {
	return &EventSettingPayload{
		Categories: defaultEventCategories(),
		Events:     defaultEventMap(),
	}
}

func normalizeEventMap(input map[string]bool) (map[string]bool, error) {
	events := defaultEventMap()
	if input == nil {
		return events, nil
	}

	for key, enabled := range input {
		if !isKnownEventKey(key) {
			return nil, fmt.Errorf("未知事件: %s", key)
		}
		events[key] = enabled
	}
	return events, nil
}

func defaultEventMap() map[string]bool {
	events := map[string]bool{}
	for _, category := range defaultEventCategories() {
		for _, item := range category.Items {
			events[item.Key] = false
		}
	}
	return events
}

func cloneEventMap(source map[string]bool) map[string]bool {
	events := defaultEventMap()
	for key, enabled := range source {
		if isKnownEventKey(key) {
			events[key] = enabled
		}
	}
	return events
}

func enabledEventKeys(events map[string]bool) []string {
	keys := make([]string, 0)
	for _, category := range defaultEventCategories() {
		for _, item := range category.Items {
			if events[item.Key] {
				keys = append(keys, item.Key)
			}
		}
	}
	return keys
}

func eventCategoryByKey(categoryKey string) (EventSettingCategory, bool) {
	for _, category := range defaultEventCategories() {
		if category.Key == categoryKey {
			return category, true
		}
	}
	return EventSettingCategory{}, false
}

func isKnownEventKey(eventKey string) bool {
	for _, category := range defaultEventCategories() {
		for _, item := range category.Items {
			if item.Key == eventKey {
				return true
			}
		}
	}
	return false
}
