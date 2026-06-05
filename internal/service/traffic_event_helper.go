package service

import (
	"strings"

	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
)

func recordTrafficEvent(db *gorm.DB, userID uint, direction string, bytes int64, source string, resourceType string, resourceID string) {
	if db == nil || bytes <= 0 {
		return
	}

	event := model.TrafficEvent{
		UserID:       userID,
		Direction:    strings.ToLower(strings.TrimSpace(direction)),
		Bytes:        bytes,
		Source:       strings.TrimSpace(source),
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
	}
	if event.Direction == "" {
		return
	}

	_ = db.Create(&event).Error
}

func sumUserFileSizes(files []*model.UserFile) int64 {
	var total int64
	for _, file := range files {
		if file == nil {
			continue
		}
		if file.FileSize > 0 {
			total += file.FileSize
			continue
		}
		if file.PhysicalFile != nil && file.PhysicalFile.FileSize > 0 {
			total += file.PhysicalFile.FileSize
		}
	}
	return total
}
