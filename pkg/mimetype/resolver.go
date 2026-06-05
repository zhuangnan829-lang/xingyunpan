package mimetype

import (
	"mime"
	"path/filepath"
	"strings"
)

var officeTypes = map[string]string{
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".pps":  "application/vnd.ms-powerpoint",
	".ppsx": "application/vnd.openxmlformats-officedocument.presentationml.slideshow",
	".pot":  "application/vnd.ms-powerpoint",
	".potx": "application/vnd.openxmlformats-officedocument.presentationml.template",
}

// FromFileName resolves a content type from a file extension with explicit Office overrides.
func FromFileName(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return "application/octet-stream"
	}

	if value := strings.TrimSpace(officeTypes[ext]); value != "" {
		return value
	}

	if value := strings.TrimSpace(mime.TypeByExtension(ext)); value != "" {
		return value
	}

	return "application/octet-stream"
}

// IsGeneric reports whether a content type lacks useful type information.
func IsGeneric(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return contentType == "" ||
		contentType == "application/octet-stream" ||
		contentType == "binary/octet-stream" ||
		contentType == "application/zip"
}
