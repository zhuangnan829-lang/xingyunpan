package model

// FileSystemSetting stores admin-managed file system experience settings.
type FileSystemSetting struct {
	BaseModel
	OnlineEditorSize             int    `gorm:"not null;default:50" json:"online_editor_size"`
	OnlineEditorUnit             string `gorm:"size:8;not null;default:'MB'" json:"online_editor_unit"`
	RecycleScanInterval          string `gorm:"size:64;not null;default:'@every 33m'" json:"recycle_scan_interval"`
	BlobRecycleInterval          string `gorm:"size:64;not null;default:'@every 15m'" json:"blob_recycle_interval"`
	StaticCacheTTL               int    `gorm:"not null;default:86400" json:"static_cache_ttl"`
	ListPaginationMode           string `gorm:"size:32;not null;default:'cursor'" json:"list_pagination_mode"`
	MaxPageSize                  int    `gorm:"not null;default:2000" json:"max_page_size"`
	MaxBatchActionSize           int    `gorm:"not null;default:3000" json:"max_batch_action_size"`
	MaxRecursiveSearch           int    `gorm:"not null;default:65535" json:"max_recursive_search"`
	MapProvider                  string `gorm:"size:64;not null;default:'osm-leaflet'" json:"map_provider"`
	MimeMap                      string `gorm:"type:longtext" json:"mime_map"`
	ImageQuery                   string `gorm:"type:text" json:"image_query"`
	VideoQuery                   string `gorm:"type:text" json:"video_query"`
	AudioQuery                   string `gorm:"type:text" json:"audio_query"`
	DocumentQuery                string `gorm:"type:text" json:"document_query"`
	FileIconRules                string `gorm:"type:longtext" json:"file_icon_rules"`
	EmojiOptions                 string `gorm:"type:text" json:"emoji_options"`
	BrowserApps                  string `gorm:"type:longtext" json:"browser_apps"`
	CustomProperties             string `gorm:"type:longtext" json:"custom_properties"`
	MasterKeyStorage             string `gorm:"size:32;not null;default:'database'" json:"master_key_storage"`
	ShowEncryptionStatus         bool   `gorm:"not null;default:true" json:"show_encryption_status"`
	EnableEventPush              bool   `gorm:"not null;default:true" json:"enable_event_push"`
	OfflineTTL                   int    `gorm:"not null;default:1209600" json:"offline_ttl"`
	DebounceDelay                int    `gorm:"not null;default:5" json:"debounce_delay"`
	ServerSideDownloadSessionTTL int    `gorm:"not null;default:600" json:"server_side_download_session_ttl"`
	UploadSessionTTL             int    `gorm:"not null;default:86400" json:"upload_session_ttl"`
	SlaveAPISignTTL              int    `gorm:"not null;default:60" json:"slave_api_sign_ttl"`
	DirectoryStatTTL             int    `gorm:"not null;default:300" json:"directory_stat_ttl"`
	MaxChunkRetry                int    `gorm:"not null;default:5" json:"max_chunk_retry"`
	CacheChunksForRetry          bool   `gorm:"not null;default:true" json:"cache_chunks_for_retry"`
	TransferParallelism          int    `gorm:"not null;default:4" json:"transfer_parallelism"`
	OAuthRefreshInterval         string `gorm:"size:64;not null;default:'@every 230h'" json:"oauth_refresh_interval"`
	WOPISessionTTL               int    `gorm:"not null;default:36000" json:"wopi_session_ttl"`
	BlobSignedURLTTL             int    `gorm:"not null;default:3600" json:"blob_signed_url_ttl"`
	BlobSignedURLReuseTTL        int    `gorm:"not null;default:600" json:"blob_signed_url_reuse_ttl"`
}

// TableName specifies the DB table name.
func (FileSystemSetting) TableName() string {
	return "file_system_settings"
}
