package service

import (
	"context"
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"

	"gorm.io/gorm"
)

// FolderService handles folder operations.
type FolderService interface {
	Create(userID uint, name string, parentID *uint) (*model.UserFile, error)
	Rename(userID, folderID uint, newName string) error
	Delete(userID, folderID uint) error
	Move(userID, folderID uint, newParentID *uint) error
	Copy(userID, folderID uint, targetParentID *uint) (*model.UserFile, error)
	Path(userID, folderID uint) ([]FolderPathItem, error)
}

type FolderPathItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type folderService struct {
	db           *gorm.DB
	userFileRepo repository.UserFileRepository
	cache        *cache.CacheService
	fileEvents   FileEventService
}

func NewFolderService(db *gorm.DB, userFileRepo repository.UserFileRepository, cache *cache.CacheService, fileEvents ...FileEventService) FolderService {
	var eventService FileEventService
	if len(fileEvents) > 0 {
		eventService = fileEvents[0]
	}
	return &folderService{
		db:           db,
		userFileRepo: userFileRepo,
		cache:        cache,
		fileEvents:   eventService,
	}
}

func (s *folderService) Create(userID uint, name string, parentID *uint) (*model.UserFile, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("folder name cannot be empty")
	}
	if parentID != nil {
		if err := s.validateFolderOwnership(*parentID, userID); err != nil {
			return nil, err
		}
	}

	folder := &model.UserFile{
		UserID:   userID,
		FileName: name,
		IsFolder: true,
		ParentID: parentID,
	}

	if err := s.userFileRepo.Create(folder); err != nil {
		return nil, err
	}

	s.invalidateFolderCaches(userID, folder.ID, parentID)
	s.publishFolderEvent(userID, parentID, "folder.created", folder.ID)
	return folder, nil
}

func (s *folderService) Rename(userID, folderID uint, newName string) error {
	folder, err := s.userFileRepo.GetByID(folderID)
	if err != nil {
		return err
	}

	if folder.UserID != userID {
		return fmt.Errorf("no permission to rename this folder")
	}
	if !folder.IsFolder {
		return fmt.Errorf("target is not a folder")
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("folder name cannot be empty")
	}

	folder.FileName = newName
	if err := s.userFileRepo.Update(folder); err != nil {
		return err
	}

	s.invalidateFolderCaches(userID, folderID, folder.ParentID)
	s.publishFolderEvent(userID, folder.ParentID, "folder.renamed", folderID)
	return nil
}

func (s *folderService) Delete(userID, folderID uint) error {
	folder, err := s.userFileRepo.GetByID(folderID)
	if err != nil {
		return err
	}

	if folder.UserID != userID {
		return fmt.Errorf("no permission to delete this folder")
	}
	if !folder.IsFolder {
		return fmt.Errorf("target is not a folder")
	}

	parentID := folder.ParentID
	descendants, err := s.userFileRepo.ListDescendants(userID, folderID)
	if err != nil {
		return err
	}

	allIDs := make([]uint, 0, len(descendants)+1)
	allIDs = append(allIDs, folderID)
	totalSize := int64(0)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range descendants {
			allIDs = append(allIDs, item.ID)
			if !item.IsFolder && item.FileSize > 0 {
				totalSize += item.FileSize
			}
			if !item.IsFolder && item.PhysicalFileID != nil {
				if err := tx.Model(&model.PhysicalFile{}).
					Where("id = ?", *item.PhysicalFileID).
					UpdateColumn("ref_count", gorm.Expr("CASE WHEN ref_count > 0 THEN ref_count - 1 ELSE 0 END")).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Delete(&model.UserFile{}, allIDs).Error; err != nil {
			return err
		}
		if totalSize > 0 {
			if err := tx.Model(&model.User{}).
				Where("id = ?", userID).
				UpdateColumn("used_size", gorm.Expr("CASE WHEN used_size >= ? THEN used_size - ? ELSE 0 END", totalSize, totalSize)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.invalidateFolderCaches(userID, folderID, parentID)
	s.publishFolderEvent(userID, parentID, "folder.deleted", folderID)
	return nil
}

func (s *folderService) Move(userID, folderID uint, newParentID *uint) error {
	folder, err := s.userFileRepo.GetByID(folderID)
	if err != nil {
		return err
	}

	if folder.UserID != userID {
		return fmt.Errorf("no permission to move this folder")
	}
	if !folder.IsFolder {
		return fmt.Errorf("target is not a folder")
	}

	if newParentID != nil {
		if err := s.validateFolderOwnership(*newParentID, userID); err != nil {
			return err
		}
		if *newParentID == folderID {
			return fmt.Errorf("涓嶈兘灏嗘枃浠跺す绉诲姩鍒拌嚜宸变笅闈?")
		}
		if err := s.checkNotDescendant(folderID, *newParentID); err != nil {
			return err
		}
	}

	oldParentID := folder.ParentID
	folder.ParentID = newParentID
	if err := s.userFileRepo.Update(folder); err != nil {
		return err
	}

	if s.cache != nil {
		ctx := context.Background()
		_ = s.cache.InvalidateFileList(ctx, userID, oldParentID)
		_ = s.cache.InvalidateFileList(ctx, userID, newParentID)
		_ = s.cache.InvalidateDirectoryStats(ctx, userID, folderID)
		if oldParentID != nil && newParentID != nil {
			_ = s.cache.InvalidateDirectoryStatsMany(ctx, userID, *oldParentID, *newParentID)
		} else if oldParentID != nil {
			_ = s.cache.InvalidateDirectoryStats(ctx, userID, *oldParentID)
		} else if newParentID != nil {
			_ = s.cache.InvalidateDirectoryStats(ctx, userID, *newParentID)
		}
	}

	s.publishFolderEvent(userID, oldParentID, "folder.moved", folderID)
	s.publishFolderEvent(userID, newParentID, "folder.moved", folderID)
	return nil
}

func (s *folderService) Copy(userID, folderID uint, targetParentID *uint) (*model.UserFile, error) {
	source, err := s.userFileRepo.GetByID(folderID)
	if err != nil {
		return nil, err
	}
	if source.UserID != userID {
		return nil, fmt.Errorf("no permission to copy this folder")
	}
	if !source.IsFolder {
		return nil, fmt.Errorf("target is not a folder")
	}
	if targetParentID != nil {
		if err := s.validateFolderOwnership(*targetParentID, userID); err != nil {
			return nil, err
		}
		if *targetParentID == folderID {
			return nil, fmt.Errorf("cannot copy folder into itself")
		}
		if err := s.checkNotDescendant(folderID, *targetParentID); err != nil {
			return nil, err
		}
	}

	descendants, err := s.userFileRepo.ListDescendants(userID, folderID)
	if err != nil {
		return nil, err
	}

	var rootCopy *model.UserFile
	parentMap := map[uint]uint{}
	totalCopiedSize := int64(0)
	for _, item := range descendants {
		if !item.IsFolder && item.FileSize > 0 {
			totalCopiedSize += item.FileSize
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := reserveUserCapacity(tx, userID, totalCopiedSize); err != nil {
			return err
		}

		rootCopy = &model.UserFile{
			UserID:   userID,
			ParentID: targetParentID,
			FileName: source.FileName,
			IsFolder: true,
		}
		if err := tx.Create(rootCopy).Error; err != nil {
			return err
		}
		parentMap[source.ID] = rootCopy.ID

		for _, item := range descendants {
			newParentID, ok := parentMap[*item.ParentID]
			if !ok {
				return fmt.Errorf("copy parent not found")
			}
			parentID := newParentID
			copied := &model.UserFile{
				UserID:         userID,
				ParentID:       &parentID,
				FileName:       item.FileName,
				PhysicalFileID: item.PhysicalFileID,
				IsFolder:       item.IsFolder,
				FileSize:       item.FileSize,
				FilePath:       item.FilePath,
			}
			if err := tx.Create(copied).Error; err != nil {
				return err
			}
			if item.IsFolder {
				parentMap[item.ID] = copied.ID
				continue
			}
			if item.PhysicalFileID != nil {
				if err := tx.Model(&model.PhysicalFile{}).
					Where("id = ?", *item.PhysicalFileID).
					UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.invalidateFolderCaches(userID, rootCopy.ID, targetParentID)
	s.publishFolderEvent(userID, targetParentID, "folder.copied", rootCopy.ID)
	return rootCopy, nil
}

func (s *folderService) Path(userID, folderID uint) ([]FolderPathItem, error) {
	folders, err := s.userFileRepo.GetFolderPath(userID, folderID)
	if err != nil {
		return nil, err
	}

	path := make([]FolderPathItem, 0, len(folders))
	for _, folder := range folders {
		path = append(path, FolderPathItem{
			ID:   folder.ID,
			Name: folder.FileName,
		})
	}
	return path, nil
}

func (s *folderService) invalidateFolderCaches(userID, folderID uint, parentID *uint) {
	if s.cache == nil {
		return
	}

	ctx := context.Background()
	_ = s.cache.InvalidateFileList(ctx, userID, parentID)
	_ = s.cache.InvalidateDirectoryStats(ctx, userID, folderID)
	if parentID != nil {
		_ = s.cache.InvalidateDirectoryStats(ctx, userID, *parentID)
	}
}

func (s *folderService) publishFolderEvent(userID uint, folderID *uint, eventType string, resourceID uint) {
	if s.fileEvents == nil {
		return
	}
	s.fileEvents.Publish(userID, folderID, eventType, "folder", resourceID)
}

func (s *folderService) validateFolderOwnership(folderID uint, userID uint) error {
	folder, err := s.userFileRepo.GetByID(folderID)
	if err != nil {
		return fmt.Errorf("folder not found")
	}

	if folder.UserID != userID {
		return fmt.Errorf("no permission to access this folder")
	}

	if !folder.IsFolder {
		return fmt.Errorf("target is not a folder")
	}

	return nil
}

func (s *folderService) checkNotDescendant(ancestorID, descendantID uint) error {
	current := descendantID
	visited := make(map[uint]bool)

	for current != 0 {
		if visited[current] {
			return fmt.Errorf("folder cycle detected")
		}
		visited[current] = true

		if current == ancestorID {
			return fmt.Errorf("cannot move to child folder")
		}

		folder, err := s.userFileRepo.GetByID(current)
		if err != nil || folder.ParentID == nil {
			break
		}
		current = *folder.ParentID
	}

	return nil
}
