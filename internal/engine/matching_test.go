package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/testutil"
)

func TestNewMatchingEngine(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	engine := NewMatchingEngine(db, cfg, logger)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.db)
	assert.NotNil(t, engine.cfg)
	assert.NotNil(t, engine.logger)
}

func TestMatchMarketBuyOrder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	t.Run("Match market buy order successfully", func(t *testing.T) {
		// Given: 用户有足够的 USDT 余额和冻结资金
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5500.0) // available: 10000, locked: 5500
		testutil.SeedBalance(t, db, user.ID, "BTC", 0, 0)

		// 创建市场价格
		bidPrice := 49990.0
		askPrice := 50010.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50000.0,
			BidPrice:  &bidPrice,
			AskPrice:  &askPrice, // 买单使用 ask 价格
		}
		err := db.Save(ticker).Error // 使用 Save 而非 Create
		require.NoError(t, err)

		// 创建市价买单 (已冻结资金)
		amount := 0.1
		price := 50010.0 // ask price
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: amount,
			Filled: 0,
			Status: "new",
		}
		err = db.Create(order).Error
		require.NoError(t, err)

		// When: 撮合订单
		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		// Then: 订单成交
		require.NoError(t, err)

		// 验证订单状态
		var updatedOrder model.Order
		err = db.First(&updatedOrder, order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "filled", updatedOrder.Status)
		assert.Equal(t, amount, updatedOrder.Filled)

		// 验证成交记录
		var trade model.Trade
		err = db.Where("order_id = ?", order.ID).First(&trade).Error
		require.NoError(t, err)
		assert.Equal(t, user.ID, trade.UserID)
		assert.Equal(t, "BTC/USDT", trade.Symbol)
		assert.Equal(t, "buy", trade.Side)
		assert.Equal(t, price, trade.Price)
		assert.Equal(t, amount, trade.Amount)
		assert.Greater(t, trade.Fee, 0.0) // 应该有手续费

		// 验证余额变化
		var btcBalance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "BTC").First(&btcBalance).Error
		require.NoError(t, err)
		// 买入后应该增加 BTC (扣除手续费)
		expectedBTC := amount - trade.Fee // 0.1 - 手续费
		assert.InDelta(t, expectedBTC, btcBalance.Available, 0.00001)

		var usdtBalance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "USDT").First(&usdtBalance).Error
		require.NoError(t, err)
		// 冻结的 USDT 应该被扣除
		expectedCost := amount * price * (1 + cfg.Trading.TakerFeeRate)
		expectedLocked := 5500.0 - expectedCost
		assert.InDelta(t, expectedLocked, usdtBalance.Locked, 0.01)
	})

	t.Run("Match market buy order with ticker not found", func(t *testing.T) {
		// Given: 没有行情数据
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5000.0)

		order := &model.Order{
			UserID: user.ID,
			Symbol: "ETH/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 1.0,
			Status: "new",
		}
		err := db.Create(order).Error
		require.NoError(t, err)

		// When: 撮合订单
		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		// Then: 应该返回错误
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ticker not found")
	})
}

func TestMatchMarketSellOrder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	t.Run("Match market sell order successfully", func(t *testing.T) {
		// Given: 用户有足够的 BTC 余额和冻结资金
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "BTC", 1.0, 0.1) // available: 1.0, locked: 0.1
		testutil.SeedBalance(t, db, user.ID, "USDT", 0, 0)

		// 创建市场价格
		bidPrice := 49990.0
		askPrice := 50010.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50000.0,
			BidPrice:  &bidPrice, // 卖单使用 bid 价格
			AskPrice:  &askPrice,
		}
		err := db.Save(ticker).Error // 使用 Save 而非 Create
		require.NoError(t, err)

		// 创建市价卖单 (已冻结资金)
		amount := 0.1
		price := 49990.0 // bid price
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "sell",
			Type:   "market",
			Amount: amount,
			Filled: 0,
			Status: "new",
		}
		err = db.Create(order).Error
		require.NoError(t, err)

		// When: 撮合订单
		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		// Then: 订单成交
		require.NoError(t, err)

		// 验证订单状态
		var updatedOrder model.Order
		err = db.First(&updatedOrder, order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "filled", updatedOrder.Status)
		assert.Equal(t, amount, updatedOrder.Filled)

		// 验证成交记录
		var trade model.Trade
		err = db.Where("order_id = ?", order.ID).First(&trade).Error
		require.NoError(t, err)
		assert.Equal(t, user.ID, trade.UserID)
		assert.Equal(t, "BTC/USDT", trade.Symbol)
		assert.Equal(t, "sell", trade.Side)
		assert.Equal(t, price, trade.Price)
		assert.Equal(t, amount, trade.Amount)
		assert.Greater(t, trade.Fee, 0.0)

		// 验证余额变化
		var btcBalance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "BTC").First(&btcBalance).Error
		require.NoError(t, err)
		// 冻结的 BTC 应该被扣除
		assert.InDelta(t, 0.0, btcBalance.Locked, 0.00001)

		var usdtBalance model.Balance
		err = db.Where("user_id = ? AND asset = ?", user.ID, "USDT").First(&usdtBalance).Error
		require.NoError(t, err)
		// 应该收到 USDT (扣除手续费)
		expectedUSDT := amount * price * (1 - cfg.Trading.TakerFeeRate)
		assert.InDelta(t, expectedUSDT, usdtBalance.Available, 0.01)
	})
}

func TestCalculateFee(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	engine := NewMatchingEngine(db, cfg, logger)

	tests := []struct {
		name     string
		price    float64
		amount   float64
		feeRate  float64
		expected float64
	}{
		{
			name:     "Basic fee calculation",
			price:    50000.0,
			amount:   0.1,
			feeRate:  0.001,
			expected: 0.1 * 0.001, // 0.0001 BTC
		},
		{
			name:     "Zero fee rate",
			price:    50000.0,
			amount:   0.1,
			feeRate:  0.0,
			expected: 0.0,
		},
		{
			name:     "Large amount",
			price:    50000.0,
			amount:   10.0,
			feeRate:  0.002,
			expected: 10.0 * 0.002, // 0.02 BTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee := engine.calculateFee(tt.amount, tt.feeRate)
			assert.InDelta(t, tt.expected, fee, 0.00000001)
		})
	}
}

func TestMatchLimitOrder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	t.Run("Match limit buy order when price is acceptable", func(t *testing.T) {
		// Given: 限价单价格高于市场价
		testutil.CleanupDB(t, db) // 清理数据库避免冲突
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5500.0)
		testutil.SeedBalance(t, db, user.ID, "BTC", 0, 0)

		bidPrice := 49990.0
		askPrice := 50010.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50000.0,
			BidPrice:  &bidPrice,
			AskPrice:  &askPrice,
		}
		err := db.Save(ticker).Error // 使用 Save 而非 Create
		require.NoError(t, err)

		limitPrice := 50100.0 // 限价高于 ask，可以成交
		amount := 0.1
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Price:  &limitPrice,
			Amount: amount,
			Filled: 0,
			Status: "new",
		}
		err = db.Create(order).Error
		require.NoError(t, err)

		// When: 撮合订单
		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		// Then: 订单成交
		require.NoError(t, err)

		var updatedOrder model.Order
		err = db.First(&updatedOrder, order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "filled", updatedOrder.Status)
	})

	t.Run("Limit buy order not matched when price too low", func(t *testing.T) {
		// Given: 限价单价格低于市场价
		testutil.CleanupDB(t, db) // 清理数据库避免冲突
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 4900.0)

		bidPrice := 49990.0
		askPrice := 50010.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50000.0,
			BidPrice:  &bidPrice,
			AskPrice:  &askPrice,
		}
		err := db.Save(ticker).Error // 使用 Save 而非 Create
		require.NoError(t, err)

		limitPrice := 49000.0 // 限价低于 ask，不能成交
		amount := 0.1
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Price:  &limitPrice,
			Amount: amount,
			Filled: 0,
			Status: "new",
		}
		err = db.Create(order).Error
		require.NoError(t, err)

		// When: 撮合订单
		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		// Then: 订单未成交，保持 new 状态
		require.NoError(t, err)

		var updatedOrder model.Order
		err = db.First(&updatedOrder, order.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "new", updatedOrder.Status)
		assert.Equal(t, 0.0, updatedOrder.Filled)
	})
}

func TestMatchOrder_OrderNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	engine := NewMatchingEngine(db, cfg, logger)

	// When: 撮合不存在的订单
	err := engine.MatchOrder(99999)

	// Then: 应该返回错误
	require.Error(t, err)
	assert.Contains(t, err.Error(), "order not found")
}

func TestMatchOrder_InvalidOrderStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	t.Run("Cannot match cancelled order", func(t *testing.T) {
		user := testutil.SeedUser(t, db)
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
			Status: "cancelled",
		}
		err := db.Create(order).Error
		require.NoError(t, err)

		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "order status is not new")
	})

	t.Run("Cannot match already filled order", func(t *testing.T) {
		user := testutil.SeedUser(t, db)
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
			Filled: 0.1,
			Status: "filled",
		}
		err := db.Create(order).Error
		require.NoError(t, err)

		engine := NewMatchingEngine(db, cfg, logger)
		err = engine.MatchOrder(order.ID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "order status is not new")
	})
}
