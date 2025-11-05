package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// NewTestDB 创建测试数据库
// 支持两种模式：
// 1. SQLite 内存模式（默认）：快速但不支持并发
// 2. PostgreSQL 模式：设置环境变量 TEST_DB=postgres
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	usePostgres := os.Getenv("TEST_DB") == "postgres"

	var db *gorm.DB
	var err error

	if usePostgres {
		// PostgreSQL 测试数据库
		dsn := os.Getenv("TEST_DATABASE_URL")
		if dsn == "" {
			dsn = "host=localhost port=5432 user=postgres password=pgdb dbname=quicksilver_test sslmode=disable"
		}

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err, "failed to create PostgreSQL test database")

		// PostgreSQL: 每个测试清理一次，避免数据残留
		// 注意：使用 t.Helper() 确保测试失败时能正确定位
		cleanupTestDB(t, db)
	} else {
		// SQLite 内存模式（默认）
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err, "failed to create SQLite test database")
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&model.User{},
		&model.Balance{},
		&model.Order{},
		&model.Trade{},
		&model.Ticker{},
	)
	require.NoError(t, err, "failed to migrate test database")

	return db
}

// cleanupTestDB 清理测试数据库中的所有数据
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 按照外键依赖顺序删除
	tables := []string{"trades", "orders", "balances", "tickers", "users"}
	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error
		if err != nil {
			// 如果表不存在，忽略错误
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// SetupTestDB 创建内存测试数据库（别名）
func SetupTestDB(t *testing.T) *gorm.DB {
	return NewTestDB(t)
}

// NewTestLogger 创建测试日志记录器
func NewTestLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to create test logger: %v", err))
	}
	return logger
}

// NewTestConfig 创建测试配置
func NewTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:    8080,
			Mode:    "test",
			Name:    "quicksilver-test",
			Version: "0.1.0-test",
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Name:            "quicksilver_test",
			User:            "test",
			Password:        "test",
			SSLMode:         "disable",
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: 300,
		},
		Market: config.MarketConfig{
			UpdateInterval: "1s",
			DataSource:     "hyperliquid",
			APIURL:         "https://api.hyperliquid.xyz",
			Symbols:        []string{"BTC/USDT", "ETH/USDT"},
			Hyperliquid: config.HyperliquidConfig{
				InfoEndpoint: "/info",
				WSEndpoint:   "/ws",
			},
		},
		Trading: config.TradingConfig{
			DefaultFeeRate: 0.001,
			MakerFeeRate:   0.0005,
			TakerFeeRate:   0.001,
			MinOrderAmount: 0.0001,
		},
		Auth: config.AuthConfig{
			JWTSecret:   "test-secret-key",
			TokenExpire: 3600,
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
			Output: "stdout",
		},
	}
}

// LoadTestConfig 加载测试配置（别名）
func LoadTestConfig(t *testing.T) *config.Config {
	return NewTestConfig()
}

// CleanupDB 清理测试数据库
func CleanupDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 清空所有表
	db.Exec("DELETE FROM trades")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM balances")
	db.Exec("DELETE FROM tickers")
	db.Exec("DELETE FROM users")
}

// CreateTestUser 创建测试用户（每次生成唯一邮箱）
func CreateTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	// 使用时间戳确保邮箱唯一
	timestamp := time.Now().UnixNano()

	user := &model.User{
		Email:     fmt.Sprintf("test%d@example.com", timestamp),
		Username:  "testuser",
		APIKey:    fmt.Sprintf("test-api-key-%d", timestamp),
		APISecret: "test-api-secret-123456",
		Status:    "active",
	}

	err := db.Create(user).Error
	require.NoError(t, err, "failed to create test user")

	return user
}

// CreateTestBalance 创建测试余额
func CreateTestBalance(t *testing.T, db *gorm.DB, userID uint, asset string, available, locked float64) *model.Balance {
	t.Helper()

	balance := &model.Balance{
		UserID:    userID,
		Asset:     asset,
		Available: available,
		Locked:    locked,
	}

	err := db.Create(balance).Error
	require.NoError(t, err, "failed to create test balance")

	return balance
}

// CreateTestTicker 创建测试行情
func CreateTestTicker(t *testing.T, db *gorm.DB, symbol string, price float64) *model.Ticker {
	t.Helper()

	ticker := &model.Ticker{
		Symbol:    symbol,
		LastPrice: price,
		Source:    "test",
	}

	err := db.Save(ticker).Error
	require.NoError(t, err, "failed to create test ticker")

	return ticker
}

// CreateTestOrder 创建测试订单
func CreateTestOrder(t *testing.T, db *gorm.DB, userID uint, symbol, side, orderType string, amount float64, price *float64) *model.Order {
	t.Helper()

	order := &model.Order{
		UserID: userID,
		Symbol: symbol,
		Side:   side,
		Type:   orderType,
		Status: "new",
		Amount: amount,
		Price:  price,
		Filled: 0,
	}

	err := db.Create(order).Error
	require.NoError(t, err, "failed to create test order")

	return order
}

// Float64Ptr 返回 float64 指针
func Float64Ptr(v float64) *float64 {
	return &v
}

// SeedUser 创建种子用户（简化版，每次生成唯一邮箱）
func SeedUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	// 使用时间戳确保邮箱唯一
	timestamp := time.Now().UnixNano()

	user := &model.User{
		Email:     fmt.Sprintf("test%d@example.com", timestamp),
		Username:  "testuser",
		APIKey:    fmt.Sprintf("test-api-key-%d", timestamp),
		APISecret: "test-secret",
		Status:    "active",
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// SeedBalance 创建种子余额（支持 available 和 locked）
func SeedBalance(t *testing.T, db *gorm.DB, userID uint, asset string, available float64, locked ...float64) *model.Balance {
	t.Helper()

	lockedAmount := 0.0
	if len(locked) > 0 {
		lockedAmount = locked[0]
	}

	balance := &model.Balance{
		UserID:    userID,
		Asset:     asset,
		Available: available,
		Locked:    lockedAmount,
	}

	err := db.Create(balance).Error
	require.NoError(t, err)

	return balance
}

// SeedOrder 创建种子订单（简化版）
func SeedOrder(t *testing.T, db *gorm.DB, userID uint, symbol string) *model.Order {
	t.Helper()

	order := &model.Order{
		UserID: userID,
		Symbol: symbol,
		Side:   "buy",
		Type:   "limit",
		Status: "new",
		Price:  Float64Ptr(50000.0),
		Amount: 0.1,
		Filled: 0,
	}

	err := db.Create(order).Error
	require.NoError(t, err)

	return order
}
