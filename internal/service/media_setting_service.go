package service

import (
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// MediaSettingPayload is the API-facing shape for admin media settings.
type MediaSettingPayload struct {
	ImageMode          string  `json:"image_mode"`
	ImageMaxSizeGB     int     `json:"image_max_size_gb"`
	ImageQuality       int     `json:"image_quality"`
	VideoPreviewSecond float64 `json:"video_preview_second"`
	VideoStrategy      string  `json:"video_strategy"`
	MetadataDeepScan   bool    `json:"metadata_deep_scan"`
	LibreOfficePath    string  `json:"libreoffice_path"`
}

// MediaSettingService provides admin access to media settings.
type MediaSettingService interface {
	Get() (*MediaSettingPayload, error)
	Update(payload *MediaSettingPayload) (*MediaSettingPayload, error)
}

type mediaSettingService struct {
	repo repository.MediaSettingRepository
}

// NewMediaSettingService creates a media settings service.
func NewMediaSettingService(repo repository.MediaSettingRepository) MediaSettingService {
	return &mediaSettingService{repo: repo}
}

// Get returns media settings, creating a default payload when no row exists yet.
func (s *mediaSettingService) Get() (*MediaSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return defaultMediaSettingPayload(), nil
	}

	return toMediaSettingPayload(setting), nil
}

// Update persists media settings.
func (s *mediaSettingService) Update(payload *MediaSettingPayload) (*MediaSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("media settings cannot be nil")
	}

	normalized, err := normalizeMediaSettingPayload(payload)
	if err != nil {
		return nil, err
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.MediaSetting{}
	}

	setting.ImageMode = normalized.ImageMode
	setting.ImageMaxSizeGB = normalized.ImageMaxSizeGB
	setting.ImageQuality = normalized.ImageQuality
	setting.VideoPreviewSecond = normalized.VideoPreviewSecond
	setting.VideoStrategy = normalized.VideoStrategy
	setting.MetadataDeepScan = normalized.MetadataDeepScan
	setting.LibreOfficePath = normalized.LibreOfficePath

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return toMediaSettingPayload(setting), nil
}

func toMediaSettingPayload(setting *model.MediaSetting) *MediaSettingPayload {
	return &MediaSettingPayload{
		ImageMode:          setting.ImageMode,
		ImageMaxSizeGB:     setting.ImageMaxSizeGB,
		ImageQuality:       setting.ImageQuality,
		VideoPreviewSecond: setting.VideoPreviewSecond,
		VideoStrategy:      setting.VideoStrategy,
		MetadataDeepScan:   setting.MetadataDeepScan,
		LibreOfficePath:    setting.LibreOfficePath,
	}
}

func defaultMediaSettingPayload() *MediaSettingPayload {
	return &MediaSettingPayload{
		ImageMode:          "quality",
		ImageMaxSizeGB:     8,
		ImageQuality:       88,
		VideoPreviewSecond: 1,
		VideoStrategy:      "balanced",
		MetadataDeepScan:   true,
		LibreOfficePath:    "soffice",
	}
}

func normalizeMediaSettingPayload(payload *MediaSettingPayload) (*MediaSettingPayload, error) {
	normalized := *payload
	normalized.ImageMode = strings.TrimSpace(strings.ToLower(normalized.ImageMode))
	normalized.VideoStrategy = strings.TrimSpace(strings.ToLower(normalized.VideoStrategy))
	normalized.LibreOfficePath = strings.TrimSpace(normalized.LibreOfficePath)

	switch normalized.ImageMode {
	case "quality", "compatibility":
	default:
		return nil, fmt.Errorf("unsupported image mode")
	}

	switch normalized.VideoStrategy {
	case "smooth", "balanced", "quality":
	default:
		return nil, fmt.Errorf("unsupported video strategy")
	}

	if normalized.ImageMaxSizeGB < 1 {
		normalized.ImageMaxSizeGB = 1
	}
	if normalized.ImageMaxSizeGB > 20 {
		normalized.ImageMaxSizeGB = 20
	}

	if normalized.ImageQuality < 40 {
		normalized.ImageQuality = 40
	}
	if normalized.ImageQuality > 100 {
		normalized.ImageQuality = 100
	}

	if normalized.VideoPreviewSecond < 0 {
		normalized.VideoPreviewSecond = 0
	}
	if normalized.VideoPreviewSecond > 12 {
		normalized.VideoPreviewSecond = 12
	}

	if normalized.LibreOfficePath == "" {
		return nil, fmt.Errorf("LibreOffice path cannot be empty")
	}

	return &normalized, nil
}
