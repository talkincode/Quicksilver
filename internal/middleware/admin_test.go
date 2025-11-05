package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/testutil"
)

// TestAdminOnly_AdminUser 测试管理员用户访问
func TestAdminOnly_AdminUser(t *testing.T) {
	// Given: 创建管理员用户
	db := testutil.NewTestDB(t)
	user := testutil.SeedUser(t, db)
	user.Role = "admin"
	db.Save(user)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟 Auth 中间件设置的 Context
	c.Set("user_id", user.ID)
	c.Set("user", user)

	// 创建测试 handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When: 使用管理员中间件
	middleware := AdminOnly()
	handler := middleware(testHandler)

	// Then: 应该允许访问
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestAdminOnly_RegularUser 测试普通用户被拦截
func TestAdminOnly_RegularUser(t *testing.T) {
	// Given: 创建普通用户
	db := testutil.NewTestDB(t)
	user := testutil.SeedUser(t, db)
	user.Role = "user" // 明确设置为普通用户
	db.Save(user)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟 Auth 中间件设置的 Context
	c.Set("user_id", user.ID)
	c.Set("user", user)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When: 使用管理员中间件
	middleware := AdminOnly()
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 403 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusForbidden, he.Code)
		assert.Contains(t, he.Message, "Admin privileges required")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAdminOnly_NoUser 测试未认证用户
func TestAdminOnly_NoUser(t *testing.T) {
	// Given: 没有用户信息
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 不设置 user 到 Context

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When: 使用管理员中间件
	middleware := AdminOnly()
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 401 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
		assert.Contains(t, he.Message, "Authentication required")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAdminOnly_InvalidUserContext 测试无效的用户 Context
func TestAdminOnly_InvalidUserContext(t *testing.T) {
	// Given: Context 中的用户数据类型错误
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 设置错误类型的数据
	c.Set("user", "invalid-type")

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When: 使用管理员中间件
	middleware := AdminOnly()
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 500 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		assert.Contains(t, he.Message, "Invalid user context")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAdminOnly_ChainedWithAuth 测试与 Auth 中间件链式使用
func TestAdminOnly_ChainedWithAuth(t *testing.T) {
	// Given: 完整的认证链
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	// 创建管理员用户
	user := testutil.SeedUser(t, db)
	user.Role = "admin"
	db.Save(user)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	req.Header.Set("X-API-Secret", user.APISecret)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		// 验证用户信息
		userID := c.Get("user_id")
		require.NotNil(t, userID)
		assert.Equal(t, user.ID, userID.(uint))

		contextUser := c.Get("user")
		require.NotNil(t, contextUser)
		assert.Equal(t, "admin", contextUser.(*model.User).Role)

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When: 链式使用两个中间件
	authMiddleware := Auth(db, cfg)
	adminMiddleware := AdminOnly()
	handler := authMiddleware(adminMiddleware(testHandler))

	// Then: 应该成功通过认证和管理员验证
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
