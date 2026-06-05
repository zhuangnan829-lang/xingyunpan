package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal HTTP 请求总数
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xingyunpan_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HttpRequestDuration HTTP 请求延迟
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "xingyunpan_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// FileUploadsTotal 文件上传总数
	FileUploadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_file_uploads_total",
			Help: "Total number of file uploads",
		},
	)

	// FileUploadBytes 文件上传字节数
	FileUploadBytes = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_file_upload_bytes_total",
			Help: "Total bytes uploaded",
		},
	)

	// ActiveUsers 活跃用户数
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_active_users",
			Help: "Number of active users",
		},
	)

	// Business Monitoring Metrics

	// TotalUsers 总用户数
	TotalUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_total_users",
			Help: "Total number of registered users",
		},
	)

	// TotalFiles 总文件数
	TotalFiles = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_total_files",
			Help: "Total number of files in the system",
		},
	)

	// TotalStorageBytes 总存储使用量（字节）
	TotalStorageBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_total_storage_bytes",
			Help: "Total storage used in bytes",
		},
	)

	// FileUploadSuccessTotal 文件上传成功总数
	FileUploadSuccessTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_file_upload_success_total",
			Help: "Total number of successful file uploads",
		},
	)

	// FileUploadFailureTotal 文件上传失败总数
	FileUploadFailureTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_file_upload_failure_total",
			Help: "Total number of failed file uploads",
		},
	)

	// ApiRequestsByEndpoint API 请求分布（按端点）
	ApiRequestsByEndpoint = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xingyunpan_api_requests_by_endpoint",
			Help: "API requests distribution by endpoint",
		},
		[]string{"endpoint"},
	)
)

// UpdateBusinessMetrics 更新业务监控指标的辅助函数
// 这些函数应该在适当的时机被调用（例如定时任务、数据变更时）

// SetTotalUsers 设置总用户数
func SetTotalUsers(count float64) {
	TotalUsers.Set(count)
}

// SetTotalFiles 设置总文件数
func SetTotalFiles(count float64) {
	TotalFiles.Set(count)
}

// SetTotalStorageBytes 设置总存储使用量
func SetTotalStorageBytes(bytes float64) {
	TotalStorageBytes.Set(bytes)
}

// RecordFileUploadSuccess 记录文件上传成功
func RecordFileUploadSuccess() {
	FileUploadSuccessTotal.Inc()
	FileUploadsTotal.Inc()
}

// RecordFileUploadFailure 记录文件上传失败
func RecordFileUploadFailure() {
	FileUploadFailureTotal.Inc()
}

// RecordApiRequest 记录 API 请求（按端点）
func RecordApiRequest(endpoint string) {
	ApiRequestsByEndpoint.WithLabelValues(endpoint).Inc()
}

// Phase 5 Metrics

var (
	// SharesCreatedTotal 创建的分享总数
	SharesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_shares_created_total",
			Help: "Total number of shares created",
		},
	)

	// ShareAccessTotal 分享访问总数
	ShareAccessTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_share_access_total",
			Help: "Total number of share accesses",
		},
	)

	// ShareDownloadsTotal 分享下载总数
	ShareDownloadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_share_downloads_total",
			Help: "Total number of share downloads",
		},
	)

	// SharePasswordVerifyFailures 分享密码验证失败次数
	SharePasswordVerifyFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_share_password_verify_failures_total",
			Help: "Total number of share password verification failures",
		},
	)

	// ActiveShares 活跃分享数（未过期）
	ActiveShares = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_active_shares",
			Help: "Number of active (non-expired) shares",
		},
	)

	// SearchQueriesTotal 搜索查询总数
	SearchQueriesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_search_queries_total",
			Help: "Total number of search queries",
		},
	)

	// SearchDuration 搜索查询延迟
	SearchDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "xingyunpan_search_duration_seconds",
			Help:    "Search query duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
		},
	)

	// SearchCacheHits 搜索缓存命中次数
	SearchCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_search_cache_hits_total",
			Help: "Total number of search cache hits",
		},
	)

	// SearchCacheMisses 搜索缓存未命中次数
	SearchCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_search_cache_misses_total",
			Help: "Total number of search cache misses",
		},
	)

	// RecycleBinItemsTotal 回收站项目总数
	RecycleBinItemsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_recycle_bin_items_total",
			Help: "Total number of items in recycle bin",
		},
	)

	// RecycleBinRestoresTotal 回收站恢复总数
	RecycleBinRestoresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_recycle_bin_restores_total",
			Help: "Total number of recycle bin restores",
		},
	)

	// RecycleBinPermanentDeletesTotal 回收站永久删除总数
	RecycleBinPermanentDeletesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_recycle_bin_permanent_deletes_total",
			Help: "Total number of permanent deletes from recycle bin",
		},
	)

	// RecycleBinCleanupTotal 回收站自动清理总数
	RecycleBinCleanupTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_recycle_bin_cleanup_total",
			Help: "Total number of items cleaned up automatically from recycle bin",
		},
	)

	// FileVersionsTotal 文件版本总数
	FileVersionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_file_versions_total",
			Help: "Total number of file versions",
		},
	)

	// VersionRestoresTotal 版本恢复总数
	VersionRestoresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_version_restores_total",
			Help: "Total number of version restores",
		},
	)

	// VersionDownloadsTotal 版本下载总数
	VersionDownloadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_version_downloads_total",
			Help: "Total number of version downloads",
		},
	)

	// CollaborationsTotal 协作关系总数
	CollaborationsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "xingyunpan_collaborations_total",
			Help: "Total number of active collaborations",
		},
	)

	// CollaboratorsAddedTotal 添加协作者总数
	CollaboratorsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_collaborators_added_total",
			Help: "Total number of collaborators added",
		},
	)

	// CollaboratorsRemovedTotal 移除协作者总数
	CollaboratorsRemovedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "xingyunpan_collaborators_removed_total",
			Help: "Total number of collaborators removed",
		},
	)

	// RateLimitTriggersTotal 限流触发总数
	RateLimitTriggersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xingyunpan_rate_limit_triggers_total",
			Help: "Total number of rate limit triggers by type",
		},
		[]string{"type"}, // type: search, file_ops, share_verify
	)
)

// Phase 5 Metrics Helper Functions

// RecordShareCreated 记录分享创建
func RecordShareCreated() {
	SharesCreatedTotal.Inc()
}

// RecordShareAccess 记录分享访问
func RecordShareAccess() {
	ShareAccessTotal.Inc()
}

// RecordShareDownload 记录分享下载
func RecordShareDownload() {
	ShareDownloadsTotal.Inc()
}

// RecordSharePasswordFailure 记录分享密码验证失败
func RecordSharePasswordFailure() {
	SharePasswordVerifyFailures.Inc()
}

// SetActiveShares 设置活跃分享数
func SetActiveShares(count float64) {
	ActiveShares.Set(count)
}

// RecordSearchQuery 记录搜索查询
func RecordSearchQuery() {
	SearchQueriesTotal.Inc()
}

// ObserveSearchDuration 记录搜索延迟
func ObserveSearchDuration(seconds float64) {
	SearchDuration.Observe(seconds)
}

// RecordSearchCacheHit 记录搜索缓存命中
func RecordSearchCacheHit() {
	SearchCacheHits.Inc()
}

// RecordSearchCacheMiss 记录搜索缓存未命中
func RecordSearchCacheMiss() {
	SearchCacheMisses.Inc()
}

// SetRecycleBinItems 设置回收站项目数
func SetRecycleBinItems(count float64) {
	RecycleBinItemsTotal.Set(count)
}

// RecordRecycleBinRestore 记录回收站恢复
func RecordRecycleBinRestore() {
	RecycleBinRestoresTotal.Inc()
}

// RecordRecycleBinPermanentDelete 记录回收站永久删除
func RecordRecycleBinPermanentDelete() {
	RecycleBinPermanentDeletesTotal.Inc()
}

// RecordRecycleBinCleanup 记录回收站自动清理
func RecordRecycleBinCleanup(count int) {
	RecycleBinCleanupTotal.Add(float64(count))
}

// SetFileVersions 设置文件版本总数
func SetFileVersions(count float64) {
	FileVersionsTotal.Set(count)
}

// RecordVersionRestore 记录版本恢复
func RecordVersionRestore() {
	VersionRestoresTotal.Inc()
}

// RecordVersionDownload 记录版本下载
func RecordVersionDownload() {
	VersionDownloadsTotal.Inc()
}

// SetCollaborations 设置协作关系总数
func SetCollaborations(count float64) {
	CollaborationsTotal.Set(count)
}

// RecordCollaboratorAdded 记录添加协作者
func RecordCollaboratorAdded() {
	CollaboratorsAddedTotal.Inc()
}

// RecordCollaboratorRemoved 记录移除协作者
func RecordCollaboratorRemoved() {
	CollaboratorsRemovedTotal.Inc()
}

// RecordRateLimitTrigger 记录限流触发
func RecordRateLimitTrigger(limitType string) {
	RateLimitTriggersTotal.WithLabelValues(limitType).Inc()
}
