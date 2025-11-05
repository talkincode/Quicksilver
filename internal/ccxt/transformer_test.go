package ccxt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/quicksilver/internal/model"
)

func TestTransformKline(t *testing.T) {
	t.Run("Transform kline to CCXT OHLCV format", func(t *testing.T) {
		openTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		kline := &model.Kline{
			Symbol:    "BTC/USDT",
			Interval:  "1h",
			OpenTime:  openTime,
			CloseTime: openTime.Add(1 * time.Hour),
			Open:      50000.0,
			High:      51000.0,
			Low:       49500.0,
			Close:     50500.0,
			Volume:    123.456,
		}

		result := TransformKline(kline)

		require.Len(t, result, 6)
		assert.Equal(t, openTime.UnixMilli(), result[0])
		assert.Equal(t, 50000.0, result[1])
		assert.Equal(t, 51000.0, result[2])
		assert.Equal(t, 49500.0, result[3])
		assert.Equal(t, 50500.0, result[4])
		assert.Equal(t, 123.456, result[5])
	})
}

func TestTransformKlines(t *testing.T) {
	t.Run("Transform multiple klines", func(t *testing.T) {
		openTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		klines := []model.Kline{
			{
				Symbol:    "BTC/USDT",
				Interval:  "1h",
				OpenTime:  openTime,
				CloseTime: openTime.Add(1 * time.Hour),
				Open:      50000.0,
				High:      51000.0,
				Low:       49500.0,
				Close:     50500.0,
				Volume:    100.0,
			},
			{
				Symbol:    "BTC/USDT",
				Interval:  "1h",
				OpenTime:  openTime.Add(1 * time.Hour),
				CloseTime: openTime.Add(2 * time.Hour),
				Open:      50500.0,
				High:      52000.0,
				Low:       50000.0,
				Close:     51500.0,
				Volume:    150.0,
			},
		}

		result := TransformKlines(klines)

		require.Len(t, result, 2)
		assert.Equal(t, openTime.UnixMilli(), result[0][0])
		assert.Equal(t, 50000.0, result[0][1])
		assert.Equal(t, openTime.Add(1*time.Hour).UnixMilli(), result[1][0])
		assert.Equal(t, 50500.0, result[1][1])
	})
}

func TestTransformTicker(t *testing.T) {
	t.Run("Transform ticker with all fields", func(t *testing.T) {
		// Given: 一个完整的 Ticker 模型
		now := time.Now()
		bidPrice := 49999.0
		askPrice := 50001.0
		high24h := 51000.0
		low24h := 49000.0
		volume24hBase := 123.45678901
		volume24hQuote := 6172839.50

		ticker := &model.Ticker{
			Symbol:         "BTC/USDT",
			LastPrice:      50000.12345678,
			BidPrice:       &bidPrice,
			AskPrice:       &askPrice,
			High24h:        &high24h,
			Low24h:         &low24h,
			Volume24hBase:  &volume24hBase,
			Volume24hQuote: &volume24hQuote,
			UpdatedAt:      now,
		}

		// When: 转换为 CCXT 格式
		result := TransformTicker(ticker)

		// Then: 验证所有字段正确转换
		assert.Equal(t, "BTC/USDT", result["symbol"])
		assert.Equal(t, now.UnixMilli(), result["timestamp"])
		assert.Equal(t, now.Format(time.RFC3339Nano), result["datetime"])
		assert.Equal(t, 51000.0, result["high"])
		assert.Equal(t, 49000.0, result["low"])
		assert.Equal(t, 49999.0, result["bid"])
		assert.Equal(t, 50001.0, result["ask"])
		assert.Equal(t, 50000.12345678, result["last"])
		assert.Equal(t, 50000.12345678, result["close"])
		assert.Equal(t, 123.45678901, result["baseVolume"])
		assert.Equal(t, 6172839.50, result["quoteVolume"])
		assert.NotNil(t, result["info"])
	})

	t.Run("Transform ticker with minimal fields", func(t *testing.T) {
		// Given: 只有必填字段的 Ticker
		ticker := &model.Ticker{
			Symbol:    "ETH/USDT",
			LastPrice: 3000.0,
			UpdatedAt: time.Now(),
		}

		// When: 转换
		result := TransformTicker(ticker)

		// Then: 基础字段存在
		assert.Equal(t, "ETH/USDT", result["symbol"])
		assert.Equal(t, 3000.0, result["last"])
		assert.NotZero(t, result["timestamp"])
	})
}

func TestTransformOrder(t *testing.T) {
	t.Run("Transform market buy order", func(t *testing.T) {
		// Given: 市价买单
		orderID := uint(12345)
		price := 50000.0
		now := time.Now()
		order := &model.Order{
			ID:            orderID,
			UserID:        1,
			ClientOrderID: "user_order_001",
			Symbol:        "BTC/USDT",
			Type:          "market",
			Side:          "buy",
			Price:         &price,
			Amount:        0.5,
			Filled:        0.3,
			Status:        "open",
			Fee:           0.015,
			FeeAsset:      "USDT",
			CreatedAt:     now,
		}

		// When: 转换
		result := TransformOrder(order)

		// Then: 验证 CCXT 格式
		assert.Equal(t, "12345", result["id"])
		assert.Equal(t, "user_order_001", result["clientOrderId"])
		assert.Equal(t, "BTC/USDT", result["symbol"])
		assert.Equal(t, "market", result["type"])
		assert.Equal(t, "buy", result["side"])
		assert.Equal(t, 50000.0, result["price"])
		assert.Equal(t, 0.5, result["amount"])
		assert.Equal(t, 0.3, result["filled"])
		assert.Equal(t, 0.2, result["remaining"])
		assert.Equal(t, "open", result["status"])
		assert.Equal(t, now.UnixMilli(), result["timestamp"])
		assert.Equal(t, now.Format(time.RFC3339Nano), result["datetime"])

		// 验证费用结构
		fee, ok := result["fee"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 0.015, fee["cost"])
		assert.Equal(t, "USDT", fee["currency"])
	})

	t.Run("Transform limit sell order", func(t *testing.T) {
		// Given: 限价卖单
		price := 51000.0
		order := &model.Order{
			ID:     uint(67890),
			Symbol: "BTC/USDT",
			Type:   "limit",
			Side:   "sell",
			Price:  &price,
			Amount: 1.0,
			Filled: 1.0,
			Status: "closed",
		}

		// When: 转换
		result := TransformOrder(order)

		// Then: 验证
		assert.Equal(t, "67890", result["id"])
		assert.Equal(t, "limit", result["type"])
		assert.Equal(t, "sell", result["side"])
		assert.Equal(t, 51000.0, result["price"])
		assert.Equal(t, 0.0, result["remaining"])
		assert.Equal(t, "closed", result["status"])
	})

	t.Run("Transform order without price", func(t *testing.T) {
		// Given: 市价单可能没有价格
		order := &model.Order{
			ID:     uint(111),
			Symbol: "ETH/USDT",
			Type:   "market",
			Side:   "buy",
			Amount: 2.0,
			Status: "new",
		}

		// When: 转换
		result := TransformOrder(order)

		// Then: price 应为 nil 或 0
		price := result["price"]
		assert.True(t, price == nil || price == 0.0)
	})
}

func TestTransformTrade(t *testing.T) {
	t.Run("Transform trade with all fields", func(t *testing.T) {
		// Given: 完整的成交记录
		now := time.Now()
		trade := &model.Trade{
			ID:        uint(999),
			OrderID:   uint(12345),
			UserID:    1,
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Price:     50000.0,
			Amount:    0.5,
			Fee:       0.025,
			CreatedAt: now,
		}

		// When: 转换
		result := TransformTrade(trade)

		// Then: 验证 CCXT 格式
		assert.Equal(t, "999", result["id"])
		assert.Equal(t, "12345", result["order"])
		assert.Equal(t, "BTC/USDT", result["symbol"])
		assert.Equal(t, "buy", result["side"])
		assert.Equal(t, 50000.0, result["price"])
		assert.Equal(t, 0.5, result["amount"])
		assert.Equal(t, 25000.0, result["cost"]) // price * amount
		assert.Equal(t, now.UnixMilli(), result["timestamp"])
		assert.Equal(t, now.Format(time.RFC3339Nano), result["datetime"])

		// 验证费用
		fee, ok := result["fee"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 0.025, fee["cost"])
		assert.Equal(t, "USDT", fee["currency"])
	})

	t.Run("Calculate cost correctly", func(t *testing.T) {
		// Given: 不同价格和数量的成交
		tests := []struct {
			name   string
			price  float64
			amount float64
			want   float64
		}{
			{"Small trade", 50000.0, 0.1, 5000.0},
			{"Large trade", 3000.0, 10.0, 30000.0},
			{"Fractional", 1000.5, 0.5555, 555.77775},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				trade := &model.Trade{
					Symbol: "BTC/USDT",
					Price:  tt.price,
					Amount: tt.amount,
				}

				result := TransformTrade(trade)
				assert.InDelta(t, tt.want, result["cost"], 0.0001)
			})
		}
	})
}

func TestTransformBalance(t *testing.T) {
	t.Run("Transform balance with used and free", func(t *testing.T) {
		// Given: 余额记录
		balance := &model.Balance{
			UserID:    1,
			Asset:     "USDT",
			Available: 10000.0,
			Locked:    500.0,
		}

		// When: 转换
		result := TransformBalance(balance)

		// Then: 验证 CCXT 格式
		assert.Equal(t, "USDT", result["currency"])
		assert.Equal(t, 10000.0, result["free"])
		assert.Equal(t, 500.0, result["used"])
		assert.Equal(t, 10500.0, result["total"]) // free + used
	})

	t.Run("Transform multiple balances", func(t *testing.T) {
		// Given: 多个资产余额
		balances := []*model.Balance{
			{Asset: "BTC", Available: 1.5, Locked: 0.5},
			{Asset: "ETH", Available: 10.0, Locked: 0.0},
			{Asset: "USDT", Available: 5000.0, Locked: 1000.0},
		}

		// When: 转换所有余额
		result := TransformBalances(balances)

		// Then: 验证结构
		assert.Len(t, result, 3)

		// 验证 BTC
		btc := result["BTC"].(map[string]interface{})
		assert.Equal(t, 1.5, btc["free"])
		assert.Equal(t, 0.5, btc["used"])
		assert.Equal(t, 2.0, btc["total"])

		// 验证 ETH
		eth := result["ETH"].(map[string]interface{})
		assert.Equal(t, 10.0, eth["free"])
		assert.Equal(t, 0.0, eth["used"])
		assert.Equal(t, 10.0, eth["total"])

		// 验证 USDT
		usdt := result["USDT"].(map[string]interface{})
		assert.Equal(t, 5000.0, usdt["free"])
		assert.Equal(t, 1000.0, usdt["used"])
		assert.Equal(t, 6000.0, usdt["total"])
	})
}

func TestTransformMarket(t *testing.T) {
	t.Run("Transform market info", func(t *testing.T) {
		// Given: 交易对配置
		symbol := "BTC/USDT"
		minAmount := 0.0001

		// When: 转换为 CCXT market 格式
		result := TransformMarket(symbol, minAmount)

		// Then: 验证格式
		assert.Equal(t, "BTC/USDT", result["symbol"])
		assert.Equal(t, "BTC/USDT", result["id"])
		assert.Equal(t, "BTC", result["base"])
		assert.Equal(t, "USDT", result["quote"])
		assert.Equal(t, true, result["active"])

		limits, ok := result["limits"].(map[string]interface{})
		require.True(t, ok)

		amount, ok := limits["amount"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 0.0001, amount["min"])
	})

	t.Run("Parse symbol correctly", func(t *testing.T) {
		tests := []struct {
			symbol string
			base   string
			quote  string
		}{
			{"BTC/USDT", "BTC", "USDT"},
			{"ETH/USDT", "ETH", "USDT"},
			{"SOL/USDT", "SOL", "USDT"},
		}

		for _, tt := range tests {
			t.Run(tt.symbol, func(t *testing.T) {
				result := TransformMarket(tt.symbol, 0.0001)
				assert.Equal(t, tt.base, result["base"])
				assert.Equal(t, tt.quote, result["quote"])
			})
		}
	})
}
