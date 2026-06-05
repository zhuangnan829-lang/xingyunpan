package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
)

type DashboardMetricPayload struct {
	TodayUploadCount   int64   `json:"today_upload_count"`
	TodayUploadBytes   int64   `json:"today_upload_bytes"`
	UploadChangePct    float64 `json:"upload_change_pct"`
	AverageLatencyMS   int64   `json:"average_latency_ms"`
	LatencyChangePct   float64 `json:"latency_change_pct"`
	ActiveUsers        int64   `json:"active_users"`
	ActiveUsersChange  float64 `json:"active_users_change_pct"`
	BlobCount          int64   `json:"blob_count"`
	BlobCountChangePct float64 `json:"blob_count_change_pct"`
	FileCount          int64   `json:"file_count"`
	FolderCount        int64   `json:"folder_count"`
}

type DashboardStorageBreakdownPayload struct {
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
	Tone      string `json:"tone"`
}

type DashboardStoragePayload struct {
	UsedBytes  int64                              `json:"used_bytes"`
	TotalBytes int64                              `json:"total_bytes"`
	Percent    float64                            `json:"percent"`
	Breakdown  []DashboardStorageBreakdownPayload `json:"breakdown"`
}

type DashboardTrafficPayload struct {
	Labels     []string `json:"labels"`
	Inbound    []int64  `json:"inbound"`
	Outbound   []int64  `json:"outbound"`
	PeakValue  int64    `json:"peak_value"`
	PeakWindow string   `json:"peak_window"`
}

type DashboardOnlineGroupPayload struct {
	Name    string `json:"name"`
	Value   int64  `json:"value"`
	Percent int    `json:"percent"`
}

type DashboardOnlinePayload struct {
	CurrentSessions int64                         `json:"current_sessions"`
	Groups          []DashboardOnlineGroupPayload `json:"groups"`
}

type DashboardNodePayload struct {
	Name      string `json:"name"`
	Region    string `json:"region"`
	LatencyMS int64  `json:"latency_ms"`
	Load      int    `json:"load"`
	Status    string `json:"status"`
}

type DashboardNodesPayload struct {
	Online int64                  `json:"online"`
	Total  int64                  `json:"total"`
	Items  []DashboardNodePayload `json:"items"`
}

type DashboardTaskPayload struct {
	Name       string `json:"name"`
	Detail     string `json:"detail"`
	Progress   int    `json:"progress"`
	Success    int64  `json:"success"`
	Failed     int64  `json:"failed"`
	Processing int64  `json:"processing"`
	Pending    int64  `json:"pending"`
	Submitted  int64  `json:"submitted"`
}

type DashboardOverviewPayload struct {
	GeneratedAt time.Time               `json:"generated_at"`
	Metrics     DashboardMetricPayload  `json:"metrics"`
	Storage     DashboardStoragePayload `json:"storage"`
	Traffic     DashboardTrafficPayload `json:"traffic"`
	Online      DashboardOnlinePayload  `json:"online"`
	Nodes       DashboardNodesPayload   `json:"nodes"`
	Tasks       []DashboardTaskPayload  `json:"tasks"`
}

type DashboardService interface {
	GetOverview(ctx context.Context) (*DashboardOverviewPayload, error)
}

type dashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) DashboardService {
	return &dashboardService{db: db}
}

func (s *dashboardService) GetOverview(ctx context.Context) (*DashboardOverviewPayload, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	now := time.Now()
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startYesterday := startToday.AddDate(0, 0, -1)

	metrics, err := s.loadMetrics(ctx, startToday, startYesterday)
	if err != nil {
		return nil, err
	}

	storage, err := s.loadStorage(ctx)
	if err != nil {
		return nil, err
	}

	traffic, err := s.loadTraffic(ctx, startToday)
	if err != nil {
		return nil, err
	}

	online, err := s.loadOnline(ctx)
	if err != nil {
		return nil, err
	}

	nodes, err := s.loadNodes(ctx)
	if err != nil {
		return nil, err
	}
	metrics.AverageLatencyMS = averageNodeLatency(nodes.Items)

	tasks, err := s.loadTasks(ctx)
	if err != nil {
		return nil, err
	}

	return &DashboardOverviewPayload{
		GeneratedAt: now,
		Metrics:     metrics,
		Storage:     storage,
		Traffic:     traffic,
		Online:      online,
		Nodes:       nodes,
		Tasks:       tasks,
	}, nil
}

func (s *dashboardService) loadMetrics(ctx context.Context, startToday, startYesterday time.Time) (DashboardMetricPayload, error) {
	var payload DashboardMetricPayload
	if err := s.db.WithContext(ctx).Model(&model.UserFile{}).Where("is_folder = ?", false).Count(&payload.FileCount).Error; err != nil {
		return payload, fmt.Errorf("count files failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&model.UserFile{}).Where("is_folder = ?", true).Count(&payload.FolderCount).Error; err != nil {
		return payload, fmt.Errorf("count folders failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&model.PhysicalFile{}).Count(&payload.BlobCount).Error; err != nil {
		return payload, fmt.Errorf("count blobs failed: %w", err)
	}

	var yesterdayUploads int64
	if err := s.db.WithContext(ctx).Model(&model.UserFile{}).
		Where("is_folder = ? AND created_at >= ?", false, startToday).
		Count(&payload.TodayUploadCount).Error; err != nil {
		return payload, fmt.Errorf("count today uploads failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&model.UserFile{}).
		Where("is_folder = ? AND created_at >= ? AND created_at < ?", false, startYesterday, startToday).
		Count(&yesterdayUploads).Error; err != nil {
		return payload, fmt.Errorf("count yesterday uploads failed: %w", err)
	}
	payload.UploadChangePct = percentChange(float64(payload.TodayUploadCount), float64(yesterdayUploads))

	if err := s.db.WithContext(ctx).Model(&model.UserFile{}).
		Where("is_folder = ? AND created_at >= ?", false, startToday).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&payload.TodayUploadBytes).Error; err != nil {
		return payload, fmt.Errorf("sum today upload bytes failed: %w", err)
	}

	var enabledUsers int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("enabled = ?", true).Count(&enabledUsers).Error; err != nil {
		return payload, fmt.Errorf("count active users failed: %w", err)
	}
	payload.ActiveUsers = enabledUsers

	var yesterdayUsers int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).
		Where("created_at < ?", startToday).
		Count(&yesterdayUsers).Error; err != nil {
		return payload, fmt.Errorf("count previous users failed: %w", err)
	}
	payload.ActiveUsersChange = percentChange(float64(payload.ActiveUsers), float64(yesterdayUsers))

	var previousBlobCount int64
	if err := s.db.WithContext(ctx).Model(&model.PhysicalFile{}).
		Where("created_at < ?", startToday).
		Count(&previousBlobCount).Error; err != nil {
		return payload, fmt.Errorf("count previous blobs failed: %w", err)
	}
	payload.BlobCountChangePct = percentChange(float64(payload.BlobCount), float64(previousBlobCount))

	return payload, nil
}

func (s *dashboardService) loadStorage(ctx context.Context) (DashboardStoragePayload, error) {
	var capacityBytes int64
	var currentFileBytes int64
	var mediaBytes int64
	var versionBytes int64

	if err := s.db.WithContext(ctx).Model(&model.User{}).Select("COALESCE(SUM(capacity), 0)").Scan(&capacityBytes).Error; err != nil {
		return DashboardStoragePayload{}, fmt.Errorf("sum user capacity failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Table("user_files").
		Joins("LEFT JOIN physical_files ON physical_files.id = user_files.physical_file_id").
		Where("user_files.deleted_at IS NULL AND user_files.is_folder = ?", false).
		Select("COALESCE(SUM(COALESCE(NULLIF(user_files.file_size, 0), physical_files.file_size, 0)), 0)").
		Scan(&currentFileBytes).Error; err != nil {
		return DashboardStoragePayload{}, fmt.Errorf("sum current file size failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Table("user_files").
		Joins("LEFT JOIN physical_files ON physical_files.id = user_files.physical_file_id").
		Where("user_files.deleted_at IS NULL AND user_files.is_folder = ? AND (physical_files.content_type LIKE ? OR physical_files.content_type LIKE ? OR physical_files.content_type LIKE ?)", false, "image/%", "video/%", "audio/%").
		Select("COALESCE(SUM(COALESCE(NULLIF(user_files.file_size, 0), physical_files.file_size, 0)), 0)").
		Scan(&mediaBytes).Error; err != nil {
		return DashboardStoragePayload{}, fmt.Errorf("sum media size failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&model.FileVersion{}).
		Where("is_current = ?", false).
		Select("COALESCE(SUM(file_size), 0)").Scan(&versionBytes).Error; err != nil {
		return DashboardStoragePayload{}, fmt.Errorf("sum version size failed: %w", err)
	}

	usedBytes := currentFileBytes + versionBytes
	if capacityBytes <= 0 {
		capacityBytes = maxInt64(usedBytes, 10*1024*1024*1024)
	}

	assetBytes := maxInt64(currentFileBytes-mediaBytes, 0)
	return DashboardStoragePayload{
		UsedBytes:  usedBytes,
		TotalBytes: capacityBytes,
		Percent:    dashboardClampFloat(round1(float64(usedBytes)/float64(maxInt64(capacityBytes, 1))*100), 0, 100),
		Breakdown: []DashboardStorageBreakdownPayload{
			{Name: "文件资产", SizeBytes: assetBytes, Tone: "blue"},
			{Name: "媒体缓存", SizeBytes: mediaBytes, Tone: "pink"},
			{Name: "版本备份", SizeBytes: versionBytes, Tone: "amber"},
		},
	}, nil
}

func (s *dashboardService) loadTraffic(ctx context.Context, startToday time.Time) (DashboardTrafficPayload, error) {
	labels := make([]string, 0, 8)
	inbound := make([]int64, 0, 8)
	outbound := make([]int64, 0, 8)

	var maxValue int64
	peakWindow := "00:00-03:00"

	for index := 0; index < 8; index++ {
		windowStart := startToday.Add(time.Duration(index*3) * time.Hour)
		windowEnd := windowStart.Add(3 * time.Hour)
		labels = append(labels, windowStart.Format("15:04"))

		uploadedBytes, downloadedBytes, err := s.loadTrafficBucket(ctx, windowStart, windowEnd)
		if err != nil {
			return DashboardTrafficPayload{}, err
		}

		inbound = append(inbound, uploadedBytes)
		outbound = append(outbound, downloadedBytes)

		if bucketPeak := maxInt64(uploadedBytes, downloadedBytes); bucketPeak > maxValue {
			maxValue = bucketPeak
			peakWindow = windowStart.Format("15:04") + "-" + windowEnd.Format("15:04")
		}
	}

	return DashboardTrafficPayload{
		Labels:     labels,
		Inbound:    inbound,
		Outbound:   outbound,
		PeakValue:  maxValue,
		PeakWindow: peakWindow,
	}, nil
}

func (s *dashboardService) loadTrafficBucket(ctx context.Context, windowStart, windowEnd time.Time) (int64, int64, error) {
	var uploadedBytes int64
	var downloadedBytes int64

	if s.db.Migrator().HasTable(&model.TrafficEvent{}) {
		if err := s.db.WithContext(ctx).Model(&model.TrafficEvent{}).
			Where("direction = ? AND created_at >= ? AND created_at < ?", "upload", windowStart, windowEnd).
			Select("COALESCE(SUM(bytes), 0)").
			Scan(&uploadedBytes).Error; err != nil {
			return 0, 0, fmt.Errorf("sum traffic upload events failed: %w", err)
		}
		if err := s.db.WithContext(ctx).Model(&model.TrafficEvent{}).
			Where("direction = ? AND created_at >= ? AND created_at < ?", "download", windowStart, windowEnd).
			Select("COALESCE(SUM(bytes), 0)").
			Scan(&downloadedBytes).Error; err != nil {
			return 0, 0, fmt.Errorf("sum traffic download events failed: %w", err)
		}
	}

	if uploadedBytes == 0 {
		if err := s.db.WithContext(ctx).Model(&model.UserFile{}).
			Where("is_folder = ? AND created_at >= ? AND created_at < ?", false, windowStart, windowEnd).
			Select("COALESCE(SUM(file_size), 0)").
			Scan(&uploadedBytes).Error; err != nil {
			return 0, 0, fmt.Errorf("sum traffic upload bucket failed: %w", err)
		}
	}

	return uploadedBytes, downloadedBytes, nil
}

func (s *dashboardService) loadOnline(ctx context.Context) (DashboardOnlinePayload, error) {
	var adminUsers int64
	var enabledUsers int64
	var disabledUsers int64
	var onlineUsers int64

	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("role = ?", "admin").Count(&adminUsers).Error; err != nil {
		return DashboardOnlinePayload{}, fmt.Errorf("count admin users failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("enabled = ?", true).Count(&enabledUsers).Error; err != nil {
		return DashboardOnlinePayload{}, fmt.Errorf("count enabled users failed: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("enabled = ?", false).Count(&disabledUsers).Error; err != nil {
		return DashboardOnlinePayload{}, fmt.Errorf("count disabled users failed: %w", err)
	}
	if s.db.Migrator().HasColumn(&model.User{}, "last_seen_at") {
		if err := s.db.WithContext(ctx).Model(&model.User{}).
			Where("enabled = ? AND last_seen_at >= ?", true, time.Now().Add(-5*time.Minute)).
			Count(&onlineUsers).Error; err != nil {
			return DashboardOnlinePayload{}, fmt.Errorf("count online users failed: %w", err)
		}
	}

	total := maxInt64(enabledUsers+disabledUsers, 1)
	return DashboardOnlinePayload{
		CurrentSessions: onlineUsers,
		Groups: []DashboardOnlineGroupPayload{
			{Name: "管理员", Value: adminUsers, Percent: percentOf(adminUsers, total)},
			{Name: "启用用户", Value: enabledUsers, Percent: percentOf(enabledUsers, total)},
			{Name: "停用用户", Value: disabledUsers, Percent: percentOf(disabledUsers, total)},
		},
	}, nil
}

func (s *dashboardService) loadNodes(ctx context.Context) (DashboardNodesPayload, error) {
	var nodes []model.Node
	if err := s.db.WithContext(ctx).Order("is_built_in DESC, id ASC").Find(&nodes).Error; err != nil {
		return DashboardNodesPayload{}, fmt.Errorf("list nodes failed: %w", err)
	}

	items := make([]DashboardNodePayload, 0, len(nodes))
	var online int64
	for _, node := range nodes {
		if node.HealthStatus == "online" {
			online++
		}
		items = append(items, DashboardNodePayload{
			Name:      node.Name,
			Region:    node.Type,
			LatencyMS: estimateLatency(node),
			Load:      estimateLoad(node),
			Status:    node.HealthStatus,
		})
	}

	return DashboardNodesPayload{Online: online, Total: int64(len(nodes)), Items: items}, nil
}

func (s *dashboardService) loadTasks(ctx context.Context) ([]DashboardTaskPayload, error) {
	type row struct {
		QueueKey string
		Status   string
		Count    int64
	}
	var rows []row
	if err := s.db.WithContext(ctx).Model(&model.QueueJob{}).
		Select("queue_key, status, COUNT(*) AS count").
		Group("queue_key, status").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("query queue stats failed: %w", err)
	}

	type counts struct {
		success    int64
		failed     int64
		processing int64
		pending    int64
		submitted  int64
	}
	byQueue := make(map[string]counts)
	for _, row := range rows {
		current := byQueue[row.QueueKey]
		current.submitted += row.Count
		switch row.Status {
		case model.QueueJobStatusCompleted:
			current.success += row.Count
		case model.QueueJobStatusFailed:
			current.failed += row.Count
		case model.QueueJobStatusProcessing:
			current.processing += row.Count
		case model.QueueJobStatusPending:
			current.pending += row.Count
		}
		byQueue[row.QueueKey] = current
	}

	tasks := make([]DashboardTaskPayload, 0, len(queue.Definitions()))
	for _, definition := range queue.Definitions() {
		key := string(definition.Key)
		current := byQueue[key]
		progress := 100
		if current.submitted > 0 {
			progress = dashboardClampInt(int(math.Round(float64(current.success)/float64(current.submitted)*100)), 0, 100)
		}
		tasks = append(tasks, DashboardTaskPayload{
			Name:       queueDisplayName(key),
			Detail:     fmt.Sprintf("%d 等待 / %d 运行 / %d 失败", current.pending, current.processing, current.failed),
			Progress:   progress,
			Success:    current.success,
			Failed:     current.failed,
			Processing: current.processing,
			Pending:    current.pending,
			Submitted:  current.submitted,
		})
	}

	return tasks, nil
}

func percentChange(current, previous float64) float64 {
	if previous <= 0 {
		if current <= 0 {
			return 0
		}
		return 100
	}
	return round1((current - previous) / previous * 100)
}

func percentOf(value, total int64) int {
	if total <= 0 {
		return 0
	}
	return dashboardClampInt(int(math.Round(float64(value)/float64(total)*100)), 0, 100)
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func dashboardClampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func dashboardClampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func estimateLatency(node model.Node) int64 {
	return 0
}

func estimateLoad(node model.Node) int {
	return 0
}

func averageNodeLatency(nodes []DashboardNodePayload) int64 {
	var total int64
	var count int64
	for _, node := range nodes {
		if node.LatencyMS > 0 {
			total += node.LatencyMS
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return int64(math.Round(float64(total) / float64(count)))
}

func queueDisplayName(key string) string {
	switch strings.TrimSpace(key) {
	case "metadata":
		return "元数据队列"
	case "blob":
		return "Blob 处理队列"
	case "io":
		return "文件 I/O 队列"
	case "offline":
		return "离线下载队列"
	case "thumbnail":
		return "缩略图队列"
	default:
		return key
	}
}
