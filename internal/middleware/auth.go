package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// Auth 认证中间件 - 验证 API Key 和 Secret
func Auth(db *gorm.DB, cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. 从请求头获取 API Key 和 Secret
			apiKey := c.Request().Header.Get("X-API-Key")
			apiSecret := c.Request().Header.Get("X-API-Secret")

			// 2. 验证必填字段
			if apiKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "API key required")
			}
			if apiSecret == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "API secret required")
			}

			// 3. 查询用户
			var user model.User
			if err := db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API credentials")
				}
				return echo.NewHTTPError(http.StatusInternalServerError, "Authentication failed")
			}

			// 4. 验证 API Secret
			if user.APISecret != apiSecret {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API credentials")
			}

			// 5. 检查用户状态
			if user.Status != "active" {
				return echo.NewHTTPError(http.StatusForbidden, "User account is inactive")
			}

			// 6. 更新最后登录时间
			now := time.Now()
			user.LastLogin = &now
			db.Model(&user).Update("last_login", now)

			// 7. 将用户信息存储到 Context
			c.Set("user_id", user.ID)
			c.Set("user", &user)

			// 8. 继续处理请求
			return next(c)
		}
	}
}
