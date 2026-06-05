package service

import (
	"net/url"
	"strings"

	"xingyunpan-v2/internal/repository"
)

func configuredCategoryQuery(repo repository.FileSystemSettingRepository, category string) string {
	if repo == nil {
		return ""
	}

	setting, err := repo.Get()
	if err != nil || setting == nil {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(category)) {
	case "image":
		return setting.ImageQuery
	case "video":
		return setting.VideoQuery
	case "audio", "music":
		return setting.AudioQuery
	case "document":
		return setting.DocumentQuery
	default:
		return ""
	}
}

func categoryExtensionsFromQuery(raw string, fallback []string) []string {
	extensions := parseCategoryQueryExtensions(raw)
	if len(extensions) == 0 {
		return append([]string{}, fallback...)
	}
	return extensions
}

func parseCategoryQueryExtensions(raw string) []string {
	values, err := url.ParseQuery(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}

	seen := map[string]struct{}{}
	result := make([]string, 0, len(values["name"]))
	for _, name := range values["name"] {
		for _, token := range strings.FieldsFunc(name, func(r rune) bool {
			return r == ',' || r == ';' || r == '|'
		}) {
			ext := normalizeCategoryExtension(token)
			if ext == "" {
				continue
			}
			if _, ok := seen[ext]; ok {
				continue
			}
			seen[ext] = struct{}{}
			result = append(result, ext)
		}
	}
	return result
}

func normalizeCategoryExtension(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.TrimPrefix(value, "*")
	value = strings.TrimPrefix(value, ".")
	value = strings.TrimPrefix(value, "*.")
	value = strings.Trim(value, " ")
	if value == "" || strings.ContainsAny(value, "/\\") || strings.Contains(value, "*") {
		return ""
	}
	return value
}

func categoryPredicate(contentPatterns, extensions []string) (string, []interface{}) {
	parts := make([]string, 0, len(contentPatterns)+len(extensions))
	args := make([]interface{}, 0, len(contentPatterns)+len(extensions))

	for _, pattern := range contentPatterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		parts = append(parts, "LOWER(COALESCE(physical_files.content_type, '')) LIKE ?")
		args = append(args, pattern)
	}
	for _, ext := range extensions {
		ext = normalizeCategoryExtension(ext)
		if ext == "" {
			continue
		}
		parts = append(parts, "LOWER(user_files.file_name) LIKE ?")
		args = append(args, "%."+ext)
	}

	if len(parts) == 0 {
		return "1 = 1", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}
