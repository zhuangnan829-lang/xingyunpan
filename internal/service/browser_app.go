package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type BrowserAppPlatform string

const (
	BrowserAppPlatformAll     BrowserAppPlatform = "all"
	BrowserAppPlatformDesktop BrowserAppPlatform = "desktop"
	BrowserAppPlatformMobile  BrowserAppPlatform = "mobile"
)

type BrowserAppItemPayload struct {
	ID              int                `json:"id"`
	Icon            string             `json:"icon"`
	IconURL         string             `json:"icon_url"`
	Accent          string             `json:"accent"`
	Type            string             `json:"type"`
	Name            string             `json:"name"`
	Extensions      string             `json:"extensions"`
	Platform        BrowserAppPlatform `json:"platform"`
	CreateMapping   string             `json:"create_mapping"`
	Enabled         bool               `json:"enabled"`
	OpenInNewWindow bool               `json:"open_in_new_window"`
	MaxSize         int                `json:"max_size"`
	MaxSizeUnit     string             `json:"max_size_unit"`
}

type BrowserAppGroupPayload struct {
	ID          int                     `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Items       []BrowserAppItemPayload `json:"items"`
}

type BrowserAppResolvedPayload struct {
	GroupID          int                `json:"group_id"`
	GroupName        string             `json:"group_name"`
	ID               int                `json:"id"`
	Icon             string             `json:"icon"`
	IconURL          string             `json:"icon_url"`
	Accent           string             `json:"accent"`
	Type             string             `json:"type"`
	Name             string             `json:"name"`
	Platform         BrowserAppPlatform `json:"platform"`
	CreateMapping    string             `json:"create_mapping"`
	OpenInNewWindow  bool               `json:"open_in_new_window"`
	MaxSize          int                `json:"max_size"`
	MaxSizeUnit      string             `json:"max_size_unit"`
	MatchedExtension string             `json:"matched_extension"`
	Source           string             `json:"source"`
}

func defaultBrowserAppGroupsPayload() []BrowserAppGroupPayload {
	return []BrowserAppGroupPayload{
		{
			ID:          1,
			Name:        "应用分组 #1",
			Description: "用于统一管理预览器、编辑器和新建文件映射，尽量贴近 Cloudreve 的内置应用编排方式。",
			Items: []BrowserAppItemPayload{
				{ID: 1, Icon: "PDF", Accent: "#ef4444", Type: "内置应用", Name: "PDF 阅读器", Extensions: "pdf", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: true, MaxSize: 300, MaxSizeUnit: "MB"},
				{ID: 2, Icon: "▶", Accent: "#22c3ee", Type: "内置应用", Name: "Artplayer", Extensions: "mp4,mkv,webm,avi,mov,m3u8,flv", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: false, MaxSize: 2048, MaxSizeUnit: "MB"},
				{ID: 3, Icon: "MD", Accent: "#3f3f46", Type: "内置应用", Name: "Markdown 编辑器", Extensions: "md,markdown", Platform: BrowserAppPlatformAll, CreateMapping: "1 个", Enabled: true, OpenInNewWindow: false, MaxSize: 20, MaxSizeUnit: "MB"},
				{ID: 4, Icon: "Dx", Accent: "#f59e0b", Type: "内置应用", Name: "draw.io", Extensions: "drawio,dwb", Platform: BrowserAppPlatformAll, CreateMapping: "2 个", Enabled: true, OpenInNewWindow: true, MaxSize: 50, MaxSizeUnit: "MB"},
				{ID: 5, Icon: "图", Accent: "#dc2626", Type: "内置应用", Name: "图片查看器", Extensions: "bmp,png,gif,jpg,jpeg,svg,webp,heic,heif", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: true, MaxSize: 512, MaxSizeUnit: "MB"},
				{ID: 6, Icon: "JS", Accent: "#2563eb", Type: "内置应用", Name: "Monaco 代码编辑器", Extensions: "md,txt,json,php,py,bat,c,h,cpp,hpp,cs,css,dockerfile,go,html,htm,ini,java,js,jsx,less,lua,sh,sql,xml,yaml", Platform: BrowserAppPlatformAll, CreateMapping: "1 个", Enabled: true, OpenInNewWindow: false, MaxSize: 25, MaxSizeUnit: "MB"},
				{ID: 7, Icon: "PS", Accent: "#0ea5a4", Type: "内置应用", Name: "Photopea", Extensions: "psd,ai,indd,xcf,xd,fig,kri,clip,pxd,pxz,cdr,ufo,afphoto,svg,eps,pdf,pdn,wmf,emf,png,jpg,jpeg,gif,webp,ico,icns,bmp,avif,heic,jxl,ppm,pgm,pbm,tiff,dds,iff,anim,tga,dng,nef,cr2,cr3,arw,rw2,raf,orf,gpr,3fr,fff", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: true, MaxSize: 1024, MaxSizeUnit: "MB"},
				{ID: 8, Icon: "Ex", Accent: "#6366f1", Type: "内置应用", Name: "Excalidraw", Extensions: "excalidraw", Platform: BrowserAppPlatformAll, CreateMapping: "1 个", Enabled: true, OpenInNewWindow: true, MaxSize: 20, MaxSizeUnit: "MB"},
				{ID: 9, Icon: "ZIP", Accent: "#f59e0b", Type: "内置应用", Name: "压缩包预览", Extensions: "zip,7z", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: false, MaxSize: 512, MaxSizeUnit: "MB"},
				{ID: 10, Icon: "♪", Accent: "#7c3aed", Type: "内置应用", Name: "音频播放器", Extensions: "mp3,ogg,wav,flac,m4a", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: false, MaxSize: 1024, MaxSizeUnit: "MB"},
				{ID: 11, Icon: "EP", Accent: "#65a30d", Type: "内置应用", Name: "ePub 阅读器", Extensions: "epub", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: true, MaxSize: 100, MaxSizeUnit: "MB"},
				{ID: 12, Icon: "GD", Accent: "#22c55e", Type: "自定义", Name: "Google Docs 在线预览器", Extensions: "jpeg,png,gif,tiff,bmp,webm,mpeg4,3gp,mov,avi,mpe,gps,wmv,flv,txt,css,html,php,cpp,h,pp,js,doc,docx,xls,xlsx,ppt,pptx,pdf,pages,ai,psd,tiff,dxf,svg,eps,ps,ttf,xps", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: true, MaxSize: 1024, MaxSizeUnit: "MB"},
				{ID: 13, Icon: "O", Accent: "#8b5cf6", Type: "自定义", Name: "Microsoft Office 在线预览器", Extensions: "doc,docx,docm,dotm,dotx,xlsx,xlsb,xls,xlsm,pptx,ppsx,ppt,pps,pptm,potm,ppam,potx,ppsm", Platform: BrowserAppPlatformAll, CreateMapping: "无", Enabled: true, OpenInNewWindow: true, MaxSize: 1024, MaxSizeUnit: "MB"},
			},
		},
	}
}

func parseBrowserAppsJSON(raw string, useDefaults bool) ([]BrowserAppGroupPayload, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		if useDefaults {
			return normalizeBrowserAppGroups(defaultBrowserAppGroupsPayload()), nil
		}
		return []BrowserAppGroupPayload{}, nil
	}

	var groups []BrowserAppGroupPayload
	if err := json.Unmarshal([]byte(trimmed), &groups); err != nil {
		return nil, fmt.Errorf("browser apps must be a valid json array: %w", err)
	}

	groups = normalizeBrowserAppGroups(groups)
	if len(groups) == 0 && useDefaults {
		return normalizeBrowserAppGroups(defaultBrowserAppGroupsPayload()), nil
	}

	return groups, nil
}

func normalizeBrowserAppGroups(groups []BrowserAppGroupPayload) []BrowserAppGroupPayload {
	fallback := defaultBrowserAppGroupsPayload()
	normalized := make([]BrowserAppGroupPayload, 0, len(groups))
	nextGroupID := 1
	nextItemID := 1

	for groupIndex, group := range groups {
		fallbackGroup := fallbackGroupAt(fallback, groupIndex, group.ID)
		items := make([]BrowserAppItemPayload, 0, len(group.Items))

		for itemIndex, item := range group.Items {
			preset := fallbackItemFor(fallbackGroup, itemIndex, item)
			rawMaxSize := item.MaxSize
			rawMaxSizeUnit := item.MaxSizeUnit
			normalizedExtensions := normalizeBrowserAppExtensions(item.Extensions)
			if len(normalizedExtensions) == 0 && preset != nil {
				normalizedExtensions = normalizeBrowserAppExtensions(preset.Extensions)
			}
			if len(normalizedExtensions) == 0 {
				continue
			}

			if item.ID <= 0 {
				item.ID = nextItemID
			}
			nextItemID = item.ID + 1

			item.Icon = restoreBrokenBrowserAppText(item.Icon, valueOrDefaultItem(preset, func(candidate BrowserAppItemPayload) string { return candidate.Icon }, "APP"))
			item.Icon = truncateRunes(item.Icon, 4)
			item.IconURL = strings.TrimSpace(item.IconURL)
			item.Accent = strings.TrimSpace(item.Accent)
			if item.Accent == "" {
				item.Accent = valueOrDefaultItem(preset, func(candidate BrowserAppItemPayload) string { return candidate.Accent }, "#2563eb")
			}
			item.Type = restoreBrokenBrowserAppText(item.Type, valueOrDefaultItem(preset, func(candidate BrowserAppItemPayload) string { return candidate.Type }, "内置应用"))
			item.Name = restoreBrokenBrowserAppText(item.Name, valueOrDefaultItem(preset, func(candidate BrowserAppItemPayload) string { return candidate.Name }, fmt.Sprintf("应用 %d", itemIndex+1)))
			item.Extensions = strings.Join(normalizedExtensions, ",")
			item.Platform = normalizeBrowserAppPlatform(item.Platform)
			item.CreateMapping = restoreBrokenBrowserAppText(item.CreateMapping, valueOrDefaultItem(preset, func(candidate BrowserAppItemPayload) string { return candidate.CreateMapping }, ""))
			item.MaxSize = clampInt(item.MaxSize, 1, 102400)
			if rawMaxSize <= 0 && preset != nil && preset.MaxSize > 0 {
				item.MaxSize = preset.MaxSize
			}
			item.MaxSizeUnit = normalizeBrowserAppUnit(item.MaxSizeUnit)
			if strings.TrimSpace(rawMaxSizeUnit) == "" && preset != nil {
				item.MaxSizeUnit = preset.MaxSizeUnit
			}
			if item.MaxSizeUnit == "" {
				item.MaxSizeUnit = "MB"
			}

			items = append(items, item)
		}

		if len(items) == 0 {
			continue
		}

		group.ID = clampInt(group.ID, 0, 1_000_000)
		if group.ID == 0 {
			group.ID = nextGroupID
		}
		nextGroupID = group.ID + 1
		group.Name = restoreBrokenBrowserAppText(group.Name, valueOrDefaultGroup(fallbackGroup, func(candidate BrowserAppGroupPayload) string { return candidate.Name }, fmt.Sprintf("应用分组 #%d", groupIndex+1)))
		group.Description = restoreBrokenBrowserAppText(group.Description, valueOrDefaultGroup(fallbackGroup, func(candidate BrowserAppGroupPayload) string { return candidate.Description }, "用于统一管理预览器、编辑器和新建文件映射。"))
		group.Items = items

		normalized = append(normalized, group)
	}

	return normalized
}

func normalizeBrowserAppExtensions(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		ext := strings.TrimSpace(part)
		ext = strings.TrimPrefix(ext, ".")
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == "" {
			continue
		}
		if _, exists := seen[ext]; exists {
			continue
		}
		seen[ext] = struct{}{}
		normalized = append(normalized, ext)
	}
	return normalized
}

func normalizeBrowserAppPlatform(platform BrowserAppPlatform) BrowserAppPlatform {
	switch BrowserAppPlatform(strings.ToLower(strings.TrimSpace(string(platform)))) {
	case BrowserAppPlatformDesktop:
		return BrowserAppPlatformDesktop
	case BrowserAppPlatformMobile:
		return BrowserAppPlatformMobile
	default:
		return BrowserAppPlatformAll
	}
}

func normalizeBrowserAppUnit(unit string) string {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "B", "KB", "MB", "GB", "TB":
		return strings.ToUpper(strings.TrimSpace(unit))
	default:
		return "MB"
	}
}

func ResolveBrowserAppFromGroups(groups []BrowserAppGroupPayload, filename, mimeType string, platform BrowserAppPlatform) *BrowserAppResolvedPayload {
	targetExt := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(extensionFromName(filename))), ".")
	if targetExt == "" {
		targetExt = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(filename)), ".")
	}
	if targetExt == "" && strings.Contains(mimeType, "/") {
		targetExt = strings.TrimSpace(strings.ToLower(strings.SplitN(mimeType, "/", 2)[1]))
	}
	if targetExt == "" {
		return nil
	}

	requestPlatform := normalizeBrowserAppPlatform(platform)
	for _, group := range groups {
		for _, item := range group.Items {
			if !item.Enabled {
				continue
			}
			if item.Platform != BrowserAppPlatformAll && item.Platform != requestPlatform {
				continue
			}
			for _, ext := range normalizeBrowserAppExtensions(item.Extensions) {
				if ext != targetExt {
					continue
				}
				return &BrowserAppResolvedPayload{
					GroupID:          group.ID,
					GroupName:        group.Name,
					ID:               item.ID,
					Icon:             item.Icon,
					IconURL:          item.IconURL,
					Accent:           item.Accent,
					Type:             item.Type,
					Name:             item.Name,
					Platform:         item.Platform,
					CreateMapping:    item.CreateMapping,
					OpenInNewWindow:  item.OpenInNewWindow,
					MaxSize:          item.MaxSize,
					MaxSizeUnit:      item.MaxSizeUnit,
					MatchedExtension: ext,
					Source:           "browser_apps",
				}
			}
		}
	}

	return nil
}

func extensionFromName(filename string) string {
	base := filepath.Base(strings.TrimSpace(filename))
	if base == "." || base == "" {
		return ""
	}
	return filepath.Ext(base)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func isBrokenBrowserAppText(value string) bool {
	text := strings.TrimSpace(value)
	if text == "" {
		return true
	}
	return strings.Contains(text, "?") ||
		text == "App Group #1" ||
		text == "Office Online" ||
		text == "PDF Viewer" ||
		text == "Markdown Studio" ||
		text == "Media Deck" ||
		text == "Codex Persistence Demo"
}

func restoreBrokenBrowserAppText(value, fallback string) string {
	text := strings.TrimSpace(value)
	if isBrokenBrowserAppText(text) {
		return fallback
	}
	return text
}

func fallbackGroupAt(groups []BrowserAppGroupPayload, index, id int) *BrowserAppGroupPayload {
	for _, group := range groups {
		if id > 0 && group.ID == id {
			candidate := group
			return &candidate
		}
	}
	if index >= 0 && index < len(groups) {
		candidate := groups[index]
		return &candidate
	}
	return nil
}

func fallbackItemFor(group *BrowserAppGroupPayload, index int, item BrowserAppItemPayload) *BrowserAppItemPayload {
	if group == nil {
		return nil
	}
	for _, candidate := range group.Items {
		if item.ID > 0 && candidate.ID == item.ID {
			copy := candidate
			return &copy
		}
	}
	itemName := strings.TrimSpace(strings.ToLower(item.Name))
	if itemName != "" {
		for _, candidate := range group.Items {
			if strings.TrimSpace(strings.ToLower(candidate.Name)) == itemName {
				copy := candidate
				return &copy
			}
		}
	}
	itemIcon := strings.TrimSpace(strings.ToLower(item.Icon))
	if itemIcon != "" {
		for _, candidate := range group.Items {
			if strings.TrimSpace(strings.ToLower(candidate.Icon)) == itemIcon {
				copy := candidate
				return &copy
			}
		}
	}
	if index >= 0 && index < len(group.Items) {
		copy := group.Items[index]
		return &copy
	}
	return nil
}

func valueOrDefaultItem(candidate *BrowserAppItemPayload, getter func(BrowserAppItemPayload) string, fallback string) string {
	if candidate == nil {
		return fallback
	}
	value := strings.TrimSpace(getter(*candidate))
	if value == "" {
		return fallback
	}
	return value
}

func valueOrDefaultGroup(candidate *BrowserAppGroupPayload, getter func(BrowserAppGroupPayload) string, fallback string) string {
	if candidate == nil {
		return fallback
	}
	value := strings.TrimSpace(getter(*candidate))
	if value == "" {
		return fallback
	}
	return value
}
