package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"xingyunpan-v2/internal/repository"

	"gorm.io/gorm"
)

type AdminFileListQuery struct {
	Page            int
	PageSize        int
	Cursor          uint
	UseCursor       bool
	OwnerID         uint
	Keyword         string
	StoragePolicyID uint
}

type AdminFilePayload struct {
	ID                uint      `json:"id"`
	FileName          string    `json:"file_name"`
	IsFolder          bool      `json:"is_folder"`
	FileSize          int64     `json:"file_size"`
	OccupiedSize      int64     `json:"occupied_size"`
	FilePath          string    `json:"file_path"`
	ContentType       string    `json:"-"`
	OwnerID           uint      `json:"owner_id"`
	OwnerUsername     string    `json:"owner_username"`
	OwnerEmail        string    `json:"owner_email"`
	StoragePolicyID   uint      `json:"storage_policy_id"`
	StoragePolicyName string    `json:"storage_policy_name"`
	PhysicalFileID    *uint     `json:"physical_file_id"`
	DisplayIcon       string    `json:"display_icon"`
	DisplayIconTint   string    `json:"display_icon_tint"`
	DisplayIconLabel  string    `json:"display_icon_label"`
	DisplayIconSource string    `json:"display_icon_source"`
	HasShareLink      bool      `json:"has_share_link"`
	HasDirectLink     bool      `json:"has_direct_link"`
	Uploading         bool      `json:"uploading"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AdminFileService interface {
	List(query *AdminFileListQuery) ([]AdminFilePayload, int64, error)
	Get(fileID uint) (*AdminFilePayload, error)
	Import(ownerID uint, parentID *uint, fileName string, fileSize int64, reader io.Reader) (*AdminFilePayload, error)
	Rename(fileID uint, newName string) error
	Delete(fileID uint) error
	Download(fileID uint) (io.ReadCloser, string, string, error)
	DownloadWithDelivery(fileID uint) (*FileDownloadResult, error)
	CreateShare(ctx context.Context, fileID uint, expiresIn *int, accessCode *string) (*CreateShareResponse, error)
}

type adminFileService struct {
	db           *gorm.DB
	fileRepo     repository.UserFileRepository
	userRepo     repository.UserRepository
	fileService  FileService
	shareService ShareService
}

func NewAdminFileService(
	db *gorm.DB,
	fileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	fileService FileService,
	shareService ShareService,
) AdminFileService {
	return &adminFileService{
		db:           db,
		fileRepo:     fileRepo,
		userRepo:     userRepo,
		fileService:  fileService,
		shareService: shareService,
	}
}

func (s *adminFileService) List(query *AdminFileListQuery) ([]AdminFilePayload, int64, error) {
	if query == nil {
		query = &AdminFileListQuery{}
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	base := s.buildListQuery(query)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]AdminFilePayload, 0, pageSize)
	dataQuery := base
	if query.UseCursor && query.Cursor > 0 {
		dataQuery = dataQuery.Where("user_files.id < ?", query.Cursor)
	}

	if err := dataQuery.
		Order("user_files.id DESC").
		Offset(adminFileListOffset(query.UseCursor, page, pageSize)).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func adminFileListOffset(useCursor bool, page, pageSize int) int {
	if useCursor {
		return 0
	}
	return (page - 1) * pageSize
}

func (s *adminFileService) Get(fileID uint) (*AdminFilePayload, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("file id cannot be empty")
	}

	query := s.buildListQuery(&AdminFileListQuery{})
	var item AdminFilePayload
	if err := query.Where("user_files.id = ?", fileID).Take(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *adminFileService) Import(ownerID uint, parentID *uint, fileName string, fileSize int64, reader io.Reader) (*AdminFilePayload, error) {
	if ownerID == 0 {
		return nil, fmt.Errorf("owner id cannot be empty")
	}
	if _, err := s.userRepo.GetByID(ownerID); err != nil {
		return nil, err
	}

	userFile, err := s.fileService.Upload(ownerID, strings.TrimSpace(fileName), fileSize, reader, parentID)
	if err != nil {
		return nil, err
	}

	return s.Get(userFile.ID)
}

func (s *adminFileService) Rename(fileID uint, newName string) error {
	userFile, err := s.getUserFile(fileID)
	if err != nil {
		return err
	}

	return s.fileService.Rename(userFile.UserID, fileID, newName)
}

func (s *adminFileService) Delete(fileID uint) error {
	userFile, err := s.getUserFile(fileID)
	if err != nil {
		return err
	}

	return s.fileService.Delete(userFile.UserID, fileID)
}

func (s *adminFileService) Download(fileID uint) (io.ReadCloser, string, string, error) {
	userFile, err := s.getUserFile(fileID)
	if err != nil {
		return nil, "", "", err
	}

	return s.fileService.Download(userFile.UserID, fileID)
}

func (s *adminFileService) DownloadWithDelivery(fileID uint) (*FileDownloadResult, error) {
	userFile, err := s.getUserFile(fileID)
	if err != nil {
		return nil, err
	}

	return s.fileService.DownloadWithDelivery(userFile.UserID, fileID, false)
}

func (s *adminFileService) CreateShare(ctx context.Context, fileID uint, expiresIn *int, accessCode *string) (*CreateShareResponse, error) {
	userFile, err := s.getUserFile(fileID)
	if err != nil {
		return nil, err
	}
	if userFile.IsFolder {
		return nil, fmt.Errorf("folders cannot be shared from this entry")
	}

	return s.shareService.CreateShare(ctx, userFile.UserID, []string{fmt.Sprintf("%d", fileID)}, expiresIn, accessCode, nil)
}

func (s *adminFileService) getUserFile(fileID uint) (*repositoryUserFileRef, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("file id cannot be empty")
	}

	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, err
	}

	return &repositoryUserFileRef{ID: file.ID, UserID: file.UserID, IsFolder: file.IsFolder}, nil
}

func (s *adminFileService) buildListQuery(query *AdminFileListQuery) *gorm.DB {
	keyword := strings.TrimSpace(query.Keyword)

	base := s.db.Table("user_files").
		Select(`
			user_files.id,
			user_files.file_name,
			user_files.is_folder,
			user_files.file_size,
			COALESCE(physical_files.file_size, user_files.file_size, 0) AS occupied_size,
			user_files.file_path,
			user_files.physical_file_id,
			COALESCE(physical_files.content_type, '') AS content_type,
			users.id AS owner_id,
			users.username AS owner_username,
			users.email AS owner_email,
			COALESCE(storage_policies.id, 0) AS storage_policy_id,
			COALESCE(storage_policies.name, '') AS storage_policy_name,
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM share_files
					INNER JOIN shares ON shares.id = share_files.share_id
					WHERE share_files.file_id = user_files.id
					  AND shares.deleted_at IS NULL
				) THEN TRUE
				ELSE FALSE
			END AS has_share_link,
			CASE
				WHEN user_files.is_folder = FALSE AND user_files.physical_file_id IS NOT NULL THEN TRUE
				ELSE FALSE
			END AS has_direct_link,
			FALSE AS uploading,
			user_files.created_at,
			user_files.updated_at
		`).
		Joins("LEFT JOIN users ON users.id = user_files.user_id").
		Joins("LEFT JOIN user_groups ON user_groups.id = users.user_group_id").
		Joins("LEFT JOIN storage_policies ON storage_policies.id = user_groups.storage_policy_id").
		Joins("LEFT JOIN physical_files ON physical_files.id = user_files.physical_file_id")

	if query.OwnerID > 0 {
		base = base.Where("users.id = ?", query.OwnerID)
	}
	if keyword != "" {
		base = base.Where("user_files.file_name LIKE ?", "%"+keyword+"%")
	}
	if query.StoragePolicyID > 0 {
		base = base.Where("storage_policies.id = ?", query.StoragePolicyID)
	}

	return base
}

type repositoryUserFileRef struct {
	ID       uint
	UserID   uint
	IsFolder bool
}
