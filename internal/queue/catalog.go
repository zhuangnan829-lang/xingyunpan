package queue

// Key identifies one logical admin queue.
type Key string

const (
	KeyMetadata  Key = "metadata"
	KeyBlob      Key = "blob"
	KeyIO        Key = "io"
	KeyOffline   Key = "offline"
	KeyThumbnail Key = "thumbnail"
)

// Setting is the normalized queue configuration shared by settings APIs and workers.
type Setting struct {
	QueueKey      string
	WorkerNum     int
	MaxExecution  int
	BackoffFactor int
	MaxBackoff    int
	MaxRetry      int
	RetryDelay    int
}

// Definition describes one logical queue in the current system.
type Definition struct {
	Key              Key
	Title            string
	Description      string
	Implemented      bool
	WorkerControlled bool
}

var definitions = []Definition{
	{Key: KeyMetadata, Title: "媒体元数据提取", Description: "用于提取媒体文件的元数据。", Implemented: true, WorkerControlled: true},
	{Key: KeyBlob, Title: "Blob 回收", Description: "用于删除过期的文件 Blob。", Implemented: true, WorkerControlled: true},
	{Key: KeyIO, Title: "IO 密集型", Description: "用于统一队列 runner 执行的 IO 后台任务，包括 multipart.cleanup、fulltext.rebuild、archive.create 与 archive.extract。", Implemented: true, WorkerControlled: true},
	{Key: KeyOffline, Title: "离线下载", Description: "用于处理离线下载任务。", Implemented: true, WorkerControlled: true},
	{Key: KeyThumbnail, Title: "缩略图生成", Description: "用于为文件生成缩略图。", Implemented: true, WorkerControlled: true},
}

var defaultSettings = []Setting{
	{QueueKey: string(KeyMetadata), WorkerNum: 30, MaxExecution: 3600, BackoffFactor: 2, MaxBackoff: 60, MaxRetry: 1, RetryDelay: 0},
	{QueueKey: string(KeyBlob), WorkerNum: 5, MaxExecution: 900, BackoffFactor: 2, MaxBackoff: 60, MaxRetry: 0, RetryDelay: 0},
	{QueueKey: string(KeyIO), WorkerNum: 30, MaxExecution: 2592000, BackoffFactor: 2, MaxBackoff: 600, MaxRetry: 5, RetryDelay: 0},
	{QueueKey: string(KeyOffline), WorkerNum: 5, MaxExecution: 864000, BackoffFactor: 2, MaxBackoff: 600, MaxRetry: 5, RetryDelay: 0},
	{QueueKey: string(KeyThumbnail), WorkerNum: 15, MaxExecution: 300, BackoffFactor: 2, MaxBackoff: 60, MaxRetry: 0, RetryDelay: 0},
}

// Definitions returns all logical queue definitions.
func Definitions() []Definition {
	items := make([]Definition, len(definitions))
	copy(items, definitions)
	return items
}

// DefaultSettings returns all default queue settings.
func DefaultSettings() []Setting {
	items := make([]Setting, len(defaultSettings))
	copy(items, defaultSettings)
	return items
}

// DefaultSettingByKey returns a default queue setting by queue key.
func DefaultSettingByKey(queueKey string) (Setting, bool) {
	for _, item := range defaultSettings {
		if item.QueueKey == queueKey {
			return item, true
		}
	}

	return Setting{}, false
}
