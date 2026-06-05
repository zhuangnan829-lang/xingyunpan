package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// VideoService provides the video library data used by the drive video page.
type VideoService interface {
	ListVideos(ctx context.Context, userID uint, query *VideoListQuery) (*VideoListResult, error)
}

// VideoListQuery describes video library filters and pagination.
type VideoListQuery struct {
	Keyword  string `json:"keyword"`
	Sort     string `json:"sort"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Cursor   string `json:"cursor"`
}

// VideoListResult is the paginated video library response.
type VideoListResult struct {
	Files          []*VideoItem `json:"files"`
	Total          int64        `json:"total"`
	TotalSize      int64        `json:"total_size"`
	Page           int          `json:"page"`
	PageSize       int          `json:"page_size"`
	TotalPages     int          `json:"total_pages"`
	PaginationMode string       `json:"pagination_mode"`
	NextCursor     string       `json:"next_cursor"`
	MaxPageSize    int          `json:"max_page_size"`
}

// VideoItem mirrors the frontend file item shape for videos.
type VideoItem struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	Hash           string    `json:"hash"`
	MimeType       string    `json:"mime_type"`
	ContentType    string    `json:"content_type"`
	PhysicalFileID *uint     `json:"physical_file_id,omitempty"`
	ThumbnailURL   string    `json:"thumbnail_url,omitempty"`
	FolderID       *uint     `json:"folder_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type videoService struct {
	db           *gorm.DB
	settingsRepo repository.FileSystemSettingRepository
}

// NewVideoService creates a video library service.
func NewVideoService(db *gorm.DB, settingsRepo ...repository.FileSystemSettingRepository) VideoService {
	var repo repository.FileSystemSettingRepository
	if len(settingsRepo) > 0 {
		repo = settingsRepo[0]
	}
	return &videoService{db: db, settingsRepo: repo}
}

func (s *videoService) ListVideos(ctx context.Context, userID uint, query *VideoListQuery) (*VideoListResult, error) {
	if query == nil {
		query = &VideoListQuery{}
	}

	pagination := ResolvePagination(s.settingsRepo, query.Page, query.PageSize, query.Cursor, 50)
	page := pagination.Page
	pageSize := pagination.PageSize

	var total int64
	if err := s.videoBaseQuery(ctx, userID, strings.TrimSpace(query.Keyword)).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("query video total failed: %w", err)
	}

	var totalSize int64
	if err := s.videoBaseQuery(ctx, userID, strings.TrimSpace(query.Keyword)).
		Select("COALESCE(SUM(CASE WHEN user_files.file_size > 0 THEN user_files.file_size ELSE COALESCE(physical_files.file_size, 0) END), 0)").
		Scan(&totalSize).Error; err != nil {
		return nil, fmt.Errorf("query video total size failed: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	offset := mediaListOffset(pagination.UseCursor, page, pageSize)
	if total == 0 || offset >= int(total) {
		return &VideoListResult{
			Files:          []*VideoItem{},
			Total:          total,
			TotalSize:      totalSize,
			Page:           page,
			PageSize:       pageSize,
			TotalPages:     totalPages,
			PaginationMode: pagination.Mode,
			MaxPageSize:    pagination.MaxPageSize,
		}, nil
	}

	var files []*model.UserFile
	dataQuery := s.videoBaseQuery(ctx, userID, strings.TrimSpace(query.Keyword))
	if pagination.UseCursor && pagination.Cursor > 0 {
		dataQuery = dataQuery.Where("user_files.id < ?", pagination.Cursor)
	}
	if err := dataQuery.
		Preload("PhysicalFile").
		Order(videoOrderClause(query.Sort)).
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("query videos failed: %w", err)
	}

	items := make([]*VideoItem, 0, len(files))
	for _, file := range files {
		items = append(items, buildVideoItem(file))
	}
	nextCursor := ""
	if pagination.UseCursor && len(files) > 0 {
		nextCursor = NextCursorFromUint(files[len(files)-1].ID)
	}

	return &VideoListResult{
		Files:          items,
		Total:          total,
		TotalSize:      totalSize,
		Page:           page,
		PageSize:       pageSize,
		TotalPages:     totalPages,
		PaginationMode: pagination.Mode,
		NextCursor:     nextCursor,
		MaxPageSize:    pagination.MaxPageSize,
	}, nil
}

func (s *videoService) videoBaseQuery(ctx context.Context, userID uint, keyword string) *gorm.DB {
	query := s.db.WithContext(ctx).
		Model(&model.UserFile{}).
		Joins("LEFT JOIN physical_files ON physical_files.id = user_files.physical_file_id").
		Where("user_files.user_id = ?", userID).
		Where("user_files.deleted_at IS NULL").
		Where("user_files.is_folder = ?", false)

	clause, args := s.videoFilePredicate()
	query = query.Where(clause, args...)

	if keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("(LOWER(user_files.file_name) LIKE ? OR LOWER(user_files.file_path) LIKE ?)", like, like)
	}

	return query
}

func (s *videoService) videoFilePredicate() (string, []interface{}) {
	fallbackExtensions := []string{"mp4", "mov", "mkv", "avi", "webm", "m4v", "wmv", "flv", "mpeg", "mpg", "3gp", "3g2", "ts", "mts", "m2ts", "rm", "rmvb", "vob", "ogv", "asf", "divx"}
	extensions := categoryExtensionsFromQuery(configuredCategoryQuery(s.settingsRepo, "video"), fallbackExtensions)
	return categoryPredicate([]string{"video/%"}, extensions)
}

func videoOrderClause(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "name":
		return "user_files.file_name ASC, user_files.id DESC"
	case "size":
		return "user_files.file_size DESC, user_files.updated_at DESC, user_files.id DESC"
	default:
		return "user_files.updated_at DESC, user_files.id DESC"
	}
}

func buildVideoItem(file *model.UserFile) *VideoItem {
	contentType := "application/octet-stream"
	hash := ""
	size := file.FileSize

	if file.PhysicalFile != nil {
		if strings.TrimSpace(file.PhysicalFile.ContentType) != "" {
			contentType = file.PhysicalFile.ContentType
		}
		if strings.TrimSpace(file.PhysicalFile.FileHash) != "" {
			hash = file.PhysicalFile.FileHash
		}
		if size <= 0 && file.PhysicalFile.FileSize > 0 {
			size = file.PhysicalFile.FileSize
		}
	}

	item := &VideoItem{
		ID:             file.ID,
		Name:           file.FileName,
		Size:           size,
		Hash:           hash,
		MimeType:       contentType,
		ContentType:    contentType,
		PhysicalFileID: file.PhysicalFileID,
		FolderID:       file.ParentID,
		CreatedAt:      file.CreatedAt,
		UpdatedAt:      file.UpdatedAt,
	}

	if file.PhysicalFileID != nil {
		item.ThumbnailURL = fmt.Sprintf("/api/v1/file/%d/thumbnail", file.ID)
	}

	return item
}
