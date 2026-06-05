package service

import (
	"encoding/json"
	"fmt"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type AppearanceThemePalettePayload struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

type AppearanceThemeOptionPayload struct {
	ID        int                           `json:"id"`
	IsDefault bool                          `json:"is_default"`
	Light     AppearanceThemePalettePayload `json:"light"`
	Dark      AppearanceThemePalettePayload `json:"dark"`
}

type AppearanceFeatureOptionPayload struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type AppearanceSettingPayload struct {
	ThemeOptions    []AppearanceThemeOptionPayload   `json:"theme_options"`
	FeatureOptions  []AppearanceFeatureOptionPayload `json:"feature_options"`
	SelectedThemeID int                              `json:"selected_theme_id"`
}

type AppearanceSettingService interface {
	Get() (*AppearanceSettingPayload, error)
	Update(payload *AppearanceSettingPayload) (*AppearanceSettingPayload, error)
}

type appearanceSettingService struct {
	repo repository.SiteSettingRepository
}

func NewAppearanceSettingService(repo repository.SiteSettingRepository) AppearanceSettingService {
	return &appearanceSettingService{repo: repo}
}

func (s *appearanceSettingService) Get() (*AppearanceSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil || setting.AppearanceSettingsJSON == "" {
		return defaultAppearanceSettingPayload(), nil
	}

	var payload AppearanceSettingPayload
	if err := json.Unmarshal([]byte(setting.AppearanceSettingsJSON), &payload); err != nil {
		return nil, fmt.Errorf("解析外观设置失败: %w", err)
	}

	normalized := normalizeAppearanceSettingPayload(&payload)
	return normalized, nil
}

func (s *appearanceSettingService) Update(payload *AppearanceSettingPayload) (*AppearanceSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("外观设置不能为空")
	}

	normalized := normalizeAppearanceSettingPayload(payload)
	if len(normalized.ThemeOptions) == 0 {
		return nil, fmt.Errorf("至少需要一个主题选项")
	}

	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("序列化外观设置失败: %w", err)
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.SiteSetting{}
	}

	setting.AppearanceSettingsJSON = string(data)
	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return normalized, nil
}

func defaultAppearanceSettingPayload() *AppearanceSettingPayload {
	return &AppearanceSettingPayload{
		ThemeOptions: []AppearanceThemeOptionPayload{
			{
				ID:        1,
				IsDefault: true,
				Light:     AppearanceThemePalettePayload{Primary: "#1976d2", Secondary: "#9c27b0"},
				Dark:      AppearanceThemePalettePayload{Primary: "#90caf9", Secondary: "#ce93d8"},
			},
			{
				ID:        2,
				IsDefault: false,
				Light:     AppearanceThemePalettePayload{Primary: "#3f51b5", Secondary: "#f50057"},
				Dark:      AppearanceThemePalettePayload{Primary: "#9fa8da", Secondary: "#ff4081"},
			},
		},
		FeatureOptions: []AppearanceFeatureOptionPayload{
			{Key: "neon", Label: "启用霓虹细节 (Neon Accents)", Description: "为色块、按钮和边框加入更亮的蓝色辉光。", Enabled: true},
			{Key: "glass", Label: "玻璃拟态元素 (Glassmorphism Elements)", Description: "增强纯白界面中的轻玻璃感和层次感。", Enabled: true},
			{Key: "motion", Label: "动态过渡动画 (Animated Transitions)", Description: "增加卡片切换与控件悬停时的细腻反馈。", Enabled: true},
			{Key: "data-viz", Label: "数据可视化优化", Description: "强化预览卡片中的进度条与发光效果。", Enabled: true},
		},
		SelectedThemeID: 1,
	}
}

func normalizeAppearanceSettingPayload(payload *AppearanceSettingPayload) *AppearanceSettingPayload {
	normalized := *payload
	if len(normalized.ThemeOptions) == 0 {
		normalized.ThemeOptions = defaultAppearanceSettingPayload().ThemeOptions
	}
	if len(normalized.FeatureOptions) == 0 {
		normalized.FeatureOptions = defaultAppearanceSettingPayload().FeatureOptions
	}

	if normalized.SelectedThemeID == 0 {
		for _, item := range normalized.ThemeOptions {
			if item.IsDefault {
				normalized.SelectedThemeID = item.ID
				break
			}
		}
	}
	if normalized.SelectedThemeID == 0 && len(normalized.ThemeOptions) > 0 {
		normalized.SelectedThemeID = normalized.ThemeOptions[0].ID
	}

	for index := range normalized.ThemeOptions {
		normalized.ThemeOptions[index].IsDefault = normalized.ThemeOptions[index].ID == normalized.SelectedThemeID
	}

	return &normalized
}
