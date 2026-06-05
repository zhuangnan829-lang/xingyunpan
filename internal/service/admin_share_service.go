package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"

	"gorm.io/gorm"
)

type AdminShareListQuery struct {
	Page                int
	PageSize            int
	Cursor              uint
	UseCursor           bool
	Keyword             string
	OwnerID             uint
	Status              string
	MinDownloads        *int
	ExpiringWithinDays  *int
	MaxDownloadsReached *bool
	Unavailable         *bool
}

type AdminSharePayload struct {
	ShareID       uint       `json:"share_id"`
	ShareToken    string     `json:"share_token"`
	ShareURL      string     `json:"share_url"`
	FileIDs       []string   `json:"file_ids"`
	FileNames     []string   `json:"file_names"`
	OwnerID       uint       `json:"owner_id"`
	OwnerUsername string     `json:"owner_username"`
	OwnerEmail    string     `json:"owner_email"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	HasPassword   bool       `json:"has_password"`
	MaxDownloads  *int       `json:"max_downloads"`
	DownloadCount int        `json:"download_count"`
	AccessCount   int        `json:"access_count"`
	IsExpired     bool       `json:"is_expired"`
	IsUnavailable bool       `json:"is_unavailable"`
	StatusReason  string     `json:"status_reason"`
}

type AdminShareMetricsPayload struct {
	TotalShares               int64 `json:"total_shares"`
	ActiveShares              int64 `json:"active_shares"`
	ExpiredShares             int64 `json:"expired_shares"`
	ProtectedShares           int64 `json:"protected_shares"`
	TotalAccessCount          int64 `json:"total_access_count"`
	TotalDownloadCount        int64 `json:"total_download_count"`
	ExpiringSoonCount         int64 `json:"expiring_soon_count"`
	DownloadLimitReachedCount int64 `json:"download_limit_reached_count"`
}

type AdminShareService interface {
	List(ctx context.Context, query *AdminShareListQuery) ([]AdminSharePayload, int64, error)
	Metrics(ctx context.Context, expiringWithinDays int) (*AdminShareMetricsPayload, error)
	Delete(ctx context.Context, shareID uint) error
	BatchDelete(ctx context.Context, shareIDs []uint) error
}

type adminShareService struct {
	db        *gorm.DB
	shareRepo repository.ShareRepository
	baseURL   string
}

func NewAdminShareService(db *gorm.DB, shareRepo repository.ShareRepository, baseURL string) AdminShareService {
	return &adminShareService{
		db:        db,
		shareRepo: shareRepo,
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
	}
}

func (s *adminShareService) List(ctx context.Context, query *AdminShareListQuery) ([]AdminSharePayload, int64, error) {
	if query == nil {
		query = &AdminShareListQuery{}
	}

	page := normalizeAdminSharePage(query.Page)
	pageSize := normalizeAdminSharePageSize(query.PageSize)
	base := s.buildListQuery(ctx, query)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计分享失败: %w", err)
	}

	var shares []model.Share
	dataQuery := base
	if query.UseCursor && query.Cursor > 0 {
		dataQuery = dataQuery.Where("shares.id < ?", query.Cursor)
	}
	if err := dataQuery.
		Preload("User").
		Order("shares.created_at DESC, shares.id DESC").
		Offset(adminShareListOffset(query.UseCursor, page, pageSize)).
		Limit(pageSize).
		Find(&shares).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分享列表失败: %w", err)
	}

	items := make([]AdminSharePayload, 0, len(shares))
	for _, share := range shares {
		files, err := s.shareRepo.GetShareFiles(ctx, share.ID)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s.buildPayload(&share, files))
	}

	return items, total, nil
}

func adminShareListOffset(useCursor bool, page, pageSize int) int {
	if useCursor {
		return 0
	}
	return (page - 1) * pageSize
}

func (s *adminShareService) Delete(ctx context.Context, shareID uint) error {
	if shareID == 0 {
		return fmt.Errorf("分享 ID 不能为空")
	}
	return s.shareRepo.Delete(ctx, shareID)
}

func (s *adminShareService) Metrics(ctx context.Context, expiringWithinDays int) (*AdminShareMetricsPayload, error) {
	if expiringWithinDays <= 0 {
		expiringWithinDays = 3
	}

	now := time.Now()
	soon := now.AddDate(0, 0, expiringWithinDays)
	base := func() *gorm.DB {
		return s.db.WithContext(ctx).Model(&model.Share{})
	}
	payload := &AdminShareMetricsPayload{}

	if err := base().Count(&payload.TotalShares).Error; err != nil {
		return nil, fmt.Errorf("统计分享总量失败: %w", err)
	}
	if err := base().Where(adminShareAvailableCondition(), now).Count(&payload.ActiveShares).Error; err != nil {
		return nil, fmt.Errorf("统计生效分享失败: %w", err)
	}
	if err := base().Where("shares.expires_at IS NOT NULL AND shares.expires_at <= ?", now).Count(&payload.ExpiredShares).Error; err != nil {
		return nil, fmt.Errorf("统计过期分享失败: %w", err)
	}
	if err := base().Where("shares.access_code_hash <> ''").Count(&payload.ProtectedShares).Error; err != nil {
		return nil, fmt.Errorf("统计加密分享失败: %w", err)
	}
	if err := base().Where(adminShareDownloadLimitReachedCondition()).Count(&payload.DownloadLimitReachedCount).Error; err != nil {
		return nil, fmt.Errorf("统计下载上限分享失败: %w", err)
	}
	if err := base().
		Where("shares.expires_at IS NOT NULL AND shares.expires_at > ? AND shares.expires_at <= ?", now, soon).
		Where(adminShareNotDownloadLimitReachedCondition()).
		Count(&payload.ExpiringSoonCount).Error; err != nil {
		return nil, fmt.Errorf("统计即将过期分享失败: %w", err)
	}

	var sums struct {
		Access   int64
		Download int64
	}
	if err := base().Select("COALESCE(SUM(access_count), 0) AS access, COALESCE(SUM(download_count), 0) AS download").Scan(&sums).Error; err != nil {
		return nil, fmt.Errorf("统计分享访问下载失败: %w", err)
	}
	payload.TotalAccessCount = sums.Access
	payload.TotalDownloadCount = sums.Download

	return payload, nil
}

func (s *adminShareService) BatchDelete(ctx context.Context, shareIDs []uint) error {
	if len(shareIDs) == 0 {
		return fmt.Errorf("分享 ID 列表不能为空")
	}

	seen := make(map[uint]struct{}, len(shareIDs))
	for _, shareID := range shareIDs {
		if shareID == 0 {
			return fmt.Errorf("分享 ID 不能为空")
		}
		if _, exists := seen[shareID]; exists {
			continue
		}
		seen[shareID] = struct{}{}
		if err := s.shareRepo.Delete(ctx, shareID); err != nil {
			return err
		}
	}

	return nil
}

func (s *adminShareService) buildListQuery(ctx context.Context, query *AdminShareListQuery) *gorm.DB {
	keyword := strings.TrimSpace(query.Keyword)
	status := strings.ToLower(strings.TrimSpace(query.Status))

	base := s.db.WithContext(ctx).
		Model(&model.Share{}).
		Joins("LEFT JOIN users ON users.id = shares.user_id")

	if query.OwnerID > 0 {
		base = base.Where("shares.user_id = ?", query.OwnerID)
	}

	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(`
			shares.share_token LIKE ?
			OR users.username LIKE ?
			OR users.email LIKE ?
			OR EXISTS (
				SELECT 1
				FROM share_files
				INNER JOIN user_files ON user_files.id = share_files.file_id
				WHERE share_files.share_id = shares.id
				  AND user_files.deleted_at IS NULL
				  AND user_files.file_name LIKE ?
			)
		`, like, like, like, like)
	}

	if query.MinDownloads != nil {
		base = base.Where("shares.download_count >= ?", *query.MinDownloads)
	}

	now := time.Now()
	if query.ExpiringWithinDays != nil {
		days := *query.ExpiringWithinDays
		if days < 0 {
			days = 0
		}
		base = base.
			Where("shares.expires_at IS NOT NULL AND shares.expires_at > ? AND shares.expires_at <= ?", now, now.AddDate(0, 0, days)).
			Where(adminShareNotDownloadLimitReachedCondition())
	}

	if query.MaxDownloadsReached != nil {
		if *query.MaxDownloadsReached {
			base = base.Where(adminShareDownloadLimitReachedCondition())
		} else {
			base = base.Where(adminShareNotDownloadLimitReachedCondition())
		}
	}

	if query.Unavailable != nil {
		if *query.Unavailable {
			base = base.Where(adminShareUnavailableCondition(), now)
		} else {
			base = base.Where(adminShareAvailableCondition(), now)
		}
	}

	switch status {
	case "active":
		base = base.Where(adminShareAvailableCondition(), now)
	case "expired":
		base = base.Where("shares.expires_at IS NOT NULL AND shares.expires_at <= ?", now)
	case "protected":
		base = base.Where("shares.access_code_hash <> ''")
	case "unavailable":
		base = base.Where(adminShareUnavailableCondition(), now)
	case "download_limit_reached", "max_downloads_reached":
		base = base.Where(adminShareDownloadLimitReachedCondition())
	}

	return base
}

func (s *adminShareService) buildPayload(share *model.Share, files []*model.UserFile) AdminSharePayload {
	fileIDs := make([]string, 0, len(files))
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		fileIDs = append(fileIDs, fmt.Sprintf("%d", file.ID))
		fileNames = append(fileNames, file.FileName)
	}

	timeExpired := share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now())
	downloadLimitReached := share.MaxDownloads != nil && share.DownloadCount >= *share.MaxDownloads
	statusReason := "active"
	switch {
	case timeExpired:
		statusReason = "time_expired"
	case downloadLimitReached:
		statusReason = "download_limit_reached"
	}

	return AdminSharePayload{
		ShareID:       share.ID,
		ShareToken:    share.ShareToken,
		ShareURL:      s.shareURL(share.ShareToken),
		FileIDs:       fileIDs,
		FileNames:     fileNames,
		OwnerID:       share.UserID,
		OwnerUsername: share.User.Username,
		OwnerEmail:    share.User.Email,
		CreatedAt:     share.CreatedAt,
		ExpiresAt:     share.ExpiresAt,
		HasPassword:   strings.TrimSpace(share.AccessCodeHash) != "",
		MaxDownloads:  share.MaxDownloads,
		DownloadCount: share.DownloadCount,
		AccessCount:   share.AccessCount,
		IsExpired:     timeExpired,
		IsUnavailable: timeExpired || downloadLimitReached,
		StatusReason:  statusReason,
	}
}

func adminShareDownloadLimitReachedCondition() string {
	return "shares.max_downloads IS NOT NULL AND shares.download_count >= shares.max_downloads"
}

func adminShareNotDownloadLimitReachedCondition() string {
	return "(shares.max_downloads IS NULL OR shares.download_count < shares.max_downloads)"
}

func adminShareUnavailableCondition() string {
	return "((shares.expires_at IS NOT NULL AND shares.expires_at <= ?) OR " + adminShareDownloadLimitReachedCondition() + ")"
}

func adminShareAvailableCondition() string {
	return "((shares.expires_at IS NULL OR shares.expires_at > ?) AND " + adminShareNotDownloadLimitReachedCondition() + ")"
}

func (s *adminShareService) shareURL(token string) string {
	if s.baseURL == "" {
		return "/s/" + token
	}
	return s.baseURL + "/s/" + token
}

func normalizeAdminSharePage(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func normalizeAdminSharePageSize(value int) int {
	if value <= 0 {
		return 10
	}
	return value
}
