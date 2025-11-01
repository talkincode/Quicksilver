package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// Ping 健康检查
func Ping(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// ServerTime 获取服务器时间
func ServerTime(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"datetime":  time.Now().Format(time.RFC3339),
	})
}

// GetMarkets 获取交易对信息
func GetMarkets(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: 实现获取交易对列表逻辑
		markets := []map[string]interface{}{
			{
				"symbol":     "BTC/USDT",
				"base":       "BTC",
				"quote":      "USDT",
				"active":     true,
				"min_amount": 0.00001,
			},
		}
		return c.JSON(http.StatusOK, markets)
	}
}

// GetTicker 获取行情
func GetTicker(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 获取参数并转换格式: BTC-USDT -> BTC/USDT
		symbol := c.Param("symbol")
		symbol = strings.ReplaceAll(symbol, "-", "/")

		var ticker model.Ticker
		if err := db.Where("symbol = ?", symbol).First(&ticker).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "ticker not found",
			})
		}

		return c.JSON(http.StatusOK, ticker)
	}
}

// GetTrades 获取最近成交
func GetTrades(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 获取参数并转换格式: BTC-USDT -> BTC/USDT
		symbol := c.Param("symbol")
		symbol = strings.ReplaceAll(symbol, "-", "/")

		var trades []model.Trade
		if err := db.Where("symbol = ?", symbol).
			Order("created_at DESC").
			Limit(50).
			Find(&trades).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch trades",
			})
		}

		return c.JSON(http.StatusOK, trades)
	}
}

// GetBalance 获取余额
func GetBalance(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: 从认证中间件获取 user_id
		userID := uint(1) // 临时硬编码

		var balances []model.Balance
		if err := db.Where("user_id = ?", userID).Find(&balances).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch balance",
			})
		}

		return c.JSON(http.StatusOK, balances)
	}
}

// CreateOrder 创建订单
func CreateOrder(db *gorm.DB, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: 实现订单创建逻辑
		return c.JSON(http.StatusCreated, map[string]string{
			"message": "order created (not implemented)",
		})
	}
}

// GetOrder 获取订单详情
func GetOrder(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var order model.Order
		if err := db.Preload("Trades").First(&order, id).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "order not found",
			})
		}

		return c.JSON(http.StatusOK, order)
	}
}

// CancelOrder 撤销订单
func CancelOrder(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: 实现撤单逻辑
		return c.JSON(http.StatusOK, map[string]string{
			"message": "order canceled (not implemented)",
		})
	}
}

// GetOrders 获取订单列表
func GetOrders(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: 从认证中间件获取 user_id
		userID := uint(1)

		var orders []model.Order
		if err := db.Where("user_id = ?", userID).
			Order("created_at DESC").
			Limit(100).
			Find(&orders).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch orders",
			})
		}

		return c.JSON(http.StatusOK, orders)
	}
}

// GetOpenOrders 获取未完成订单
func GetOpenOrders(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := uint(1)

		var orders []model.Order
		if err := db.Where("user_id = ? AND status IN ?", userID, []string{"open", "partially_filled"}).
			Order("created_at DESC").
			Find(&orders).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch open orders",
			})
		}

		return c.JSON(http.StatusOK, orders)
	}
}

// GetMyTrades 获取我的成交记录
func GetMyTrades(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := uint(1)

		var trades []model.Trade
		if err := db.Where("user_id = ?", userID).
			Order("created_at DESC").
			Limit(100).
			Find(&trades).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch trades",
			})
		}

		return c.JSON(http.StatusOK, trades)
	}
}
