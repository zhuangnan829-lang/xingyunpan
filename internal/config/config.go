// 路径: internal/config/config.go
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
var Config *AppConfig

// AppConfig 应用配置结构
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Email    EmailConfig    `mapstructure:"email"`
	Log      LogConfig      `mapstructure:"log"`
	Storage  StorageConfig  `mapstructure:"storage"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Upload   UploadConfig   `mapstructure:"upload"`
	User     UserConfig     `mapstructure:"user"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	BaseURL      string `mapstructure:"base_url"` // Phase 5: 分享功能需要的基础 URL
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	LogLevel        string `mapstructure:"log_level"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type QueueConfig struct {
	EmbeddedRunnerEnabled *bool `mapstructure:"embedded_runner_enabled"`
}

type WorkerConfig struct {
	Enabled *bool `mapstructure:"enabled"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	Provider            string `mapstructure:"provider"`
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	FromName            string `mapstructure:"from_name"`
	FromAddress         string `mapstructure:"from_address"`
	CodeTTLSeconds      int    `mapstructure:"code_ttl_seconds"`
	SendIntervalSeconds int    `mapstructure:"send_interval_seconds"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string      `mapstructure:"type"`
	BasePath string      `mapstructure:"base_path"`
	MinIO    MinIOConfig `mapstructure:"minio"`
	OSS      OSSConfig   `mapstructure:"oss"`
}

// MinIOConfig MinIO 配置
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// OSSConfig 阿里云 OSS 配置
type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Bucket          string `mapstructure:"bucket"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxFileSize       int64    `mapstructure:"max_file_size"`
	ChunkSize         int64    `mapstructure:"chunk_size"`
	AllowedExtensions []string `mapstructure:"allowed_extensions"`
	TempDir           string   `mapstructure:"temp_dir"`
}

// UserConfig 用户配置
type UserConfig struct {
	DefaultCapacity int64 `mapstructure:"default_capacity"`
	MaxCapacity     int64 `mapstructure:"max_capacity"`
}

// Load 加载配置文件
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&Config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	applyRuntimeDefaults(Config)
	return nil
}

// LoadDefault 加载默认配置文件
func LoadDefault() error {
	return Load("configs/config.yaml")
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
	)
}

// GetAddr 获取 Redis 地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetDialTimeout 获取 Redis 连接超时时间
func (c *RedisConfig) GetDialTimeout() time.Duration {
	return time.Duration(c.DialTimeout) * time.Second
}

// GetReadTimeout 获取 Redis 读取超时时间
func (c *RedisConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取 Redis 写入超时时间
func (c *RedisConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetConnMaxLifetime 获取数据库连接最大生命周期
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetReadTimeout 获取服务器读取超时时间
func (c *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取服务器写入超时时间
func (c *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetAddr 获取服务器监听地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func applyRuntimeDefaults(cfg *AppConfig) {
	if cfg == nil {
		return
	}
	if cfg.Queue.EmbeddedRunnerEnabled == nil {
		enabled := cfg.Server.Mode != "release"
		cfg.Queue.EmbeddedRunnerEnabled = &enabled
	}
	if cfg.Worker.Enabled == nil {
		enabled := true
		cfg.Worker.Enabled = &enabled
	}
}

func (c *QueueConfig) IsEmbeddedRunnerEnabled() bool {
	return c != nil && c.EmbeddedRunnerEnabled != nil && *c.EmbeddedRunnerEnabled
}

func (c *WorkerConfig) IsEnabled() bool {
	return c == nil || c.Enabled == nil || *c.Enabled
}
