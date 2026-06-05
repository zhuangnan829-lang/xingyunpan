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

// MusicService provides the music library data used by the drive music page.
type MusicService interface {
	ListMusic(ctx context.Context, userID uint, query *MusicListQuery) (*MusicListResult, error)
}

// MusicListQuery describes music library filters and pagination.
type MusicListQuery struct {
	Keyword  string `json:"keyword"`
	Sort     string `json:"sort"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Cursor   string `json:"cursor"`
}

// MusicListResult is the paginated music library response.
type MusicListResult struct {
	Files          []*MusicItem `json:"files"`
	Total          int64        `json:"total"`
	TotalSize      int64        `json:"total_size"`
	Page           int          `json:"page"`
	PageSize       int          `json:"page_size"`
	TotalPages     int          `json:"total_pages"`
	PaginationMode string       `json:"pagination_mode"`
	NextCursor     string       `json:"next_cursor"`
	MaxPageSize    int          `json:"max_page_size"`
}

// MusicItem mirrors the frontend file item shape for audio files.
type MusicItem struct {
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

type musicService struct {
	db           *gorm.DB
	settingsRepo repository.FileSystemSettingRepository
}

// NewMusicService creates a music library service.
func NewMusicService(db *gorm.DB, settingsRepo ...repository.FileSystemSettingRepository) MusicService {
	var repo repository.FileSystemSettingRepository
	if len(settingsRepo) > 0 {
		repo = settingsRepo[0]
	}
	return &musicService{db: db, settingsRepo: repo}
}

func (s *musicService) ListMusic(ctx context.Context, userID uint, query *MusicListQuery) (*MusicListResult, error) {
	if query == nil {
		query = &MusicListQuery{}
	}

	pagination := ResolvePagination(s.settingsRepo, query.Page, query.PageSize, query.Cursor, 50)
	page := pagination.Page
	pageSize := pagination.PageSize

	keyword := strings.TrimSpace(query.Keyword)

	var total int64
	if err := s.musicBaseQuery(ctx, userID, keyword).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("query music total failed: %w", err)
	}

	var totalSize int64
	if err := s.musicBaseQuery(ctx, userID, keyword).
		Select("COALESCE(SUM(CASE WHEN user_files.file_size > 0 THEN user_files.file_size ELSE COALESCE(physical_files.file_size, 0) END), 0)").
		Scan(&totalSize).Error; err != nil {
		return nil, fmt.Errorf("query music total size failed: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	offset := mediaListOffset(pagination.UseCursor, page, pageSize)
	if total == 0 || offset >= int(total) {
		return &MusicListResult{
			Files:          []*MusicItem{},
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
	dataQuery := s.musicBaseQuery(ctx, userID, keyword)
	if pagination.UseCursor && pagination.Cursor > 0 {
		dataQuery = dataQuery.Where("user_files.id < ?", pagination.Cursor)
	}
	if err := dataQuery.
		Preload("PhysicalFile").
		Order(musicOrderClause(query.Sort)).
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("query music files failed: %w", err)
	}

	items := make([]*MusicItem, 0, len(files))
	for _, file := range files {
		items = append(items, buildMusicItem(file))
	}
	nextCursor := ""
	if pagination.UseCursor && len(files) > 0 {
		nextCursor = NextCursorFromUint(files[len(files)-1].ID)
	}

	return &MusicListResult{
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

func (s *musicService) musicBaseQuery(ctx context.Context, userID uint, keyword string) *gorm.DB {
	query := s.db.WithContext(ctx).
		Model(&model.UserFile{}).
		Joins("LEFT JOIN physical_files ON physical_files.id = user_files.physical_file_id").
		Where("user_files.user_id = ?", userID).
		Where("user_files.deleted_at IS NULL").
		Where("user_files.is_folder = ?", false)

	clause, args := s.musicFilePredicate()
	query = query.Where(clause, args...)

	if keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("(LOWER(user_files.file_name) LIKE ? OR LOWER(user_files.file_path) LIKE ?)", like, like)
	}

	return query
}

func (s *musicService) musicFilePredicate() (string, []interface{}) {
	fallbackExtensions := []string{
		"mp3", "wav", "flac", "aac", "ogg", "oga", "m4a", "wma", "ape", "opus",
		"amr", "mid", "midi", "aiff", "aif", "alac", "mka", "mp2", "mpa", "weba",
	}
	extensions := categoryExtensionsFromQuery(configuredCategoryQuery(s.settingsRepo, "audio"), fallbackExtensions)
	return categoryPredicate([]string{"audio/%"}, extensions)
}

func musicOrderClause(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "name":
		return "user_files.file_name ASC, user_files.id DESC"
	case "size":
		return "user_files.file_size DESC, user_files.updated_at DESC, user_files.id DESC"
	default:
		return "user_files.updated_at DESC, user_files.id DESC"
	}
}

func buildMusicItem(file *model.UserFile) *MusicItem {
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

	return &MusicItem{
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
}
