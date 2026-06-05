package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/crypto"
	"xingyunpan-v2/pkg/jwt"
	"xingyunpan-v2/pkg/storage"
	"xingyunpan-v2/pkg/token"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ShareService interface {
	CreateShare(ctx context.Context, userID uint, fileIDs []string, expiresIn *int, accessCode *string, maxDownloads *int) (*CreateShareResponse, error)
	GetShareInfo(ctx context.Context, shareToken string) (*ShareInfoResponse, error)
	VerifySharePassword(ctx context.Context, shareToken string, accessCode string, clientIP string) (*VerifyPasswordResponse, error)
	GetMyShares(ctx context.Context, userID uint) ([]*MyShareResponse, error)
	DeleteShare(ctx context.Context, userID uint, shareID uint) error
	IncrementDownloadCount(ctx context.Context, shareToken string) error
	DownloadShare(ctx context.Context, shareToken string) (io.ReadCloser, string, string, error)
	DownloadShareWithDelivery(ctx context.Context, shareToken string) (*FileDownloadResult, error)
}

type shareService struct {
	shareRepo  repository.ShareRepository
	fileRepo   repository.UserFileRepository
	redis      *redis.Client
	jwtSecret  string
	baseURL    string
	storage    *storage.LocalStorage
	db         *gorm.DB
	masterKeys MasterKeyResolver
}

func NewShareService(
	shareRepo repository.ShareRepository,
	fileRepo repository.UserFileRepository,
	redis *redis.Client,
	jwtSecret string,
	baseURL string,
	stor *storage.LocalStorage,
	db ...*gorm.DB,
) ShareService {
	var database *gorm.DB
	if len(db) > 0 {
		database = db[0]
	}
	var settingsRepo repository.FileSystemSettingRepository
	if database != nil {
		settingsRepo = repository.NewFileSystemSettingRepository(database)
	}
	return &shareService{
		shareRepo:  shareRepo,
		fileRepo:   fileRepo,
		redis:      redis,
		jwtSecret:  jwtSecret,
		baseURL:    baseURL,
		storage:    stor,
		db:         database,
		masterKeys: NewFileSystemMasterKeyResolver(database, settingsRepo),
	}
}

type CreateShareResponse struct {
	ShareID      uint       `json:"share_id"`
	ShareToken   string     `json:"share_token"`
	ShareURL     string     `json:"share_url"`
	ExpiresAt    *time.Time `json:"expires_at"`
	MaxDownloads *int       `json:"max_downloads"`
}

type ShareInfoResponse struct {
	ShareID       uint       `json:"share_id"`
	FileIDs       []string   `json:"file_ids"`
	FileNames     []string   `json:"file_names"`
	CreatorName   string     `json:"creator_name"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	HasPassword   bool       `json:"has_password"`
	MaxDownloads  *int       `json:"max_downloads"`
	DownloadCount int        `json:"download_count"`
}

type VerifyPasswordResponse struct {
	Valid       bool   `json:"valid"`
	AccessToken string `json:"access_token,omitempty"`
}

type MyShareResponse struct {
	ShareID       uint       `json:"share_id"`
	ShareToken    string     `json:"share_token"`
	ShareURL      string     `json:"share_url"`
	FileIDs       []string   `json:"file_ids"`
	FileNames     []string   `json:"file_names"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	HasPassword   bool       `json:"has_password"`
	MaxDownloads  *int       `json:"max_downloads"`
	DownloadCount int        `json:"download_count"`
	AccessCount   int        `json:"access_count"`
}

func (s *shareService) CreateShare(ctx context.Context, userID uint, fileIDs []string, expiresIn *int, accessCode *string, maxDownloads *int) (*CreateShareResponse, error) {
	fileIDsUint := make([]uint, len(fileIDs))
	for i, fileIDStr := range fileIDs {
		fileID := parseFileID(fileIDStr)
		file, err := s.fileRepo.GetByID(fileID)
		if err != nil {
			return nil, fmt.Errorf("file not found: %s", fileIDStr)
		}
		if file.UserID != userID {
			return nil, fmt.Errorf("no permission to share file: %s", fileIDStr)
		}
		fileIDsUint[i] = fileID
	}

	shareToken, err := token.GenerateShareToken()
	if err != nil {
		return nil, fmt.Errorf("generate share token failed: %w", err)
	}

	var accessCodeHash string
	if accessCode != nil && *accessCode != "" {
		if len(*accessCode) < 4 || len(*accessCode) > 8 {
			return nil, fmt.Errorf("access code must be 4-8 characters")
		}
		hash, err := crypto.HashPassword(*accessCode)
		if err != nil {
			return nil, fmt.Errorf("hash access code failed: %w", err)
		}
		accessCodeHash = hash
	}

	var expiresAt *time.Time
	if expiresIn != nil && *expiresIn > 0 {
		expiry := time.Now().Add(time.Duration(*expiresIn) * time.Second)
		expiresAt = &expiry
	}

	if maxDownloads != nil && *maxDownloads <= 0 {
		return nil, fmt.Errorf("max downloads must be greater than 0")
	}

	share := &model.Share{
		UserID:         userID,
		ShareToken:     shareToken,
		AccessCodeHash: accessCodeHash,
		ExpiresAt:      expiresAt,
		MaxDownloads:   maxDownloads,
		DownloadCount:  0,
		AccessCount:    0,
	}

	if err := s.shareRepo.Create(ctx, share, fileIDsUint); err != nil {
		return nil, fmt.Errorf("create share failed: %w", err)
	}

	shareURL := fmt.Sprintf("%s/s/%s", s.baseURL, shareToken)
	return &CreateShareResponse{
		ShareID:      share.ID,
		ShareToken:   shareToken,
		ShareURL:     shareURL,
		ExpiresAt:    expiresAt,
		MaxDownloads: maxDownloads,
	}, nil
}

func (s *shareService) GetShareInfo(ctx context.Context, shareToken string) (*ShareInfoResponse, error) {
	share, err := s.shareRepo.GetByToken(ctx, shareToken)
	if err != nil {
		return nil, err
	}
	if isShareUnavailable(share) {
		return nil, fmt.Errorf("share expired")
	}

	_ = s.shareRepo.IncrementAccessCount(ctx, share.ID)

	files, err := s.shareRepo.GetShareFiles(ctx, share.ID)
	if err != nil {
		return nil, fmt.Errorf("get shared files failed: %w", err)
	}

	fileIDs := make([]string, len(files))
	fileNames := make([]string, len(files))
	for i, file := range files {
		fileIDs[i] = fmt.Sprintf("%d", file.ID)
		fileNames[i] = file.FileName
	}

	return &ShareInfoResponse{
		ShareID:       share.ID,
		FileIDs:       fileIDs,
		FileNames:     fileNames,
		CreatorName:   share.User.Username,
		CreatedAt:     share.CreatedAt,
		ExpiresAt:     share.ExpiresAt,
		HasPassword:   share.AccessCodeHash != "",
		MaxDownloads:  share.MaxDownloads,
		DownloadCount: share.DownloadCount,
	}, nil
}

func (s *shareService) VerifySharePassword(ctx context.Context, shareToken string, accessCode string, clientIP string) (*VerifyPasswordResponse, error) {
	share, err := s.shareRepo.GetByToken(ctx, shareToken)
	if err != nil {
		return nil, err
	}
	if isShareUnavailable(share) {
		return nil, fmt.Errorf("share expired")
	}

	rateLimitKey := fmt.Sprintf("share_verify:%s:%d", clientIP, share.ID)
	count, err := s.redis.Get(ctx, rateLimitKey).Int()
	if err != nil && err != redis.Nil {
		count = 0
	}
	if count >= 5 {
		return nil, fmt.Errorf("too many verification attempts")
	}

	if err := crypto.VerifyPassword(share.AccessCodeHash, accessCode); err != nil {
		s.redis.Incr(ctx, rateLimitKey)
		s.redis.Expire(ctx, rateLimitKey, 15*time.Minute)
		return &VerifyPasswordResponse{Valid: false}, nil
	}

	accessToken, err := jwt.GenerateToken(share.UserID, share.User.Username, s.jwtSecret, 1)
	if err != nil {
		return nil, fmt.Errorf("generate access token failed: %w", err)
	}

	return &VerifyPasswordResponse{
		Valid:       true,
		AccessToken: accessToken,
	}, nil
}

func (s *shareService) GetMyShares(ctx context.Context, userID uint) ([]*MyShareResponse, error) {
	shares, err := s.shareRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*MyShareResponse, len(shares))
	for i, share := range shares {
		files, err := s.shareRepo.GetShareFiles(ctx, share.ID)
		if err != nil {
			return nil, fmt.Errorf("get shared files failed: %w", err)
		}

		fileIDs := make([]string, len(files))
		fileNames := make([]string, len(files))
		for j, file := range files {
			fileIDs[j] = fmt.Sprintf("%d", file.ID)
			fileNames[j] = file.FileName
		}

		responses[i] = &MyShareResponse{
			ShareID:       share.ID,
			ShareToken:    share.ShareToken,
			ShareURL:      fmt.Sprintf("%s/s/%s", s.baseURL, share.ShareToken),
			FileIDs:       fileIDs,
			FileNames:     fileNames,
			CreatedAt:     share.CreatedAt,
			ExpiresAt:     share.ExpiresAt,
			HasPassword:   share.AccessCodeHash != "",
			MaxDownloads:  share.MaxDownloads,
			DownloadCount: share.DownloadCount,
			AccessCount:   share.AccessCount,
		}
	}

	return responses, nil
}

func (s *shareService) DeleteShare(ctx context.Context, userID uint, shareID uint) error {
	share, err := s.shareRepo.GetByID(ctx, shareID)
	if err != nil {
		return err
	}
	if share.UserID != userID {
		return fmt.Errorf("no permission to delete this share")
	}
	return s.shareRepo.Delete(ctx, shareID)
}

func (s *shareService) IncrementDownloadCount(ctx context.Context, shareToken string) error {
	share, err := s.shareRepo.GetByToken(ctx, shareToken)
	if err != nil {
		return err
	}
	if isShareUnavailable(share) {
		return fmt.Errorf("share expired")
	}
	return s.shareRepo.IncrementDownloadCount(ctx, share.ID)
}

func (s *shareService) DownloadShare(ctx context.Context, shareToken string) (io.ReadCloser, string, string, error) {
	result, err := s.downloadShareWithDelivery(ctx, shareToken, false)
	if err != nil {
		return nil, "", "", err
	}
	if result == nil || result.Reader == nil {
		return nil, "", "", fmt.Errorf("download is not available")
	}
	return result.Reader, result.FileName, result.ContentType, nil
}

func (s *shareService) DownloadShareWithDelivery(ctx context.Context, shareToken string) (*FileDownloadResult, error) {
	return s.downloadShareWithDelivery(ctx, shareToken, true)
}

func (s *shareService) downloadShareWithDelivery(ctx context.Context, shareToken string, allowCDN bool) (*FileDownloadResult, error) {
	share, err := s.shareRepo.GetByToken(ctx, shareToken)
	if err != nil {
		return nil, err
	}
	if isShareUnavailable(share) {
		return nil, fmt.Errorf("share expired")
	}

	files, err := s.shareRepo.GetShareFiles(ctx, share.ID)
	if err != nil {
		return nil, fmt.Errorf("get shared files failed: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("share has no downloadable files")
	}

	downloadFiles := make([]*model.UserFile, 0, len(files))
	for _, file := range files {
		if file == nil || file.IsFolder {
			continue
		}
		loaded, err := s.fileRepo.GetByIDWithPhysicalFile(file.ID)
		if err != nil {
			return nil, fmt.Errorf("get shared file failed: %w", err)
		}
		if loaded.PhysicalFile == nil {
			return nil, fmt.Errorf("shared file is missing storage metadata")
		}
		downloadFiles = append(downloadFiles, loaded)
	}
	if len(downloadFiles) == 0 {
		return nil, fmt.Errorf("share has no downloadable files")
	}

	if len(downloadFiles) == 1 {
		file := downloadFiles[0]
		if allowCDN && !file.PhysicalFile.Encrypted {
			if redirectURL, err := newStoragePolicyRuntime(s.db, share.UserID).CDNDownloadURL(file.PhysicalFile.StoragePath); err != nil {
				return nil, err
			} else if redirectURL != "" {
				if err := s.shareRepo.IncrementDownloadCount(ctx, share.ID); err != nil {
					return nil, err
				}
				recordTrafficEvent(s.db, share.UserID, "download", file.FileSize, "share", "share", fmt.Sprintf("%d", share.ID))
				recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
					Action:       "share_download",
					UserID:       share.UserID,
					FileID:       file.ID,
					FileName:     file.FileName,
					FileSize:     file.FileSize,
					ResourceType: "share",
					ResourceID:   fmt.Sprintf("%d", share.ID),
					Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
						"delivery":    "cdn",
						"share_token": share.ShareToken,
					}),
				})
				return &FileDownloadResult{
					FileName:    file.FileName,
					ContentType: resolveSharedContentType(file),
					RedirectURL: redirectURL,
				}, nil
			}
		}
		reader, err := readPhysicalBlob(s.storage, file.PhysicalFile, s.masterKeys)
		if err != nil {
			return nil, fmt.Errorf("read shared file failed: %w", err)
		}
		if err := s.shareRepo.IncrementDownloadCount(ctx, share.ID); err != nil {
			_ = reader.Close()
			return nil, err
		}
		recordTrafficEvent(s.db, share.UserID, "download", file.FileSize, "share", "share", fmt.Sprintf("%d", share.ID))
		recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
			Action:       "share_download",
			UserID:       share.UserID,
			FileID:       file.ID,
			FileName:     file.FileName,
			FileSize:     file.FileSize,
			ResourceType: "share",
			ResourceID:   fmt.Sprintf("%d", share.ID),
			Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
				"delivery":    "local_stream",
				"share_token": share.ShareToken,
			}),
		})
		return &FileDownloadResult{
			Reader:      reader,
			FileName:    file.FileName,
			ContentType: resolveSharedContentType(file),
		}, nil
	}

	reader, writer := io.Pipe()
	go s.writeSharedZip(writer, downloadFiles)

	if err := s.shareRepo.IncrementDownloadCount(ctx, share.ID); err != nil {
		_ = reader.Close()
		return nil, err
	}

	recordTrafficEvent(s.db, share.UserID, "download", sumUserFileSizes(downloadFiles), "share", "share", fmt.Sprintf("%d", share.ID))
	recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
		Action:       "share_download",
		UserID:       share.UserID,
		FileName:     fmt.Sprintf("share-%s.zip", share.ShareToken),
		FileSize:     sumUserFileSizes(downloadFiles),
		ResourceType: "share",
		ResourceID:   fmt.Sprintf("%d", share.ID),
		Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
			"delivery":    "zip_stream",
			"share_token": share.ShareToken,
			"file_count":  len(downloadFiles),
		}),
	})

	return &FileDownloadResult{
		Reader:      reader,
		FileName:    fmt.Sprintf("share-%s.zip", share.ShareToken),
		ContentType: "application/zip",
	}, nil
}

func (s *shareService) writeSharedZip(pipeWriter *io.PipeWriter, files []*model.UserFile) {
	zipWriter := zip.NewWriter(pipeWriter)
	usedNames := make(map[string]int, len(files))

	for _, file := range files {
		reader, err := readPhysicalBlob(s.storage, file.PhysicalFile, s.masterKeys)
		if err != nil {
			_ = zipWriter.Close()
			_ = pipeWriter.CloseWithError(err)
			return
		}

		entryName := uniqueArchiveName(safeArchiveName(file.FileName), usedNames)
		entry, err := zipWriter.Create(entryName)
		if err == nil {
			_, err = io.Copy(entry, reader)
		}
		_ = reader.Close()
		if err != nil {
			_ = zipWriter.Close()
			_ = pipeWriter.CloseWithError(err)
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		_ = pipeWriter.CloseWithError(err)
		return
	}
	_ = pipeWriter.Close()
}

func isShareUnavailable(share *model.Share) bool {
	if share == nil {
		return true
	}
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return true
	}
	return share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads
}

func resolveSharedContentType(file *model.UserFile) string {
	if file != nil && file.PhysicalFile != nil {
		contentType := strings.TrimSpace(file.PhysicalFile.ContentType)
		if contentType != "" && contentType != "application/octet-stream" {
			return contentType
		}
	}

	if file != nil {
		if value := strings.TrimSpace(mime.TypeByExtension(strings.ToLower(filepath.Ext(file.FileName)))); value != "" {
			return value
		}
	}

	return "application/octet-stream"
}

func safeArchiveName(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "download"
	}
	return name
}

func uniqueArchiveName(name string, used map[string]int) string {
	count := used[name]
	used[name] = count + 1
	if count == 0 {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	if base == "" {
		base = "download"
	}
	return fmt.Sprintf("%s (%d)%s", base, count, ext)
}

func parseFileID(fileID string) uint {
	var id uint
	fmt.Sscanf(fileID, "%d", &id)
	return id
}
