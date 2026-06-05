package controller

import (
	"strconv"
	"strings"

	"xingyunpan-v2/internal/service"
)

func resolveControllerPagination(settings *service.FileSystemSettingPayload, page, pageSize int, cursor string, defaultPageSize int) service.PaginationRuntime {
	if page < 1 {
		page = 1
	}
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}

	maxPageSize := 2000
	mode := service.PaginationModeCursor
	if settings != nil {
		if settings.MaxPageSize > 0 {
			maxPageSize = settings.MaxPageSize
		}
		switch strings.ToLower(strings.TrimSpace(settings.ListPaginationMode)) {
		case service.PaginationModeOffset, service.PaginationModeCursor, service.PaginationModeHybrid:
			mode = strings.ToLower(strings.TrimSpace(settings.ListPaginationMode))
		}
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	cursor = strings.TrimSpace(cursor)
	cursorID := uint(0)
	if parsed, err := strconv.ParseUint(cursor, 10, 64); err == nil {
		cursorID = uint(parsed)
	}

	return service.PaginationRuntime{
		Page:            page,
		PageSize:        pageSize,
		MaxPageSize:     maxPageSize,
		Mode:            mode,
		Cursor:          cursorID,
		UseCursor:       mode == service.PaginationModeCursor || (mode == service.PaginationModeHybrid && cursorID > 0),
		RequestedCursor: cursor,
	}
}
