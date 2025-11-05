package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/talkincode/quicksilver/internal/model"
)

// AdminOnly 管理员权限验证中间件
// 必须在 Auth 中间件之后使用
func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. 从 Context 获取用户信息（由 Auth 中间件设置）
			userInterface := c.Get("user")
			if userInterface == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
			}

			// 2. 类型断言获取用户对象
			user, ok := userInterface.(*model.User)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user context")
			}

			// 3. 检查是否为管理员
			if user.Role != "admin" {
				return echo.NewHTTPError(http.StatusForbidden, "Admin privileges required")
			}

			// 4. 继续处理请求
			return next(c)
		}
	}
}
