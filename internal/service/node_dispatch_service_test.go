package service

import (
	"fmt"
	"testing"

	"xingyunpan-v2/internal/model"
)

type fakeNodeRepository struct {
	nodes []model.Node
}

func (r *fakeNodeRepository) EnsureSchema() error {
	return nil
}

func (r *fakeNodeRepository) List() ([]model.Node, error) {
	items := make([]model.Node, len(r.nodes))
	copy(items, r.nodes)
	return items, nil
}

func (r *fakeNodeRepository) GetByID(id uint) (*model.Node, error) {
	for i := range r.nodes {
		if r.nodes[i].ID == id {
			return &r.nodes[i], nil
		}
	}
	return nil, nil
}

func (r *fakeNodeRepository) Save(node *model.Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}
	if node.ID == 0 {
		node.ID = uint(len(r.nodes) + 1)
		r.nodes = append(r.nodes, *node)
		return nil
	}
	for i := range r.nodes {
		if r.nodes[i].ID == node.ID {
			r.nodes[i] = *node
			return nil
		}
	}
	r.nodes = append(r.nodes, *node)
	return nil
}

func (r *fakeNodeRepository) Delete(id uint) error {
	for i := range r.nodes {
		if r.nodes[i].ID == id {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			return nil
		}
	}
	return nil
}

func TestNodeDispatchSkipsDisabledNodes(t *testing.T) {
	repo := &fakeNodeRepository{nodes: []model.Node{
		nodeFixture(1, "disabled", true, true, true, 100),
		nodeFixture(2, "enabled", true, true, true, 1),
	}}
	repo.nodes[0].Enabled = false
	selector := NewNodeDispatchService(repo)

	for i := 0; i < 10; i++ {
		selected, err := selector.SelectCreateArchiveNode()
		if err != nil {
			t.Fatalf("select node: %v", err)
		}
		if selected.ID != 2 {
			t.Fatalf("disabled node was selected: got %d", selected.ID)
		}
	}
}

func TestNodeDispatchSkipsNodesWithoutCapability(t *testing.T) {
	repo := &fakeNodeRepository{nodes: []model.Node{
		nodeFixture(1, "no-offline", false, true, true, 100),
		nodeFixture(2, "offline", false, true, true, 1),
	}}
	repo.nodes[0].OfflineDownload = false
	repo.nodes[1].OfflineDownload = true
	selector := NewNodeDispatchService(repo)

	for i := 0; i < 10; i++ {
		selected, err := selector.SelectOfflineDownloadNode()
		if err != nil {
			t.Fatalf("select node: %v", err)
		}
		if selected.ID != 2 {
			t.Fatalf("node without offline_download capability was selected: got %d", selected.ID)
		}
	}
}

func TestNodeDispatchWeightedRoundRobin(t *testing.T) {
	repo := &fakeNodeRepository{nodes: []model.Node{
		nodeFixture(1, "low", true, true, true, 1),
		nodeFixture(2, "high", true, true, true, 3),
	}}
	selector := NewNodeDispatchService(repo)
	counts := map[uint]int{}

	for i := 0; i < 8; i++ {
		selected, err := selector.SelectExtractArchiveNode()
		if err != nil {
			t.Fatalf("select node: %v", err)
		}
		counts[selected.ID]++
	}

	if counts[1] != 2 || counts[2] != 6 {
		t.Fatalf("unexpected weighted round-robin counts: low=%d high=%d", counts[1], counts[2])
	}
}

func TestNodeDispatchReflectsRuntimeNodeConfigChanges(t *testing.T) {
	repo := &fakeNodeRepository{nodes: []model.Node{
		nodeFixture(1, "first", true, true, true, 1),
		nodeFixture(2, "second", true, true, true, 1),
	}}
	selector := NewNodeDispatchService(repo)

	selected, err := selector.SelectCreateArchiveNode()
	if err != nil {
		t.Fatalf("select node: %v", err)
	}
	if selected.ID != 1 {
		t.Fatalf("expected first node before config change, got %d", selected.ID)
	}

	repo.nodes[0].Enabled = false
	repo.nodes[1].CreateArchive = false
	if _, err := selector.SelectCreateArchiveNode(); err == nil {
		t.Fatalf("expected no eligible node after disabling first and removing second capability")
	}

	repo.nodes[1].CreateArchive = true
	repo.nodes[1].Weight = 4
	selected, err = selector.SelectCreateArchiveNode()
	if err != nil {
		t.Fatalf("select node after config change: %v", err)
	}
	if selected.ID != 2 || selected.Weight != 4 {
		t.Fatalf("expected updated second node, got id=%d weight=%d", selected.ID, selected.Weight)
	}
}

func nodeFixture(id uint, name string, createArchive bool, extractArchive bool, offlineDownload bool, weight int) model.Node {
	return model.Node{
		BaseModel:       model.BaseModel{ID: id},
		Name:            name,
		Type:            "worker",
		Enabled:         true,
		Weight:          weight,
		CreateArchive:   createArchive,
		ExtractArchive:  extractArchive,
		OfflineDownload: offlineDownload,
	}
}
