package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	redisclient "xingyunpan-v2/pkg/redis"
)

// FileSystemSettingPayload is the API-facing shape for admin file system settings.
type FileSystemSettingPayload struct {
	OnlineEditorSize             int    `json:"online_editor_size"`
	OnlineEditorUnit             string `json:"online_editor_unit"`
	RecycleScanInterval          string `json:"recycle_scan_interval"`
	BlobRecycleInterval          string `json:"blob_recycle_interval"`
	StaticCacheTTL               int    `json:"static_cache_ttl"`
	ListPaginationMode           string `json:"list_pagination_mode"`
	MaxPageSize                  int    `json:"max_page_size"`
	MaxBatchActionSize           int    `json:"max_batch_action_size"`
	MaxRecursiveSearch           int    `json:"max_recursive_search"`
	MapProvider                  string `json:"map_provider"`
	MimeMap                      string `json:"mime_map"`
	ImageQuery                   string `json:"image_query"`
	VideoQuery                   string `json:"video_query"`
	AudioQuery                   string `json:"audio_query"`
	DocumentQuery                string `json:"document_query"`
	FileIconRules                string `json:"file_icon_rules"`
	EmojiOptions                 string `json:"emoji_options"`
	BrowserApps                  string `json:"browser_apps"`
	CustomProperties             string `json:"custom_properties"`
	MasterKeyStorage             string `json:"master_key_storage"`
	ShowEncryptionStatus         bool   `json:"show_encryption_status"`
	EnableEventPush              bool   `json:"enable_event_push"`
	OfflineTTL                   int    `json:"offline_ttl"`
	DebounceDelay                int    `json:"debounce_delay"`
	ServerSideDownloadSessionTTL int    `json:"server_side_download_session_ttl"`
	UploadSessionTTL             int    `json:"upload_session_ttl"`
	SlaveAPISignTTL              int    `json:"slave_api_sign_ttl"`
	DirectoryStatTTL             int    `json:"directory_stat_ttl"`
	MaxChunkRetry                int    `json:"max_chunk_retry"`
	CacheChunksForRetry          bool   `json:"cache_chunks_for_retry"`
	TransferParallelism          int    `json:"transfer_parallelism"`
	OAuthRefreshInterval         string `json:"oauth_refresh_interval"`
	WOPISessionTTL               int    `json:"wopi_session_ttl"`
	BlobSignedURLTTL             int    `json:"blob_signed_url_ttl"`
	BlobSignedURLReuseTTL        int    `json:"blob_signed_url_reuse_ttl"`
}

type FileSystemIconSettingsPayload struct {
	FileIconRules string `json:"file_icon_rules"`
	EmojiOptions  string `json:"emoji_options"`
}

// FileSystemSettingService provides admin access to file system settings.
type FileSystemSettingService interface {
	Get() (*FileSystemSettingPayload, error)
	Update(payload *FileSystemSettingPayload) (*FileSystemSettingPayload, error)
	UpdateIcons(payload *FileSystemIconSettingsPayload) (*FileSystemSettingPayload, error)
	ClearBlobURLCache() error
	GetBrowserApps() ([]BrowserAppGroupPayload, error)
	ResolveBrowserApp(filename, mimeType string, platform BrowserAppPlatform) (*BrowserAppResolvedPayload, error)
}

type fileSystemSettingService struct {
	repo           repository.FileSystemSettingRepository
	redisMultipart *redisclient.MultipartRedis
}

var fileIconTintPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// NewFileSystemSettingService creates a file system settings service.
func NewFileSystemSettingService(repo repository.FileSystemSettingRepository, redisMultipart *redisclient.MultipartRedis) FileSystemSettingService {
	return &fileSystemSettingService{repo: repo, redisMultipart: redisMultipart}
}

// Get returns file system settings, creating a default payload when no row exists yet.
func (s *fileSystemSettingService) Get() (*FileSystemSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return defaultFileSystemSettingPayload(), nil
	}

	payload := toFileSystemSettingPayload(setting)
	if isEmptyJSONArray(payload.FileIconRules) {
		payload.FileIconRules = defaultFileIconRulesJSON()
	} else if shouldUpgradeFileIconRules(payload.FileIconRules) {
		payload.FileIconRules = defaultFileIconRulesJSON()
	}
	if shouldUpgradeEmojiOptions(payload.EmojiOptions) {
		payload.EmojiOptions = defaultEmojiOptionsJSON()
	}
	if isEmptyJSONArray(payload.CustomProperties) {
		payload.CustomProperties = defaultCustomPropertiesJSON()
	}
	payload.MimeMap = upgradeDefaultMimeMap(payload.MimeMap)
	payload.ImageQuery = upgradeDefaultCategoryQuery("image", payload.ImageQuery)
	payload.VideoQuery = upgradeDefaultCategoryQuery("video", payload.VideoQuery)
	payload.AudioQuery = upgradeDefaultCategoryQuery("audio", payload.AudioQuery)
	payload.DocumentQuery = upgradeDefaultCategoryQuery("document", payload.DocumentQuery)
	return payload, nil
}

// Update persists file system settings.
func (s *fileSystemSettingService) Update(payload *FileSystemSettingPayload) (*FileSystemSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("file system settings cannot be nil")
	}

	normalized, err := normalizeFileSystemSettingPayload(payload)
	if err != nil {
		return nil, err
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.FileSystemSetting{}
	}

	setting.OnlineEditorSize = normalized.OnlineEditorSize
	setting.OnlineEditorUnit = normalized.OnlineEditorUnit
	setting.RecycleScanInterval = normalized.RecycleScanInterval
	setting.BlobRecycleInterval = normalized.BlobRecycleInterval
	setting.StaticCacheTTL = normalized.StaticCacheTTL
	setting.ListPaginationMode = normalized.ListPaginationMode
	setting.MaxPageSize = normalized.MaxPageSize
	setting.MaxBatchActionSize = normalized.MaxBatchActionSize
	setting.MaxRecursiveSearch = normalized.MaxRecursiveSearch
	setting.MapProvider = normalized.MapProvider
	setting.MimeMap = normalized.MimeMap
	setting.ImageQuery = normalized.ImageQuery
	setting.VideoQuery = normalized.VideoQuery
	setting.AudioQuery = normalized.AudioQuery
	setting.DocumentQuery = normalized.DocumentQuery
	setting.FileIconRules = normalized.FileIconRules
	setting.EmojiOptions = normalized.EmojiOptions
	setting.BrowserApps = normalized.BrowserApps
	setting.CustomProperties = normalized.CustomProperties
	setting.MasterKeyStorage = normalized.MasterKeyStorage
	setting.ShowEncryptionStatus = normalized.ShowEncryptionStatus
	setting.EnableEventPush = normalized.EnableEventPush
	setting.OfflineTTL = normalized.OfflineTTL
	setting.DebounceDelay = normalized.DebounceDelay
	setting.ServerSideDownloadSessionTTL = normalized.ServerSideDownloadSessionTTL
	setting.UploadSessionTTL = normalized.UploadSessionTTL
	setting.SlaveAPISignTTL = normalized.SlaveAPISignTTL
	setting.DirectoryStatTTL = normalized.DirectoryStatTTL
	setting.MaxChunkRetry = normalized.MaxChunkRetry
	setting.CacheChunksForRetry = normalized.CacheChunksForRetry
	setting.TransferParallelism = normalized.TransferParallelism
	setting.OAuthRefreshInterval = normalized.OAuthRefreshInterval
	setting.WOPISessionTTL = normalized.WOPISessionTTL
	setting.BlobSignedURLTTL = normalized.BlobSignedURLTTL
	setting.BlobSignedURLReuseTTL = normalized.BlobSignedURLReuseTTL

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return toFileSystemSettingPayload(setting), nil
}

func (s *fileSystemSettingService) UpdateIcons(payload *FileSystemIconSettingsPayload) (*FileSystemSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("file system icon settings cannot be nil")
	}

	fileIconRules := strings.TrimSpace(payload.FileIconRules)
	if fileIconRules == "" {
		fileIconRules = defaultFileIconRulesJSON()
	}
	if shouldUpgradeFileIconRules(fileIconRules) {
		fileIconRules = defaultFileIconRulesJSON()
	}
	normalizedFileIconRules, err := normalizeFileIconRulesJSON(fileIconRules)
	if err != nil {
		return nil, err
	}

	emojiOptions := strings.TrimSpace(payload.EmojiOptions)
	if emojiOptions == "" || shouldUpgradeEmojiOptions(emojiOptions) {
		emojiOptions = defaultEmojiOptionsJSON()
	}
	if !json.Valid([]byte(emojiOptions)) {
		return nil, fmt.Errorf("emoji options must be valid json")
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.FileSystemSetting{}
	}
	setting.FileIconRules = normalizedFileIconRules
	setting.EmojiOptions = emojiOptions

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return s.Get()
}

func (s *fileSystemSettingService) ClearBlobURLCache() error {
	if s.redisMultipart == nil {
		return nil
	}

	return s.redisMultipart.ClearPresignedURLCaches(context.Background())
}

func (s *fileSystemSettingService) GetBrowserApps() ([]BrowserAppGroupPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return parseBrowserAppsJSON("", true)
	}

	return parseBrowserAppsJSON(setting.BrowserApps, true)
}

func (s *fileSystemSettingService) ResolveBrowserApp(filename, mimeType string, platform BrowserAppPlatform) (*BrowserAppResolvedPayload, error) {
	groups, err := s.GetBrowserApps()
	if err != nil {
		return nil, err
	}

	return ResolveBrowserAppFromGroups(groups, filename, mimeType, platform), nil
}

func toFileSystemSettingPayload(setting *model.FileSystemSetting) *FileSystemSettingPayload {
	return &FileSystemSettingPayload{
		OnlineEditorSize:             setting.OnlineEditorSize,
		OnlineEditorUnit:             setting.OnlineEditorUnit,
		RecycleScanInterval:          setting.RecycleScanInterval,
		BlobRecycleInterval:          setting.BlobRecycleInterval,
		StaticCacheTTL:               setting.StaticCacheTTL,
		ListPaginationMode:           setting.ListPaginationMode,
		MaxPageSize:                  setting.MaxPageSize,
		MaxBatchActionSize:           setting.MaxBatchActionSize,
		MaxRecursiveSearch:           setting.MaxRecursiveSearch,
		MapProvider:                  setting.MapProvider,
		MimeMap:                      setting.MimeMap,
		ImageQuery:                   setting.ImageQuery,
		VideoQuery:                   setting.VideoQuery,
		AudioQuery:                   setting.AudioQuery,
		DocumentQuery:                setting.DocumentQuery,
		FileIconRules:                setting.FileIconRules,
		EmojiOptions:                 setting.EmojiOptions,
		BrowserApps:                  setting.BrowserApps,
		CustomProperties:             setting.CustomProperties,
		MasterKeyStorage:             setting.MasterKeyStorage,
		ShowEncryptionStatus:         setting.ShowEncryptionStatus,
		EnableEventPush:              setting.EnableEventPush,
		OfflineTTL:                   setting.OfflineTTL,
		DebounceDelay:                setting.DebounceDelay,
		ServerSideDownloadSessionTTL: setting.ServerSideDownloadSessionTTL,
		UploadSessionTTL:             setting.UploadSessionTTL,
		SlaveAPISignTTL:              setting.SlaveAPISignTTL,
		DirectoryStatTTL:             setting.DirectoryStatTTL,
		MaxChunkRetry:                setting.MaxChunkRetry,
		CacheChunksForRetry:          setting.CacheChunksForRetry,
		TransferParallelism:          setting.TransferParallelism,
		OAuthRefreshInterval:         setting.OAuthRefreshInterval,
		WOPISessionTTL:               setting.WOPISessionTTL,
		BlobSignedURLTTL:             setting.BlobSignedURLTTL,
		BlobSignedURLReuseTTL:        setting.BlobSignedURLReuseTTL,
	}
}

func defaultFileSystemSettingPayload() *FileSystemSettingPayload {
	return &FileSystemSettingPayload{
		OnlineEditorSize:             50,
		OnlineEditorUnit:             "MB",
		RecycleScanInterval:          "@every 33m",
		BlobRecycleInterval:          "@every 15m",
		StaticCacheTTL:               86400,
		ListPaginationMode:           "cursor",
		MaxPageSize:                  2000,
		MaxBatchActionSize:           3000,
		MaxRecursiveSearch:           65535,
		MapProvider:                  "osm-leaflet",
		MimeMap:                      defaultMimeMapJSON(),
		ImageQuery:                   defaultImageCategoryQuery(),
		VideoQuery:                   defaultVideoCategoryQuery(),
		AudioQuery:                   defaultAudioCategoryQuery(),
		DocumentQuery:                defaultDocumentCategoryQuery(),
		FileIconRules:                defaultFileIconRulesJSON(),
		EmojiOptions:                 defaultEmojiOptionsJSON(),
		BrowserApps:                  "[]",
		CustomProperties:             defaultCustomPropertiesJSON(),
		MasterKeyStorage:             "database",
		ShowEncryptionStatus:         true,
		EnableEventPush:              true,
		OfflineTTL:                   1209600,
		DebounceDelay:                5,
		ServerSideDownloadSessionTTL: 600,
		UploadSessionTTL:             86400,
		SlaveAPISignTTL:              60,
		DirectoryStatTTL:             300,
		MaxChunkRetry:                5,
		CacheChunksForRetry:          true,
		TransferParallelism:          4,
		OAuthRefreshInterval:         "@every 230h",
		WOPISessionTTL:               36000,
		BlobSignedURLTTL:             3600,
		BlobSignedURLReuseTTL:        600,
	}
}

const legacyDefaultMimeMapJSON = `{".doc":"application/msword",".docx":"application/vnd.openxmlformats-officedocument.wordprocessingml.document",".xls":"application/vnd.ms-excel",".xlsx":"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",".xltx":"application/vnd.openxmlformats-officedocument.spreadsheetml.template",".ppt":"application/vnd.ms-powerpoint",".pptx":"application/vnd.openxmlformats-officedocument.presentationml.presentation",".potx":"application/vnd.openxmlformats-officedocument.presentationml.template",".ppsx":"application/vnd.openxmlformats-officedocument.presentationml.slideshow"}`

const legacyDefaultImageQuery = "type=file&case_folding&use_or&name=*.bmp&name=*.iff&name=*.png&name=*.gif&name=*.jpg&name=*.jpeg&name=*.psd"
const legacyDefaultVideoQuery = "type=file&case_folding&use_or&name=*.mp4&name=*.m3u8&name=*.flv&name=*.avi&name=*.wmv&name=*.mkv&name=*.rm"
const legacyDefaultAudioQuery = "type=file&case_folding&use_or&name=*.mp3&name=*.flac&name=*.ape&name=*.wav&name=*.acc&name=*.ogg&name=*.m4a"
const legacyDefaultDocumentQuery = "type=file&case_folding&use_or&name=*.pdf&name=*.doc&name=*.docx&name=*.ppt&name=*.pptx&name=*.xls&name=*.xlsx"

func defaultMimeMapJSON() string {
	return `{".3gp":"video/3gpp",".7z":"application/x-7z-compressed",".aac":"audio/aac",".aif":"audio/aiff",".aiff":"audio/aiff",".amr":"audio/amr",".ape":"audio/ape",".arw":"image/x-sony-arw",".avif":"image/avif",".avi":"video/x-msvideo",".azw3":"application/vnd.amazon.ebook",".bmp":"image/bmp",".bz2":"application/x-bzip2",".c":"text/x-c",".cpp":"text/x-c++",".cr2":"image/x-canon-cr2",".css":"text/css",".csv":"text/csv",".doc":"application/msword",".docx":"application/vnd.openxmlformats-officedocument.wordprocessingml.document",".dng":"image/x-adobe-dng",".epub":"application/epub+zip",".flac":"audio/flac",".flv":"video/x-flv",".gif":"image/gif",".go":"text/x-go",".gz":"application/gzip",".heic":"image/heic",".heif":"image/heif",".htm":"text/html",".html":"text/html",".iff":"image/x-iff",".jpeg":"image/jpeg",".jpg":"image/jpeg",".js":"application/javascript",".json":"application/json",".m3u8":"application/vnd.apple.mpegurl",".m4a":"audio/mp4",".m4v":"video/x-m4v",".md":"text/markdown",".mkv":"video/x-matroska",".mobi":"application/x-mobipocket-ebook",".mov":"video/quicktime",".mp3":"audio/mpeg",".mp4":"video/mp4",".mpeg":"video/mpeg",".mpg":"video/mpeg",".nef":"image/x-nikon-nef",".odp":"application/vnd.oasis.opendocument.presentation",".ods":"application/vnd.oasis.opendocument.spreadsheet",".odt":"application/vnd.oasis.opendocument.text",".ogg":"audio/ogg",".opus":"audio/ogg",".pdf":"application/pdf",".png":"image/png",".pot":"application/vnd.ms-powerpoint",".potx":"application/vnd.openxmlformats-officedocument.presentationml.template",".pps":"application/vnd.ms-powerpoint",".ppsx":"application/vnd.openxmlformats-officedocument.presentationml.slideshow",".ppt":"application/vnd.ms-powerpoint",".pptx":"application/vnd.openxmlformats-officedocument.presentationml.presentation",".psd":"image/vnd.adobe.photoshop",".py":"text/x-python",".rar":"application/vnd.rar",".raw":"image/x-panasonic-raw",".rm":"application/vnd.rn-realmedia",".rmvb":"application/vnd.rn-realmedia-vbr",".rs":"text/rust",".rtf":"application/rtf",".svg":"image/svg+xml",".tar":"application/x-tar",".tif":"image/tiff",".tiff":"image/tiff",".ts":"video/mp2t",".txt":"text/plain",".wav":"audio/wav",".webm":"video/webm",".webp":"image/webp",".wma":"audio/x-ms-wma",".wmv":"video/x-ms-wmv",".xls":"application/vnd.ms-excel",".xlsx":"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",".xltx":"application/vnd.openxmlformats-officedocument.spreadsheetml.template",".xml":"application/xml",".yaml":"application/yaml",".yml":"application/yaml",".zip":"application/zip"}`
}

func defaultImageCategoryQuery() string {
	return "type=file&case_folding&use_or&name=*.bmp&name=*.iff&name=*.png&name=*.gif&name=*.jpg&name=*.jpeg&name=*.psd&name=*.webp&name=*.svg&name=*.tif&name=*.tiff&name=*.heic&name=*.heif&name=*.avif&name=*.raw&name=*.dng&name=*.cr2&name=*.nef&name=*.arw"
}

func defaultVideoCategoryQuery() string {
	return "type=file&case_folding&use_or&name=*.mp4&name=*.m3u8&name=*.flv&name=*.avi&name=*.wmv&name=*.mkv&name=*.rm&name=*.rmvb&name=*.mov&name=*.webm&name=*.m4v&name=*.3gp&name=*.ts&name=*.mpg&name=*.mpeg"
}

func defaultAudioCategoryQuery() string {
	return "type=file&case_folding&use_or&name=*.mp3&name=*.flac&name=*.ape&name=*.wav&name=*.aac&name=*.ogg&name=*.m4a&name=*.opus&name=*.wma&name=*.aiff&name=*.aif&name=*.amr"
}

func defaultDocumentCategoryQuery() string {
	return "type=file&case_folding&use_or&name=*.pdf&name=*.doc&name=*.docx&name=*.ppt&name=*.pptx&name=*.xls&name=*.xlsx&name=*.txt&name=*.md&name=*.rtf&name=*.csv&name=*.odt&name=*.ods&name=*.odp&name=*.epub&name=*.mobi&name=*.azw3&name=*.json&name=*.xml&name=*.yaml&name=*.yml"
}

func defaultFileIconRulesJSON() string {
	rules := []fileIconRulePayload{
		{Label: "音频文件", Icon: "🎧", Match: "mp3,flac,ape,wav,aac,ogg,m4a", Tint: "#7c3aed"},
		{Label: "视频文件", Icon: "🎞️", Match: "m3u8,mp4,flv,avi,wmv,mkv,rm,rmvb,mov,webm", Tint: "#dc2626"},
		{Label: "图片文件", Icon: "🖼️", Match: "bmp,iff,png,gif,jpg,jpeg,psd,svg,webp,heif,heic,tiff,avif", Tint: "#ea580c"},
		{Label: "RAW 图片", Icon: "📷", Match: "3fr,ari,arw,bay,braw,crw,cr2,cr3,cap,dcs,dcr,dng,drf,eip,erf,fff,gpr,iiq,k25,kdc,mdc,mef,mos,mrw,nef,nrw,obm,orf,pef,ptx,pxn,r3d,raf,raw,rwl,rw2,rwz,sr2,srf,srw,tif,x3f", Tint: "#ef4444"},
		{Label: "PDF 文档", Icon: "📕", Match: "pdf", Tint: "#ef4444"},
		{Label: "Word 文档", Icon: "📘", Match: "doc,docx", Tint: "#3b82f6"},
		{Label: "PowerPoint 演示", Icon: "📙", Match: "ppt,pptx", Tint: "#f97316"},
		{Label: "Excel 表格", Icon: "📗", Match: "xls,xlsx,csv", Tint: "#22c55e"},
		{Label: "文本与网页", Icon: "📄", Match: "txt,md,markdown,html,htm,xml,json,yaml,yml", Tint: "#64748b"},
		{Label: "压缩包", Icon: "🗜️", Match: "zip,gz,xz,tar,rar,7z,bz2", Tint: "#f59e0b"},
		{Label: "程序安装包", Icon: "⚙️", Match: "exe,msi,bat,cmd", Tint: "#1e3a8a"},
		{Label: "Android 安装包", Icon: "🤖", Match: "apk", Tint: "#84cc16"},
		{Label: "Go 源码", Icon: "Go", Match: "go", Tint: "#06b6d4"},
		{Label: "Python 源码", Icon: "Py", Match: "py", Tint: "#3776ab"},
		{Label: "C 源码", Icon: "C", Match: "c,h", Tint: "#84cc16"},
		{Label: "C++ 源码", Icon: "C++", Match: "cpp,cxx,hpp,hxx", Tint: "#ec4899"},
		{Label: "JavaScript / TypeScript", Icon: "JS", Match: "js,jsx,ts,tsx", Tint: "#facc15"},
		{Label: "Rust 源码", Icon: "Rs", Match: "rs", Tint: "#111827"},
		{Label: "电子书", Icon: "📗", Match: "epub,mobi,azw3", Tint: "#65a30d"},
		{Label: "磁力 / 种子", Icon: "🧲", Match: "torrent", Tint: "#6366f1"},
		{Label: "流程图 / 白板", Icon: "✏️", Match: "drawio,dwb,excalidraw", Tint: "#f97316"},
	}
	data, err := json.Marshal(rules)
	if err != nil {
		return "[]"
	}
	return string(data)
}
func defaultCustomPropertiesJSON() string {
	return `[
  {
    "id": 1,
    "key": "description",
    "name": "描述",
    "icon": "📝",
    "type": "text",
    "defaultValue": "",
    "minLength": 0,
    "maxLength": 500
  },
  {
    "id": 2,
    "key": "rating",
    "name": "评级",
    "icon": "★",
    "type": "rating",
    "defaultValue": "0",
    "maxValue": 5
  }
]`
}

func isEmptyJSONArray(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == "[]"
}

func upgradeDefaultMimeMap(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultMimeMapJSON()
	}
	if sameJSONMap(trimmed, legacyDefaultMimeMapJSON) {
		return defaultMimeMapJSON()
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return trimmed
	}
	if len(parsed) <= 10 && parsed[".docx"] != "" && parsed[".xlsx"] != "" && parsed[".pptx"] != "" && parsed[".pdf"] == "" {
		return defaultMimeMapJSON()
	}
	return trimmed
}

func upgradeDefaultCategoryQuery(category, raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultCategoryQuery(category)
	}
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "image":
		if trimmed == legacyDefaultImageQuery {
			return defaultImageCategoryQuery()
		}
	case "video":
		if trimmed == legacyDefaultVideoQuery {
			return defaultVideoCategoryQuery()
		}
	case "audio":
		if trimmed == legacyDefaultAudioQuery || (strings.Contains(trimmed, "name=*.acc") && !strings.Contains(trimmed, "name=*.aac")) {
			return defaultAudioCategoryQuery()
		}
	case "document":
		if trimmed == legacyDefaultDocumentQuery {
			return defaultDocumentCategoryQuery()
		}
	}
	return trimmed
}

func shouldUpgradeFileIconRules(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return true
	}
	if isLegacyMojibakeFileIconRules(trimmed) {
		return true
	}
	if isLegacyVisualDefaultFileIconRules(trimmed) {
		return true
	}
	return false
}

func isLegacyVisualDefaultFileIconRules(raw string) bool {
	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return false
	}
	expected := map[string]fileIconRulePayload{
		"音频文件":                    {Icon: "🎵", Match: "mp3,flac,ape,wav,aac,ogg,m4a"},
		"视频文件":                    {Icon: "🎬", Match: "m3u8,mp4,flv,avi,wmv,mkv,rm,rmvb,mov,webm"},
		"图片文件":                    {Icon: "🖼️", Match: "bmp,iff,png,gif,jpg,jpeg,psd,svg,webp,heif,heic,tiff,avif"},
		"RAW 图片":                  {Icon: "RAW", Match: "3fr,ari,arw,bay,braw,crw,cr2,cr3,cap,dcs,dcr,dng,drf,eip,erf,fff,gpr,iiq,k25,kdc,mdc,mef,mos,mrw,nef,nrw,obm,orf,pef,ptx,pxn,r3d,raf,raw,rwl,rw2,rwz,sr2,srf,srw,tif,x3f"},
		"PDF 文档":                  {Icon: "📕", Match: "pdf"},
		"Word 文档":                 {Icon: "W", Match: "doc,docx"},
		"PowerPoint 演示":           {Icon: "P", Match: "ppt,pptx"},
		"Excel 表格":                {Icon: "X", Match: "xls,xlsx,csv"},
		"文本与网页":                   {Icon: "📝", Match: "txt,md,markdown,html,htm,xml,json,yaml,yml"},
		"压缩包":                     {Icon: "🗜️", Match: "zip,gz,xz,tar,rar,7z,bz2"},
		"程序安装包":                   {Icon: "⚙️", Match: "exe,msi,bat,cmd"},
		"Android 安装包":             {Icon: "🤖", Match: "apk"},
		"Go 源码":                   {Icon: "Go", Match: "go"},
		"Python 源码":               {Icon: "Py", Match: "py"},
		"C 源码":                    {Icon: "C", Match: "c,h"},
		"C++ 源码":                  {Icon: "C++", Match: "cpp,cxx,hpp,hxx"},
		"JavaScript / TypeScript": {Icon: "JS", Match: "js,jsx,ts,tsx"},
		"Rust 源码":                 {Icon: "Rs", Match: "rs"},
		"电子书":                     {Icon: "📗", Match: "epub,mobi,azw3"},
		"磁力 / 种子":                 {Icon: "🧲", Match: "torrent"},
		"流程图 / 白板":                {Icon: "✏️", Match: "drawio,dwb,excalidraw"},
	}
	if len(rules) != len(expected) {
		return false
	}
	matches := 0
	for _, rule := range rules {
		want, ok := expected[strings.TrimSpace(rule.Label)]
		if !ok {
			continue
		}
		if strings.TrimSpace(rule.Icon) != want.Icon || normalizeFileIconRuleMatch(rule.Match) != normalizeFileIconRuleMatch(want.Match) {
			return false
		}
		matches++
	}
	return matches == len(expected)
}

func isLegacyMojibakeFileIconRules(raw string) bool {
	legacyMarkers := []string{
		"\u95ca\u62bd",
		"\u7459\u55db",
		"\u9365\u5267",
		"PDF \u93c2",
		"Word \u93c2",
		"Excel \u741b",
		"Go \u5a67",
		"mp3,flac,ape,wav,aac,ogg,m4a",
		"m3u8,mp4,flv,avi,wmv,mkv,rm,rmvb,mov,webm",
		"drawio,dwb,excalidraw",
	}
	for _, marker := range legacyMarkers {
		if !strings.Contains(raw, marker) {
			return false
		}
	}
	return true
}

func defaultCategoryQuery(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "image":
		return defaultImageCategoryQuery()
	case "video":
		return defaultVideoCategoryQuery()
	case "audio", "music":
		return defaultAudioCategoryQuery()
	case "document":
		return defaultDocumentCategoryQuery()
	default:
		return ""
	}
}

func sameJSONMap(left, right string) bool {
	var leftMap map[string]string
	var rightMap map[string]string
	if err := json.Unmarshal([]byte(left), &leftMap); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(right), &rightMap); err != nil {
		return false
	}
	if len(leftMap) != len(rightMap) {
		return false
	}
	for key, leftValue := range leftMap {
		if rightMap[key] != leftValue {
			return false
		}
	}
	return true
}

func normalizeFileSystemSettingPayload(payload *FileSystemSettingPayload) (*FileSystemSettingPayload, error) {
	normalized := *payload
	normalized.OnlineEditorUnit = strings.ToUpper(strings.TrimSpace(normalized.OnlineEditorUnit))
	normalized.RecycleScanInterval = strings.TrimSpace(normalized.RecycleScanInterval)
	normalized.BlobRecycleInterval = strings.TrimSpace(normalized.BlobRecycleInterval)
	normalized.ListPaginationMode = strings.ToLower(strings.TrimSpace(normalized.ListPaginationMode))
	normalized.MapProvider = strings.ToLower(strings.TrimSpace(normalized.MapProvider))
	normalized.MimeMap = strings.TrimSpace(normalized.MimeMap)
	normalized.ImageQuery = strings.TrimSpace(normalized.ImageQuery)
	normalized.VideoQuery = strings.TrimSpace(normalized.VideoQuery)
	normalized.AudioQuery = strings.TrimSpace(normalized.AudioQuery)
	normalized.DocumentQuery = strings.TrimSpace(normalized.DocumentQuery)
	normalized.MimeMap = upgradeDefaultMimeMap(normalized.MimeMap)
	normalized.ImageQuery = upgradeDefaultCategoryQuery("image", normalized.ImageQuery)
	normalized.VideoQuery = upgradeDefaultCategoryQuery("video", normalized.VideoQuery)
	normalized.AudioQuery = upgradeDefaultCategoryQuery("audio", normalized.AudioQuery)
	normalized.DocumentQuery = upgradeDefaultCategoryQuery("document", normalized.DocumentQuery)
	normalized.FileIconRules = strings.TrimSpace(normalized.FileIconRules)
	normalized.EmojiOptions = strings.TrimSpace(normalized.EmojiOptions)
	normalized.BrowserApps = strings.TrimSpace(normalized.BrowserApps)
	normalized.CustomProperties = strings.TrimSpace(normalized.CustomProperties)
	normalized.MasterKeyStorage = strings.ToLower(strings.TrimSpace(normalized.MasterKeyStorage))
	normalized.OAuthRefreshInterval = strings.TrimSpace(normalized.OAuthRefreshInterval)

	switch normalized.OnlineEditorUnit {
	case "B", "KB", "MB", "GB", "TB":
	default:
		return nil, fmt.Errorf("unsupported online editor size unit")
	}

	switch normalized.ListPaginationMode {
	case "cursor", "offset", "hybrid":
	default:
		return nil, fmt.Errorf("unsupported list pagination mode")
	}

	switch normalized.MapProvider {
	case "google-leaflet", "osm-leaflet", "osm-mapbox":
	default:
		return nil, fmt.Errorf("unsupported map provider")
	}

	switch normalized.MasterKeyStorage {
	case "database", "file", "env":
	default:
		return nil, fmt.Errorf("unsupported master key storage mode")
	}

	if normalized.RecycleScanInterval == "" {
		return nil, fmt.Errorf("recycle scan interval cannot be empty")
	}
	if normalized.BlobRecycleInterval == "" {
		return nil, fmt.Errorf("blob recycle interval cannot be empty")
	}
	if normalized.MimeMap == "" {
		return nil, fmt.Errorf("mime map cannot be empty")
	}
	if normalized.ImageQuery == "" || normalized.VideoQuery == "" || normalized.AudioQuery == "" || normalized.DocumentQuery == "" {
		return nil, fmt.Errorf("category query cannot be empty")
	}
	if normalized.FileIconRules == "" {
		normalized.FileIconRules = defaultFileIconRulesJSON()
	}
	if shouldUpgradeFileIconRules(normalized.FileIconRules) {
		normalized.FileIconRules = defaultFileIconRulesJSON()
	}
	if normalized.EmojiOptions == "" {
		normalized.EmojiOptions = defaultEmojiOptionsJSON()
	}
	if shouldUpgradeEmojiOptions(normalized.EmojiOptions) {
		normalized.EmojiOptions = defaultEmojiOptionsJSON()
	}
	if normalized.BrowserApps == "" {
		normalized.BrowserApps = "[]"
	}
	if normalized.CustomProperties == "" {
		normalized.CustomProperties = "[]"
	}
	if normalized.OAuthRefreshInterval == "" {
		return nil, fmt.Errorf("oauth refresh interval cannot be empty")
	}
	normalizedFileIconRules, err := normalizeFileIconRulesJSON(normalized.FileIconRules)
	if err != nil {
		return nil, err
	}
	normalized.FileIconRules = normalizedFileIconRules
	if !json.Valid([]byte(normalized.EmojiOptions)) {
		return nil, fmt.Errorf("emoji options must be valid json")
	}
	if !json.Valid([]byte(normalized.BrowserApps)) {
		return nil, fmt.Errorf("browser apps must be valid json")
	}
	if !json.Valid([]byte(normalized.CustomProperties)) {
		return nil, fmt.Errorf("custom properties must be valid json")
	}
	browserApps, err := parseBrowserAppsJSON(normalized.BrowserApps, false)
	if err != nil {
		return nil, err
	}
	browserAppsJSON, err := json.Marshal(browserApps)
	if err != nil {
		return nil, fmt.Errorf("browser apps normalization failed: %w", err)
	}
	normalized.BrowserApps = string(browserAppsJSON)
	customPropertiesJSON, err := normalizeCustomPropertiesJSON(normalized.CustomProperties)
	if err != nil {
		return nil, err
	}
	normalized.CustomProperties = customPropertiesJSON

	normalized.OnlineEditorSize = clampInt(normalized.OnlineEditorSize, 1, 102400)
	normalized.StaticCacheTTL = clampInt(normalized.StaticCacheTTL, 0, 31536000)
	normalized.MaxPageSize = clampInt(normalized.MaxPageSize, 1, 10000)
	normalized.MaxBatchActionSize = clampInt(normalized.MaxBatchActionSize, 1, 10000)
	normalized.MaxRecursiveSearch = clampInt(normalized.MaxRecursiveSearch, 1, 1000000)
	normalized.OfflineTTL = clampInt(normalized.OfflineTTL, 0, 31536000)
	normalized.DebounceDelay = clampInt(normalized.DebounceDelay, 0, 3600)
	normalized.ServerSideDownloadSessionTTL = clampInt(normalized.ServerSideDownloadSessionTTL, 1, 86400)
	normalized.UploadSessionTTL = clampInt(normalized.UploadSessionTTL, 1, 2592000)
	normalized.SlaveAPISignTTL = clampInt(normalized.SlaveAPISignTTL, 1, 86400)
	normalized.DirectoryStatTTL = clampInt(normalized.DirectoryStatTTL, 0, 86400)
	normalized.MaxChunkRetry = clampInt(normalized.MaxChunkRetry, 0, 50)
	normalized.TransferParallelism = clampInt(normalized.TransferParallelism, 1, 64)
	normalized.WOPISessionTTL = clampInt(normalized.WOPISessionTTL, 1, 2592000)
	normalized.BlobSignedURLTTL = clampInt(normalized.BlobSignedURLTTL, 1, 86400)
	normalized.BlobSignedURLReuseTTL = clampInt(normalized.BlobSignedURLReuseTTL, 0, normalized.BlobSignedURLTTL)

	return &normalized, nil
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

type fileIconRulePayload struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
	Match string `json:"match"`
	Tint  string `json:"tint"`
}

func normalizeFileIconRulesJSON(raw string) (string, error) {
	if !json.Valid([]byte(raw)) {
		return "", fmt.Errorf("file icon rules must be valid json")
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		var probe interface{}
		if probeErr := json.Unmarshal([]byte(raw), &probe); probeErr == nil {
			if _, ok := probe.([]interface{}); !ok {
				return "", fmt.Errorf("file icon rules must be a json array")
			}
		}
		return "", fmt.Errorf("file icon rules contain invalid item fields: %w", err)
	}

	normalized := make([]fileIconRulePayload, 0, len(rules))
	for index, rule := range rules {
		item := fileIconRulePayload{
			Label: strings.TrimSpace(rule.Label),
			Icon:  strings.TrimSpace(rule.Icon),
			Match: normalizeFileIconRuleMatch(rule.Match),
			Tint:  strings.TrimSpace(rule.Tint),
		}
		if item.Tint != "" && !fileIconTintPattern.MatchString(item.Tint) {
			return "", fmt.Errorf("file icon rule %d tint must be #RGB, #RRGGBB, or empty", index+1)
		}
		normalized = append(normalized, item)
	}

	result, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("file icon rules normalization failed: %w", err)
	}
	return string(result), nil
}

func normalizeFileIconRuleMatch(raw string) string {
	seen := make(map[string]struct{})
	tokens := make([]string, 0)
	for _, token := range splitFileIconRuleMatch(raw) {
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	return strings.Join(tokens, ",")
}

func splitFileIconRuleMatch(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '，', ';', '；', '|':
			return true
		default:
			return r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}
	})

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.ToLower(strings.TrimSpace(part))
		token = strings.TrimPrefix(token, ".")
		if token != "" {
			result = append(result, token)
		}
	}
	return result
}

type customPropertyPayload struct {
	ID           int      `json:"id"`
	Key          string   `json:"key"`
	Name         string   `json:"name"`
	Icon         string   `json:"icon"`
	Type         string   `json:"type"`
	MinLength    *int     `json:"minLength"`
	MaxLength    *int     `json:"maxLength"`
	MaxValue     *int     `json:"maxValue"`
	Options      []string `json:"options"`
	DefaultValue string   `json:"defaultValue"`
}

func normalizeCustomPropertiesJSON(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultCustomPropertiesJSON(), nil
	}

	var parsed []customPropertyPayload
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", fmt.Errorf("custom properties normalization failed: %w", err)
	}
	if len(parsed) == 0 {
		return defaultCustomPropertiesJSON(), nil
	}

	normalized := make([]customPropertyPayload, 0, len(parsed))
	for index, item := range parsed {
		propertyType := strings.ToLower(strings.TrimSpace(item.Type))
		switch propertyType {
		case "rating", "switch", "date", "tags", "multi_select":
		default:
			propertyType = "text"
		}

		normalizedItem := customPropertyPayload{
			ID:           item.ID,
			Key:          strings.TrimSpace(item.Key),
			Name:         strings.TrimSpace(item.Name),
			Icon:         strings.TrimSpace(item.Icon),
			Type:         propertyType,
			DefaultValue: strings.TrimSpace(item.DefaultValue),
		}
		if normalizedItem.ID <= 0 {
			normalizedItem.ID = index + 1
		}
		if normalizedItem.Key == "" {
			normalizedItem.Key = fmt.Sprintf("property_%d", index+1)
		}
		if normalizedItem.Name == "" {
			normalizedItem.Name = fmt.Sprintf("属性 %d", index+1)
		}

		if propertyType == "rating" {
			maxValue := 5
			if item.MaxValue != nil {
				maxValue = clampInt(*item.MaxValue, 1, 10)
			}
			normalizedItem.MaxValue = &maxValue
			defaultValue := clampInt(parseOptionalInt(normalizedItem.DefaultValue, 0), 0, maxValue)
			normalizedItem.DefaultValue = fmt.Sprintf("%d", defaultValue)
		} else if propertyType == "text" {
			if item.MinLength != nil {
				minLength := clampInt(*item.MinLength, 0, 10000)
				normalizedItem.MinLength = &minLength
			}
			if item.MaxLength != nil {
				maxLength := clampInt(*item.MaxLength, 0, 10000)
				normalizedItem.MaxLength = &maxLength
			}
			if normalizedItem.MinLength != nil && normalizedItem.MaxLength != nil && *normalizedItem.MinLength > *normalizedItem.MaxLength {
				maxLength := *normalizedItem.MinLength
				normalizedItem.MaxLength = &maxLength
			}
		} else if propertyType == "switch" {
			if normalizedItem.DefaultValue != "true" {
				normalizedItem.DefaultValue = "false"
			}
		} else if propertyType == "date" {
			normalizedItem.DefaultValue = strings.TrimSpace(normalizedItem.DefaultValue)
		} else {
			normalizedItem.Options = normalizeCustomPropertyOptions(item.Options)
			normalizedItem.DefaultValue = normalizeCustomPropertyArrayValue(normalizedItem.DefaultValue, normalizedItem.Options)
		}

		normalized = append(normalized, normalizedItem)
	}

	normalized = ensureDefaultCustomPropertyItems(normalized)

	result, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("custom properties normalization failed: %w", err)
	}
	return string(result), nil
}

func ensureDefaultCustomPropertyItems(items []customPropertyPayload) []customPropertyPayload {
	next := append([]customPropertyPayload{}, items...)
	hasDescription := false
	hasRating := false
	maxID := 0

	for _, item := range next {
		if item.ID > maxID {
			maxID = item.ID
		}

		lowerKey := strings.ToLower(strings.TrimSpace(item.Key))
		lowerName := strings.ToLower(strings.TrimSpace(item.Name))
		if lowerKey == "description" || lowerName == "描述" {
			hasDescription = true
		}
		if item.Type == "rating" || lowerKey == "rating" || lowerName == "评级" || lowerName == "评分" {
			hasRating = true
		}
	}

	if !hasDescription {
		maxID++
		minLength := 0
		maxLength := 500
		next = append([]customPropertyPayload{
			{
				ID:           maxID,
				Key:          "description",
				Name:         "描述",
				Icon:         "📝",
				Type:         "text",
				MinLength:    &minLength,
				MaxLength:    &maxLength,
				DefaultValue: "",
			},
		}, next...)
	}

	if !hasRating {
		maxID++
		maxValue := 5
		next = append(next, customPropertyPayload{
			ID:           maxID,
			Key:          "rating",
			Name:         "评级",
			Icon:         "★",
			Type:         "rating",
			MaxValue:     &maxValue,
			DefaultValue: "0",
		})
	}

	return next
}

func parseOptionalInt(value string, fallback int) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(trimmed, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}

func normalizeCustomPropertyOptions(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func normalizeCustomPropertyArrayValue(raw string, options []string) string {
	var parsed []string
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &parsed)
	}

	allowed := map[string]struct{}{}
	for _, item := range options {
		allowed[item] = struct{}{}
	}

	result := make([]string, 0, len(parsed))
	seen := map[string]struct{}{}
	for _, item := range parsed {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[trimmed]; !ok {
				continue
			}
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "[]"
	}
	return string(data)
}
