package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type fakeFileSystemSettings struct {
	payload *FileSystemSettingPayload
}

func (s *fakeFileSystemSettings) Get() (*FileSystemSettingPayload, error) {
	if s.payload == nil {
		return defaultFileSystemSettingPayload(), nil
	}
	return s.payload, nil
}

func (s *fakeFileSystemSettings) Update(payload *FileSystemSettingPayload) (*FileSystemSettingPayload, error) {
	s.payload = payload
	return payload, nil
}
func (s *fakeFileSystemSettings) UpdateIcons(payload *FileSystemIconSettingsPayload) (*FileSystemSettingPayload, error) {
	current, _ := s.Get()
	current.FileIconRules = payload.FileIconRules
	current.EmojiOptions = payload.EmojiOptions
	s.payload = current
	return current, nil
}

func (s *fakeFileSystemSettings) ClearBlobURLCache() error { return nil }
func (s *fakeFileSystemSettings) GetBrowserApps() ([]BrowserAppGroupPayload, error) {
	return nil, nil
}
func (s *fakeFileSystemSettings) ResolveBrowserApp(filename, mimeType string, platform BrowserAppPlatform) (*BrowserAppResolvedPayload, error) {
	return nil, nil
}

type fakeOfflineTaskRepository struct {
	mu     sync.Mutex
	nextID uint
	tasks  map[uint]model.OfflineDownloadTask
}

func newFakeOfflineTaskRepository() *fakeOfflineTaskRepository {
	return &fakeOfflineTaskRepository{nextID: 1, tasks: map[uint]model.OfflineDownloadTask{}}
}

func (r *fakeOfflineTaskRepository) EnsureSchema() error { return nil }

func (r *fakeOfflineTaskRepository) ListByUser(ctx context.Context, filter repository.OfflineDownloadTaskListFilter) ([]model.OfflineDownloadTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []model.OfflineDownloadTask
	for _, task := range r.tasks {
		if task.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		result = append(result, task)
	}
	return result, nil
}

func (r *fakeOfflineTaskRepository) GetByIDForUser(ctx context.Context, userID uint, id uint) (*model.OfflineDownloadTask, error) {
	task, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, fmt.Errorf("offline download task not found")
	}
	return task, nil
}

func (r *fakeOfflineTaskRepository) GetByID(ctx context.Context, id uint) (*model.OfflineDownloadTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("offline download task not found")
	}
	return &task, nil
}

func (r *fakeOfflineTaskRepository) Create(ctx context.Context, task *model.OfflineDownloadTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.ID = r.nextID
	r.nextID++
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	r.tasks[task.ID] = *task
	return nil
}

func (r *fakeOfflineTaskRepository) Save(ctx context.Context, task *model.OfflineDownloadTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = *task
	return nil
}

func (r *fakeOfflineTaskRepository) Delete(ctx context.Context, userID uint, id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, id)
	return nil
}

func (r *fakeOfflineTaskRepository) BatchDelete(ctx context.Context, userID uint, ids []uint) (int64, error) {
	for _, id := range ids {
		_ = r.Delete(ctx, userID, id)
	}
	return int64(len(ids)), nil
}

func TestOfflineDownloadCreateSubmitsAria2WithNodeSnapshot(t *testing.T) {
	addURIRequests := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode aria2 request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch req["method"] {
		case "aria2.addUri":
			addURIRequests <- req
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":"gid-1"}`))
		case "aria2.tellStatus":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":{"status":"complete","completedLength":"4096","totalLength":"4096","downloadSpeed":"0"}}`))
		default:
			t.Errorf("unexpected aria2 method: %v", req["method"])
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":null}`))
		}
	}))
	defer server.Close()

	nodeID := uint(7)
	nodeRepo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:              model.BaseModel{ID: nodeID},
		Name:                   "aria2-node",
		Type:                   "worker",
		Enabled:                true,
		Weight:                 1,
		OfflineDownload:        true,
		OfflineDownloader:      "Aria2",
		OfflineRPCURL:          server.URL,
		OfflineRPCSecret:       "secret-token",
		OfflineTaskOptions:     `{"max-connection-per-server":"8"}`,
		OfflineTempDir:         "/tmp/offline",
		OfflineRefreshInterval: 3,
		OfflineWaitForSeeding:  true,
	}}}
	repo := newFakeOfflineTaskRepository()
	svc := NewOfflineDownloadService(repo, nil, nil, nil, NewNodeDispatchService(nodeRepo))

	payload, err := svc.Create(context.Background(), 42, OfflineDownloadCreateRequest{URL: "https://example.com/file.iso", Name: "file.iso"})
	if err != nil {
		t.Fatalf("create offline download: %v", err)
	}
	if payload.DispatchNode == nil || payload.DispatchNode.ID != nodeID {
		t.Fatalf("expected dispatch node %d, got %#v", nodeID, payload.DispatchNode)
	}

	var addReq map[string]interface{}
	select {
	case addReq = <-addURIRequests:
	case <-time.After(2 * time.Second):
		t.Fatalf("aria2.addUri was not called")
	}

	params := addReq["params"].([]interface{})
	if params[0] != "token:secret-token" {
		t.Fatalf("expected aria2 token auth, got %#v", params[0])
	}
	options := params[2].(map[string]interface{})
	if options["max-connection-per-server"] != "8" {
		t.Fatalf("task_options were not forwarded: %#v", options)
	}
	if options["dir"] != "/tmp/offline" {
		t.Fatalf("temp_dir was not forwarded as dir: %#v", options)
	}

	task := waitOfflineTask(t, repo, payload.ID, func(task *model.OfflineDownloadTask) bool {
		return task.RemoteTaskID == "gid-1"
	})
	if task.NodeID == nil || *task.NodeID != nodeID || task.RPCURL != server.URL || task.RPCSecret != "secret-token" {
		t.Fatalf("node runtime snapshot not saved: %#v", task)
	}
	if task.TaskOptions != `{"max-connection-per-server":"8"}` || task.TempDir != "/tmp/offline" || task.RefreshInterval != 3 || !task.WaitForSeeding {
		t.Fatalf("runtime options snapshot not saved: %#v", task)
	}
	if task.RemoteTaskID != "gid-1" {
		t.Fatalf("remote task id not saved: %q", task.RemoteTaskID)
	}
}

func TestOfflineDownloadSubmitsHTTPAndMagnetToAria2(t *testing.T) {
	addURIRequests := make(chan map[string]interface{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode aria2 request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch req["method"] {
		case "aria2.addUri":
			addURIRequests <- req
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":"gid-submit"}`))
		case "aria2.tellStatus":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":{"status":"active","completedLength":"0","totalLength":"10","downloadSpeed":"0"}}`))
		default:
			t.Errorf("unexpected aria2 method: %v", req["method"])
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":null}`))
		}
	}))
	defer server.Close()

	nodeRepo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:              model.BaseModel{ID: 8},
		Name:                   "aria2-submit",
		Type:                   "worker",
		Enabled:                true,
		Weight:                 1,
		OfflineDownload:        true,
		OfflineDownloader:      "Aria2",
		OfflineRPCURL:          server.URL,
		OfflineTaskOptions:     `{"max-connection-per-server":"16"}`,
		OfflineTempDir:         "/tmp/changed",
		OfflineRefreshInterval: 60,
	}}}
	svc := NewOfflineDownloadService(newFakeOfflineTaskRepository(), nil, nil, nil, NewNodeDispatchService(nodeRepo))

	if _, err := svc.Create(context.Background(), 42, OfflineDownloadCreateRequest{URL: "https://example.com/file.iso"}); err != nil {
		t.Fatalf("create HTTP offline download: %v", err)
	}
	if _, err := svc.Create(context.Background(), 42, OfflineDownloadCreateRequest{URL: "magnet:?xt=urn:btih:abcdef"}); err != nil {
		t.Fatalf("create magnet offline download: %v", err)
	}

	submittedURLs := map[string]bool{}
	for i := 0; i < 2; i++ {
		var addReq map[string]interface{}
		select {
		case addReq = <-addURIRequests:
		case <-time.After(2 * time.Second):
			t.Fatalf("aria2.addUri call %d was not received", i+1)
		}
		params := addReq["params"].([]interface{})
		urls := params[0].([]interface{})
		submittedURLs[urls[0].(string)] = true
		options := params[1].(map[string]interface{})
		if options["max-connection-per-server"] != "16" {
			t.Fatalf("updated task_options were not submitted to aria2: %#v", options)
		}
		if options["dir"] != "/tmp/changed" {
			t.Fatalf("updated temp_dir was not submitted to aria2: %#v", options)
		}
	}
	if !submittedURLs["https://example.com/file.iso"] || !submittedURLs["magnet:?xt=urn:btih:abcdef"] {
		t.Fatalf("HTTP and magnet URLs were not both submitted: %#v", submittedURLs)
	}
}

func TestAria2RPCSecretFailureAndSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		params := req["params"].([]interface{})
		if len(params) == 0 || params[0] != "token:correct-secret" {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","error":{"code":1,"message":"unauthorized"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":"gid-secret"}`))
	}))
	defer server.Close()

	downloader := &aria2Downloader{client: server.Client()}
	wrong := &modelOfflineRuntime{
		URL:       "https://example.com/file.iso",
		RPCURL:    server.URL,
		RPCSecret: "wrong-secret",
	}
	if _, err := downloader.Submit(context.Background(), wrong); err == nil {
		t.Fatalf("expected wrong rpc_secret to fail")
	}

	correct := &modelOfflineRuntime{
		URL:       "https://example.com/file.iso",
		RPCURL:    server.URL,
		RPCSecret: "correct-secret",
	}
	gid, err := downloader.Submit(context.Background(), correct)
	if err != nil {
		t.Fatalf("expected correct rpc_secret to succeed: %v", err)
	}
	if gid != "gid-secret" {
		t.Fatalf("unexpected gid: %q", gid)
	}
}

func waitOfflineTask(t *testing.T, repo *fakeOfflineTaskRepository, id uint, predicate func(*model.OfflineDownloadTask) bool) *model.OfflineDownloadTask {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		task, err := repo.GetByID(context.Background(), id)
		if err == nil && predicate(task) {
			return task
		}
		time.Sleep(10 * time.Millisecond)
	}
	task, _ := repo.GetByID(context.Background(), id)
	t.Fatalf("task did not reach expected state: %#v", task)
	return nil
}

func TestOfflineDownloadCreateReturnsErrorWithoutEligibleNode(t *testing.T) {
	nodeRepo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:       model.BaseModel{ID: 1},
		Name:            "disabled",
		Type:            "worker",
		Enabled:         false,
		OfflineDownload: true,
		Weight:          1,
	}}}
	svc := NewOfflineDownloadService(newFakeOfflineTaskRepository(), nil, nil, nil, NewNodeDispatchService(nodeRepo))

	_, err := svc.Create(context.Background(), 42, OfflineDownloadCreateRequest{URL: "https://example.com/file.iso"})
	if err == nil {
		t.Fatalf("expected no eligible node error")
	}
}

func TestOfflineDownloadTTLReflectsUpdatedSettings(t *testing.T) {
	repo := newFakeOfflineTaskRepository()
	createdAt := time.Now().Add(-2 * time.Hour)
	task := model.OfflineDownloadTask{
		BaseModel: model.BaseModel{
			ID:        1,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		UserID:    42,
		TaskToken: "old-task",
		Name:      "old.iso",
		SourceURL: "https://example.com/old.iso",
		SavePath:  "/offline-downloads",
		Status:    model.OfflineTaskStatusQueued,
		SpeedText: "waiting",
		SizeText:  "unknown",
	}
	repo.tasks[task.ID] = task
	repo.nextID = 2

	settings := &fakeFileSystemSettings{payload: &FileSystemSettingPayload{OfflineTTL: 3600}}
	svc := NewOfflineDownloadServiceWithSettings(repo, nil, nil, nil, settings)

	items, err := svc.Refresh(context.Background(), 42)
	if err != nil {
		t.Fatalf("refresh expired tasks: %v", err)
	}
	if len(items) != 1 || !items[0].Expired || items[0].TTLSeconds != 3600 || items[0].ExpiresAt == nil {
		t.Fatalf("expected expired payload with ttl, got %#v", items)
	}
	stored, _ := repo.GetByID(context.Background(), task.ID)
	if stored.Status != model.OfflineTaskStatusExpired {
		t.Fatalf("expired task status = %q, want %q", stored.Status, model.OfflineTaskStatusExpired)
	}

	stored.Status = model.OfflineTaskStatusQueued
	stored.ErrorMessage = ""
	if err := repo.Save(context.Background(), stored); err != nil {
		t.Fatalf("reset task: %v", err)
	}
	settings.payload.OfflineTTL = 0

	items, err = svc.Refresh(context.Background(), 42)
	if err != nil {
		t.Fatalf("refresh no-ttl tasks: %v", err)
	}
	if len(items) != 1 || items[0].Expired || items[0].TTLSeconds != 0 || items[0].ExpiresAt != nil {
		t.Fatalf("expected non-expiring payload after ttl=0, got %#v", items)
	}
}

func TestOfflineDownloadCreateRejectsUnsupportedDownloader(t *testing.T) {
	nodeRepo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:         model.BaseModel{ID: 1},
		Name:              "qb",
		Type:              "worker",
		Enabled:           true,
		Weight:            1,
		OfflineDownload:   true,
		OfflineDownloader: "qBittorrent",
	}}}
	svc := NewOfflineDownloadService(newFakeOfflineTaskRepository(), nil, nil, nil, NewNodeDispatchService(nodeRepo))

	_, err := svc.Create(context.Background(), 42, OfflineDownloadCreateRequest{URL: "magnet:?xt=urn:btih:test"})
	if err == nil {
		t.Fatalf("expected unsupported downloader error")
	}
}

func TestOfflineDownloadRefreshIntervalControlsPolling(t *testing.T) {
	statusCalls := make(chan time.Time, 3)
	var mu sync.Mutex
	statusCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		switch req["method"] {
		case "aria2.addUri":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":"gid-interval"}`))
		case "aria2.tellStatus":
			mu.Lock()
			statusCount++
			currentCount := statusCount
			mu.Unlock()
			statusCalls <- time.Now()
			if currentCount >= 2 {
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":{"status":"complete","completedLength":"10","totalLength":"10","downloadSpeed":"0"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":{"status":"active","completedLength":"1","totalLength":"10","downloadSpeed":"1"}}`))
		}
	}))
	defer server.Close()

	nodeRepo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:              model.BaseModel{ID: 2},
		Name:                   "aria2-node",
		Type:                   "worker",
		Enabled:                true,
		Weight:                 1,
		OfflineDownload:        true,
		OfflineDownloader:      "Aria2",
		OfflineRPCURL:          server.URL,
		OfflineRefreshInterval: 1,
	}}}
	svc := NewOfflineDownloadService(newFakeOfflineTaskRepository(), nil, nil, nil, NewNodeDispatchService(nodeRepo))

	if _, err := svc.Create(context.Background(), 42, OfflineDownloadCreateRequest{URL: "https://example.com/file.iso"}); err != nil {
		t.Fatalf("create offline download: %v", err)
	}
	first := <-statusCalls
	second := <-statusCalls
	if second.Sub(first) < 900*time.Millisecond {
		t.Fatalf("refresh_interval was not used, got interval %s", second.Sub(first))
	}
}

func TestAria2WaitForSeedingKeepsMagnetActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":{"status":"active","completedLength":"10","totalLength":"10","downloadSpeed":"0","uploadSpeed":"1"}}`))
	}))
	defer server.Close()

	status, err := (&aria2Downloader{client: server.Client()}).Status(context.Background(), &modelOfflineRuntime{
		URL:            "magnet:?xt=urn:btih:test",
		RPCURL:         server.URL,
		RemoteTaskID:   "gid",
		WaitForSeeding: true,
	})
	if err != nil {
		t.Fatalf("query aria2 status: %v", err)
	}
	if status.Completed {
		t.Fatalf("magnet task was completed while still active and waiting for seeding")
	}
	if !status.WaitingSeeding {
		t.Fatalf("expected waiting seeding state")
	}
}

func TestAria2WithoutWaitForSeedingCompletesMagnetWhenDownloaded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"xingyunpan","result":{"status":"active","completedLength":"10","totalLength":"10","downloadSpeed":"0","uploadSpeed":"1"}}`))
	}))
	defer server.Close()

	status, err := (&aria2Downloader{client: server.Client()}).Status(context.Background(), &modelOfflineRuntime{
		URL:            "magnet:?xt=urn:btih:test",
		RPCURL:         server.URL,
		RemoteTaskID:   "gid",
		WaitForSeeding: false,
	})
	if err != nil {
		t.Fatalf("query aria2 status: %v", err)
	}
	if !status.Completed {
		t.Fatalf("magnet task should complete when wait_for_seeding is disabled and bytes are complete")
	}
	if status.WaitingSeeding {
		t.Fatalf("did not expect waiting seeding state when wait_for_seeding is disabled")
	}
}
