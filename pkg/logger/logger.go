// 路径: pkg/logger/logger.go
package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *zap.Logger

// Config 日志配置
type Config struct {
	Level      string // 日志级别: debug, info, warn, error
	Format     string // 日志格式: console, json
	Output     string // 输出位置: stdout, stderr, 文件路径
	FilePath   string // 日志文件路径
	MaxSize    int    // 单个日志文件最大大小（MB）
	MaxBackups int    // 保留的旧日志文件最大数量
	MaxAge     int    // 保留旧日志文件的最大天数
	Compress   bool   // 是否压缩旧日志文件
}

// RequestContext 请求上下文信息
type RequestContext struct {
	RequestID string
	UserID    uint
	Method    string
	Path      string
	Status    int
	Duration  float64
}

// Init 初始化日志
func Init(cfg *Config) error {
	// 设置日志级别
	level := parseLevel(cfg.Level)

	// 设置编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色日志级别
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据格式选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		// JSON 格式：使用标准配置便于日志聚合工具解析
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 不使用彩色
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // ISO8601 时间格式
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置输出位置
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	default:
		// 输出到文件，使用 lumberjack 进行日志轮转
		writeSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		})
	}

	// 创建 Core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// customTimeEncoder 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// Debug 输出 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 输出 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 输出 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 输出 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 输出 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// WithRequestContext 创建带请求上下文的日志记录器
func WithRequestContext(ctx *RequestContext) *zap.Logger {
	if ctx == nil {
		return Logger
	}

	fields := []zap.Field{}

	if ctx.RequestID != "" {
		fields = append(fields, zap.String("request_id", ctx.RequestID))
	}
	if ctx.UserID > 0 {
		fields = append(fields, zap.Uint("user_id", ctx.UserID))
	}
	if ctx.Method != "" {
		fields = append(fields, zap.String("method", ctx.Method))
	}
	if ctx.Path != "" {
		fields = append(fields, zap.String("path", ctx.Path))
	}
	if ctx.Status > 0 {
		fields = append(fields, zap.Int("status", ctx.Status))
	}
	if ctx.Duration > 0 {
		fields = append(fields, zap.Float64("duration", ctx.Duration))
	}

	return Logger.With(fields...)
}

// InfoWithContext 输出带上下文的 Info 日志
func InfoWithContext(ctx *RequestContext, msg string, fields ...zap.Field) {
	WithRequestContext(ctx).Info(msg, fields...)
}

// ErrorWithContext 输出带上下文的 Error 日志（包含 stack trace）
func ErrorWithContext(ctx *RequestContext, msg string, fields ...zap.Field) {
	WithRequestContext(ctx).Error(msg, fields...)
}

// WarnWithContext 输出带上下文的 Warn 日志
func WarnWithContext(ctx *RequestContext, msg string, fields ...zap.Field) {
	WithRequestContext(ctx).Warn(msg, fields...)
}
