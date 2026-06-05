package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

func TestQueueSettingsControllerGetAndPut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newQueueSettingsControllerDB(t)
	repo := repository.NewQueueSettingRepository(db)
	settings := service.NewQueueSettingService(repo)
	controller := NewQueueController(settings, nil, nil)

	router := gin.New()
	router.GET("/api/v1/admin/queue-settings", controller.GetSettings)
	router.PUT("/api/v1/admin/queue-settings", controller.UpdateSettings)

	getResp := requestQueueSettings(t, router, http.MethodGet, "/api/v1/admin/queue-settings", nil)
	if getResp.Code != response.CodeSuccess {
		t.Fatalf("GET response code = %d message=%s", getResp.Code, getResp.Message)
	}
	payload := decodeQueueSettingsPayload(t, getResp.Data)
	if len(payload) == 0 {
		t.Fatalf("GET returned empty queue settings")
	}

	for i := range payload {
		if payload[i].QueueKey == "offline" {
			payload[i].WorkerNum = 9
		}
	}

	putResp := requestQueueSettings(t, router, http.MethodPut, "/api/v1/admin/queue-settings", payload)
	if putResp.Code != response.CodeSuccess {
		t.Fatalf("PUT response code = %d message=%s", putResp.Code, putResp.Message)
	}

	var saved model.QueueSetting
	if err := db.Where("queue_key = ?", "offline").First(&saved).Error; err != nil {
		t.Fatalf("query saved offline setting: %v", err)
	}
	if saved.WorkerNum != 9 {
		t.Fatalf("offline worker_num = %d, want 9", saved.WorkerNum)
	}
}

func TestQueueRuntimeControllerReturnsStatusWithoutRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewQueueController(nil, nil, nil)
	controller.SetRuntimeService(service.NewQueueRuntimeService(nil))

	router := gin.New()
	router.GET("/api/v1/admin/queue-runtime", controller.GetRuntime)

	resp := requestQueueSettings(t, router, http.MethodGet, "/api/v1/admin/queue-runtime", nil)
	if resp.Code != response.CodeSuccess {
		t.Fatalf("GET runtime response code = %d message=%s", resp.Code, resp.Message)
	}

	raw, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal runtime response data: %v", err)
	}
	var payload service.QueueRuntimeStatusPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode runtime response data: %v", err)
	}
	if payload.HeartbeatAvailable {
		t.Fatalf("expected heartbeat unavailable without redis")
	}
	if payload.Message == "" {
		t.Fatalf("expected runtime status message")
	}
}

func newQueueSettingsControllerDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "queue-settings-controller.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sqlite sql db: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close sqlite db: %v", err)
		}
	})

	return db
}

func requestQueueSettings(t *testing.T, router *gin.Engine, method, path string, body interface{}) response.Response {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s %s HTTP status = %d body=%s", method, path, rec.Code, rec.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func decodeQueueSettingsPayload(t *testing.T, data interface{}) []service.QueueSettingItemPayload {
	t.Helper()

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var payload []service.QueueSettingItemPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode queue settings payload: %v", err)
	}
	return payload
}
