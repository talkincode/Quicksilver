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

// TestAuth_ValidAPIKey 测试有效的 API Key
func TestAuth_ValidAPIKey(t *testing.T) {
	// Given: 创建测试环境
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	// 创建测试用户
	user := testutil.SeedUser(t, db)

	// 创建 Echo 实例和测试请求
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	req.Header.Set("X-API-Secret", user.APISecret)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 创建测试 handler
	testHandler := func(c echo.Context) error {
		// 验证用户 ID 是否被正确设置到 Context
		userID := c.Get("user_id")
		require.NotNil(t, userID)
		assert.Equal(t, user.ID, userID.(uint))

		// 验证用户对象是否被设置
		contextUser := c.Get("user")
		require.NotNil(t, contextUser)
		assert.Equal(t, user.ID, contextUser.(*model.User).ID)

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When: 使用认证中间件
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)

	// Then: 认证应该成功
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestAuth_MissingAPIKey 测试缺少 API Key
func TestAuth_MissingAPIKey(t *testing.T) {
	// Given
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// 不设置 X-API-Key header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 401 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
		assert.Contains(t, he.Message, "API key required")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAuth_MissingAPISecret 测试缺少 API Secret
func TestAuth_MissingAPISecret(t *testing.T) {
	// Given
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "test-key")
	// 不设置 X-API-Secret header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 401 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
		assert.Contains(t, he.Message, "API secret required")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAuth_InvalidAPIKey 测试无效的 API Key
func TestAuth_InvalidAPIKey(t *testing.T) {
	// Given
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	req.Header.Set("X-API-Secret", "invalid-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 401 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
		assert.Contains(t, he.Message, "Invalid API credentials")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAuth_InvalidAPISecret 测试 API Key 正确但 Secret 错误
func TestAuth_InvalidAPISecret(t *testing.T) {
	// Given
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	user := testutil.SeedUser(t, db)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	req.Header.Set("X-API-Secret", "wrong-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 401 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
		assert.Contains(t, he.Message, "Invalid API credentials")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAuth_InactiveUser 测试被禁用的用户
func TestAuth_InactiveUser(t *testing.T) {
	// Given
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	user := testutil.SeedUser(t, db)
	// 禁用用户
	user.Status = "inactive"
	db.Save(user)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	req.Header.Set("X-API-Secret", user.APISecret)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 应该返回 403 错误
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusForbidden, he.Code)
		assert.Contains(t, he.Message, "User account is inactive")
	} else {
		t.Fatalf("Expected echo.HTTPError, got %T", err)
	}
}

// TestAuth_UpdateLastLogin 测试最后登录时间更新
func TestAuth_UpdateLastLogin(t *testing.T) {
	// Given
	db := testutil.NewTestDB(t)
	cfg := testutil.NewTestConfig()

	user := testutil.SeedUser(t, db)
	require.Nil(t, user.LastLogin)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	req.Header.Set("X-API-Secret", user.APISecret)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	// When
	middleware := Auth(db, cfg)
	handler := middleware(testHandler)
	err := handler(c)

	// Then: 认证成功且更新了 LastLogin
	require.NoError(t, err)

	var updatedUser model.User
	db.First(&updatedUser, user.ID)
	assert.NotNil(t, updatedUser.LastLogin)
}
