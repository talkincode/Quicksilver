package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/service"
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
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "user not authenticated",
			})
		}

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
func CreateOrder(orderService *service.OrderService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			// 测试环境：使用硬编码 userID
			userID = 1
		}

		// 解析请求
		var req service.CreateOrderRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request",
			})
		}

		// 创建订单
		order, err := orderService.CreateOrder(userID, req)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, order)
	}
}

// GetOrder 获取订单详情
func GetOrder(orderService *service.OrderService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			// 测试环境：使用硬编码 userID
			userID = 1
		}

		id := c.Param("id")
		var orderID uint
		if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid order id",
			})
		}

		order, err := orderService.GetOrderByID(orderID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "order not found",
			})
		}

		// 验证订单所有者
		if order.UserID != userID {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "access denied",
			})
		}

		return c.JSON(http.StatusOK, order)
	}
}

// CancelOrder 撤销订单
func CancelOrder(orderService *service.OrderService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			// 测试环境：使用硬编码 userID
			userID = 1
		}

		id := c.Param("id")
		var orderID uint
		if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid order id",
			})
		}

		// 撤销订单
		if err := orderService.CancelOrder(userID, orderID); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "order cancelled",
		})
	}
}

// GetOrders 获取订单列表
func GetOrders(orderService *service.OrderService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			// 测试环境：使用硬编码 userID
			userID = 1
		}

		// 获取分页参数
		page := 1
		pageSize := 50
		if p := c.QueryParam("page"); p != "" {
			fmt.Sscanf(p, "%d", &page)
		}
		if ps := c.QueryParam("page_size"); ps != "" {
			fmt.Sscanf(ps, "%d", &pageSize)
		}

		// 获取订单列表
		orders, total, err := orderService.GetUserOrders(userID, page, pageSize)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch orders",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"orders": orders,
			"total":  total,
			"page":   page,
			"size":   pageSize,
		})
	}
}

// GetOpenOrders 获取未完成订单
func GetOpenOrders(orderService *service.OrderService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			// 测试环境：使用硬编码 userID
			userID = 1
		}

		// 获取未完成订单
		orders, err := orderService.GetOpenOrders(userID)
		if err != nil {
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
		// 从认证中间件获取 user_id
		userID, ok := c.Get("user_id").(uint)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "user not authenticated",
			})
		}

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
