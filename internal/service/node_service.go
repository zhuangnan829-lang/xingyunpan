package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// NodeFeaturePayload is the API-facing node feature shape.
type NodeFeaturePayload struct {
	CreateArchive   bool `json:"create_archive"`
	ExtractArchive  bool `json:"extract_archive"`
	OfflineDownload bool `json:"offline_download"`
}

// NodeOfflinePayload is the API-facing node offline-download shape.
type NodeOfflinePayload struct {
	Downloader      string `json:"downloader"`
	RPCURL          string `json:"rpc_url"`
	RPCSecret       string `json:"rpc_secret"`
	TaskOptions     string `json:"task_options"`
	TempDir         string `json:"temp_dir"`
	RefreshInterval int    `json:"refresh_interval"`
	WaitForSeeding  bool   `json:"wait_for_seeding"`
}

// NodePayload is the API-facing node shape.
type NodePayload struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Enabled   bool               `json:"enabled"`
	Weight    int                `json:"weight"`
	IsBuiltIn bool               `json:"is_built_in"`
	Health    NodeHealthPayload  `json:"health"`
	Features  NodeFeaturePayload `json:"features"`
	Offline   NodeOfflinePayload `json:"offline"`
}

// NodeHealthPayload is the API-facing node health shape.
type NodeHealthPayload struct {
	Status          string  `json:"status"`
	Message         string  `json:"message"`
	LastHeartbeatAt *string `json:"last_heartbeat_at"`
	LastCheckedAt   *string `json:"last_checked_at"`
}

// NodeService provides admin CRUD for nodes.
type NodeService interface {
	List() ([]NodePayload, error)
	Get(id uint) (*NodePayload, error)
	Create(payload *NodePayload) (*NodePayload, error)
	Update(id uint, payload *NodePayload) (*NodePayload, error)
	Delete(id uint) error
	CheckHealth(id uint) (*NodePayload, error)
}

type nodeService struct {
	repo repository.NodeRepository
}

// NewNodeService creates a node service.
func NewNodeService(repo repository.NodeRepository) NodeService {
	return &nodeService{repo: repo}
}

func (s *nodeService) List() ([]NodePayload, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		seed := defaultNodePayload()
		created, err := s.Create(&seed)
		if err != nil {
			return nil, err
		}
		return []NodePayload{*created}, nil
	}

	result := make([]NodePayload, 0, len(items))
	for _, item := range items {
		result = append(result, toNodePayload(&item))
	}
	return result, nil
}

func (s *nodeService) Get(id uint) (*NodePayload, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("node not found")
	}
	payload := toNodePayload(item)
	return &payload, nil
}

func (s *nodeService) Create(payload *NodePayload) (*NodePayload, error) {
	normalized, err := normalizeNodePayload(payload, false)
	if err != nil {
		return nil, err
	}

	item := toNodeModel(normalized, nil)
	item.IsBuiltIn = false
	if item.Type == "master" {
		item.IsBuiltIn = true
	}

	if err := s.repo.Save(item); err != nil {
		return nil, err
	}

	result := toNodePayload(item)
	return &result, nil
}

func (s *nodeService) Update(id uint, payload *NodePayload) (*NodePayload, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("node not found")
	}

	normalized, err := normalizeNodePayload(payload, existing.IsBuiltIn)
	if err != nil {
		return nil, err
	}

	item := toNodeModel(normalized, existing)
	item.ID = id
	item.Type = existing.Type
	item.IsBuiltIn = existing.IsBuiltIn

	if err := s.repo.Save(item); err != nil {
		return nil, err
	}

	result := toNodePayload(item)
	return &result, nil
}

func (s *nodeService) Delete(id uint) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("node not found")
	}
	if existing.IsBuiltIn || existing.Type == "master" {
		return fmt.Errorf("built-in master node cannot be deleted")
	}

	return s.repo.Delete(id)
}

func (s *nodeService) CheckHealth(id uint) (*NodePayload, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("node not found")
	}

	now := time.Now()
	existing.LastCheckedAt = &now

	if !existing.Enabled {
		existing.HealthStatus = "disabled"
		existing.HealthMessage = "node disabled: enabled=false; capabilities were not checked"
		if err := s.repo.Save(existing); err != nil {
			return nil, err
		}
		payload := toNodePayload(existing)
		return &payload, nil
	}

	health := checkNodeCapabilities(existing)
	existing.HealthStatus = health.status
	existing.HealthMessage = strings.Join(health.messages, "; ")
	if health.healthy {
		existing.LastHeartbeatAt = &now
	}

	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}

	payload := toNodePayload(existing)
	return &payload, nil
}

type nodeCapabilityHealth struct {
	status   string
	healthy  bool
	messages []string
}

func checkNodeCapabilities(node *model.Node) nodeCapabilityHealth {
	result := nodeCapabilityHealth{
		status:   "idle",
		healthy:  true,
		messages: []string{},
	}
	checked := 0

	if node.CreateArchive {
		checked++
		if ok, message := checkCreateArchiveCapability(node); ok {
			result.messages = append(result.messages, "create_archive: ok ("+message+")")
		} else {
			result.healthy = false
			result.messages = append(result.messages, "create_archive: failed ("+message+")")
		}
	}

	if node.ExtractArchive {
		checked++
		if ok, message := checkExtractArchiveCapability(node); ok {
			result.messages = append(result.messages, "extract_archive: ok ("+message+")")
		} else {
			result.healthy = false
			result.messages = append(result.messages, "extract_archive: failed ("+message+")")
		}
	}

	if node.OfflineDownload {
		checked++
		connectivity, err := TestNodeOfflineConnectivity(NodeOfflinePayload{
			Downloader:      node.OfflineDownloader,
			RPCURL:          node.OfflineRPCURL,
			RPCSecret:       node.OfflineRPCSecret,
			TaskOptions:     node.OfflineTaskOptions,
			TempDir:         node.OfflineTempDir,
			RefreshInterval: node.OfflineRefreshInterval,
			WaitForSeeding:  node.OfflineWaitForSeeding,
		})
		if err != nil {
			result.healthy = false
			result.messages = append(result.messages, "offline_download: failed ("+err.Error()+")")
		} else {
			result.messages = append(result.messages, "offline_download: ok ("+connectivity.Message+")")
		}
	}

	if checked == 0 {
		result.status = "idle"
		result.messages = append(result.messages, "node enabled: no runtime capabilities are enabled")
		return result
	}
	if result.healthy {
		result.status = "online"
	} else {
		result.status = "offline"
	}
	return result
}

func checkCreateArchiveCapability(node *model.Node) (bool, string) {
	if !canNodeRunLocalArchive(node) {
		return false, fmt.Sprintf("remote archive execution is not supported yet for %s nodes", nodeTypeLabel(node))
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	entry, err := zipWriter.Create("health-check.txt")
	if err != nil {
		_ = zipWriter.Close()
		return false, fmt.Sprintf("create zip entry failed: %v", err)
	}
	if _, err := entry.Write([]byte("ok")); err != nil {
		_ = zipWriter.Close()
		return false, fmt.Sprintf("write zip entry failed: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		return false, fmt.Sprintf("close zip writer failed: %v", err)
	}
	if buf.Len() == 0 {
		return false, "zip writer produced empty output"
	}
	return true, "local zip writer can create archives"
}

func checkExtractArchiveCapability(node *model.Node) (bool, string) {
	if !canNodeRunLocalArchive(node) {
		return false, fmt.Sprintf("remote archive extraction is not supported yet for %s nodes", nodeTypeLabel(node))
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	entry, err := zipWriter.Create("health-check.txt")
	if err != nil {
		_ = zipWriter.Close()
		return false, fmt.Sprintf("build probe archive failed: %v", err)
	}
	if _, err := entry.Write([]byte("ok")); err != nil {
		_ = zipWriter.Close()
		return false, fmt.Sprintf("write probe archive failed: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		return false, fmt.Sprintf("close probe archive failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return false, fmt.Sprintf("open zip reader failed: %v", err)
	}
	if len(reader.File) != 1 {
		return false, fmt.Sprintf("expected 1 probe entry, got %d", len(reader.File))
	}
	stream, err := reader.File[0].Open()
	if err != nil {
		return false, fmt.Sprintf("open zip entry failed: %v", err)
	}
	defer stream.Close()
	content, err := io.ReadAll(stream)
	if err != nil {
		return false, fmt.Sprintf("read zip entry failed: %v", err)
	}
	if string(content) != "ok" {
		return false, "zip entry content verification failed"
	}
	return true, "local zip reader can extract archives"
}

func canNodeRunLocalArchive(node *model.Node) bool {
	return node != nil && node.Type == "master"
}

func nodeTypeLabel(node *model.Node) string {
	if node == nil || strings.TrimSpace(node.Type) == "" {
		return "unknown"
	}
	return node.Type
}
func defaultNodePayload() NodePayload {
	return NodePayload{
		Name:      "Master",
		Type:      "master",
		Enabled:   true,
		Weight:    1,
		IsBuiltIn: true,
		Health: NodeHealthPayload{
			Status:  "online",
			Message: "主节点运行中",
		},
		Features: NodeFeaturePayload{
			CreateArchive:   true,
			ExtractArchive:  true,
			OfflineDownload: true,
		},
		Offline: NodeOfflinePayload{
			Downloader:      "Aria2",
			RPCURL:          "http://127.0.0.1:6800/jsonrpc",
			RPCSecret:       "",
			TaskOptions:     "{\n  \"max-connection-per-server\": \"8\"\n}",
			TempDir:         "",
			RefreshInterval: 5,
			WaitForSeeding:  false,
		},
	}
}

func normalizeNodePayload(payload *NodePayload, keepBuiltIn bool) (*NodePayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("node payload cannot be nil")
	}

	normalized := *payload
	normalized.Name = strings.TrimSpace(normalized.Name)
	normalized.Type = strings.ToLower(strings.TrimSpace(normalized.Type))
	normalized.Offline.Downloader = strings.TrimSpace(normalized.Offline.Downloader)
	normalized.Offline.RPCURL = strings.TrimSpace(normalized.Offline.RPCURL)
	normalized.Offline.RPCSecret = strings.TrimSpace(normalized.Offline.RPCSecret)
	normalized.Offline.TaskOptions = strings.TrimSpace(normalized.Offline.TaskOptions)
	normalized.Offline.TempDir = strings.TrimSpace(normalized.Offline.TempDir)

	if normalized.Name == "" {
		return nil, fmt.Errorf("node name cannot be empty")
	}

	if normalized.Type == "" {
		normalized.Type = "worker"
	}

	switch normalized.Type {
	case "master", "worker":
	default:
		return nil, fmt.Errorf("unsupported node type")
	}

	if keepBuiltIn {
		normalized.Type = "master"
		normalized.IsBuiltIn = true
	}

	if normalized.Offline.Downloader == "" {
		normalized.Offline.Downloader = "Aria2"
	}

	switch normalized.Offline.Downloader {
	case "Aria2", "qBittorrent", "Transmission":
	default:
		return nil, fmt.Errorf("unsupported downloader type")
	}

	normalized.Weight = clampInt(normalized.Weight, 1, 1024)
	normalized.Offline.RefreshInterval = clampInt(normalized.Offline.RefreshInterval, 1, 3600)

	if normalized.Offline.TaskOptions == "" {
		normalized.Offline.TaskOptions = "{}"
	}

	return &normalized, nil
}

func toNodePayload(item *model.Node) NodePayload {
	if item == nil {
		return defaultNodePayload()
	}

	return NodePayload{
		ID:        item.ID,
		Name:      item.Name,
		Type:      item.Type,
		Enabled:   item.Enabled,
		Weight:    item.Weight,
		IsBuiltIn: item.IsBuiltIn,
		Health: NodeHealthPayload{
			Status:          item.HealthStatus,
			Message:         item.HealthMessage,
			LastHeartbeatAt: formatNodeTime(item.LastHeartbeatAt),
			LastCheckedAt:   formatNodeTime(item.LastCheckedAt),
		},
		Features: NodeFeaturePayload{
			CreateArchive:   item.CreateArchive,
			ExtractArchive:  item.ExtractArchive,
			OfflineDownload: item.OfflineDownload,
		},
		Offline: NodeOfflinePayload{
			Downloader:      item.OfflineDownloader,
			RPCURL:          item.OfflineRPCURL,
			RPCSecret:       item.OfflineRPCSecret,
			TaskOptions:     item.OfflineTaskOptions,
			TempDir:         item.OfflineTempDir,
			RefreshInterval: item.OfflineRefreshInterval,
			WaitForSeeding:  item.OfflineWaitForSeeding,
		},
	}
}

func toNodeModel(payload *NodePayload, existing *model.Node) *model.Node {
	item := &model.Node{}
	if existing != nil {
		*item = *existing
	}

	item.Name = payload.Name
	item.Type = payload.Type
	item.Enabled = payload.Enabled
	item.Weight = payload.Weight
	item.IsBuiltIn = payload.IsBuiltIn
	item.CreateArchive = payload.Features.CreateArchive
	item.ExtractArchive = payload.Features.ExtractArchive
	item.OfflineDownload = payload.Features.OfflineDownload
	item.OfflineDownloader = payload.Offline.Downloader
	item.OfflineRPCURL = payload.Offline.RPCURL
	item.OfflineRPCSecret = payload.Offline.RPCSecret
	item.OfflineTaskOptions = payload.Offline.TaskOptions
	item.OfflineTempDir = payload.Offline.TempDir
	item.OfflineRefreshInterval = payload.Offline.RefreshInterval
	item.OfflineWaitForSeeding = payload.Offline.WaitForSeeding
	if item.HealthStatus == "" {
		item.HealthStatus = "unknown"
	}

	return item
}

func formatNodeTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}
