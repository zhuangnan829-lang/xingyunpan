// 路径: internal/config/database.go
package config

import (
	"context"
	"fmt"
	"time"

	"xingyunpan-v2/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	var err error
	var gormLogger gormlogger.Interface

	// 根据配置设置 GORM 日志级别
	switch Config.Database.LogLevel {
	case "silent":
		gormLogger = gormlogger.Default.LogMode(gormlogger.Silent)
	case "error":
		gormLogger = gormlogger.Default.LogMode(gormlogger.Error)
	case "warn":
		gormLogger = gormlogger.Default.LogMode(gormlogger.Warn)
	case "info":
		gormLogger = gormlogger.Default.LogMode(gormlogger.Info)
	default:
		gormLogger = gormlogger.Default.LogMode(gormlogger.Info)
	}

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(Config.Database.GetDSN()), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(Config.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(Config.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(Config.Database.GetConnMaxLifetime())

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	logger.Info("数据库连接成功",
		zap.String("host", Config.Database.Host),
		zap.String("database", Config.Database.Database),
		zap.Int("max_idle_conns", Config.Database.MaxIdleConns),
		zap.Int("max_open_conns", Config.Database.MaxOpenConns),
	)

	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	logger.Info("数据库连接已关闭")
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if err := DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.Info("数据库迁移成功", zap.Int("models_count", len(models)))
	return nil
}

// MigrateAll 迁移所有模型
func MigrateAll() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 导入模型包
	// 注意：这里需要在实际使用时导入 model 包
	// 为了避免循环依赖，这个函数应该在 main.go 中调用
	// 并传入所有需要迁移的模型

	logger.Info("开始数据库迁移...")
	return nil
}

// GetDB 获取数据库连接（用于测试或特殊场景）
func GetDB() *gorm.DB {
	return DB
}

// PingDatabase 测试数据库连接
func PingDatabase() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return nil
}
