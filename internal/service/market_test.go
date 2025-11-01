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

		// 验证所有交易对都已更新
		var tickers []model.Ticker
		db.Find(&tickers)
		assert.Equal(t, 3, len(tickers))

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
