package router

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/api"
	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/middleware"
)

// SetupRoutes 设置路由
func SetupRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config, logger *zap.Logger) {
	// 健康检查
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// API v1 路由组
	v1 := e.Group("/v1")

	// 公开接口
	public := v1.Group("")
	{
		public.GET("/ping", api.Ping)
		public.GET("/time", api.ServerTime)
		public.GET("/markets", api.GetMarkets(db))
		public.GET("/ticker/:symbol", api.GetTicker(db))
		public.GET("/trades/:symbol", api.GetTrades(db))
	}

	// 私有接口（需要认证）
	private := v1.Group("")
	private.Use(middleware.Auth(db, cfg)) // ✅ 启用认证中间件
	{
		private.GET("/balance", api.GetBalance(db))
		private.POST("/order", api.CreateOrder(db, cfg))
		private.GET("/order/:id", api.GetOrder(db))
		private.DELETE("/order/:id", api.CancelOrder(db))
		private.GET("/orders", api.GetOrders(db))
		private.GET("/orders/open", api.GetOpenOrders(db))
		private.GET("/myTrades", api.GetMyTrades(db))
	}
}
