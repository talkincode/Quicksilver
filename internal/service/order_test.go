package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to create test database")

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Balance{},
		&model.Order{},
		&model.Trade{},
		&model.Ticker{},
	)
	require.NoError(t, err)

	return db
}

// setupTestConfig 创建测试配置
func setupTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		Trading: config.TradingConfig{
			DefaultFeeRate: 0.001,
			MakerFeeRate:   0.0005,
			TakerFeeRate:   0.001,
			MinOrderAmount: 0.0001,
		},
		Market: config.MarketConfig{
			Symbols: []string{"BTC/USDT", "ETH/USDT"},
		},
	}
}

// createTestUser 创建测试用户
func createTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	user := &model.User{
		Email:     fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		APIKey:    fmt.Sprintf("test-api-key-%d", time.Now().UnixNano()),
		APISecret: "test-api-secret",
		Status:    "active",
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// createTestBalance 创建测试余额
func createTestBalance(t *testing.T, db *gorm.DB, userID uint, asset string, available, locked float64) *model.Balance {
	t.Helper()

	balance := &model.Balance{
		UserID:    userID,
		Asset:     asset,
		Available: available,
		Locked:    locked,
	}

	err := db.Create(balance).Error
	require.NoError(t, err)

	return balance
}

// createTestTicker 创建测试行情
func createTestTicker(t *testing.T, db *gorm.DB, symbol string, price float64) *model.Ticker {
	t.Helper()

	ticker := &model.Ticker{
		Symbol:    symbol,
		LastPrice: price,
		Source:    "test",
	}

	err := db.Save(ticker).Error
	require.NoError(t, err)

	return ticker
}

// TestNewOrderService 测试服务创建
func TestNewOrderService(t *testing.T) {
	db := setupTestDB(t)
	cfg := setupTestConfig(t)
	logger := zap.NewNop()

	balanceService := NewBalanceService(db, cfg, logger)
	orderService := NewOrderService(db, cfg, logger, balanceService)

	assert.NotNil(t, orderService)
	assert.NotNil(t, orderService.db)
	assert.NotNil(t, orderService.cfg)
	assert.NotNil(t, orderService.logger)
	assert.NotNil(t, orderService.balanceService)
}

// TestCreateMarketBuyOrder 测试创建市价买单
func TestCreateMarketBuyOrder(t *testing.T) {
	t.Run("Create market buy order successfully", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 创建用户和余额
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "USDT", 10000.0, 0)

		// 创建行情数据
		createTestTicker(t, db, "BTC/USDT", 50000.0)

		// 创建市价买单
		req := CreateOrderRequest{
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
		}

		order, err := orderService.CreateOrder(user.ID, req)

		// 验证订单创建成功
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, user.ID, order.UserID)
		assert.Equal(t, "BTC/USDT", order.Symbol)
		assert.Equal(t, "buy", order.Side)
		assert.Equal(t, "market", order.Type)
		assert.Equal(t, 0.1, order.Amount)
		assert.Equal(t, "new", order.Status)
		assert.Nil(t, order.Price) // 市价单无价格

		// 验证资金被冻结
		var balance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "USDT").First(&balance).Error
		require.NoError(t, err)
		assert.Greater(t, balance.Locked, 0.0) // 应该冻结了一定金额
	})

	t.Run("Create market buy order with insufficient balance", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 创建用户和余额（余额不足）
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "USDT", 100.0, 0)

		// 创建行情数据
		createTestTicker(t, db, "BTC/USDT", 50000.0)

		// 创建市价买单（需要 5000 USDT，但只有 100）
		req := CreateOrderRequest{
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
		}

		order, err := orderService.CreateOrder(user.ID, req)

		// 验证订单创建失败
		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "insufficient balance")
	})
}

// TestCreateMarketSellOrder 测试创建市价卖单
func TestCreateMarketSellOrder(t *testing.T) {
	t.Run("Create market sell order successfully", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 创建用户和余额
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 1.0, 0)

		// 创建行情数据
		createTestTicker(t, db, "BTC/USDT", 50000.0)

		// 创建市价卖单
		req := CreateOrderRequest{
			Symbol: "BTC/USDT",
			Side:   "sell",
			Type:   "market",
			Amount: 0.1,
		}

		order, err := orderService.CreateOrder(user.ID, req)

		// 验证订单创建成功
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, "sell", order.Side)

		// 验证 BTC 被冻结
		var balance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "BTC").First(&balance).Error
		require.NoError(t, err)
		assert.Equal(t, 0.1, balance.Locked)
		assert.Equal(t, 0.9, balance.Available)
	})
}

// TestCreateLimitOrder 测试创建限价单
func TestCreateLimitOrder(t *testing.T) {
	t.Run("Create limit buy order successfully", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 创建用户和余额
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "USDT", 10000.0, 0)

		// 创建限价买单
		price := 49000.0
		req := CreateOrderRequest{
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Amount: 0.1,
			Price:  &price,
		}

		order, err := orderService.CreateOrder(user.ID, req)

		// 验证订单创建成功
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, "limit", order.Type)
		assert.NotNil(t, order.Price)
		assert.Equal(t, 49000.0, *order.Price)

		// 验证资金被冻结（按限价计算）
		var balance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "USDT").First(&balance).Error
		require.NoError(t, err)
		expectedLocked := 0.1 * 49000.0 // 金额 * 价格
		assert.InDelta(t, expectedLocked, balance.Locked, 0.01)
	})

	t.Run("Create limit order without price", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "USDT", 10000.0, 0)

		// 限价单但未提供价格
		req := CreateOrderRequest{
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Amount: 0.1,
			// Price 为 nil
		}

		order, err := orderService.CreateOrder(user.ID, req)

		// 验证订单创建失败
		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "price is required")
	})
}

// TestValidateOrderRequest 测试订单参数验证
func TestValidateOrderRequest(t *testing.T) {
	db := setupTestDB(t)
	cfg := setupTestConfig(t)
	logger := zap.NewNop()

	balanceService := NewBalanceService(db, cfg, logger)
	orderService := NewOrderService(db, cfg, logger, balanceService)
	user := createTestUser(t, db)

	tests := []struct {
		name    string
		req     CreateOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid market buy order",
			req: CreateOrderRequest{
				Symbol: "BTC/USDT",
				Side:   "buy",
				Type:   "market",
				Amount: 0.1,
			},
			wantErr: false,
		},
		{
			name: "Invalid symbol",
			req: CreateOrderRequest{
				Symbol: "",
				Side:   "buy",
				Type:   "market",
				Amount: 0.1,
			},
			wantErr: true,
			errMsg:  "symbol is required",
		},
		{
			name: "Invalid side",
			req: CreateOrderRequest{
				Symbol: "BTC/USDT",
				Side:   "invalid",
				Type:   "market",
				Amount: 0.1,
			},
			wantErr: true,
			errMsg:  "side must be buy or sell",
		},
		{
			name: "Invalid type",
			req: CreateOrderRequest{
				Symbol: "BTC/USDT",
				Side:   "buy",
				Type:   "invalid",
				Amount: 0.1,
			},
			wantErr: true,
			errMsg:  "type must be market or limit",
		},
		{
			name: "Amount too small",
			req: CreateOrderRequest{
				Symbol: "BTC/USDT",
				Side:   "buy",
				Type:   "market",
				Amount: 0.00001,
			},
			wantErr: true,
			errMsg:  "amount is too small",
		},
		{
			name: "Negative amount",
			req: CreateOrderRequest{
				Symbol: "BTC/USDT",
				Side:   "buy",
				Type:   "market",
				Amount: -0.1,
			},
			wantErr: true,
			errMsg:  "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := orderService.CreateOrder(user.ID, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				// 如果期望成功，需要先准备余额
				if !tt.wantErr {
					createTestBalance(t, db, user.ID, "USDT", 100000.0, 0)
					createTestTicker(t, db, "BTC/USDT", 50000.0)
				}
			}
		})
	}
}

// TestGetOrderByID 测试查询订单
func TestGetOrderByID(t *testing.T) {
	t.Run("Get existing order", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 创建测试订单
		user := createTestUser(t, db)
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
			Status: "new",
		}
		err := db.Create(order).Error
		require.NoError(t, err)

		// 查询订单
		found, err := orderService.GetOrderByID(order.ID)

		require.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
		assert.Equal(t, "BTC/USDT", found.Symbol)
	})

	t.Run("Get non-existent order", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 查询不存在的订单
		found, err := orderService.GetOrderByID(99999)

		require.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "order not found")
	})
}

// TestCancelOrder 测试撤销订单
func TestCancelOrder(t *testing.T) {
	t.Run("Cancel open order successfully", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		// 创建用户和余额
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "USDT", 10000.0, 5000.0) // 5000 已冻结

		// 创建未成交订单
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Amount: 0.1,
			Status: "new",
		}
		price := 50000.0
		order.Price = &price
		err := db.Create(order).Error
		require.NoError(t, err)

		// 撤销订单
		err = orderService.CancelOrder(user.ID, order.ID)

		require.NoError(t, err)

		// 验证订单状态
		var updated model.Order
		err = db.First(&updated, order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "cancelled", updated.Status)

		// 验证资金被解冻
		var balance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "USDT").First(&balance).Error
		require.NoError(t, err)
		assert.Equal(t, 0.0, balance.Locked) // 冻结金额应该被释放
		assert.Equal(t, 15000.0, balance.Available)
	})

	t.Run("Cannot cancel filled order", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		user := createTestUser(t, db)

		// 创建已成交订单
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
			Status: "filled",
		}
		err := db.Create(order).Error
		require.NoError(t, err)

		// 尝试撤销已成交订单
		err = orderService.CancelOrder(user.ID, order.ID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot cancel")
	})
}

// TestGetUserOrders 测试获取用户订单列表
func TestGetUserOrders(t *testing.T) {
	t.Run("Get user orders with pagination", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		user := createTestUser(t, db)

		// 创建多个订单
		for i := 0; i < 5; i++ {
			order := &model.Order{
				UserID: user.ID,
				Symbol: "BTC/USDT",
				Side:   "buy",
				Type:   "market",
				Amount: 0.1,
				Status: "new",
			}
			err := db.Create(order).Error
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond) // 确保时间戳不同
		}

		// 获取订单列表
		orders, total, err := orderService.GetUserOrders(user.ID, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, orders, 5)
	})
}

// TestGetOpenOrders 测试获取未完成订单
func TestGetOpenOrders(t *testing.T) {
	t.Run("Get only open orders", func(t *testing.T) {
		db := setupTestDB(t)
		cfg := setupTestConfig(t)
		logger := zap.NewNop()

		balanceService := NewBalanceService(db, cfg, logger)
		orderService := NewOrderService(db, cfg, logger, balanceService)

		user := createTestUser(t, db)

		// 创建不同状态的订单
		statuses := []string{"new", "new", "filled", "cancelled"}
		for _, status := range statuses {
			order := &model.Order{
				UserID: user.ID,
				Symbol: "BTC/USDT",
				Side:   "buy",
				Type:   "market",
				Amount: 0.1,
				Status: status,
			}
			err := db.Create(order).Error
			require.NoError(t, err)
		}

		// 获取未完成订单
		orders, err := orderService.GetOpenOrders(user.ID)

		require.NoError(t, err)
		assert.Len(t, orders, 2) // 只有 2 个 new 状态的订单
		for _, order := range orders {
			assert.Equal(t, "new", order.Status)
		}
	})
}

// TestCreateStopLossOrder 测试创建止损单
func TestCreateStopLossOrder(t *testing.T) {
	db := setupTestDB(t)
	cfg := setupTestConfig(t)
	logger := zap.NewNop()
	balanceService := NewBalanceService(db, cfg, logger)
	orderService := NewOrderService(db, cfg, logger, balanceService)

	t.Run("Create stop loss sell order", func(t *testing.T) {
		// Given: 用户持有 BTC
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 1.0, 0)

		// When: 创建止损卖单（当价格跌破 48000 时卖出）
		stopPrice := 48000.0
		order, err := orderService.CreateStopLossOrder(user.ID, "BTC/USDT", "sell", 0.5, stopPrice)

		// Then: 订单创建成功
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, "stop_loss", order.Type)
		assert.Equal(t, "new", order.Status)
		assert.Equal(t, stopPrice, *order.StopPrice)
		assert.Equal(t, "<=", order.TriggerCondition) // 价格 <= 止损价时触发
		assert.Equal(t, 0.5, order.Amount)
	})

	t.Run("Create stop loss with insufficient balance", func(t *testing.T) {
		cleanupTestDB(t, db)
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 0.1, 0)

		stopPrice := 48000.0
		_, err := orderService.CreateStopLossOrder(user.ID, "BTC/USDT", "sell", 1.0, stopPrice)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})
}

// TestCreateTakeProfitOrder 测试创建止盈单
func TestCreateTakeProfitOrder(t *testing.T) {
	db := setupTestDB(t)
	cfg := setupTestConfig(t)
	logger := zap.NewNop()
	balanceService := NewBalanceService(db, cfg, logger)
	orderService := NewOrderService(db, cfg, logger, balanceService)

	t.Run("Create take profit sell order", func(t *testing.T) {
		// Given: 用户持有 BTC
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 1.0, 0)

		// When: 创建止盈卖单（当价格涨到 52000 时卖出）
		takeProfitPrice := 52000.0
		order, err := orderService.CreateTakeProfitOrder(user.ID, "BTC/USDT", "sell", 0.5, takeProfitPrice)

		// Then: 订单创建成功
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, "take_profit", order.Type)
		assert.Equal(t, "new", order.Status)
		assert.Equal(t, takeProfitPrice, *order.StopPrice)
		assert.Equal(t, ">=", order.TriggerCondition) // 价格 >= 止盈价时触发
		assert.Equal(t, 0.5, order.Amount)
	})
}

// TestTriggerStopOrders 测试止盈止损触发逻辑
func TestTriggerStopOrders(t *testing.T) {
	db := setupTestDB(t)
	cfg := setupTestConfig(t)
	logger := zap.NewNop()

	balanceService := NewBalanceService(db, cfg, logger)
	orderService := NewOrderService(db, cfg, logger, balanceService)
	marketService := NewMarketService(db, cfg, logger)

	t.Run("Trigger stop loss when price drops", func(t *testing.T) {
		// Given: 创建止损单
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 1.0, 0.5) // 0.5 已冻结
		createTestBalance(t, db, user.ID, "USDT", 0, 0)

		stopPrice := 48000.0
		stopOrder, _ := orderService.CreateStopLossOrder(user.ID, "BTC/USDT", "sell", 0.5, stopPrice)

		// 创建行情（价格跌破止损价）
		bidPrice := 47500.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 47500.0,
			BidPrice:  &bidPrice,
		}
		db.Save(ticker)

		// When: 触发止盈止损检查
		err := marketService.TriggerStopOrders()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // 等待异步处理

		// Then: 止损单被触发，创建市价卖单
		var updatedOrder model.Order
		db.First(&updatedOrder, stopOrder.ID)
		assert.Equal(t, "triggered", updatedOrder.Status)
		assert.NotNil(t, updatedOrder.TriggeredAt)

		// 应该有一个市价单被创建并成交
		var marketOrders []model.Order
		db.Where("user_id = ? AND type = ? AND parent_order_id = ?",
			user.ID, "market", stopOrder.ID).Find(&marketOrders)
		assert.Len(t, marketOrders, 1)
		assert.Equal(t, "filled", marketOrders[0].Status)
	})

	t.Run("Trigger take profit when price rises", func(t *testing.T) {
		// Given: 创建止盈单
		cleanupTestDB(t, db)
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 1.0, 0.5)
		createTestBalance(t, db, user.ID, "USDT", 0, 0)

		takeProfitPrice := 52000.0
		takeProfitOrder, _ := orderService.CreateTakeProfitOrder(user.ID, "BTC/USDT", "sell", 0.5, takeProfitPrice)

		// 创建行情（价格突破止盈价）
		bidPrice := 52500.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 52500.0,
			BidPrice:  &bidPrice,
		}
		db.Save(ticker)

		// When: 触发止盈止损检查
		err := marketService.TriggerStopOrders()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Then: 止盈单被触发
		var updatedOrder model.Order
		db.First(&updatedOrder, takeProfitOrder.ID)
		assert.Equal(t, "triggered", updatedOrder.Status)
	})

	t.Run("Do not trigger when price condition not met", func(t *testing.T) {
		// Given: 创建止损单
		cleanupTestDB(t, db)
		user := createTestUser(t, db)
		createTestBalance(t, db, user.ID, "BTC", 1.0, 0.5)

		stopPrice := 48000.0
		stopOrder, _ := orderService.CreateStopLossOrder(user.ID, "BTC/USDT", "sell", 0.5, stopPrice)

		// 创建行情（价格仍高于止损价）
		bidPrice := 49000.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 49000.0,
			BidPrice:  &bidPrice,
		}
		db.Save(ticker)

		// When: 触发止盈止损检查
		err := marketService.TriggerStopOrders()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Then: 订单状态不变
		var updatedOrder model.Order
		db.First(&updatedOrder, stopOrder.ID)
		assert.Equal(t, "new", updatedOrder.Status)
		assert.Nil(t, updatedOrder.TriggeredAt)
	})
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("DELETE FROM trades")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM balances")
	db.Exec("DELETE FROM tickers")
}
