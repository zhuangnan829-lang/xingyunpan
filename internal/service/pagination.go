package service

import (
	"strconv"
	"strings"

	"xingyunpan-v2/internal/repository"
)

const (
	PaginationModeOffset = "offset"
	PaginationModeCursor = "cursor"
	PaginationModeHybrid = "hybrid"
)

type PaginationRuntime struct {
	Page            int
	PageSize        int
	MaxPageSize     int
	Mode            string
	Cursor          uint
	UseCursor       bool
	RequestedCursor string
}

type PaginationResponseMeta struct {
	PaginationMode string `json:"pagination_mode"`
	NextCursor     string `json:"next_cursor"`
	PageSize       int    `json:"page_size"`
	MaxPageSize    int    `json:"max_page_size"`
}

func ResolvePagination(settingsRepo repository.FileSystemSettingRepository, page, pageSize int, cursor string, defaultPageSize int) PaginationRuntime {
	maxPageSize := defaultFileSystemSettingPayload().MaxPageSize
	mode := defaultFileSystemSettingPayload().ListPaginationMode
	if settingsRepo != nil {
		if setting, err := settingsRepo.Get(); err == nil && setting != nil {
			if setting.MaxPageSize > 0 {
				maxPageSize = setting.MaxPageSize
			}
			if normalizedMode := normalizePaginationMode(setting.ListPaginationMode); normalizedMode != "" {
				mode = normalizedMode
			}
		}
	}

	if page < 1 {
		page = 1
	}
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	cursor = strings.TrimSpace(cursor)
	cursorID := uint(0)
	if parsed, err := strconv.ParseUint(cursor, 10, 64); err == nil {
		cursorID = uint(parsed)
	}
	useCursor := mode == PaginationModeCursor || (mode == PaginationModeHybrid && cursorID > 0)

	return PaginationRuntime{
		Page:            page,
		PageSize:        pageSize,
		MaxPageSize:     maxPageSize,
		Mode:            mode,
		Cursor:          cursorID,
		UseCursor:       useCursor,
		RequestedCursor: cursor,
	}
}

func PaginationMeta(runtime PaginationRuntime, nextCursor string) PaginationResponseMeta {
	return PaginationResponseMeta{
		PaginationMode: runtime.Mode,
		NextCursor:     nextCursor,
		PageSize:       runtime.PageSize,
		MaxPageSize:    runtime.MaxPageSize,
	}
}

func NextCursorFromUint(id uint) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(id), 10)
}

func normalizePaginationMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case PaginationModeOffset, PaginationModeCursor, PaginationModeHybrid:
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return ""
	}
}
