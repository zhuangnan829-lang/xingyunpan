package service

import (
	"fmt"
	"sync"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type NodeCapability string

const (
	NodeCapabilityCreateArchive   NodeCapability = "create_archive"
	NodeCapabilityExtractArchive  NodeCapability = "extract_archive"
	NodeCapabilityOfflineDownload NodeCapability = "offline_download"
)

type SelectedNodePayload struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Weight          int    `json:"weight"`
	Downloader      string `json:"downloader,omitempty"`
	RPCURL          string `json:"rpc_url,omitempty"`
	TaskOptions     string `json:"task_options,omitempty"`
	TempDir         string `json:"temp_dir,omitempty"`
	RefreshInterval int    `json:"refresh_interval,omitempty"`
	WaitForSeeding  bool   `json:"wait_for_seeding,omitempty"`
	rpcSecret       string
}

type NodeSelector interface {
	Select(capability NodeCapability) (*SelectedNodePayload, error)
}

type NodeDispatchService interface {
	NodeSelector
	SelectCreateArchiveNode() (*SelectedNodePayload, error)
	SelectExtractArchiveNode() (*SelectedNodePayload, error)
	SelectOfflineDownloadNode() (*SelectedNodePayload, error)
}

type nodeDispatchService struct {
	repo    repository.NodeRepository
	mu      sync.Mutex
	cursors map[NodeCapability]int
}

func NewNodeDispatchService(repo repository.NodeRepository) NodeDispatchService {
	return &nodeDispatchService{
		repo:    repo,
		cursors: map[NodeCapability]int{},
	}
}

func (s *nodeDispatchService) SelectCreateArchiveNode() (*SelectedNodePayload, error) {
	return s.Select(NodeCapabilityCreateArchive)
}

func (s *nodeDispatchService) SelectExtractArchiveNode() (*SelectedNodePayload, error) {
	return s.Select(NodeCapabilityExtractArchive)
}

func (s *nodeDispatchService) SelectOfflineDownloadNode() (*SelectedNodePayload, error) {
	return s.Select(NodeCapabilityOfflineDownload)
}

func (s *nodeDispatchService) Select(capability NodeCapability) (*SelectedNodePayload, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("node dispatch service is not initialized")
	}
	if !isSupportedNodeCapability(capability) {
		return nil, fmt.Errorf("unsupported node capability: %s", capability)
	}

	nodes, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		payload := defaultNodePayload()
		seed := toNodeModel(&payload, nil)
		seed.IsBuiltIn = true
		if err := s.repo.Save(seed); err != nil {
			return nil, err
		}
		nodes = []model.Node{*seed}
	}

	weighted := make([]model.Node, 0, len(nodes))
	for _, node := range nodes {
		if !node.Enabled || !nodeSupportsCapability(node, capability) {
			continue
		}
		weight := clampInt(node.Weight, 1, 1024)
		for i := 0; i < weight; i++ {
			weighted = append(weighted, node)
		}
	}
	if len(weighted) == 0 {
		return nil, fmt.Errorf("no enabled node supports capability: %s", capability)
	}

	s.mu.Lock()
	index := s.cursors[capability] % len(weighted)
	s.cursors[capability] = (index + 1) % len(weighted)
	s.mu.Unlock()

	return selectedNodePayload(weighted[index]), nil
}

func isSupportedNodeCapability(capability NodeCapability) bool {
	switch capability {
	case NodeCapabilityCreateArchive, NodeCapabilityExtractArchive, NodeCapabilityOfflineDownload:
		return true
	default:
		return false
	}
}

func nodeSupportsCapability(node model.Node, capability NodeCapability) bool {
	switch capability {
	case NodeCapabilityCreateArchive:
		return node.CreateArchive
	case NodeCapabilityExtractArchive:
		return node.ExtractArchive
	case NodeCapabilityOfflineDownload:
		return node.OfflineDownload
	default:
		return false
	}
}

func selectedNodePayload(node model.Node) *SelectedNodePayload {
	return &SelectedNodePayload{
		ID:              node.ID,
		Name:            node.Name,
		Type:            node.Type,
		Weight:          node.Weight,
		Downloader:      node.OfflineDownloader,
		RPCURL:          node.OfflineRPCURL,
		TaskOptions:     node.OfflineTaskOptions,
		TempDir:         node.OfflineTempDir,
		RefreshInterval: node.OfflineRefreshInterval,
		WaitForSeeding:  node.OfflineWaitForSeeding,
		rpcSecret:       node.OfflineRPCSecret,
	}
}
