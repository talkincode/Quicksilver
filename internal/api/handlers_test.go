package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/service"
	"github.com/talkincode/quicksilver/internal/testutil"
)

// TestPing 测试健康检查端点
func TestPing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Ping(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.NotEmpty(t, response["time"])

	// 验证时间格式
	_, err = time.Parse(time.RFC3339, response["time"])
	assert.NoError(t, err, "time should be in RFC3339 format")
}

// TestServerTime 测试服务器时间端点
func TestServerTime(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/time", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ServerTime(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["timestamp"])
	assert.NotNil(t, response["datetime"])

	// 验证时间戳是合理的（近期时间）
	timestamp := int64(response["timestamp"].(float64))
	now := time.Now().Unix()
	assert.InDelta(t, now, timestamp, 2.0, "timestamp should be close to current time")
}

// TestGetMarkets 测试获取交易对列表
func TestGetMarkets(t *testing.T) {
	cfg := testutil.NewTestConfig()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/markets", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := GetMarkets(cfg)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var markets []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &markets)
	require.NoError(t, err)

	assert.NotEmpty(t, markets)
	assert.Equal(t, "BTC/USDT", markets[0]["symbol"])
	assert.Equal(t, "BTC", markets[0]["base"])
	assert.Equal(t, "USDT", markets[0]["quote"])
	assert.Equal(t, true, markets[0]["active"])
}

// TestGetTicker 测试获取行情数据
func TestGetTicker(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 准备测试数据
	ticker := &model.Ticker{
		Symbol:    "BTC/USDT",
		LastPrice: 50000.5,
		BidPrice:  testutil.Float64Ptr(50000.0),
		AskPrice:  testutil.Float64Ptr(50001.0),
		High24h:   testutil.Float64Ptr(51000.0),
		Low24h:    testutil.Float64Ptr(49000.0),
		Source:    "hyperliquid",
	}
	db.Create(ticker)

	tests := []struct {
		name           string
		symbol         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Get existing ticker with slash format",
			symbol:         "BTC/USDT",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Get existing ticker with dash format",
			symbol:         "BTC-USDT",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Get non-existing ticker",
			symbol:         "ETH/USDT",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/v1/ticker/"+tt.symbol, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("symbol")
			c.SetParamValues(tt.symbol)

			handler := GetTicker(db)
			err := handler(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectError {
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)
				assert.Contains(t, response["error"], "not found")
			} else {
				// 期望 CCXT 格式的响应
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "BTC/USDT", response["symbol"])
				assert.Equal(t, 50000.5, response["last"])
			}
		})
	}
}

// TestGetTrades 测试获取成交记录
func TestGetTrades(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 准备测试数据
	user := testutil.SeedUser(t, db)
	order := testutil.SeedOrder(t, db, user.ID, "BTC/USDT")

	// 创建多条成交记录
	for i := 0; i < 5; i++ {
		trade := &model.Trade{
			OrderID: order.ID,
			UserID:  user.ID,
			Symbol:  "BTC/USDT",
			Side:    "buy",
			Price:   50000.0 + float64(i)*10,
			Amount:  0.1,
		}
		db.Create(trade)
	}

	tests := []struct {
		name           string
		symbol         string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get trades for existing symbol",
			symbol:         "BTC/USDT",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			name:           "Get trades with dash format",
			symbol:         "BTC-USDT",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			name:           "Get trades for non-existing symbol",
			symbol:         "ETH/USDT",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/v1/trades/"+tt.symbol, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("symbol")
			c.SetParamValues(tt.symbol)

			handler := GetTrades(db)
			err := handler(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// 期望 CCXT 格式的响应（数组of maps）
			var trades []map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &trades)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(trades))
		})
	}
}

// TestGetBalance 测试获取余额
func TestGetBalance(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 准备测试数据
	user := testutil.SeedUser(t, db)
	testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)
	testutil.SeedBalance(t, db, user.ID, "BTC", 0.5)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/balance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟认证中间件设置的用户信息
	c.Set("user_id", user.ID)
	c.Set("user", user)

	handler := GetBalance(db)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 期望 CCXT 格式的响应（map of asset->balance）
	var balances map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &balances)
	require.NoError(t, err)

	// 应该返回 2 个资产
	assert.Equal(t, 2, len(balances))

	// 验证 USDT 余额
	usdtBalance, ok := balances["USDT"].(map[string]interface{})
	require.True(t, ok, "USDT balance should exist")
	assert.Equal(t, 10000.0, usdtBalance["free"])
	assert.Equal(t, 0.0, usdtBalance["used"])
	assert.Equal(t, 10000.0, usdtBalance["total"])

	// 验证 BTC 余额
	btcBalance, ok := balances["BTC"].(map[string]interface{})
	require.True(t, ok, "BTC balance should exist")
	assert.Equal(t, 0.5, btcBalance["free"])
	assert.Equal(t, 0.0, btcBalance["used"])
	assert.Equal(t, 0.5, btcBalance["total"])
}

// TestCreateOrder 测试创建订单
func TestCreateOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := zap.NewNop()

	// 初始化服务
	balanceService := service.NewBalanceService(db, cfg, logger)
	orderService := service.NewOrderService(db, cfg, logger, balanceService)

	e := echo.New()

	// 测试创建订单
	orderJSON := `{"symbol":"BTC/USDT","side":"buy","type":"market","amount":0.1}`
	req := httptest.NewRequest(http.MethodPost, "/v1/order", strings.NewReader(orderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CreateOrder(orderService)
	err := handler(c)

	require.NoError(t, err)
	// 由于没有余额和 ticker 数据，应该返回错误
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestGetOrder 测试获取订单详情
func TestGetOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := zap.NewNop()

	// 初始化服务
	balanceService := service.NewBalanceService(db, cfg, logger)
	orderService := service.NewOrderService(db, cfg, logger, balanceService)

	// 准备测试数据
	user := testutil.SeedUser(t, db)
	order := testutil.SeedOrder(t, db, user.ID, "BTC/USDT")

	tests := []struct {
		name           string
		orderID        string
		userID         uint
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Get existing order",
			orderID:        strconv.Itoa(int(order.ID)),
			userID:         user.ID,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Get non-existing order",
			orderID:        "99999",
			userID:         user.ID,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/v1/order/"+tt.orderID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.orderID)
			c.Set("user_id", tt.userID) // 模拟认证中间件

			handler := GetOrder(orderService)
			err := handler(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectError {
				var response map[string]string
				json.Unmarshal(rec.Body.Bytes(), &response)
				assert.Contains(t, response["error"], "not found")
			} else {
				// 期望 CCXT 格式的响应
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("%d", order.ID), response["id"])
				assert.Equal(t, "BTC/USDT", response["symbol"])
			}
		})
	}
}

// TestCancelOrder 测试撤销订单
func TestCancelOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := zap.NewNop()

	// 初始化服务
	balanceService := service.NewBalanceService(db, cfg, logger)
	orderService := service.NewOrderService(db, cfg, logger, balanceService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/v1/order/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	handler := CancelOrder(orderService)
	err := handler(c)

	require.NoError(t, err)
	// 订单不存在，应该返回错误
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestGetOrders 测试获取订单列表
func TestGetOrders(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := zap.NewNop()

	// 初始化服务
	balanceService := service.NewBalanceService(db, cfg, logger)
	orderService := service.NewOrderService(db, cfg, logger, balanceService)

	// 准备测试数据
	user := testutil.SeedUser(t, db)

	// 创建多个订单
	for i := 0; i < 3; i++ {
		testutil.SeedOrder(t, db, user.ID, "BTC/USDT")
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/orders", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟认证中间件设置的用户信息
	c.Set("user_id", user.ID)
	c.Set("user", user)

	handler := GetOrders(orderService)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(3), response["total"])
}

// TestGetOpenOrders 测试获取未完成订单
func TestGetOpenOrders(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()
	logger := zap.NewNop()

	// 初始化服务
	balanceService := service.NewBalanceService(db, cfg, logger)
	orderService := service.NewOrderService(db, cfg, logger, balanceService)

	// 准备测试数据
	user := testutil.SeedUser(t, db)

	// 创建不同状态的订单
	newOrder := &model.Order{
		UserID: user.ID,
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "limit",
		Status: "new",
		Price:  testutil.Float64Ptr(50000.0),
		Amount: 0.1,
	}
	db.Create(newOrder)

	filledOrder := &model.Order{
		UserID: user.ID,
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "market",
		Status: "filled",
		Amount: 0.1,
		Filled: 0.1,
	}
	db.Create(filledOrder)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/orders/open", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟认证中间件设置的用户信息
	c.Set("user_id", user.ID)
	c.Set("user", user)

	handler := GetOpenOrders(orderService)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var orders []model.Order
	err = json.Unmarshal(rec.Body.Bytes(), &orders)
	require.NoError(t, err)
	assert.Equal(t, 1, len(orders)) // 只有1个 new 状态的订单
	assert.Equal(t, "new", orders[0].Status)
}

// TestGetMyTrades 测试获取我的成交记录
func TestGetMyTrades(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 准备测试数据
	user1 := testutil.SeedUser(t, db)

	// 创建第二个用户
	user2 := &model.User{
		Email:     "test2@example.com",
		Username:  "testuser2",
		APIKey:    "test-api-key-2",
		APISecret: "test-secret-2",
		Status:    "active",
	}
	db.Create(user2)

	order1 := testutil.SeedOrder(t, db, user1.ID, "BTC/USDT")
	order2 := testutil.SeedOrder(t, db, user2.ID, "BTC/USDT")

	// user1 的成交记录
	for i := 0; i < 3; i++ {
		trade := &model.Trade{
			OrderID: order1.ID,
			UserID:  user1.ID,
			Symbol:  "BTC/USDT",
			Side:    "buy",
			Price:   50000.0,
			Amount:  0.1,
		}
		db.Create(trade)
	}

	// user2 的成交记录
	trade := &model.Trade{
		OrderID: order2.ID,
		UserID:  user2.ID,
		Symbol:  "BTC/USDT",
		Side:    "sell",
		Price:   50000.0,
		Amount:  0.1,
	}
	db.Create(trade)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/myTrades", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟认证中间件设置的用户信息
	c.Set("user_id", user1.ID)
	c.Set("user", user1)

	handler := GetMyTrades(db)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var trades []model.Trade
	err = json.Unmarshal(rec.Body.Bytes(), &trades)
	require.NoError(t, err)

	// 应该只返回 user1 的 3 条成交记录（硬编码 userID = 1）
	assert.Equal(t, 3, len(trades))

	// 验证所有成交都属于 user1
	for _, trade := range trades {
		assert.Equal(t, user1.ID, trade.UserID)
	}
}
