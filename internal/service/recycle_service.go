package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/storage"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RecycleService recycle bin operations.
type RecycleService interface {
	MoveToRecycle(ctx context.Context, userID uint, fileIDs []uint) error
	GetRecycleList(ctx context.Context, userID uint, query RecycleListQuery) (*RecycleListResponse, error)
	RestoreFiles(ctx context.Context, userID uint, itemIDs []uint) error
	PermanentDelete(ctx context.Context, userID uint, itemIDs []uint) error
	EmptyRecycleBin(ctx context.Context, userID uint) error
}

type recycleService struct {
	recycleRepo   repository.RecycleRepository
	fileRepo      repository.UserFileRepository
	storage       storage.Storage
	db            *gorm.DB
	queueDispatch QueueDispatchService
}

func NewRecycleService(
	recycleRepo repository.RecycleRepository,
	fileRepo repository.UserFileRepository,
	storage storage.Storage,
	db *gorm.DB,
	queueDispatch QueueDispatchService,
) RecycleService {
	return &recycleService{
		recycleRepo:   recycleRepo,
		fileRepo:      fileRepo,
		storage:       storage,
		db:            db,
		queueDispatch: queueDispatch,
	}
}

type RecycleListResponse struct {
	Items          []*RecycleItem `json:"items"`
	Total          int64          `json:"total"`
	Page           int            `json:"page"`
	PageSize       int            `json:"page_size"`
	PaginationMode string         `json:"pagination_mode"`
	NextCursor     string         `json:"next_cursor"`
	MaxPageSize    int            `json:"max_page_size"`
	Stats          RecycleStats   `json:"stats"`
	Sort           string         `json:"sort"`
	Keyword        string         `json:"keyword,omitempty"`
	FileType       string         `json:"file_type,omitempty"`
	TotalSize      int64          `json:"total_size"`
}

type RecycleItem struct {
	ID           uint      `json:"id"`
	FileID       uint      `json:"file_id"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	FileType     string    `json:"file_type"`
	OriginalPath string    `json:"original_path"`
	DeletedAt    time.Time `json:"deleted_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type RecycleListQuery struct {
	Page           int
	PageSize       int
	Cursor         uint
	UseCursor      bool
	PaginationMode string
	MaxPageSize    int
	Sort           string
	Keyword        string
	FileType       string
}

type RecycleStats struct {
	TotalSize    int64          `json:"total_size"`
	ExpiringSoon int64          `json:"expiring_soon"`
	Expired      int64          `json:"expired"`
	CountByType  map[string]int `json:"count_by_type"`
}

func (s *recycleService) MoveToRecycle(ctx context.Context, userID uint, fileIDs []uint) error {
	fileIDs = uniqueUintIDs(fileIDs)
	if len(fileIDs) == 0 {
		return fmt.Errorf("file list cannot be empty")
	}

	items := make([]*model.RecycleBin, 0, len(fileIDs))
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		softDeleteIDs := make([]uint, 0, len(fileIDs))
		for _, fileID := range fileIDs {
			file, err := s.getActiveUserFile(tx, fileID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					deletedFile, deletedErr := s.getUserFileUnscoped(tx, fileID)
					if deletedErr == nil && deletedFile.UserID == userID && deletedFile.DeletedAt.Valid {
						continue
					}
				}
				return fmt.Errorf("file not found: %d", fileID)
			}
			if file.UserID != userID {
				return fmt.Errorf("no permission to recycle file: %d", fileID)
			}

			originalPath, err := s.buildFilePathWithDB(ctx, tx, file)
			if err != nil {
				return fmt.Errorf("build original path failed: %w", err)
			}

			item := &model.RecycleBin{
				UserID:        userID,
				FileID:        fileID,
				FileName:      file.FileName,
				FileSize:      file.FileSize,
				FileType:      s.getFileType(file),
				OriginalPath:  originalPath,
				FileDeletedAt: now,
				ExpiresAt:     now.Add(30 * 24 * time.Hour),
			}
			items = append(items, item)

			subtreeIDs, err := s.collectSubtreeIDs(ctx, tx, userID, fileID, false)
			if err != nil {
				return err
			}
			softDeleteIDs = append(softDeleteIDs, fileID)
			softDeleteIDs = append(softDeleteIDs, subtreeIDs...)
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("create recycle records failed: %w", err)
			}
		}

		if err := tx.Model(&model.UserFile{}).
			Where("id IN ?", uniqueUintIDs(softDeleteIDs)).
			Update("deleted_at", now).Error; err != nil {
			return fmt.Errorf("soft delete file failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.queueDispatch != nil {
		for _, item := range items {
			if err := s.queueDispatch.EnqueueRecycleCleanup(item.ID, item.ExpiresAt); err != nil {
				logger.Warn("enqueue recycle cleanup job failed", zap.Uint("recycle_id", item.ID), zap.Error(err))
			}
		}
	}

	return nil
}

func (s *recycleService) GetRecycleList(ctx context.Context, userID uint, query RecycleListQuery) (*RecycleListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	query.Sort = normalizeRecycleSort(query.Sort)
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.FileType = strings.TrimSpace(strings.ToLower(query.FileType))

	items, total, stats, err := s.queryRecycleItems(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	recycleItems := make([]*RecycleItem, len(items))
	for i, item := range items {
		recycleItems[i] = &RecycleItem{
			ID:           item.ID,
			FileID:       item.FileID,
			FileName:     item.FileName,
			FileSize:     item.FileSize,
			FileType:     item.FileType,
			OriginalPath: item.OriginalPath,
			DeletedAt:    item.FileDeletedAt,
			ExpiresAt:    item.ExpiresAt,
		}
	}
	nextCursor := ""
	if query.UseCursor && len(items) > 0 {
		nextCursor = NextCursorFromUint(items[len(items)-1].ID)
	}

	return &RecycleListResponse{
		Items:          recycleItems,
		Total:          total,
		Page:           query.Page,
		PageSize:       query.PageSize,
		PaginationMode: query.PaginationMode,
		NextCursor:     nextCursor,
		MaxPageSize:    query.MaxPageSize,
		Stats:          stats,
		Sort:           query.Sort,
		Keyword:        query.Keyword,
		FileType:       query.FileType,
		TotalSize:      stats.TotalSize,
	}, nil
}

func (s *recycleService) RestoreFiles(ctx context.Context, userID uint, itemIDs []uint) error {
	itemIDs = uniqueUintIDs(itemIDs)
	if len(itemIDs) == 0 {
		return fmt.Errorf("recycle item list cannot be empty")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items, err := s.getRecycleItemsForUser(ctx, tx, userID, itemIDs)
		if err != nil {
			return err
		}
		if len(items) != len(itemIDs) {
			return fmt.Errorf("some recycle items do not exist")
		}

		for _, item := range items {
			file, err := s.getUserFileUnscoped(tx, item.FileID)
			if err != nil {
				return fmt.Errorf("file not found: %d", item.FileID)
			}
			if file.UserID != userID {
				return fmt.Errorf("no permission to restore file: %d", item.FileID)
			}

			parentID, err := s.resolveParentID(ctx, tx, userID, item.OriginalPath)
			if err != nil {
				parentID = nil
			}

			newFileName := s.resolveFileNameConflict(ctx, tx, userID, parentID, item.FileName)
			if err := tx.Unscoped().Model(&model.UserFile{}).
				Where("id = ?", item.FileID).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"parent_id":  parentID,
					"file_name":  newFileName,
				}).Error; err != nil {
				return fmt.Errorf("restore file failed: %w", err)
			}

			subtreeIDs, err := s.collectSubtreeIDsUnscoped(ctx, tx, userID, item.FileID, false)
			if err != nil {
				return err
			}
			if len(subtreeIDs) > 0 {
				if err := tx.Unscoped().
					Model(&model.UserFile{}).
					Where("id IN ?", subtreeIDs).
					Update("deleted_at", nil).Error; err != nil {
					return fmt.Errorf("restore child files failed: %w", err)
				}
			}
		}

		if err := tx.Where("id IN ?", itemIDs).Delete(&model.RecycleBin{}).Error; err != nil {
			return fmt.Errorf("delete recycle records failed: %w", err)
		}

		return nil
	})
}

func (s *recycleService) PermanentDelete(ctx context.Context, userID uint, itemIDs []uint) error {
	itemIDs = uniqueUintIDs(itemIDs)
	if len(itemIDs) == 0 {
		return fmt.Errorf("recycle item list cannot be empty")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items, err := s.getRecycleItemsForUser(ctx, tx, userID, itemIDs)
		if err != nil {
			return err
		}
		if len(items) != len(itemIDs) {
			return fmt.Errorf("some recycle items do not exist")
		}

		userFileIDs := make([]uint, 0, len(items))
		physicalRefCounts := make(map[uint]int)
		totalSize := int64(0)

		for _, item := range items {
			files, err := s.collectSubtreeFilesUnscoped(ctx, tx, userID, item.FileID)
			if err != nil {
				return err
			}
			for _, file := range files {
				userFileIDs = append(userFileIDs, file.ID)
				if file.IsFolder {
					continue
				}
				if file.FileSize > 0 {
					totalSize += file.FileSize
				}
				if file.PhysicalFileID != nil {
					physicalRefCounts[*file.PhysicalFileID]++
				}
			}
		}

		userFileIDs = uniqueUintIDs(userFileIDs)
		for physicalID, removedRefs := range physicalRefCounts {
			var physicalFile model.PhysicalFile
			if err := tx.Unscoped().First(&physicalFile, physicalID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				return fmt.Errorf("query physical file failed: %w", err)
			}

			if physicalFile.RefCount <= removedRefs {
				if physicalFile.StoragePath != "" && s.storage != nil && s.storage.Exists(physicalFile.StoragePath) {
					if err := s.storage.Delete(physicalFile.StoragePath); err != nil {
						logger.Warn("delete recycled physical file failed",
							zap.Uint("physical_file_id", physicalID),
							zap.String("path", physicalFile.StoragePath),
							zap.Error(err))
					}
				}
				if err := tx.Unscoped().Delete(&model.PhysicalFile{}, physicalID).Error; err != nil {
					return fmt.Errorf("delete physical file record failed: %w", err)
				}
				continue
			}

			if err := tx.Model(&model.PhysicalFile{}).
				Where("id = ?", physicalID).
				UpdateColumn("ref_count", gorm.Expr("ref_count - ?", removedRefs)).Error; err != nil {
				return fmt.Errorf("decrement physical file ref count failed: %w", err)
			}
		}

		if len(userFileIDs) > 0 {
			if err := tx.Unscoped().Delete(&model.UserFile{}, userFileIDs).Error; err != nil {
				return fmt.Errorf("delete file records failed: %w", err)
			}
		}

		if totalSize > 0 {
			if err := tx.Model(&model.User{}).
				Where("id = ?", userID).
				UpdateColumn("used_size", gorm.Expr("CASE WHEN used_size >= ? THEN used_size - ? ELSE 0 END", totalSize, totalSize)).Error; err != nil {
				return fmt.Errorf("update user storage usage failed: %w", err)
			}
		}

		if err := tx.Where("id IN ?", itemIDs).Delete(&model.RecycleBin{}).Error; err != nil {
			return fmt.Errorf("delete recycle records failed: %w", err)
		}

		return nil
	})
}

func (s *recycleService) EmptyRecycleBin(ctx context.Context, userID uint) error {
	for {
		items, _, _, err := s.queryRecycleItems(ctx, userID, RecycleListQuery{
			Page:     1,
			PageSize: 500,
			Sort:     "deleted-asc",
		})
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}

		itemIDs := make([]uint, len(items))
		for i, item := range items {
			itemIDs[i] = item.ID
		}
		if err := s.PermanentDelete(ctx, userID, itemIDs); err != nil {
			return err
		}
	}
}

func (s *recycleService) queryRecycleItems(ctx context.Context, userID uint, query RecycleListQuery) ([]*model.RecycleBin, int64, RecycleStats, error) {
	var items []*model.RecycleBin
	var total int64
	stats := RecycleStats{CountByType: map[string]int{}}

	if err := s.recycleListBaseQuery(ctx, userID, query).Count(&total).Error; err != nil {
		return nil, 0, stats, fmt.Errorf("query recycle total failed: %w", err)
	}

	var aggregates []struct {
		FileType string
		Count    int
		Size     int64
	}
	if err := s.recycleListBaseQuery(ctx, userID, query).
		Select("file_type, COUNT(*) AS count, COALESCE(SUM(file_size), 0) AS size").
		Group("file_type").
		Scan(&aggregates).Error; err != nil {
		return nil, 0, stats, fmt.Errorf("query recycle stats failed: %w", err)
	}
	for _, row := range aggregates {
		fileType := row.FileType
		if fileType == "" {
			fileType = "other"
		}
		stats.CountByType[fileType] = row.Count
		stats.TotalSize += row.Size
	}

	now := time.Now()
	if err := s.recycleListBaseQuery(ctx, userID, query).
		Where("expires_at <= ?", now.Add(3*24*time.Hour)).
		Count(&stats.ExpiringSoon).Error; err != nil {
		return nil, 0, stats, fmt.Errorf("query expiring recycle count failed: %w", err)
	}
	if err := s.recycleListBaseQuery(ctx, userID, query).
		Where("expires_at <= ?", now).
		Count(&stats.Expired).Error; err != nil {
		return nil, 0, stats, fmt.Errorf("query expired recycle count failed: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	dataQuery := s.recycleListBaseQuery(ctx, userID, query)
	if query.UseCursor && query.Cursor > 0 {
		dataQuery = dataQuery.Where("id < ?", query.Cursor)
	}
	if err := dataQuery.
		Order(recycleOrderClause(query.Sort)).
		Offset(recycleListOffset(query.UseCursor, offset)).
		Limit(query.PageSize).
		Find(&items).Error; err != nil {
		return nil, 0, stats, fmt.Errorf("query recycle list failed: %w", err)
	}

	return items, total, stats, nil
}

func recycleListOffset(useCursor bool, offset int) int {
	if useCursor {
		return 0
	}
	return offset
}

func (s *recycleService) recycleListBaseQuery(ctx context.Context, userID uint, query RecycleListQuery) *gorm.DB {
	return applyRecycleFilters(
		s.db.WithContext(ctx).Model(&model.RecycleBin{}).Where("user_id = ?", userID),
		query,
	)
}

func applyRecycleFilters(db *gorm.DB, query RecycleListQuery) *gorm.DB {
	if query.Keyword != "" {
		keyword := "%" + strings.ReplaceAll(query.Keyword, "%", "\\%") + "%"
		db = db.Where("(file_name LIKE ? OR original_path LIKE ?)", keyword, keyword)
	}
	if query.FileType != "" && query.FileType != "all" {
		db = db.Where("file_type = ?", query.FileType)
	}
	return db
}

func normalizeRecycleSort(sort string) string {
	switch sort {
	case "deleted-asc", "name-asc", "name-desc", "size-asc", "size-desc", "expires-asc", "expires-desc":
		return sort
	default:
		return "deleted-desc"
	}
}

func recycleOrderClause(sort string) string {
	switch sort {
	case "deleted-asc":
		return "deleted_at ASC, id ASC"
	case "name-asc":
		return "file_name ASC, id DESC"
	case "name-desc":
		return "file_name DESC, id DESC"
	case "size-asc":
		return "file_size ASC, id DESC"
	case "size-desc":
		return "file_size DESC, id DESC"
	case "expires-asc":
		return "expires_at ASC, id DESC"
	case "expires-desc":
		return "expires_at DESC, id DESC"
	default:
		return "deleted_at DESC, id DESC"
	}
}

func uniqueUintIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (s *recycleService) getActiveUserFile(tx *gorm.DB, fileID uint) (*model.UserFile, error) {
	var file model.UserFile
	if err := tx.First(&file, fileID).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *recycleService) getUserFileUnscoped(tx *gorm.DB, fileID uint) (*model.UserFile, error) {
	var file model.UserFile
	if err := tx.Unscoped().First(&file, fileID).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *recycleService) getRecycleItemsForUser(ctx context.Context, tx *gorm.DB, userID uint, itemIDs []uint) ([]*model.RecycleBin, error) {
	var items []*model.RecycleBin
	if err := tx.WithContext(ctx).
		Where("user_id = ? AND id IN ?", userID, itemIDs).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("query recycle items failed: %w", err)
	}
	return items, nil
}

func (s *recycleService) collectSubtreeIDs(ctx context.Context, tx *gorm.DB, userID uint, rootID uint, includeRoot bool) ([]uint, error) {
	return s.collectSubtreeIDsWithScope(ctx, tx, userID, rootID, includeRoot, false)
}

func (s *recycleService) collectSubtreeIDsUnscoped(ctx context.Context, tx *gorm.DB, userID uint, rootID uint, includeRoot bool) ([]uint, error) {
	return s.collectSubtreeIDsWithScope(ctx, tx, userID, rootID, includeRoot, true)
}

func (s *recycleService) collectSubtreeIDsWithScope(ctx context.Context, tx *gorm.DB, userID uint, rootID uint, includeRoot bool, unscoped bool) ([]uint, error) {
	result := make([]uint, 0)
	if includeRoot {
		result = append(result, rootID)
	}

	queue := []uint{rootID}
	visited := map[uint]struct{}{}
	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		if _, ok := visited[parentID]; ok {
			continue
		}
		visited[parentID] = struct{}{}

		var children []model.UserFile
		query := tx.WithContext(ctx)
		if unscoped {
			query = query.Unscoped()
		}
		if err := query.
			Where("user_id = ? AND parent_id = ?", userID, parentID).
			Find(&children).Error; err != nil {
			return nil, fmt.Errorf("query child files failed: %w", err)
		}

		for _, child := range children {
			result = append(result, child.ID)
			if child.IsFolder {
				queue = append(queue, child.ID)
			}
		}
	}

	return uniqueUintIDs(result), nil
}

func (s *recycleService) collectSubtreeFilesUnscoped(ctx context.Context, tx *gorm.DB, userID uint, rootID uint) ([]model.UserFile, error) {
	ids, err := s.collectSubtreeIDsUnscoped(ctx, tx, userID, rootID, true)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	var files []model.UserFile
	if err := tx.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND id IN ?", userID, ids).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("query subtree files failed: %w", err)
	}

	return files, nil
}

func (s *recycleService) buildFilePath(ctx context.Context, file *model.UserFile) (string, error) {
	return s.buildFilePathWithDB(ctx, s.db, file)
}

func (s *recycleService) buildFilePathWithDB(ctx context.Context, db *gorm.DB, file *model.UserFile) (string, error) {
	if file.ParentID == nil {
		return "/" + file.FileName, nil
	}

	pathParts := []string{file.FileName}
	currentParentID := file.ParentID
	for currentParentID != nil {
		var parent model.UserFile
		if err := db.WithContext(ctx).First(&parent, *currentParentID).Error; err != nil {
			break
		}
		pathParts = append([]string{parent.FileName}, pathParts...)
		currentParentID = parent.ParentID
	}

	return "/" + strings.Join(pathParts, "/"), nil
}

func (s *recycleService) resolveParentID(ctx context.Context, db *gorm.DB, userID uint, originalPath string) (*uint, error) {
	if originalPath == "" || originalPath == "/" {
		return nil, nil
	}

	pathParts := strings.Split(strings.Trim(originalPath, "/"), "/")
	if len(pathParts) <= 1 {
		return nil, nil
	}
	pathParts = pathParts[:len(pathParts)-1]

	var parentID *uint
	for _, part := range pathParts {
		var folder model.UserFile
		query := db.Where("user_id = ? AND file_name = ? AND is_folder = ?", userID, part, true)
		if parentID != nil {
			query = query.Where("parent_id = ?", *parentID)
		} else {
			query = query.Where("parent_id IS NULL")
		}
		if err := query.First(&folder).Error; err != nil {
			return nil, fmt.Errorf("parent folder not found: %s", part)
		}
		parentID = &folder.ID
	}

	return parentID, nil
}

func (s *recycleService) resolveFileNameConflict(ctx context.Context, tx *gorm.DB, userID uint, parentID *uint, fileName string) string {
	var count int64
	query := tx.Model(&model.UserFile{}).
		Where("user_id = ? AND file_name = ? AND deleted_at IS NULL", userID, fileName)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	query.Count(&count)
	if count == 0 {
		return fileName
	}

	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)
	for i := 1; i <= 100; i++ {
		newFileName := fmt.Sprintf("%s(%d)%s", nameWithoutExt, i, ext)
		query := tx.Model(&model.UserFile{}).
			Where("user_id = ? AND file_name = ? AND deleted_at IS NULL", userID, newFileName)
		if parentID != nil {
			query = query.Where("parent_id = ?", *parentID)
		} else {
			query = query.Where("parent_id IS NULL")
		}
		query.Count(&count)
		if count == 0 {
			return newFileName
		}
	}

	return fmt.Sprintf("%s_%d%s", nameWithoutExt, time.Now().Unix(), ext)
}

func (s *recycleService) getFileType(file *model.UserFile) string {
	if file.IsFolder {
		return "folder"
	}

	ext := strings.ToLower(filepath.Ext(file.FileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "word"
	case ".xls", ".xlsx":
		return "excel"
	case ".ppt", ".pptx":
		return "powerpoint"
	case ".txt", ".md":
		return "text"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	default:
		return "other"
	}
}
