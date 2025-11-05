package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	// 根据应用日志级别配置 GORM 日志级别
	var logLevel logger.LogLevel
	switch cfg.Logging.Level {
	case "debug":
		logLevel = logger.Info // Debug 模式显示所有 SQL
	case "info":
		logLevel = logger.Warn // Info 模式只显示慢查询和错误
	case "warn", "error":
		logLevel = logger.Error // Warn/Error 模式只显示错误
	default:
		logLevel = logger.Silent // 其他情况静默
	}

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	return db, nil
}

// AutoMigrate 自动迁移数据表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Balance{},
		&model.Order{},
		&model.Trade{},
		&model.Ticker{},
		&model.Kline{},
	)
}
