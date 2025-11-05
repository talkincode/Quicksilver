package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/testutil"
)

func TestNewMarketService(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := testutil.NewTestLogger()

	service := NewMarketService(db, cfg, logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.cfg)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.client)
}

func TestUpdateHyperliquidTickers(t *testing.T) {
	db := testutil.NewTestDB(t)
	logger := testutil.NewTestLogger()

	// 创建模拟 Hyperliquid API 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/info", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 返回模拟数据 (扁平的键值对格式，不是嵌套的 mids 对象)
		response := HyperliquidAllMidsResponse{
			"BTC": "50000.5",
			"ETH": "3000.25",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testutil.NewTestConfig()
	cfg.Market.APIURL = server.URL
	cfg.Market.Symbols = []string{"BTC/USDT", "ETH/USDT"}

	service := NewMarketService(db, cfg, logger)

	t.Run("Update tickers successfully", func(t *testing.T) {
		err := service.updateHyperliquidTickers()
		require.NoError(t, err)

		// 验证数据库中的 ticker
		var btcTicker model.Ticker
		err = db.Where("symbol = ?", "BTC/USDT").First(&btcTicker).Error
		require.NoError(t, err)
		assert.Equal(t, "BTC/USDT", btcTicker.Symbol)
		assert.Equal(t, 50000.5, btcTicker.LastPrice)
		assert.Equal(t, "hyperliquid", btcTicker.Source)

		var ethTicker model.Ticker
		err = db.Where("symbol = ?", "ETH/USDT").First(&ethTicker).Error
		require.NoError(t, err)
		assert.Equal(t, 3000.25, ethTicker.LastPrice)
	})

	t.Run("Update existing ticker", func(t *testing.T) {
		// 第一次更新
		err := service.updateHyperliquidTickers()
		require.NoError(t, err)

		var ticker1 model.Ticker
		db.Where("symbol = ?", "BTC/USDT").First(&ticker1)
		firstUpdate := ticker1.UpdatedAt

		// 等待一小段时间
		time.Sleep(time.Millisecond * 10)

		// 第二次更新
		err = service.updateHyperliquidTickers()
		require.NoError(t, err)

		var ticker2 model.Ticker
		db.Where("symbol = ?", "BTC/USDT").First(&ticker2)

		// 验证时间戳更新
		assert.True(t, ticker2.UpdatedAt.After(firstUpdate))
	})
}

func TestUpdateTickersError(t *testing.T) {
	db := testutil.NewTestDB(t)
	logger := testutil.NewTestLogger()

	t.Run("API server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := testutil.NewTestConfig()
		cfg.Market.APIURL = server.URL

		service := NewMarketService(db, cfg, logger)
		err := service.updateHyperliquidTickers()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		cfg := testutil.NewTestConfig()
		cfg.Market.APIURL = server.URL

		service := NewMarketService(db, cfg, logger)
		err := service.updateHyperliquidTickers()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("Unsupported data source", func(t *testing.T) {
		cfg := testutil.NewTestConfig()
		cfg.Market.DataSource = "unknown"

		service := NewMarketService(db, cfg, logger)
		err := service.UpdateTickers()

		// 等待异步 goroutine 完成
		time.Sleep(100 * time.Millisecond)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported data source")
	})
}

func TestConvertSymbolToCoin(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		expected string
	}{
		{"BTC/USDT", "BTC/USDT", "BTC"},
		{"ETH/USDT", "ETH/USDT", "ETH"},
		{"SOL/USDT", "SOL/USDT", "SOL"},
		{"Short symbol", "AB", "AB"},
		{"Single char", "A", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSymbolToCoin(tt.symbol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarketServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := testutil.NewTestDB(t)
	logger := testutil.NewTestLogger()

	// 创建完整的模拟 API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := HyperliquidAllMidsResponse{
			"BTC": "109965.50",
			"ETH": "3456.78",
			"SOL": "234.56",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testutil.NewTestConfig()
	cfg.Market.APIURL = server.URL
	cfg.Market.Symbols = []string{"BTC/USDT", "ETH/USDT", "SOL/USDT"}

	service := NewMarketService(db, cfg, logger)

	t.Run("Full update cycle", func(t *testing.T) {
		// 执行更新
		err := service.UpdateTickers()
		require.NoError(t, err)

		// 注意：异步 goroutine 可能在 SQLite 内存数据库中失败
		// 这是预期行为，因为 SQLite :memory: 不支持跨 goroutine 共享

		// 直接从数据库查询（同步）
		var tickers []model.Ticker
		err = db.Find(&tickers).Error
		require.NoError(t, err, "Should query tickers successfully")

		// 验证至少有数据被写入
		assert.GreaterOrEqual(t, len(tickers), 1, "Should have at least one ticker")

		// 验证每个 ticker 的数据
		symbols := map[string]float64{
			"BTC/USDT": 109965.50,
			"ETH/USDT": 3456.78,
			"SOL/USDT": 234.56,
		}

		for symbol, expectedPrice := range symbols {
			var ticker model.Ticker
			err := db.Where("symbol = ?", symbol).First(&ticker).Error
			require.NoError(t, err)
			assert.Equal(t, expectedPrice, ticker.LastPrice)
			assert.False(t, ticker.UpdatedAt.IsZero())
		}
	})
}

func BenchmarkUpdateTickers(b *testing.B) {
	db := testutil.NewTestDB(&testing.T{})
	logger := testutil.NewTestLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := HyperliquidAllMidsResponse{
			"BTC": "50000",
			"ETH": "3000",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testutil.NewTestConfig()
	cfg.Market.APIURL = server.URL

	service := NewMarketService(db, cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.UpdateTickers()
	}
}

// TestTriggerPendingOrdersMatching 测试价格更新后触发限价单撮合
func TestTriggerPendingOrdersMatching(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := testutil.NewTestLogger()

	t.Run("Trigger limit buy order when price drops", func(t *testing.T) {
		// Given: 创建限价买单（价格低于市场价）
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5000.0)
		testutil.SeedBalance(t, db, user.ID, "BTC", 0, 0)

		// 创建初始行情（价格较高）
		bidPrice := 50000.0
		askPrice := 50100.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50050.0,
			BidPrice:  &bidPrice,
			AskPrice:  &askPrice,
		}
		db.Save(ticker)

		// 创建限价买单（限价 49500，低于当前价）
		limitPrice := 49500.0
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Price:  &limitPrice,
			Amount: 0.1,
			Status: "new",
		}
		db.Create(order)

		// When: 价格下跌到 49000（低于限价）
		newAskPrice := 49000.0
		ticker.AskPrice = &newAskPrice
		ticker.LastPrice = 49000.0
		db.Save(ticker)

		// 触发限价单撮合
		service := NewMarketService(db, cfg, logger)
		err := service.TriggerPendingOrdersMatching()
		require.NoError(t, err)

		// 等待异步撮合完成
		time.Sleep(100 * time.Millisecond)

		// Then: 订单应该成交
		var updatedOrder model.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, "filled", updatedOrder.Status)
		assert.Equal(t, 0.1, updatedOrder.Filled)

		// 验证成交记录
		var trade model.Trade
		err = db.Where("order_id = ?", order.ID).First(&trade).Error
		require.NoError(t, err)
		assert.Equal(t, 49000.0, trade.Price)
	})

	t.Run("Do not trigger limit buy order when price is still high", func(t *testing.T) {
		// Given: 创建限价买单
		testutil.CleanupDB(t, db)
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5000.0)

		askPrice := 50100.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50050.0,
			AskPrice:  &askPrice,
		}
		db.Save(ticker)

		limitPrice := 49500.0
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Price:  &limitPrice,
			Amount: 0.1,
			Status: "new",
		}
		db.Create(order)

		// When: 价格仍然高于限价（50100 > 49500）
		service := NewMarketService(db, cfg, logger)
		err := service.TriggerPendingOrdersMatching()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Then: 订单应该保持未成交状态
		var updatedOrder model.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, "new", updatedOrder.Status)
		assert.Equal(t, 0.0, updatedOrder.Filled)
	})

	t.Run("Trigger limit sell order when price rises", func(t *testing.T) {
		// Given: 创建限价卖单
		testutil.CleanupDB(t, db)
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "BTC", 1.0, 0.1)
		testutil.SeedBalance(t, db, user.ID, "USDT", 0, 0)

		bidPrice := 49900.0
		ticker := &model.Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 49950.0,
			BidPrice:  &bidPrice,
		}
		db.Save(ticker)

		// 限价卖单：限价 50500
		limitPrice := 50500.0
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "sell",
			Type:   "limit",
			Price:  &limitPrice,
			Amount: 0.1,
			Status: "new",
		}
		db.Create(order)

		// When: 价格上涨到 51000（高于限价）
		newBidPrice := 51000.0
		ticker.BidPrice = &newBidPrice
		ticker.LastPrice = 51000.0
		db.Save(ticker)

		service := NewMarketService(db, cfg, logger)
		err := service.TriggerPendingOrdersMatching()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Then: 订单应该成交
		var updatedOrder model.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, "filled", updatedOrder.Status)
		assert.Equal(t, 0.1, updatedOrder.Filled)
	})

	t.Run("Only trigger limit orders, not market orders", func(t *testing.T) {
		// Given: 创建市价单和限价单
		testutil.CleanupDB(t, db)
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5000.0)

		askPrice := 50000.0
		ticker := &model.Ticker{
			Symbol:   "BTC/USDT",
			AskPrice: &askPrice,
		}
		db.Save(ticker)

		// 创建一个市价单（状态为 new，不应该被触发）
		marketOrder := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Amount: 0.1,
			Status: "new",
		}
		db.Create(marketOrder)

		// When: 触发撮合
		service := NewMarketService(db, cfg, logger)
		err := service.TriggerPendingOrdersMatching()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Then: 市价单不应该被触发（只触发限价单）
		var updatedMarketOrder model.Order
		db.First(&updatedMarketOrder, marketOrder.ID)
		assert.Equal(t, "new", updatedMarketOrder.Status)
	})
}

// TestUpdateTickersWithMatching 测试价格更新后自动触发撮合
func TestUpdateTickersWithMatching(t *testing.T) {
	t.Skip("Skipping async matching test (SQLite memory database limitations)")

	db := testutil.NewTestDB(t)
	logger := testutil.NewTestLogger()

	t.Run("Price update triggers pending orders matching", func(t *testing.T) {
		// Given: 创建限价买单
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0, 5000.0)
		testutil.SeedBalance(t, db, user.ID, "BTC", 0, 0)

		limitPrice := 49500.0
		order := &model.Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Price:  &limitPrice,
			Amount: 0.1,
			Status: "new",
		}
		db.Create(order)

		// 创建模拟服务器（返回低价）
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := HyperliquidAllMidsResponse{
				"BTC": "49000",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := testutil.NewTestConfig()
		cfg.Market.APIURL = server.URL
		cfg.Market.Symbols = []string{"BTC/USDT"}

		// When: 更新行情
		service := NewMarketService(db, cfg, logger)
		err := service.UpdateTickers()
		require.NoError(t, err)

		// 等待异步撮合完成（增加时间）
		time.Sleep(500 * time.Millisecond)

		// Then: 订单应该自动成交
		var updatedOrder model.Order
		err = db.First(&updatedOrder, order.ID).Error
		require.NoError(t, err, "Order should exist")

		t.Logf("Order status: %s, filled: %.8f", updatedOrder.Status, updatedOrder.Filled)
		assert.Equal(t, "filled", updatedOrder.Status)
	})
}
