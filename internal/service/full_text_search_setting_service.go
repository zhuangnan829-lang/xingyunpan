package service

import (
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// FullTextSearchSettingPayload is the API-facing shape for admin full text search settings.
type FullTextSearchSettingPayload struct {
	Enabled         bool   `json:"enabled"`
	MeiliEndpoint   string `json:"meili_endpoint"`
	APIKey          string `json:"api_key"`
	ResultPageSize  int    `json:"result_page_size"`
	AISearch        bool   `json:"ai_search"`
	TikaEndpoint    string `json:"tika_endpoint"`
	Extensions      string `json:"extensions"`
	MaxFileSize     int    `json:"max_file_size"`
	MaxFileSizeUnit string `json:"max_file_size_unit"`
	ChunkSize       int    `json:"chunk_size"`
	ChunkUnit       string `json:"chunk_unit"`
	IndexNotes      string `json:"index_notes"`
}

// FullTextSearchSettingService provides admin access to full text search settings.
type FullTextSearchSettingService interface {
	Get() (*FullTextSearchSettingPayload, error)
	Update(payload *FullTextSearchSettingPayload) (*FullTextSearchSettingPayload, error)
	TriggerRebuild(triggeredBy uint) (*FullTextSearchRebuildPayload, error)
}

type fullTextSearchSettingService struct {
	repo       repository.FullTextSearchSettingRepository
	dispatcher QueueDispatchService
}

// NewFullTextSearchSettingService creates a service instance.
func NewFullTextSearchSettingService(repo repository.FullTextSearchSettingRepository, dispatcher QueueDispatchService) FullTextSearchSettingService {
	return &fullTextSearchSettingService{repo: repo, dispatcher: dispatcher}
}

// FullTextSearchRebuildPayload is the API response after triggering rebuild.
type FullTextSearchRebuildPayload struct {
	JobID       uint   `json:"job_id"`
	QueueKey    string `json:"queue_key"`
	Status      string `json:"status"`
	ResourceID  string `json:"resource_id"`
	ScheduledAt int64  `json:"scheduled_at"`
}

// Get returns settings, providing defaults when no row exists.
func (s *fullTextSearchSettingService) Get() (*FullTextSearchSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return defaultFullTextSearchSettingPayload(), nil
	}

	return toFullTextSearchSettingPayload(setting), nil
}

// Update persists settings.
func (s *fullTextSearchSettingService) Update(payload *FullTextSearchSettingPayload) (*FullTextSearchSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("full text search settings cannot be nil")
	}

	normalized, err := normalizeFullTextSearchSettingPayload(payload)
	if err != nil {
		return nil, err
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.FullTextSearchSetting{}
	}

	setting.Enabled = normalized.Enabled
	setting.MeiliEndpoint = normalized.MeiliEndpoint
	setting.APIKey = normalized.APIKey
	setting.ResultPageSize = normalized.ResultPageSize
	setting.AISearch = normalized.AISearch
	setting.TikaEndpoint = normalized.TikaEndpoint
	setting.Extensions = normalized.Extensions
	setting.MaxFileSize = normalized.MaxFileSize
	setting.MaxFileSizeUnit = normalized.MaxFileSizeUnit
	setting.ChunkSize = normalized.ChunkSize
	setting.ChunkUnit = normalized.ChunkUnit
	setting.IndexNotes = normalized.IndexNotes

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return toFullTextSearchSettingPayload(setting), nil
}

func toFullTextSearchSettingPayload(setting *model.FullTextSearchSetting) *FullTextSearchSettingPayload {
	return &FullTextSearchSettingPayload{
		Enabled:         setting.Enabled,
		MeiliEndpoint:   setting.MeiliEndpoint,
		APIKey:          setting.APIKey,
		ResultPageSize:  setting.ResultPageSize,
		AISearch:        setting.AISearch,
		TikaEndpoint:    setting.TikaEndpoint,
		Extensions:      setting.Extensions,
		MaxFileSize:     setting.MaxFileSize,
		MaxFileSizeUnit: setting.MaxFileSizeUnit,
		ChunkSize:       setting.ChunkSize,
		ChunkUnit:       setting.ChunkUnit,
		IndexNotes:      setting.IndexNotes,
	}
}

func defaultFullTextSearchSettingPayload() *FullTextSearchSettingPayload {
	return &FullTextSearchSettingPayload{
		Enabled:         false,
		MeiliEndpoint:   "http://localhost:7700",
		APIKey:          "",
		ResultPageSize:  5,
		AISearch:        false,
		TikaEndpoint:    "http://localhost:9998",
		Extensions:      "pdf,doc,docx,xls,xlsx,ppt,pptx,odt,ods,odp,rtf,txt,md,html,htm,epub,csv",
		MaxFileSize:     25,
		MaxFileSizeUnit: "MB",
		ChunkSize:       2000,
		ChunkUnit:       "B",
		IndexNotes:      "",
	}
}

func normalizeFullTextSearchSettingPayload(payload *FullTextSearchSettingPayload) (*FullTextSearchSettingPayload, error) {
	normalized := *payload
	normalized.MeiliEndpoint = strings.TrimSpace(normalized.MeiliEndpoint)
	normalized.APIKey = strings.TrimSpace(normalized.APIKey)
	normalized.TikaEndpoint = strings.TrimSpace(normalized.TikaEndpoint)
	normalized.Extensions = strings.TrimSpace(normalized.Extensions)
	normalized.MaxFileSizeUnit = strings.ToUpper(strings.TrimSpace(normalized.MaxFileSizeUnit))
	normalized.ChunkUnit = strings.ToUpper(strings.TrimSpace(normalized.ChunkUnit))
	normalized.IndexNotes = strings.TrimSpace(normalized.IndexNotes)

	if normalized.MeiliEndpoint == "" {
		return nil, fmt.Errorf("meilisearch endpoint cannot be empty")
	}
	if normalized.TikaEndpoint == "" {
		return nil, fmt.Errorf("tika endpoint cannot be empty")
	}
	if normalized.Extensions == "" {
		return nil, fmt.Errorf("supported extensions cannot be empty")
	}

	switch normalized.MaxFileSizeUnit {
	case "B", "KB", "MB", "GB", "TB":
	default:
		return nil, fmt.Errorf("unsupported max file size unit")
	}

	switch normalized.ChunkUnit {
	case "B", "KB", "MB":
	default:
		return nil, fmt.Errorf("unsupported chunk unit")
	}

	if normalized.ResultPageSize < 1 {
		normalized.ResultPageSize = 1
	}
	if normalized.ResultPageSize > 100 {
		normalized.ResultPageSize = 100
	}
	if normalized.MaxFileSize < 1 {
		normalized.MaxFileSize = 1
	}
	if normalized.MaxFileSize > 10240 {
		normalized.MaxFileSize = 10240
	}
	if normalized.ChunkSize < 1 {
		normalized.ChunkSize = 1
	}
	if normalized.ChunkSize > 1048576 {
		normalized.ChunkSize = 1048576
	}

	return &normalized, nil
}

// TriggerRebuild enqueues a full text search rebuild job.
func (s *fullTextSearchSettingService) TriggerRebuild(triggeredBy uint) (*FullTextSearchRebuildPayload, error) {
	if s.dispatcher == nil {
		return nil, fmt.Errorf("queue dispatcher is not initialized")
	}

	job, err := s.dispatcher.EnqueueFullTextRebuild(triggeredBy)
	if err != nil {
		return nil, err
	}

	return &FullTextSearchRebuildPayload{
		JobID:       job.ID,
		QueueKey:    job.QueueKey,
		Status:      job.Status,
		ResourceID:  job.ResourceID,
		ScheduledAt: job.ScheduledAt.Unix(),
	}, nil
}
