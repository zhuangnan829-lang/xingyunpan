// 路径: internal/repository/user_file_repository.go
package repository

import (
	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// UserFileRepository 用户文件仓储接口
type UserFileRepository interface {
	Create(file *model.UserFile) error
	GetByID(id uint) (*model.UserFile, error)
	GetByIDWithPhysicalFile(id uint) (*model.UserFile, error)
	List(userID uint, parentID *uint, page, pageSize int) ([]*model.UserFile, int64, error)
	ListAfterID(userID uint, parentID *uint, afterID uint, pageSize int) ([]*model.UserFile, int64, error)
	ListChildren(userID uint, parentID uint) ([]*model.UserFile, error)
	ListDescendants(userID uint, folderID uint) ([]*model.UserFile, error)
	GetFolderPath(userID uint, folderID uint) ([]*model.UserFile, error)
	GetImmediateChildStats(userID uint, folderIDs []uint) (map[uint]FolderChildStats, error)
	Update(file *model.UserFile) error
	Delete(id uint) error
}

type FolderChildStats struct {
	FolderID    uint
	ChildCount  int64
	FileCount   int64
	FolderCount int64
	TotalSize   int64
}

type userFileRepository struct {
	db *gorm.DB
}

func NewUserFileRepository(db *gorm.DB) UserFileRepository {
	return &userFileRepository{db: db}
}

func (r *userFileRepository) Create(file *model.UserFile) error {
	return r.db.Create(file).Error
}

func (r *userFileRepository) GetByID(id uint) (*model.UserFile, error) {
	var file model.UserFile
	if err := r.db.First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByIDWithPhysicalFile 获取文件并预加载物理文件信息（避免 N+1 查询）
func (r *userFileRepository) GetByIDWithPhysicalFile(id uint) (*model.UserFile, error) {
	var file model.UserFile
	if err := r.db.Preload("PhysicalFile").First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *userFileRepository) List(userID uint, parentID *uint, page, pageSize int) ([]*model.UserFile, int64, error) {
	var files []*model.UserFile
	var total int64

	query := r.db.Model(&model.UserFile{}).Where("user_id = ?", userID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	// ✅ 任务 4: 预加载 PhysicalFile 关联,避免 N+1 查询
	// ✅ 添加排序: 文件夹优先,按创建时间倒序
	if err := query.
		Preload("PhysicalFile").
		Order("is_folder DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (r *userFileRepository) ListAfterID(userID uint, parentID *uint, afterID uint, pageSize int) ([]*model.UserFile, int64, error) {
	var files []*model.UserFile
	var total int64

	query := r.db.Model(&model.UserFile{}).Where("user_id = ?", userID)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := query
	if afterID > 0 {
		dataQuery = dataQuery.Where("id < ?", afterID)
	}
	if err := dataQuery.
		Preload("PhysicalFile").
		Order("id DESC").
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (r *userFileRepository) ListChildren(userID uint, parentID uint) ([]*model.UserFile, error) {
	var files []*model.UserFile
	if err := r.db.
		Preload("PhysicalFile").
		Where("user_id = ? AND parent_id = ?", userID, parentID).
		Order("is_folder DESC, created_at ASC").
		Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (r *userFileRepository) ListDescendants(userID uint, folderID uint) ([]*model.UserFile, error) {
	result := make([]*model.UserFile, 0)
	queue := []uint{folderID}
	visited := make(map[uint]struct{})

	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		if _, ok := visited[parentID]; ok {
			continue
		}
		visited[parentID] = struct{}{}

		children, err := r.ListChildren(userID, parentID)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			result = append(result, child)
			if child.IsFolder {
				queue = append(queue, child.ID)
			}
		}
	}

	return result, nil
}

func (r *userFileRepository) GetFolderPath(userID uint, folderID uint) ([]*model.UserFile, error) {
	path := make([]*model.UserFile, 0)
	visited := make(map[uint]struct{})
	currentID := folderID

	for currentID != 0 {
		if _, ok := visited[currentID]; ok {
			break
		}
		visited[currentID] = struct{}{}

		folder, err := r.GetByID(currentID)
		if err != nil {
			return nil, err
		}
		if folder.UserID != userID || !folder.IsFolder {
			return nil, gorm.ErrRecordNotFound
		}

		path = append(path, folder)
		if folder.ParentID == nil {
			break
		}
		currentID = *folder.ParentID
	}

	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}

	return path, nil
}

func (r *userFileRepository) GetImmediateChildStats(userID uint, folderIDs []uint) (map[uint]FolderChildStats, error) {
	if len(folderIDs) == 0 {
		return map[uint]FolderChildStats{}, nil
	}

	var rows []struct {
		ParentID    uint
		ChildCount  int64
		FileCount   int64
		FolderCount int64
		TotalSize   int64
	}

	if err := r.db.Model(&model.UserFile{}).
		Select(`
			parent_id,
			COUNT(*) AS child_count,
			SUM(CASE WHEN is_folder = ? THEN 0 ELSE 1 END) AS file_count,
			SUM(CASE WHEN is_folder = ? THEN 1 ELSE 0 END) AS folder_count,
			COALESCE(SUM(CASE WHEN is_folder = ? THEN 0 ELSE file_size END), 0) AS total_size
		`, true, true, true).
		Where("user_id = ? AND parent_id IN ?", userID, folderIDs).
		Group("parent_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	stats := make(map[uint]FolderChildStats, len(folderIDs))
	for _, folderID := range folderIDs {
		stats[folderID] = FolderChildStats{FolderID: folderID}
	}

	for _, row := range rows {
		stats[row.ParentID] = FolderChildStats{
			FolderID:    row.ParentID,
			ChildCount:  row.ChildCount,
			FileCount:   row.FileCount,
			FolderCount: row.FolderCount,
			TotalSize:   row.TotalSize,
		}
	}

	return stats, nil
}

func (r *userFileRepository) Update(file *model.UserFile) error {
	return r.db.Save(file).Error
}

func (r *userFileRepository) Delete(id uint) error {
	return r.db.Delete(&model.UserFile{}, id).Error
}
