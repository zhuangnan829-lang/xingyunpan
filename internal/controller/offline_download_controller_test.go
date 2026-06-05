package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"xingyunpan-v2/internal/service"
)

type fakeOfflineDownloadService struct {
	created service.OfflineDownloadCreateRequest
}

func (s *fakeOfflineDownloadService) List(ctx context.Context, userID uint, status string, keyword string) ([]service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}

func (s *fakeOfflineDownloadService) Create(ctx context.Context, userID uint, req service.OfflineDownloadCreateRequest) (*service.OfflineDownloadTaskPayload, error) {
	s.created = req
	return &service.OfflineDownloadTaskPayload{
		ID:        99,
		TaskToken: "task-token",
		Name:      req.Name,
		URL:       req.URL,
		SavePath:  req.SavePath,
		Status:    "queued",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DispatchNode: &service.SelectedNodePayload{
			ID:         7,
			Name:       "aria2-node",
			Type:       "worker",
			Downloader: "Aria2",
			RPCURL:     "http://aria2/jsonrpc",
		},
	}, nil
}

func (s *fakeOfflineDownloadService) Refresh(ctx context.Context, userID uint) ([]service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}

func (s *fakeOfflineDownloadService) Pause(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}

func (s *fakeOfflineDownloadService) Resume(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}

func (s *fakeOfflineDownloadService) Retry(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error) {
	return nil, nil
}

func (s *fakeOfflineDownloadService) ExecuteOfflineDownload(ctx context.Context, taskID uint) (string, error) {
	return "", nil
}

func (s *fakeOfflineDownloadService) Delete(ctx context.Context, userID uint, id uint) error {
	return nil
}

func (s *fakeOfflineDownloadService) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	return 0, nil
}

func TestOfflineDownloadCreateEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fakeService := &fakeOfflineDownloadService{}
	controller := NewOfflineDownloadController(fakeService)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(42))
		ctx.Next()
	})
	router.POST("/offline-downloads", controller.Create)

	body := bytes.NewBufferString(`{"url":"https://example.com/file.iso","name":"file.iso","save_path":"/downloads"}`)
	req := httptest.NewRequest(http.MethodPost, "/offline-downloads", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if fakeService.created.URL != "https://example.com/file.iso" || fakeService.created.SavePath != "/downloads" {
		t.Fatalf("request was not passed to service: %#v", fakeService.created)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["code"].(float64) != 200 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}
