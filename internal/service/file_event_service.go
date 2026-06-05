package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"xingyunpan-v2/internal/repository"
)

// FileEventPayload is pushed to connected clients after file-system mutations.
type FileEventPayload struct {
	Type       string `json:"type"`
	UserID     uint   `json:"user_id"`
	FolderID   *uint  `json:"folder_id,omitempty"`
	Resource   string `json:"resource,omitempty"`
	ResourceID uint   `json:"resource_id,omitempty"`
	At         int64  `json:"at"`
}

// FileEventService publishes and streams file-system refresh events.
type FileEventService interface {
	Publish(userID uint, folderID *uint, eventType string, resource string, resourceID uint)
	Subscribe(ctx context.Context, userID uint) (<-chan FileEventPayload, func())
}

type fileEventService struct {
	settingsRepo repository.FileSystemSettingRepository
	mu           sync.Mutex
	subscribers  map[uint]map[chan FileEventPayload]struct{}
	timers       map[string]*time.Timer
	pending      map[string]FileEventPayload
}

func NewFileEventService(settingsRepo repository.FileSystemSettingRepository) FileEventService {
	return &fileEventService{
		settingsRepo: settingsRepo,
		subscribers:  make(map[uint]map[chan FileEventPayload]struct{}),
		timers:       make(map[string]*time.Timer),
		pending:      make(map[string]FileEventPayload),
	}
}

func (s *fileEventService) Publish(userID uint, folderID *uint, eventType string, resource string, resourceID uint) {
	if userID == 0 || !s.enabled() {
		return
	}

	payload := FileEventPayload{
		Type:       strings.TrimSpace(eventType),
		UserID:     userID,
		FolderID:   folderID,
		Resource:   strings.TrimSpace(resource),
		ResourceID: resourceID,
		At:         time.Now().Unix(),
	}
	if payload.Type == "" {
		payload.Type = "file.changed"
	}

	delay := s.debounceDelay()
	if delay <= 0 {
		s.broadcast(payload)
		return
	}

	key := eventDebounceKey(userID, folderID)
	s.mu.Lock()
	s.pending[key] = payload
	if timer, ok := s.timers[key]; ok {
		timer.Reset(delay)
		s.mu.Unlock()
		return
	}
	s.timers[key] = time.AfterFunc(delay, func() {
		s.flush(key)
	})
	s.mu.Unlock()
}

func (s *fileEventService) Subscribe(ctx context.Context, userID uint) (<-chan FileEventPayload, func()) {
	ch := make(chan FileEventPayload, 16)

	s.mu.Lock()
	if s.subscribers[userID] == nil {
		s.subscribers[userID] = make(map[chan FileEventPayload]struct{})
	}
	s.subscribers[userID][ch] = struct{}{}
	s.mu.Unlock()

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			s.mu.Lock()
			if subscribers := s.subscribers[userID]; subscribers != nil {
				delete(subscribers, ch)
				if len(subscribers) == 0 {
					delete(s.subscribers, userID)
				}
			}
			s.mu.Unlock()
			close(ch)
		})
	}

	go func() {
		<-ctx.Done()
		cleanup()
	}()

	return ch, cleanup
}

func (s *fileEventService) flush(key string) {
	s.mu.Lock()
	payload, ok := s.pending[key]
	delete(s.pending, key)
	delete(s.timers, key)
	s.mu.Unlock()

	if ok {
		s.broadcast(payload)
	}
}

func (s *fileEventService) broadcast(payload FileEventPayload) {
	s.mu.Lock()
	targets := make([]chan FileEventPayload, 0, len(s.subscribers[payload.UserID]))
	for ch := range s.subscribers[payload.UserID] {
		targets = append(targets, ch)
	}
	s.mu.Unlock()

	for _, ch := range targets {
		select {
		case ch <- payload:
		default:
		}
	}
}

func (s *fileEventService) enabled() bool {
	if s.settingsRepo == nil {
		return true
	}
	setting, err := s.settingsRepo.Get()
	return err != nil || setting == nil || setting.EnableEventPush
}

func (s *fileEventService) debounceDelay() time.Duration {
	if s.settingsRepo == nil {
		return 0
	}
	setting, err := s.settingsRepo.Get()
	if err != nil || setting == nil || setting.DebounceDelay <= 0 {
		return 0
	}
	return time.Duration(setting.DebounceDelay) * time.Second
}

func eventDebounceKey(userID uint, folderID *uint) string {
	if folderID == nil {
		return fmt.Sprintf("%d:root", userID)
	}
	return fmt.Sprintf("%d:%d", userID, *folderID)
}

func MarshalFileEvent(payload FileEventPayload) string {
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}
