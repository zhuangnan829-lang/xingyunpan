package service

import (
	"encoding/json"
	"testing"

	"xingyunpan-v2/internal/model"
)

type memoryFileSystemSettingRepository struct {
	setting *model.FileSystemSetting
}

func (r *memoryFileSystemSettingRepository) Get() (*model.FileSystemSetting, error) {
	if r.setting == nil {
		return nil, nil
	}
	copy := *r.setting
	return &copy, nil
}

func (r *memoryFileSystemSettingRepository) Save(setting *model.FileSystemSetting) error {
	copy := *setting
	if copy.ID == 0 {
		copy.ID = 1
	}
	r.setting = &copy
	return nil
}

func TestUpdateIconsOnlyChangesIconFields(t *testing.T) {
	initial := defaultFileSystemSettingPayload()
	initial.OnlineEditorSize = 2048
	initial.ListPaginationMode = "offset"
	initial.BrowserApps = `[{"id":1,"name":"keep","items":[]}]`
	initial.CustomProperties = `[{"id":9,"key":"keep","name":"Keep","type":"text"}]`
	initial.MasterKeyStorage = "env"

	repo := &memoryFileSystemSettingRepository{setting: fileSystemSettingModelFromPayload(initial)}
	svc := NewFileSystemSettingService(repo, nil)

	updated, err := svc.UpdateIcons(&FileSystemIconSettingsPayload{
		FileIconRules: `[{"label":" PDF ","icon":" P ","match":".PDF pdf","tint":"#abc"}]`,
		EmojiOptions:  `{"enabled":true,"showInList":true,"fallbackUnknown":true,"folderEmoji":"📁","unknownEmoji":"🗂️","categories":[{"icon":"P","label":"PDF","match":"pdf","emojis":["P"]}]}`,
	})
	if err != nil {
		t.Fatalf("UpdateIcons returned error: %v", err)
	}

	if updated.OnlineEditorSize != initial.OnlineEditorSize ||
		updated.ListPaginationMode != initial.ListPaginationMode ||
		updated.BrowserApps != initial.BrowserApps ||
		updated.CustomProperties != initial.CustomProperties ||
		updated.MasterKeyStorage != initial.MasterKeyStorage {
		t.Fatalf("non-icon fields changed: updated=%#v initial=%#v", updated, initial)
	}
	if updated.FileIconRules == initial.FileIconRules || updated.EmojiOptions == initial.EmojiOptions {
		t.Fatalf("icon fields were not updated")
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(updated.FileIconRules), &rules); err != nil {
		t.Fatalf("updated icon rules are invalid: %v", err)
	}
	if len(rules) != 1 || rules[0].Label != "PDF" || rules[0].Icon != "P" || rules[0].Match != "pdf" || rules[0].Tint != "#abc" {
		t.Fatalf("icon rules were not normalized: %#v", rules)
	}
}

func fileSystemSettingModelFromPayload(payload *FileSystemSettingPayload) *model.FileSystemSetting {
	return &model.FileSystemSetting{
		BaseModel:                    model.BaseModel{ID: 1},
		OnlineEditorSize:             payload.OnlineEditorSize,
		OnlineEditorUnit:             payload.OnlineEditorUnit,
		RecycleScanInterval:          payload.RecycleScanInterval,
		BlobRecycleInterval:          payload.BlobRecycleInterval,
		StaticCacheTTL:               payload.StaticCacheTTL,
		ListPaginationMode:           payload.ListPaginationMode,
		MaxPageSize:                  payload.MaxPageSize,
		MaxBatchActionSize:           payload.MaxBatchActionSize,
		MaxRecursiveSearch:           payload.MaxRecursiveSearch,
		MapProvider:                  payload.MapProvider,
		MimeMap:                      payload.MimeMap,
		ImageQuery:                   payload.ImageQuery,
		VideoQuery:                   payload.VideoQuery,
		AudioQuery:                   payload.AudioQuery,
		DocumentQuery:                payload.DocumentQuery,
		FileIconRules:                payload.FileIconRules,
		EmojiOptions:                 payload.EmojiOptions,
		BrowserApps:                  payload.BrowserApps,
		CustomProperties:             payload.CustomProperties,
		MasterKeyStorage:             payload.MasterKeyStorage,
		ShowEncryptionStatus:         payload.ShowEncryptionStatus,
		EnableEventPush:              payload.EnableEventPush,
		OfflineTTL:                   payload.OfflineTTL,
		DebounceDelay:                payload.DebounceDelay,
		ServerSideDownloadSessionTTL: payload.ServerSideDownloadSessionTTL,
		UploadSessionTTL:             payload.UploadSessionTTL,
		SlaveAPISignTTL:              payload.SlaveAPISignTTL,
		DirectoryStatTTL:             payload.DirectoryStatTTL,
		MaxChunkRetry:                payload.MaxChunkRetry,
		CacheChunksForRetry:          payload.CacheChunksForRetry,
		TransferParallelism:          payload.TransferParallelism,
		OAuthRefreshInterval:         payload.OAuthRefreshInterval,
		WOPISessionTTL:               payload.WOPISessionTTL,
		BlobSignedURLTTL:             payload.BlobSignedURLTTL,
		BlobSignedURLReuseTTL:        payload.BlobSignedURLReuseTTL,
	}
}
