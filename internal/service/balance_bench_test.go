package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// setupBenchDB 为基准测试创建数据库
func setupBenchDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&model.User{},
		&model.Balance{},
		&model.Order{},
		&model.Trade{},
		&model.Ticker{},
	)

	return db
}

// BenchmarkFreezeBalance 冻结余额性能基准测试
func BenchmarkFreezeBalance(b *testing.B) {
	db := setupBenchDB()
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()

	// 创建测试用户和余额
	user := &model.User{
		Email:     fmt.Sprintf("bench-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user)

	balance := &model.Balance{
		UserID:    user.ID,
		Asset:     "USDT",
		Available: 1000000.0,
		Locked:    0,
	}
	db.Create(balance)

	service := NewBalanceService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.FreezeBalance(user.ID, "USDT", 1.0)
	}
}

// BenchmarkUnfreezeBalance 解冻余额性能基准测试
func BenchmarkUnfreezeBalance(b *testing.B) {
	db := setupBenchDB()
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()

	user := &model.User{
		Email:     fmt.Sprintf("bench-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user)

	balance := &model.Balance{
		UserID:    user.ID,
		Asset:     "USDT",
		Available: 0,
		Locked:    1000000.0,
	}
	db.Create(balance)

	service := NewBalanceService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.UnfreezeBalance(user.ID, "USDT", 1.0)
	}
}

// BenchmarkGetBalance 获取余额性能基准测试
func BenchmarkGetBalance(b *testing.B) {
	db := setupBenchDB()
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()

	user := &model.User{
		Email:     fmt.Sprintf("bench-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user)

	balance := &model.Balance{
		UserID:    user.ID,
		Asset:     "USDT",
		Available: 1000.0,
		Locked:    0,
	}
	db.Create(balance)

	service := NewBalanceService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetBalance(user.ID, "USDT")
	}
}

// BenchmarkTransferBalance 转账性能基准测试
func BenchmarkTransferBalance(b *testing.B) {
	db := setupBenchDB()
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()

	user1 := &model.User{
		Email:     fmt.Sprintf("bench1-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key1-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user1)

	user2 := &model.User{
		Email:     fmt.Sprintf("bench2-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key2-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user2)

	db.Create(&model.Balance{UserID: user1.ID, Asset: "USDT", Available: 1000000.0, Locked: 0})
	db.Create(&model.Balance{UserID: user2.ID, Asset: "USDT", Available: 0, Locked: 0})

	service := NewBalanceService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.TransferBalance(user1.ID, user2.ID, "USDT", 1.0)
	}
}

// BenchmarkDeductBalance 扣除余额性能基准测试
func BenchmarkDeductBalance(b *testing.B) {
	db := setupBenchDB()
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()

	user := &model.User{
		Email:     fmt.Sprintf("bench-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user)

	balance := &model.Balance{
		UserID:    user.ID,
		Asset:     "USDT",
		Available: 0,
		Locked:    1000000.0,
	}
	db.Create(balance)

	service := NewBalanceService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.DeductBalance(user.ID, "USDT", 1.0)
	}
}

// BenchmarkAddBalance 添加余额性能基准测试
func BenchmarkAddBalance(b *testing.B) {
	db := setupBenchDB()
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()

	user := &model.User{
		Email:     fmt.Sprintf("bench-%d@test.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
		APISecret: "secret",
		Status:    "active",
	}
	db.Create(user)

	balance := &model.Balance{
		UserID:    user.ID,
		Asset:     "USDT",
		Available: 0,
		Locked:    0,
	}
	db.Create(balance)

	service := NewBalanceService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.AddBalance(user.ID, "USDT", 1.0)
	}
}
