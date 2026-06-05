// 路径: internal/middleware/file_validator.go
package middleware

import (
	"mime/multipart"
	"net/http"
	"strings"

	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AllowedExtensions 允许的文件扩展名白名单
var AllowedExtensions = []string{
	// 图片
	".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".ico",
	// 视频
	".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v",
	// 音频
	".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a",
	// 文档
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".txt", ".md", ".csv", ".rtf", ".odt", ".ods", ".odp",
	// 压缩包
	".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz",
	// 代码
	".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".hpp",
	".html", ".css", ".json", ".xml", ".yaml", ".yml", ".sql",
}

// AllowedMimeTypes 允许的 MIME 类型映射
var AllowedMimeTypes = map[string][]string{
	// 图片
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".png":  {"image/png"},
	".gif":  {"image/gif"},
	".webp": {"image/webp"},
	".bmp":  {"image/bmp"},
	".svg":  {"image/svg+xml"},
	".ico":  {"image/x-icon", "image/vnd.microsoft.icon"},
	// 视频
	".mp4":  {"video/mp4"},
	".avi":  {"video/x-msvideo"},
	".mkv":  {"video/x-matroska"},
	".mov":  {"video/quicktime"},
	".wmv":  {"video/x-ms-wmv"},
	".flv":  {"video/x-flv"},
	".webm": {"video/webm"},
	".m4v":  {"video/x-m4v"},
	// 音频
	".mp3":  {"audio/mpeg"},
	".wav":  {"audio/wav", "audio/x-wav"},
	".flac": {"audio/flac"},
	".aac":  {"audio/aac"},
	".ogg":  {"audio/ogg"},
	".wma":  {"audio/x-ms-wma"},
	".m4a":  {"audio/mp4", "audio/x-m4a"},
	// 文档
	".pdf":  {"application/pdf"},
	".doc":  {"application/msword"},
	".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	".xls":  {"application/vnd.ms-excel"},
	".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	".ppt":  {"application/vnd.ms-powerpoint"},
	".pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	".txt":  {"text/plain"},
	".md":   {"text/markdown", "text/plain"},
	".csv":  {"text/csv"},
	".rtf":  {"application/rtf"},
	".odt":  {"application/vnd.oasis.opendocument.text"},
	".ods":  {"application/vnd.oasis.opendocument.spreadsheet"},
	".odp":  {"application/vnd.oasis.opendocument.presentation"},
	// 压缩包
	".zip": {"application/zip"},
	".rar": {"application/x-rar-compressed", "application/vnd.rar"},
	".7z":  {"application/x-7z-compressed"},
	".tar": {"application/x-tar"},
	".gz":  {"application/gzip"},
	".bz2": {"application/x-bzip2"},
	".xz":  {"application/x-xz"},
	// 代码
	".go":   {"text/plain"},
	".js":   {"application/javascript", "text/javascript"},
	".ts":   {"application/typescript", "text/plain"},
	".py":   {"text/x-python", "text/plain"},
	".java": {"text/x-java-source", "text/plain"},
	".c":    {"text/x-c", "text/plain"},
	".cpp":  {"text/x-c++", "text/plain"},
	".h":    {"text/x-c", "text/plain"},
	".hpp":  {"text/x-c++", "text/plain"},
	".html": {"text/html"},
	".css":  {"text/css"},
	".json": {"application/json"},
	".xml":  {"application/xml", "text/xml"},
	".yaml": {"application/x-yaml", "text/yaml"},
	".yml":  {"application/x-yaml", "text/yaml"},
	".sql":  {"application/sql", "text/plain"},
}

const (
	// MaxFileSize 最大文件大小 10GB
	MaxFileSize = 10 * 1024 * 1024 * 1024 // 10GB in bytes
)

// FileValidationMiddleware 文件验证中间件
// 验证文件扩展名、MIME 类型和文件大小
func FileValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			// 如果没有文件，继续执行（可能是其他类型的请求）
			c.Next()
			return
		}

		// 验证文件
		if err := validateFile(file, c); err != nil {
			// 记录恶意文件上传尝试
			userID, exists := c.Get("user_id")
			if exists {
				logger.Warn("文件验证失败",
					zap.Uint("user_id", userID.(uint)),
					zap.String("filename", file.Filename),
					zap.Int64("size", file.Size),
					zap.String("error", err.Error()))
			} else {
				logger.Warn("文件验证失败（未认证用户）",
					zap.String("filename", file.Filename),
					zap.Int64("size", file.Size),
					zap.String("error", err.Error()))
			}

			response.Error(c, http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateFile 验证文件
func validateFile(file *multipart.FileHeader, c *gin.Context) error {
	// 1. 检查文件大小
	if file.Size > MaxFileSize {
		return &ValidationError{Message: "文件大小超过 10GB 限制"}
	}

	// 2. 检查文件扩展名
	filename := strings.ToLower(file.Filename)
	ext := getFileExtension(filename)
	if !isExtensionAllowed(ext) {
		return &ValidationError{Message: "不支持的文件类型"}
	}

	// 3. 验证 MIME 类型
	contentType := file.Header.Get("Content-Type")
	if !isMimeTypeValid(ext, contentType) {
		return &ValidationError{Message: "文件 MIME 类型与扩展名不匹配"}
	}

	return nil
}

// validateMultipartFile 验证分片上传文件信息
func validateMultipartFile(filename string, fileSize int64) error {
	// 1. 检查文件大小
	if fileSize > MaxFileSize {
		return &ValidationError{Message: "文件大小超过 10GB 限制"}
	}

	// 2. 检查文件扩展名
	filename = strings.ToLower(filename)
	ext := getFileExtension(filename)
	if !isExtensionAllowed(ext) {
		return &ValidationError{Message: "不支持的文件类型"}
	}

	return nil
}

// ValidateFileName 公开的文件名验证函数（供控制器使用）
func ValidateFileName(filename string) error {
	filename = strings.ToLower(filename)
	ext := getFileExtension(filename)
	if !isExtensionAllowed(ext) {
		return &ValidationError{Message: "不支持的文件类型"}
	}
	return nil
}

// ValidateFileSize 公开的文件大小验证函数（供控制器使用）
func ValidateFileSize(fileSize int64) error {
	if fileSize > MaxFileSize {
		return &ValidationError{Message: "文件大小超过 10GB 限制"}
	}
	return nil
}

// getFileExtension 获取文件扩展名
func getFileExtension(filename string) string {
	// 查找最后一个点的位置
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return ""
	}
	return filename[lastDot:]
}

// isExtensionAllowed 检查扩展名是否在白名单中
func isExtensionAllowed(ext string) bool {
	for _, allowed := range AllowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// isMimeTypeValid 验证 MIME 类型是否与扩展名匹配
func isMimeTypeValid(ext, contentType string) bool {
	// 如果没有 Content-Type，跳过 MIME 验证
	if contentType == "" {
		return true
	}

	// 获取该扩展名允许的 MIME 类型列表
	allowedMimes, exists := AllowedMimeTypes[ext]
	if !exists {
		// 如果扩展名在白名单中但没有定义 MIME 类型，允许通过
		return true
	}

	// 检查 Content-Type 是否在允许的列表中
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	// 处理带参数的 Content-Type，如 "text/plain; charset=utf-8"
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	for _, allowed := range allowedMimes {
		if contentType == allowed {
			return true
		}
	}

	return false
}

// ValidationError 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
