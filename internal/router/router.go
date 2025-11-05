package router

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/api"
	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/middleware"
	"github.com/talkincode/quicksilver/internal/service"
)

// SetupRoutes 设置路由
func SetupRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config, logger *zap.Logger) {
	// 初始化服务层
	balanceService := service.NewBalanceService(db, cfg, logger)
	orderService := service.NewOrderService(db, cfg, logger, balanceService)
	userService := service.NewUserService(db, cfg, logger)
	klineService := service.NewKlineService(db, cfg, logger)

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
		public.GET("/markets", api.GetMarkets(cfg))
		public.GET("/ticker/:symbol", api.GetTicker(db))
		public.GET("/trades/:symbol", api.GetTrades(db))
		public.GET("/ohlcv/:symbol", api.GetOHLCV(klineService)) // K线数据
	}

	// 私有接口（需要认证）
	private := v1.Group("")
	private.Use(middleware.Auth(db, cfg)) // ✅ 启用认证中间件
	{
		private.GET("/balance", api.GetBalance(db))
		private.POST("/order", api.CreateOrder(orderService))
		private.GET("/order/:id", api.GetOrder(orderService))
		private.DELETE("/order/:id", api.CancelOrder(orderService))
		private.GET("/orders", api.GetOrders(orderService))
		private.GET("/orders/open", api.GetOpenOrders(orderService))
		private.GET("/myTrades", api.GetMyTrades(db))
	}

	// 管理员接口（需要认证 + 管理员权限）
	admin := v1.Group("/admin")
	admin.Use(middleware.Auth(db, cfg)) // 先验证身份
	admin.Use(middleware.AdminOnly())   // 再验证管理员权限
	{
		// 用户管理
		admin.POST("/users", api.AdminCreateUser(userService))
		admin.GET("/users", api.AdminListUsers(userService))
		admin.GET("/users/:id", api.AdminGetUser(userService))
		admin.PUT("/users/:id", api.AdminUpdateUser(userService))
		admin.DELETE("/users/:id", api.AdminDeleteUser(userService))

		// 余额管理
		admin.GET("/users/:id/balances", api.AdminGetUserBalances(balanceService))
		admin.GET("/balances", api.AdminGetAllBalances(balanceService))
		admin.POST("/users/:id/balance/adjust", api.AdminAdjustBalance(balanceService))
	}
}
