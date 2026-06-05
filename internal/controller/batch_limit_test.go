package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type fakeBatchSettings struct {
	limit int
}

func (s *fakeBatchSettings) Get() (*service.FileSystemSettingPayload, error) {
	return &service.FileSystemSettingPayload{MaxBatchActionSize: s.limit}, nil
}

func (s *fakeBatchSettings) Update(payload *service.FileSystemSettingPayload) (*service.FileSystemSettingPayload, error) {
	return payload, nil
}
func (s *fakeBatchSettings) UpdateIcons(payload *service.FileSystemIconSettingsPayload) (*service.FileSystemSettingPayload, error) {
	return &service.FileSystemSettingPayload{
		FileIconRules: payload.FileIconRules,
		EmojiOptions:  payload.EmojiOptions,
	}, nil
}
func (s *fakeBatchSettings) ClearBlobURLCache() error { return nil }
func (s *fakeBatchSettings) GetBrowserApps() ([]service.BrowserAppGroupPayload, error) {
	return nil, nil
}
func (s *fakeBatchSettings) ResolveBrowserApp(filename, mimeType string, platform service.BrowserAppPlatform) (*service.BrowserAppResolvedPayload, error) {
	return nil, nil
}

type fakeBatchRecycleService struct {
	moveCalls    int
	restoreCalls int
	deleteCalls  int
}

func (s *fakeBatchRecycleService) MoveToRecycle(ctx context.Context, userID uint, fileIDs []uint) error {
	s.moveCalls++
	return nil
}
func (s *fakeBatchRecycleService) GetRecycleList(ctx context.Context, userID uint, query service.RecycleListQuery) (*service.RecycleListResponse, error) {
	return nil, nil
}
func (s *fakeBatchRecycleService) RestoreFiles(ctx context.Context, userID uint, itemIDs []uint) error {
	s.restoreCalls++
	return nil
}
func (s *fakeBatchRecycleService) PermanentDelete(ctx context.Context, userID uint, itemIDs []uint) error {
	s.deleteCalls++
	return nil
}
func (s *fakeBatchRecycleService) EmptyRecycleBin(ctx context.Context, userID uint) error {
	return nil
}

type fakeBatchAdminShareService struct {
	batchDeleteCalls int
}

func (s *fakeBatchAdminShareService) List(ctx context.Context, query *service.AdminShareListQuery) ([]service.AdminSharePayload, int64, error) {
	return nil, 0, nil
}
func (s *fakeBatchAdminShareService) Metrics(ctx context.Context, expiringWithinDays int) (*service.AdminShareMetricsPayload, error) {
	return nil, nil
}
func (s *fakeBatchAdminShareService) Delete(ctx context.Context, shareID uint) error { return nil }
func (s *fakeBatchAdminShareService) BatchDelete(ctx context.Context, shareIDs []uint) error {
	s.batchDeleteCalls++
	return nil
}

type fakeBatchQueueJobService struct {
	batchDeleteCalls int
}

func (s *fakeBatchQueueJobService) List(queueKey, status string, nodeID uint, page, pageSize int, cursor uint, useCursor bool) (*service.QueueJobListPayload, error) {
	return nil, nil
}
func (s *fakeBatchQueueJobService) Get(id uint) (*service.QueueJobListItemPayload, error) {
	return nil, nil
}
func (s *fakeBatchQueueJobService) Retry(id uint) (*service.QueueJobListItemPayload, error) {
	return nil, nil
}
func (s *fakeBatchQueueJobService) RecoverStale(queueKey string) (*service.QueueJobStaleRecoveryPayload, error) {
	return nil, nil
}
func (s *fakeBatchQueueJobService) Delete(id uint) (*service.QueueJobMutationPayload, error) {
	return nil, nil
}
func (s *fakeBatchQueueJobService) BatchDelete(payload *service.QueueJobBatchDeletePayload) (*service.QueueJobMutationPayload, error) {
	s.batchDeleteCalls++
	return &service.QueueJobMutationPayload{Deleted: int64(len(payload.JobIDs))}, nil
}
func (s *fakeBatchQueueJobService) Clear(payload *service.QueueJobClearPayload) (*service.QueueJobMutationPayload, error) {
	return nil, nil
}

type fakeBatchOfflineDownloadService struct {
	batchDeleteCalls int
}

func (s *fakeBatchOfflineDownloadService) List(ctx context.Context, userID uint, status string, keyword string) ([]service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}
func (s *fakeBatchOfflineDownloadService) Create(ctx context.Context, userID uint, req service.OfflineDownloadCreateRequest) (*service.OfflineDownloadTaskPayload, error) {
	return &service.OfflineDownloadTaskPayload{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}
func (s *fakeBatchOfflineDownloadService) Refresh(ctx context.Context, userID uint) ([]service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}
func (s *fakeBatchOfflineDownloadService) Pause(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}
func (s *fakeBatchOfflineDownloadService) Resume(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}
func (s *fakeBatchOfflineDownloadService) Retry(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}
func (s *fakeBatchOfflineDownloadService) ExecuteOfflineDownload(ctx context.Context, taskID uint) (string, error) {
	return "", nil
}
func (s *fakeBatchOfflineDownloadService) Delete(ctx context.Context, userID uint, id uint) error {
	return nil
}
func (s *fakeBatchOfflineDownloadService) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	s.batchDeleteCalls++
	return int64(len(ids)), nil
}

func TestMaxBatchActionSizeBlocksMultipleBatchControllers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settings := &fakeBatchSettings{limit: 2}
	recycleSvc := &fakeBatchRecycleService{}
	shareSvc := &fakeBatchAdminShareService{}
	queueSvc := &fakeBatchQueueJobService{}
	offlineSvc := &fakeBatchOfflineDownloadService{}

	recycleController := NewRecycleController(recycleSvc, settings)
	shareController := NewAdminShareController(shareSvc, settings)
	queueController := NewQueueController(nil, nil, queueSvc, settings)
	offlineController := NewOfflineDownloadController(offlineSvc, settings)

	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(42))
		ctx.Next()
	})
	router.POST("/recycle", recycleController.MoveToRecycle)
	router.POST("/recycle/restore", recycleController.RestoreFiles)
	router.DELETE("/recycle", recycleController.PermanentDelete)
	router.POST("/admin/shares/batch-delete", shareController.BatchDelete)
	router.POST("/admin/queue-jobs/batch-delete", queueController.BatchDeleteJobs)
	router.POST("/offline-downloads/batch-delete", offlineController.BatchDelete)

	cases := []struct {
		name   string
		method string
		path   string
		body   map[string]interface{}
	}{
		{name: "file recycle delete", method: http.MethodPost, path: "/recycle", body: map[string]interface{}{"file_ids": []uint{1, 2, 3}}},
		{name: "recycle restore", method: http.MethodPost, path: "/recycle/restore", body: map[string]interface{}{"item_ids": []uint{1, 2, 3}}},
		{name: "share batch delete", method: http.MethodPost, path: "/admin/shares/batch-delete", body: map[string]interface{}{"share_ids": []uint{1, 2, 3}}},
		{name: "queue job batch delete", method: http.MethodPost, path: "/admin/queue-jobs/batch-delete", body: map[string]interface{}{"job_ids": []uint{1, 2, 3}}},
		{name: "offline batch delete", method: http.MethodPost, path: "/offline-downloads/batch-delete", body: map[string]interface{}{"ids": []uint{1, 2, 3}}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp := batchJSON(t, router, tt.method, tt.path, tt.body)
			if resp.Code != response.CodeInvalidParams || !strings.Contains(resp.Message, "最多允许 2 项") {
				t.Fatalf("response = code:%d message:%q, want max batch error", resp.Code, resp.Message)
			}
		})
	}

	if recycleSvc.moveCalls != 0 || recycleSvc.restoreCalls != 0 || recycleSvc.deleteCalls != 0 {
		t.Fatalf("recycle service was called despite limit: %#v", recycleSvc)
	}
	if shareSvc.batchDeleteCalls != 0 || queueSvc.batchDeleteCalls != 0 || offlineSvc.batchDeleteCalls != 0 {
		t.Fatalf("batch services were called despite limit: share=%d queue=%d offline=%d", shareSvc.batchDeleteCalls, queueSvc.batchDeleteCalls, offlineSvc.batchDeleteCalls)
	}
}

func batchJSON(t *testing.T, router *gin.Engine, method, path string, body interface{}) response.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response HTTP %d body=%q: %v", rec.Code, rec.Body.String(), err)
	}
	return resp
}
