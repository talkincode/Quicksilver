package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/service"
	"github.com/talkincode/quicksilver/internal/testutil"
)

// TestAdminCreateUser 测试创建用户
func TestAdminCreateUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	userService := service.NewUserService(db, cfg, logger)
	handler := AdminCreateUser(userService)

	t.Run("Create user successfully", func(t *testing.T) {
		// Given: 有效的请求体
		reqBody := map[string]interface{}{
			"email": "admin@example.com",
		}
		body, _ := json.Marshal(reqBody)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// When: 调用处理器
		err := handler(c)

		// Then: 返回成功
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// And: 返回用户信息
		assert.Contains(t, response, "id")
		assert.Equal(t, "admin@example.com", response["email"])
		assert.Contains(t, response, "api_key")
		assert.Contains(t, response, "api_secret")
		assert.Equal(t, "active", response["status"])
	})

	t.Run("Create user with duplicate email", func(t *testing.T) {
		// Given: 先创建一个用户
		email := "duplicate@example.com"
		userService.CreateUser(service.CreateUserRequest{Email: email})

		// When: 尝试创建相同邮箱的用户
		reqBody := map[string]interface{}{
			"email": email,
		}
		body, _ := json.Marshal(reqBody)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// When: 尝试创建重复邮箱
		err := handler(c)

		// Then: 返回错误
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)

		var response map[string]string
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "already exists")
	})

	t.Run("Create user with invalid email", func(t *testing.T) {
		// Given: 无效邮箱
		reqBody := map[string]interface{}{
			"email": "invalid-email",
		}
		body, _ := json.Marshal(reqBody)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// When: 尝试创建
		err := handler(c)

		// Then: 返回验证错误
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestAdminListUsers 测试获取用户列表
func TestAdminListUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	userService := service.NewUserService(db, cfg, logger)
	handler := AdminListUsers(userService)

	t.Run("List users with pagination", func(t *testing.T) {
		// Given: 创建多个用户
		for i := 0; i < 5; i++ {
			testutil.SeedUser(t, db)
		}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&limit=2", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.QueryParam("page")
		c.QueryParam("limit")

		// When: 调用处理器
		err := handler(c)

		// Then: 返回成功
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// And: 返回分页数据
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "total")
		assert.Contains(t, response, "page")
		assert.Contains(t, response, "limit")

		users := response["data"].([]interface{})
		assert.LessOrEqual(t, len(users), 2)
	})

	t.Run("List users with search", func(t *testing.T) {
		// Given: 创建特定邮箱用户
		user := testutil.SeedUser(t, db)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/users?search="+user.Email, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// When: 搜索用户
		err := handler(c)

		// Then: 找到用户
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		users := response["data"].([]interface{})
		assert.Equal(t, 1, len(users))
	})
}

// TestAdminGetUser 测试获取单个用户
func TestAdminGetUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	userService := service.NewUserService(db, cfg, logger)
	handler := AdminGetUser(userService)

	t.Run("Get user successfully", func(t *testing.T) {
		// Given: 用户存在
		user := testutil.SeedUser(t, db)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/users/"+fmt.Sprint(user.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(user.ID))

		// When: 获取用户
		err := handler(c)

		// Then: 返回成功
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response model.User
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Email, response.Email)
	})

	t.Run("Get non-existent user", func(t *testing.T) {
		// Given: 用户不存在
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/users/99999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("99999")

		// When: 获取用户
		err := handler(c)

		// Then: 返回 404
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// TestAdminUpdateUser 测试更新用户
func TestAdminUpdateUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	userService := service.NewUserService(db, cfg, logger)
	handler := AdminUpdateUser(userService)

	t.Run("Update user status", func(t *testing.T) {
		// Given: 用户存在
		user := testutil.SeedUser(t, db)

		reqBody := map[string]interface{}{
			"status": "inactive",
		}
		body, _ := json.Marshal(reqBody)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/admin/users/"+fmt.Sprint(user.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(user.ID))

		// When: 更新状态
		err := handler(c)

		// Then: 返回成功
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// And: 状态已更新
		var updated model.User
		db.First(&updated, user.ID)
		assert.Equal(t, "inactive", updated.Status)
	})

	t.Run("Regenerate API key", func(t *testing.T) {
		// Given: 用户存在
		user := testutil.SeedUser(t, db)
		oldAPIKey := user.APIKey

		reqBody := map[string]interface{}{
			"regenerate_api_key": true,
		}
		body, _ := json.Marshal(reqBody)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/admin/users/"+fmt.Sprint(user.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(user.ID))

		// When: 重新生成 API Key
		err := handler(c)

		// Then: 返回新的 API Key
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NotEqual(t, oldAPIKey, response["api_key"])
		assert.Contains(t, response, "api_secret")
	})
}

// TestAdminDeleteUser 测试删除用户
func TestAdminDeleteUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	userService := service.NewUserService(db, cfg, logger)
	handler := AdminDeleteUser(userService)

	t.Run("Delete user (soft delete)", func(t *testing.T) {
		// Given: 用户存在
		user := testutil.SeedUser(t, db)

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+fmt.Sprint(user.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(user.ID))

		// When: 删除用户
		err := handler(c)

		// Then: 返回成功
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// And: 用户状态变为 inactive
		var deleted model.User
		db.First(&deleted, user.ID)
		assert.Equal(t, "inactive", deleted.Status)
	})
}
