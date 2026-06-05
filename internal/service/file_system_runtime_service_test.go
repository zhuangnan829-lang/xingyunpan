package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

func TestParseEveryDuration(t *testing.T) {
	duration, err := parseEveryDuration("@every 230h")
	if err != nil {
		t.Fatalf("parseEveryDuration returned error: %v", err)
	}
	if duration != 230*time.Hour {
		t.Fatalf("expected 230h, got %s", duration)
	}

	duration, err = parseEveryDuration("15m")
	if err != nil {
		t.Fatalf("parseEveryDuration without prefix returned error: %v", err)
	}
	if duration != 15*time.Minute {
		t.Fatalf("expected 15m, got %s", duration)
	}
}

func TestSlaveSignatureRoundTrip(t *testing.T) {
	key := []byte("test-master-key")
	timestamp := time.Now().Unix()
	signature := signSlavePayload(key, "post", "/api/v1/slave/ping", "abc123", timestamp)
	if signature == "" {
		t.Fatal("expected signature")
	}

	secondSignature := signSlavePayload(key, "POST", "/api/v1/slave/ping", "abc123", timestamp)
	if signature != secondSignature {
		t.Fatalf("signature should normalize method casing")
	}

	tampered := signSlavePayload(key, "POST", "/api/v1/slave/ping", "def456", timestamp)
	if signature == tampered {
		t.Fatal("signature should change when body hash changes")
	}
}

func TestRandomTokenIsURLSafe(t *testing.T) {
	token, err := randomToken(24)
	if err != nil {
		t.Fatalf("randomToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if strings.ContainsAny(token, "+/=") {
		t.Fatalf("token should be raw URL-safe base64, got %q", token)
	}
}

type fakeFileSystemSettingRepository struct{}

func (r *fakeFileSystemSettingRepository) Get() (*model.FileSystemSetting, error) {
	return &model.FileSystemSetting{ServerSideDownloadSessionTTL: 600}, nil
}

func (r *fakeFileSystemSettingRepository) Save(setting *model.FileSystemSetting) error {
	return nil
}

type fakeRuntimeUserFileRepository struct {
	nextID uint
	files  map[uint]*model.UserFile
}

func (r *fakeRuntimeUserFileRepository) Create(file *model.UserFile) error {
	if r.nextID == 0 {
		r.nextID = 100
	}
	file.ID = r.nextID
	r.nextID++
	copyFile := *file
	r.files[file.ID] = &copyFile
	return nil
}

func (r *fakeRuntimeUserFileRepository) GetByID(id uint) (*model.UserFile, error) {
	return r.GetByIDWithPhysicalFile(id)
}

func (r *fakeRuntimeUserFileRepository) GetByIDWithPhysicalFile(id uint) (*model.UserFile, error) {
	file, ok := r.files[id]
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	copyFile := *file
	return &copyFile, nil
}

func (r *fakeRuntimeUserFileRepository) List(userID uint, parentID *uint, page, pageSize int) ([]*model.UserFile, int64, error) {
	return nil, 0, nil
}

func (r *fakeRuntimeUserFileRepository) ListAfterID(userID uint, parentID *uint, afterID uint, pageSize int) ([]*model.UserFile, int64, error) {
	return nil, 0, nil
}

func (r *fakeRuntimeUserFileRepository) ListChildren(userID uint, parentID uint) ([]*model.UserFile, error) {
	return nil, nil
}

func (r *fakeRuntimeUserFileRepository) ListDescendants(userID uint, folderID uint) ([]*model.UserFile, error) {
	return nil, nil
}

func (r *fakeRuntimeUserFileRepository) GetFolderPath(userID uint, folderID uint) ([]*model.UserFile, error) {
	return nil, nil
}

func (r *fakeRuntimeUserFileRepository) GetImmediateChildStats(userID uint, folderIDs []uint) (map[uint]repository.FolderChildStats, error) {
	return nil, nil
}

func (r *fakeRuntimeUserFileRepository) Update(file *model.UserFile) error { return nil }

func (r *fakeRuntimeUserFileRepository) Delete(id uint) error { return nil }

type fakeRuntimeStorage struct {
	objects map[string][]byte
}

func (s *fakeRuntimeStorage) Save(reader io.Reader, relativePath string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.objects[relativePath] = data
	return nil
}

func (s *fakeRuntimeStorage) Read(relativePath string) (io.ReadCloser, error) {
	data, ok := s.objects[relativePath]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", relativePath)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *fakeRuntimeStorage) Delete(relativePath string) error {
	delete(s.objects, relativePath)
	return nil
}

func (s *fakeRuntimeStorage) Exists(relativePath string) bool {
	_, ok := s.objects[relativePath]
	return ok
}

func TestCreatePackageDownloadSessionRejectsWithoutCreateArchiveCapability(t *testing.T) {
	service := newTestFileSystemRuntimeService([]model.Node{
		archiveNode(1, "Master", "master", true, false, 10),
	})

	if _, err := service.CreatePackageDownloadSession(context.Background(), 42, []uint{1}); err == nil {
		t.Fatalf("expected create_archive capability error")
	}
}

func TestCreatePackageDownloadSessionSkipsDisabledArchiveNode(t *testing.T) {
	service := newTestFileSystemRuntimeService([]model.Node{
		archiveNode(1, "Master", "master", false, true, 10),
	})

	if _, err := service.CreatePackageDownloadSession(context.Background(), 42, []uint{1}); err == nil {
		t.Fatalf("expected enabled=false node to be skipped")
	}
}

func TestCreatePackageDownloadSessionWeightedArchiveSelection(t *testing.T) {
	service := newTestFileSystemRuntimeService([]model.Node{
		archiveNode(1, "Master", "master", true, true, 1),
		archiveNode(2, "Archive Worker", "worker", true, true, 3),
	})
	counts := map[uint]int{}

	for i := 0; i < 8; i++ {
		payload, err := service.CreatePackageDownloadSession(context.Background(), 42, []uint{1})
		if err != nil {
			t.Fatalf("create package session: %v", err)
		}
		counts[payload.NodeID]++
		if payload.NodeName == "" || payload.DispatchNode == nil {
			t.Fatalf("node record was not returned: %#v", payload)
		}
	}

	if counts[1] != 2 || counts[2] != 6 {
		t.Fatalf("unexpected archive node distribution: master=%d worker=%d", counts[1], counts[2])
	}
}

func TestCreatePackageDownloadSessionCanSwitchAwayFromDisabledMasterCapability(t *testing.T) {
	service := newTestFileSystemRuntimeService([]model.Node{
		archiveNode(1, "Master", "master", true, false, 100),
		archiveNode(2, "Archive Worker", "worker", true, true, 1),
	})

	payload, err := service.CreatePackageDownloadSession(context.Background(), 42, []uint{1})
	if err != nil {
		t.Fatalf("create package session: %v", err)
	}
	if payload.NodeID != 2 {
		t.Fatalf("expected worker when master create_archive=false, got %d", payload.NodeID)
	}
}

func TestWritePackageDownloadSucceedsAndRecordsMasterNode(t *testing.T) {
	service := newTestPackageRuntimeServiceWithStorage([]model.Node{
		archiveNode(1, "Master", "master", true, true, 1),
	})

	payload, err := service.CreatePackageDownloadSession(context.Background(), 42, []uint{1})
	if err != nil {
		t.Fatalf("create package session: %v", err)
	}
	if payload.NodeID != 1 || payload.NodeName != "Master" || payload.DispatchNode == nil || payload.DispatchNode.ID != 1 {
		t.Fatalf("archive session did not record master node: %#v", payload)
	}

	var archive bytes.Buffer
	completed, err := service.WritePackageDownload(context.Background(), 42, payload.SessionID, &archive)
	if err != nil {
		t.Fatalf("write package download: %v", err)
	}
	if completed.NodeID != 1 || completed.NodeName != "Master" {
		t.Fatalf("completed package did not keep node record: %#v", completed)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		t.Fatalf("open generated zip: %v", err)
	}
	if len(zipReader.File) != 1 || zipReader.File[0].Name != "1-report.txt" {
		t.Fatalf("unexpected generated zip entries: %#v", zipReader.File)
	}
}

func TestWritePackageDownloadRejectsWorkerUntilRemoteArchiveExecutionExists(t *testing.T) {
	service := newTestFileSystemRuntimeService([]model.Node{
		archiveNode(2, "Archive Worker", "worker", true, true, 1),
	})
	payload, err := service.CreatePackageDownloadSession(context.Background(), 42, []uint{1})
	if err != nil {
		t.Fatalf("create package session: %v", err)
	}

	if _, err := service.WritePackageDownload(context.Background(), 42, payload.SessionID, &strings.Builder{}); err == nil {
		t.Fatalf("expected remote archive execution unsupported error")
	}
}

func TestExtractArchiveRejectsWithoutExtractCapability(t *testing.T) {
	service, _ := newTestExtractRuntimeService([]model.Node{
		extractNode(1, "Master", "master", true, false, 1),
	})

	result, err := service.ExtractArchive(context.Background(), 42, 1, nil)
	if err == nil {
		t.Fatalf("expected extract_archive capability error, got result %#v", result)
	}
}

func TestExtractArchiveRejectsDisabledNode(t *testing.T) {
	service, _ := newTestExtractRuntimeService([]model.Node{
		extractNode(1, "Master", "master", false, true, 1),
	})

	result, err := service.ExtractArchive(context.Background(), 42, 1, nil)
	if err == nil {
		t.Fatalf("expected enabled=false node error, got result %#v", result)
	}
}

func TestExtractArchiveZipCreatesFilesInTargetFolder(t *testing.T) {
	targetID := uint(2)
	service, files := newTestExtractRuntimeService([]model.Node{
		extractNode(1, "Master", "master", true, true, 1),
	})

	result, err := service.ExtractArchive(context.Background(), 42, 1, &targetID)
	if err != nil {
		t.Fatalf("extract archive: %v", err)
	}
	if result.Status != "processing" {
		t.Fatalf("expected queued processing task, got %#v", result)
	}
	if result.NodeID != 1 || result.NodeName != "Master" || result.SourceFileID != 1 {
		t.Fatalf("node/source task fields not recorded: %#v", result)
	}

	raw, err := service.ExecuteArchiveExtract(context.Background(), 42, result.TaskID)
	if err != nil {
		t.Fatalf("execute archive extract: %v", err)
	}
	if !strings.Contains(raw, `"status":"completed"`) {
		t.Fatalf("unexpected extract runner result: %s", raw)
	}
	if len(files.files) < 3 {
		t.Fatalf("expected extracted file to be created, files=%#v", files.files)
	}
	var extracted *model.UserFile
	for id, file := range files.files {
		if id != 1 && id != targetID {
			extracted = file
			break
		}
	}
	if extracted == nil {
		t.Fatalf("extracted file was not created")
	}
	if extracted.FileName != "hello.txt" || extracted.ParentID == nil || *extracted.ParentID != targetID {
		t.Fatalf("extracted file target/name mismatch: %#v", extracted)
	}
}

func newTestFileSystemRuntimeService(nodes []model.Node) FileSystemRuntimeService {
	physicalID := uint(10)
	files := &fakeRuntimeUserFileRepository{files: map[uint]*model.UserFile{
		1: {
			BaseModel:      model.BaseModel{ID: 1},
			UserID:         42,
			FileName:       "report.txt",
			PhysicalFileID: &physicalID,
			PhysicalFile:   &model.PhysicalFile{BaseModel: model.BaseModel{ID: physicalID}, StoragePath: "files/report.txt", StorageType: "local"},
		},
	}}
	return NewFileSystemRuntimeService(nil, &fakeFileSystemSettingRepository{}, files, nil, "", NewNodeDispatchService(&fakeNodeRepository{nodes: nodes}))
}

func newTestPackageRuntimeServiceWithStorage(nodes []model.Node) FileSystemRuntimeService {
	physicalID := uint(10)
	files := &fakeRuntimeUserFileRepository{files: map[uint]*model.UserFile{
		1: {
			BaseModel:      model.BaseModel{ID: 1},
			UserID:         42,
			FileName:       "report.txt",
			PhysicalFileID: &physicalID,
			PhysicalFile:   &model.PhysicalFile{BaseModel: model.BaseModel{ID: physicalID}, StoragePath: "files/report.txt", StorageType: "local"},
		},
	}}
	stor := &fakeRuntimeStorage{objects: map[string][]byte{"files/report.txt": []byte("archive content")}}
	return NewFileSystemRuntimeService(nil, &fakeFileSystemSettingRepository{}, files, stor, "", NewNodeDispatchService(&fakeNodeRepository{nodes: nodes}))
}

func newTestExtractRuntimeService(nodes []model.Node) (FileSystemRuntimeService, *fakeRuntimeUserFileRepository) {
	physicalID := uint(10)
	targetID := uint(2)
	stor := &fakeRuntimeStorage{objects: map[string][]byte{
		"files/source.zip": buildTestZipBytes(map[string]string{"hello.txt": "hello from zip"}),
	}}
	files := &fakeRuntimeUserFileRepository{files: map[uint]*model.UserFile{
		1: {
			BaseModel:      model.BaseModel{ID: 1},
			UserID:         42,
			FileName:       "source.zip",
			PhysicalFileID: &physicalID,
			PhysicalFile:   &model.PhysicalFile{BaseModel: model.BaseModel{ID: physicalID}, StoragePath: "files/source.zip", StorageType: "local"},
		},
		2: {
			BaseModel: model.BaseModel{ID: targetID},
			UserID:    42,
			FileName:  "target",
			IsFolder:  true,
		},
	}}
	return NewFileSystemRuntimeService(nil, &fakeFileSystemSettingRepository{}, files, stor, "", NewNodeDispatchService(&fakeNodeRepository{nodes: nodes})), files
}

func buildTestZipBytes(entries map[string]string) []byte {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	for name, content := range entries {
		writer, err := zipWriter.Create(name)
		if err != nil {
			panic(err)
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			panic(err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func archiveNode(id uint, name string, nodeType string, enabled bool, createArchive bool, weight int) model.Node {
	return model.Node{
		BaseModel:     model.BaseModel{ID: id},
		Name:          name,
		Type:          nodeType,
		Enabled:       enabled,
		Weight:        weight,
		CreateArchive: createArchive,
	}
}

func extractNode(id uint, name string, nodeType string, enabled bool, extractArchive bool, weight int) model.Node {
	return model.Node{
		BaseModel:      model.BaseModel{ID: id},
		Name:           name,
		Type:           nodeType,
		Enabled:        enabled,
		Weight:         weight,
		ExtractArchive: extractArchive,
	}
}
