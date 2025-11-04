package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// NewTestDB 创建内存测试数据库
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to create test database")

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

// SetupTestDB 创建内存测试数据库（别名）
func SetupTestDB(t *testing.T) *gorm.DB {
	return NewTestDB(t)
}

// NewTestLogger 创建测试日志记录器
func NewTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
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

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	user := &model.User{
		Email:     "test@example.com",
		Username:  "testuser",
		APIKey:    "test-api-key-123456",
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

// SeedUser 创建种子用户（简化版）
func SeedUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	user := &model.User{
		Email:     "test@example.com",
		Username:  "testuser",
		APIKey:    "test-api-key",
		APISecret: "test-secret",
		Status:    "active",
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// SeedBalance 创建种子余额（简化版）
func SeedBalance(t *testing.T, db *gorm.DB, userID uint, asset string, amount float64) *model.Balance {
	t.Helper()

	balance := &model.Balance{
		UserID:    userID,
		Asset:     asset,
		Available: amount,
		Locked:    0,
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
