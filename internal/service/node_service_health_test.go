package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xingyunpan-v2/internal/model"
)

func TestNodeHealthOfflineWhenRPCUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rpcURL := server.URL
	server.Close()

	repo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:         model.BaseModel{ID: 1},
		Name:              "aria2-node",
		Type:              "worker",
		Enabled:           true,
		Weight:            1,
		OfflineDownload:   true,
		OfflineDownloader: "Aria2",
		OfflineRPCURL:     rpcURL,
	}}}
	payload, err := NewNodeService(repo).CheckHealth(1)
	if err != nil {
		t.Fatalf("check health: %v", err)
	}
	if payload.Health.Status != "offline" {
		t.Fatalf("expected offline, got %#v", payload.Health)
	}
	if !strings.Contains(payload.Health.Message, "offline_download: failed") {
		t.Fatalf("expected rpc failure in message, got %q", payload.Health.Message)
	}
	if payload.Health.LastCheckedAt == nil {
		t.Fatalf("last_checked_at was not updated")
	}
	if payload.Health.LastHeartbeatAt != nil {
		t.Fatalf("last_heartbeat_at should not update on failed health check")
	}
}

func TestNodeHealthDisabledNode(t *testing.T) {
	repo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:       model.BaseModel{ID: 1},
		Name:            "disabled",
		Type:            "master",
		Enabled:         false,
		CreateArchive:   true,
		ExtractArchive:  true,
		OfflineDownload: true,
	}}}
	payload, err := NewNodeService(repo).CheckHealth(1)
	if err != nil {
		t.Fatalf("check health: %v", err)
	}
	if payload.Health.Status != "disabled" {
		t.Fatalf("expected disabled, got %#v", payload.Health)
	}
	if !strings.Contains(payload.Health.Message, "enabled=false") {
		t.Fatalf("expected disabled reason, got %q", payload.Health.Message)
	}
	if payload.Health.LastCheckedAt == nil {
		t.Fatalf("last_checked_at was not updated")
	}
}

func TestNodeHealthCapabilityFailureHasReason(t *testing.T) {
	repo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:      model.BaseModel{ID: 1},
		Name:           "worker",
		Type:           "worker",
		Enabled:        true,
		Weight:         1,
		CreateArchive:  true,
		ExtractArchive: true,
	}}}
	payload, err := NewNodeService(repo).CheckHealth(1)
	if err != nil {
		t.Fatalf("check health: %v", err)
	}
	if payload.Health.Status != "offline" {
		t.Fatalf("expected offline for failed capability checks, got %#v", payload.Health)
	}
	if !strings.Contains(payload.Health.Message, "create_archive: failed") || !strings.Contains(payload.Health.Message, "extract_archive: failed") {
		t.Fatalf("expected archive capability reasons, got %q", payload.Health.Message)
	}
}

func TestNodeHealthOnlineWhenRPCRecovers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req["method"] != "aria2.getVersion" {
			t.Fatalf("unexpected method: %#v", req["method"])
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"node-connectivity","result":{"version":"1.37.0"}}`))
	}))
	defer server.Close()

	repo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:         model.BaseModel{ID: 1},
		Name:              "aria2-node",
		Type:              "worker",
		Enabled:           true,
		Weight:            1,
		OfflineDownload:   true,
		OfflineDownloader: "Aria2",
		OfflineRPCURL:     server.URL,
	}}}
	payload, err := NewNodeService(repo).CheckHealth(1)
	if err != nil {
		t.Fatalf("check health: %v", err)
	}
	if payload.Health.Status != "online" {
		t.Fatalf("expected online, got %#v", payload.Health)
	}
	if !strings.Contains(payload.Health.Message, "offline_download: ok") {
		t.Fatalf("expected rpc success in message, got %q", payload.Health.Message)
	}
	if payload.Health.LastHeartbeatAt == nil || payload.Health.LastCheckedAt == nil {
		t.Fatalf("heartbeat and checked timestamps should be updated on healthy node")
	}
}

func TestNodeHealthFailsWithWrongRPCAndSucceedsAfterFix(t *testing.T) {
	closedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	wrongRPCURL := closedServer.URL
	closedServer.Close()

	workingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"node-connectivity","result":{"version":"1.37.0"}}`))
	}))
	defer workingServer.Close()

	repo := &fakeNodeRepository{nodes: []model.Node{{
		BaseModel:         model.BaseModel{ID: 1},
		Name:              "aria2-node",
		Type:              "master",
		Enabled:           true,
		Weight:            1,
		OfflineDownload:   true,
		OfflineDownloader: "Aria2",
		OfflineRPCURL:     wrongRPCURL,
	}}}
	nodeService := NewNodeService(repo)

	failed, err := nodeService.CheckHealth(1)
	if err != nil {
		t.Fatalf("check wrong rpc health: %v", err)
	}
	if failed.Health.Status != "offline" {
		t.Fatalf("expected offline with wrong rpc url, got %#v", failed.Health)
	}
	if !strings.Contains(failed.Health.Message, "offline_download: failed") {
		t.Fatalf("expected failed rpc reason, got %q", failed.Health.Message)
	}

	repo.nodes[0].OfflineRPCURL = workingServer.URL
	recovered, err := nodeService.CheckHealth(1)
	if err != nil {
		t.Fatalf("check fixed rpc health: %v", err)
	}
	if recovered.Health.Status != "online" {
		t.Fatalf("expected online after fixing rpc url, got %#v", recovered.Health)
	}
	if !strings.Contains(recovered.Health.Message, "offline_download: ok") {
		t.Fatalf("expected successful rpc result, got %q", recovered.Health.Message)
	}
	if recovered.Health.LastHeartbeatAt == nil || recovered.Health.LastCheckedAt == nil {
		t.Fatalf("heartbeat and checked timestamps should be updated after recovery")
	}
}
