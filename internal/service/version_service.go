// 路径: internal/service/version_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"

	"gorm.io/gorm"
)

// VersionService 版本管理服务接口
type VersionService interface {
	// GetVersionHistory 获取版本历史
	GetVersionHistory(ctx context.Context, userID uint, fileID uint) ([]*VersionInfo, error)

	// DownloadVersion 下载版本
	DownloadVersion(ctx context.Context, userID uint, fileID uint, versionID uint) (string, error)

	// RestoreVersion 恢复版本
	RestoreVersion(ctx context.Context, userID uint, fileID uint, versionID uint) (*RestoreVersionResponse, error)

	// DeleteVersion 删除版本
	DeleteVersion(ctx context.Context, userID uint, fileID uint, versionID uint) error
}

// VersionInfo 版本信息
type VersionInfo struct {
	VersionID     uint      `json:"version_id"`
	VersionNumber int       `json:"version_number"`
	FileSize      int64     `json:"file_size"`
	CreatedAt     time.Time `json:"created_at"`
	UploaderID    uint      `json:"uploader_id"`
	UploaderName  string    `json:"uploader_name"`
	IsCurrent     bool      `json:"is_current"`
}

// RestoreVersionResponse 恢复版本响应
type RestoreVersionResponse struct {
	NewVersionID  uint `json:"new_version_id"`
	VersionNumber int  `json:"version_number"`
}

// versionService 版本管理服务实现
type versionService struct {
	versionRepo      repository.VersionRepository
	userFileRepo     repository.UserFileRepository
	physicalFileRepo repository.PhysicalFileRepository
	userRepo         repository.UserRepository
	cache            *cache.CacheService
	db               *gorm.DB
}

// NewVersionService 创建版本管理服务实例
func NewVersionService(
	versionRepo repository.VersionRepository,
	userFileRepo repository.UserFileRepository,
	physicalFileRepo repository.PhysicalFileRepository,
	userRepo repository.UserRepository,
	cache *cache.CacheService,
	db *gorm.DB,
) VersionService {
	return &versionService{
		versionRepo:      versionRepo,
		userFileRepo:     userFileRepo,
		physicalFileRepo: physicalFileRepo,
		userRepo:         userRepo,
		cache:            cache,
		db:               db,
	}
}

// GetVersionHistory 获取版本历史
func (s *versionService) GetVersionHistory(ctx context.Context, userID uint, fileID uint) ([]*VersionInfo, error) {
	// 1. 验证文件存在且用户有权限
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("查询文件失败: %w", err)
	}

	// 验证用户权限（所有者或协作者）
	if file.UserID != userID {
		// TODO: 检查协作者权限
		return nil, fmt.Errorf("无权限访问此文件")
	}

	// 2. 尝试从缓存获取
	cacheKey := fmt.Sprintf("versions:%d", fileID)
	var cachedVersions []*VersionInfo
	if s.cache != nil {
		if err := s.cache.Get(ctx, cacheKey, &cachedVersions); err == nil {
			return cachedVersions, nil
		}
	}

	// 3. 查询版本历史
	versions, err := s.versionRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("查询版本历史失败: %w", err)
	}

	// 4. 构建响应数据
	result := make([]*VersionInfo, 0, len(versions))
	for _, v := range versions {
		// 查询上传者信息
		uploader, err := s.userRepo.GetByID(v.UploaderID)
		if err != nil {
			return nil, fmt.Errorf("查询上传者信息失败: %w", err)
		}

		result = append(result, &VersionInfo{
			VersionID:     v.ID,
			VersionNumber: v.VersionNumber,
			FileSize:      v.FileSize,
			CreatedAt:     v.CreatedAt,
			UploaderID:    v.UploaderID,
			UploaderName:  uploader.Username,
			IsCurrent:     v.IsCurrent,
		})
	}

	// 5. 缓存结果（TTL=1分钟）
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, result, 1*time.Minute)
	}

	return result, nil
}

// DownloadVersion 下载版本
func (s *versionService) DownloadVersion(ctx context.Context, userID uint, fileID uint, versionID uint) (string, error) {
	// 1. 验证文件存在且用户有权限
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("文件不存在")
		}
		return "", fmt.Errorf("查询文件失败: %w", err)
	}

	// 验证用户权限（所有者或有 download 权限的协作者）
	if file.UserID != userID {
		// TODO: 检查协作者权限（需要有 download 或 edit 权限）
		return "", fmt.Errorf("无权限下载此文件")
	}

	// 2. 查询版本信息
	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return "", fmt.Errorf("查询版本失败: %w", err)
	}

	// 验证版本属于该文件
	if version.FileID != fileID {
		return "", fmt.Errorf("版本不属于该文件")
	}

	// 3. 查询物理文件信息
	physicalFile, err := s.physicalFileRepo.GetByID(version.PhysicalFileID)
	if err != nil {
		return "", fmt.Errorf("查询物理文件失败: %w", err)
	}

	// 4. 返回物理文件路径
	return physicalFile.StoragePath, nil
}

// RestoreVersion 恢复版本
func (s *versionService) RestoreVersion(ctx context.Context, userID uint, fileID uint, versionID uint) (*RestoreVersionResponse, error) {
	// 1. 验证文件存在且用户有权限
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("查询文件失败: %w", err)
	}

	// 验证用户权限（所有者或有 edit 权限的协作者）
	if file.UserID != userID {
		// TODO: 检查协作者权限（需要有 edit 权限）
		return nil, fmt.Errorf("无权限恢复此文件版本")
	}

	// 2. 查询要恢复的版本
	oldVersion, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("查询版本失败: %w", err)
	}

	// 验证版本属于该文件
	if oldVersion.FileID != fileID {
		return nil, fmt.Errorf("版本不属于该文件")
	}

	// 3. 使用事务执行恢复操作
	var newVersion *model.FileVersion
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1 获取当前最大版本号
		versions, err := s.versionRepo.GetByFileID(ctx, fileID)
		if err != nil {
			return fmt.Errorf("查询版本列表失败: %w", err)
		}

		maxVersionNumber := 0
		for _, v := range versions {
			if v.VersionNumber > maxVersionNumber {
				maxVersionNumber = v.VersionNumber
			}
		}

		// 3.2 创建新版本（复用旧版本的 physical_file_id）
		newVersion = &model.FileVersion{
			FileID:         fileID,
			VersionNumber:  maxVersionNumber + 1,
			PhysicalFileID: oldVersion.PhysicalFileID,
			FileSize:       oldVersion.FileSize,
			UploaderID:     userID,
			IsCurrent:      true,
		}

		if err := s.versionRepo.Create(ctx, newVersion); err != nil {
			return fmt.Errorf("创建新版本失败: %w", err)
		}

		// 3.3 将旧的当前版本标记为非当前版本
		if err := s.versionRepo.SetCurrentVersion(ctx, fileID, newVersion.ID); err != nil {
			return fmt.Errorf("更新当前版本失败: %w", err)
		}

		// 3.4 增加物理文件引用计数
		if err := s.physicalFileRepo.IncrementRefCount(oldVersion.PhysicalFileID); err != nil {
			return fmt.Errorf("更新物理文件引用计数失败: %w", err)
		}

		// 3.5 检查版本数量限制（最多 10 个）
		count, err := s.versionRepo.CountByFileID(ctx, fileID)
		if err != nil {
			return fmt.Errorf("查询版本数量失败: %w", err)
		}

		if count > 10 {
			// 删除最旧的版本
			oldestVersion, err := s.versionRepo.GetOldestVersion(ctx, fileID)
			if err != nil {
				return fmt.Errorf("查询最旧版本失败: %w", err)
			}

			// 检查物理文件引用计数
			physicalFile, err := s.physicalFileRepo.GetByID(oldestVersion.PhysicalFileID)
			if err != nil {
				return fmt.Errorf("查询物理文件失败: %w", err)
			}

			// 删除版本记录
			if err := s.versionRepo.Delete(ctx, oldestVersion.ID); err != nil {
				return fmt.Errorf("删除最旧版本失败: %w", err)
			}

			// 减少物理文件引用计数
			if err := s.physicalFileRepo.UpdateRefCount(oldestVersion.PhysicalFileID, -1); err != nil {
				return fmt.Errorf("更新物理文件引用计数失败: %w", err)
			}

			// 如果没有其他引用，删除物理文件
			if physicalFile.RefCount <= 1 {
				// TODO: 删除物理文件（调用 storage.Delete）
				if err := s.physicalFileRepo.Delete(oldestVersion.PhysicalFileID); err != nil {
					return fmt.Errorf("删除物理文件失败: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 4. 清除缓存
	cacheKey := fmt.Sprintf("versions:%d", fileID)
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, nil, 0)
	}

	return &RestoreVersionResponse{
		NewVersionID:  newVersion.ID,
		VersionNumber: newVersion.VersionNumber,
	}, nil
}

// DeleteVersion 删除版本
func (s *versionService) DeleteVersion(ctx context.Context, userID uint, fileID uint, versionID uint) error {
	// 1. 验证文件存在且用户有权限（只有所有者可以删除版本）
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("文件不存在")
		}
		return fmt.Errorf("查询文件失败: %w", err)
	}

	// 验证用户是所有者
	if file.UserID != userID {
		return fmt.Errorf("只有文件所有者可以删除版本")
	}

	// 2. 查询版本信息
	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("查询版本失败: %w", err)
	}

	// 验证版本属于该文件
	if version.FileID != fileID {
		return fmt.Errorf("版本不属于该文件")
	}

	// 3. 验证不是当前版本
	if version.IsCurrent {
		return fmt.Errorf("不能删除当前版本")
	}

	// 4. 使用事务执行删除操作
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 4.1 查询物理文件信息
		physicalFile, err := s.physicalFileRepo.GetByID(version.PhysicalFileID)
		if err != nil {
			return fmt.Errorf("查询物理文件失败: %w", err)
		}

		// 4.2 删除版本记录
		if err := s.versionRepo.Delete(ctx, versionID); err != nil {
			return fmt.Errorf("删除版本失败: %w", err)
		}

		// 4.3 减少物理文件引用计数
		if err := s.physicalFileRepo.UpdateRefCount(version.PhysicalFileID, -1); err != nil {
			return fmt.Errorf("更新物理文件引用计数失败: %w", err)
		}

		// 4.4 如果没有其他引用，删除物理文件
		if physicalFile.RefCount <= 1 {
			// TODO: 删除物理文件（调用 storage.Delete）
			if err := s.physicalFileRepo.Delete(version.PhysicalFileID); err != nil {
				return fmt.Errorf("删除物理文件失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 5. 清除缓存
	cacheKey := fmt.Sprintf("versions:%d", fileID)
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, nil, 0)
	}

	return nil
}
