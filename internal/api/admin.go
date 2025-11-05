package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/quicksilver/internal/service"
)

// AdminCreateUser 创建新用户 (管理员接口)
func AdminCreateUser(userService *service.UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req service.CreateUserRequest

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		// 创建用户
		user, apiSecret, err := userService.CreateUser(req)
		if err != nil {
			if err.Error() == "email already exists" {
				return c.JSON(http.StatusConflict, map[string]string{
					"error": "user with this email already exists",
				})
			}
			if err.Error() == "invalid email format" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid email format",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to create user",
			})
		}

		// 返回用户信息（包含 API Secret，仅显示一次）
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"api_key":    user.APIKey,
			"api_secret": apiSecret, // 仅创建时返回
			"status":     user.Status,
			"created_at": user.CreatedAt,
		})
	}
}

// AdminListUsers 获取用户列表 (管理员接口)
func AdminListUsers(userService *service.UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析分页参数
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		if limit < 1 || limit > 100 {
			limit = 20
		}

		search := c.QueryParam("search")
		status := c.QueryParam("status")

		// 获取用户列表
		users, total, err := userService.ListUsers(page, limit, search, status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch users",
			})
		}

		// 返回分页数据
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":  users,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// AdminGetUser 获取单个用户详情 (管理员接口)
func AdminGetUser(userService *service.UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析用户 ID
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		// 获取用户
		user, err := userService.GetUserByID(uint(id))
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
		}

		return c.JSON(http.StatusOK, user)
	}
}

// AdminUpdateUser 更新用户信息 (管理员接口)
func AdminUpdateUser(userService *service.UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析用户 ID
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		var req struct {
			Status           *string `json:"status"`
			RegenerateAPIKey bool    `json:"regenerate_api_key"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		// 更新用户状态
		if req.Status != nil {
			if _, err := userService.UpdateUserStatus(uint(id), *req.Status); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to update user status",
				})
			}
		}

		// 重新生成 API Key
		if req.RegenerateAPIKey {
			user, apiSecret, err := userService.RegenerateAPIKey(uint(id))
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to regenerate API key",
				})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"id":         user.ID,
				"email":      user.Email,
				"api_key":    user.APIKey,
				"api_secret": apiSecret, // 仅重新生成时返回
				"status":     user.Status,
				"updated_at": user.UpdatedAt,
			})
		}

		// 返回更新后的用户
		user, err := userService.GetUserByID(uint(id))
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
		}

		return c.JSON(http.StatusOK, user)
	}
}

// AdminDeleteUser 删除用户 (软删除，设置为 inactive) (管理员接口)
func AdminDeleteUser(userService *service.UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析用户 ID
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		// 软删除：设置状态为 inactive
		if _, err := userService.UpdateUserStatus(uint(id), "inactive"); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to delete user",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "user deleted successfully",
		})
	}
}
