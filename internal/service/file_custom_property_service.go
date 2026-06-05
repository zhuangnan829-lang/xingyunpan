package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// FileCustomPropertyDefinitionPayload is the API-facing custom property definition.
type FileCustomPropertyDefinitionPayload struct {
	ID           int      `json:"id"`
	Key          string   `json:"key"`
	Name         string   `json:"name"`
	Icon         string   `json:"icon"`
	Type         string   `json:"type"`
	MinLength    *int     `json:"minLength"`
	MaxLength    *int     `json:"maxLength"`
	MaxValue     *int     `json:"maxValue"`
	Options      []string `json:"options"`
	DefaultValue string   `json:"defaultValue"`
}

// FileCustomPropertyPayload is the API response for one file's custom property values.
type FileCustomPropertyPayload struct {
	FileID       uint                                  `json:"file_id"`
	Definitions  []FileCustomPropertyDefinitionPayload `json:"definitions"`
	Values       map[string]string                     `json:"values"`
	LastModified *string                               `json:"last_modified,omitempty"`
}

// UpdateFileCustomPropertyPayload is the update request for one file's custom property values.
type UpdateFileCustomPropertyPayload struct {
	Values map[string]string `json:"values"`
}

// FileCustomPropertyService handles per-file custom property values.
type FileCustomPropertyService interface {
	Get(userID, fileID uint) (*FileCustomPropertyPayload, error)
	Update(userID, fileID uint, payload *UpdateFileCustomPropertyPayload) (*FileCustomPropertyPayload, error)
}

type fileCustomPropertyService struct {
	userFileRepo           repository.UserFileRepository
	fileSystemSettingRepo  repository.FileSystemSettingRepository
	fileCustomPropertyRepo repository.FileCustomPropertyValueRepository
}

// NewFileCustomPropertyService creates a file custom property service.
func NewFileCustomPropertyService(
	userFileRepo repository.UserFileRepository,
	fileSystemSettingRepo repository.FileSystemSettingRepository,
	fileCustomPropertyRepo repository.FileCustomPropertyValueRepository,
) FileCustomPropertyService {
	return &fileCustomPropertyService{
		userFileRepo:           userFileRepo,
		fileSystemSettingRepo:  fileSystemSettingRepo,
		fileCustomPropertyRepo: fileCustomPropertyRepo,
	}
}

func (s *fileCustomPropertyService) Get(userID, fileID uint) (*FileCustomPropertyPayload, error) {
	file, definitions, err := s.loadContext(userID, fileID)
	if err != nil {
		return nil, err
	}

	valueRecord, err := s.fileCustomPropertyRepo.GetByFileID(file.ID)
	if err != nil {
		return nil, err
	}

	values := s.normalizeValues(definitions, "")
	var lastModified *string
	if valueRecord != nil {
		values = s.normalizeValues(definitions, valueRecord.ValuesJSON)
		timestamp := valueRecord.UpdatedAt.Format("2006-01-02 15:04:05")
		lastModified = &timestamp
	}

	return &FileCustomPropertyPayload{
		FileID:       file.ID,
		Definitions:  definitions,
		Values:       values,
		LastModified: lastModified,
	}, nil
}

func (s *fileCustomPropertyService) Update(userID, fileID uint, payload *UpdateFileCustomPropertyPayload) (*FileCustomPropertyPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("custom property payload cannot be nil")
	}

	file, definitions, err := s.loadContext(userID, fileID)
	if err != nil {
		return nil, err
	}

	normalizedValues, err := normalizeCustomPropertyValueMap(definitions, payload.Values)
	if err != nil {
		return nil, err
	}

	valuesJSON, err := json.Marshal(normalizedValues)
	if err != nil {
		return nil, fmt.Errorf("marshal custom property values failed: %w", err)
	}

	valueRecord, err := s.fileCustomPropertyRepo.GetByFileID(file.ID)
	if err != nil {
		return nil, err
	}
	if valueRecord == nil {
		valueRecord = &model.FileCustomPropertyValue{
			FileID: file.ID,
			UserID: userID,
		}
	}
	valueRecord.ValuesJSON = string(valuesJSON)

	if err := s.fileCustomPropertyRepo.Save(valueRecord); err != nil {
		return nil, err
	}

	timestamp := valueRecord.UpdatedAt.Format("2006-01-02 15:04:05")
	return &FileCustomPropertyPayload{
		FileID:       file.ID,
		Definitions:  definitions,
		Values:       normalizedValues,
		LastModified: &timestamp,
	}, nil
}

func (s *fileCustomPropertyService) loadContext(userID, fileID uint) (*model.UserFile, []FileCustomPropertyDefinitionPayload, error) {
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("file not found")
	}
	if file.UserID != userID {
		return nil, nil, fmt.Errorf("no permission to access this file")
	}
	if file.IsFolder {
		return nil, nil, fmt.Errorf("custom properties are only supported for files")
	}

	setting, err := s.fileSystemSettingRepo.Get()
	if err != nil {
		return nil, nil, err
	}

	definitions, err := parseConfiguredCustomProperties("")
	if setting != nil {
		definitions, err = parseConfiguredCustomProperties(setting.CustomProperties)
	}
	if err != nil {
		return nil, nil, err
	}

	return file, definitions, nil
}

func parseConfiguredCustomProperties(raw string) ([]FileCustomPropertyDefinitionPayload, error) {
	if strings.TrimSpace(raw) == "" {
		return []FileCustomPropertyDefinitionPayload{}, nil
	}

	var parsed []FileCustomPropertyDefinitionPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("parse configured custom properties failed: %w", err)
	}

	definitions := make([]FileCustomPropertyDefinitionPayload, 0, len(parsed))
	for index, item := range parsed {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			key = fmt.Sprintf("property_%d", index+1)
		}

		next := item
		next.Key = key
		next.Name = strings.TrimSpace(next.Name)
		if next.Name == "" {
			next.Name = fmt.Sprintf("属性 %d", index+1)
		}
		switch next.Type {
		case "rating", "switch", "date", "tags", "multi_select":
		default:
			next.Type = "text"
		}
		next.Options = normalizeCustomPropertyOptions(next.Options)
		definitions = append(definitions, next)
	}

	return definitions, nil
}

func normalizeCustomPropertyValueMap(definitions []FileCustomPropertyDefinitionPayload, source map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		rawValue := ""
		if source != nil {
			rawValue = strings.TrimSpace(source[definition.Key])
		}

		if definition.Type == "rating" {
			maxValue := 5
			if definition.MaxValue != nil {
				maxValue = clampInt(*definition.MaxValue, 1, 10)
			}
			value := clampInt(parseOptionalInt(rawValue, parseOptionalInt(definition.DefaultValue, 0)), 0, maxValue)
			result[definition.Key] = fmt.Sprintf("%d", value)
			continue
		}

		if definition.Type == "switch" {
			if rawValue != "true" {
				rawValue = "false"
			}
			if source != nil {
				if source[definition.Key] == "true" {
					rawValue = "true"
				}
			}
			if rawValue == "" {
				rawValue = definition.DefaultValue
				if rawValue != "true" {
					rawValue = "false"
				}
			}
			result[definition.Key] = rawValue
			continue
		}

		if definition.Type == "date" {
			value := rawValue
			if value == "" {
				value = strings.TrimSpace(definition.DefaultValue)
			}
			result[definition.Key] = value
			continue
		}

		if definition.Type == "tags" || definition.Type == "multi_select" {
			result[definition.Key] = normalizeCustomPropertyArrayValue(rawValue, definition.Options)
			if strings.TrimSpace(rawValue) == "" && strings.TrimSpace(definition.DefaultValue) != "" {
				result[definition.Key] = normalizeCustomPropertyArrayValue(definition.DefaultValue, definition.Options)
			}
			continue
		}

		value := rawValue
		if value == "" {
			value = strings.TrimSpace(definition.DefaultValue)
		}
		if definition.MinLength != nil && len(value) < *definition.MinLength && value != "" {
			return nil, fmt.Errorf("%s 长度不能少于 %d", definition.Name, *definition.MinLength)
		}
		if definition.MaxLength != nil && len(value) > *definition.MaxLength {
			return nil, fmt.Errorf("%s 长度不能超过 %d", definition.Name, *definition.MaxLength)
		}
		result[definition.Key] = value
	}

	return result, nil
}

func (s *fileCustomPropertyService) normalizeValues(definitions []FileCustomPropertyDefinitionPayload, raw string) map[string]string {
	values := map[string]string{}
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &values)
	}

	normalized, err := normalizeCustomPropertyValueMap(definitions, values)
	if err != nil {
		return map[string]string{}
	}
	return normalized
}
