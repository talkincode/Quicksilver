package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/quicksilver/internal/testutil"
)

func TestNewUserService(t *testing.T) {
	// Given
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	// When
	userService := NewUserService(db, cfg, logger)

	// Then
	assert.NotNil(t, userService)
	assert.NotNil(t, userService.db)
	assert.NotNil(t, userService.cfg)
	assert.NotNil(t, userService.logger)
}

func TestCreateUser(t *testing.T) {
	t.Run("Create user successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		req := CreateUserRequest{
			Email: "newuser@example.com",
		}

		// When
		user, _, err := userService.CreateUser(req)

		// Then: 用户创建成功
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "newuser@example.com", user.Email)
		assert.Equal(t, "active", user.Status)

		// And: API Key 和 Secret 已生成
		assert.NotEmpty(t, user.APIKey)
		assert.NotEmpty(t, user.APISecret)
		assert.Greater(t, len(user.APIKey), 20, "API Key should be at least 20 characters")
		assert.Greater(t, len(user.APISecret), 30, "API Secret should be at least 30 characters")
	})

	t.Run("Create user with duplicate email", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// 先创建一个用户
		req := CreateUserRequest{Email: "duplicate@example.com"}
		_, _, err := userService.CreateUser(req)
		require.NoError(t, err)

		// When: 尝试用相同邮箱创建用户
		_, _, err = userService.CreateUser(req)

		// Then: 应该返回错误
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})

	t.Run("Create user with invalid email", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		req := CreateUserRequest{Email: "invalid-email"}

		// When
		_, _, err := userService.CreateUser(req)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("Create user with empty email", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		req := CreateUserRequest{Email: ""}

		// When
		_, _, err := userService.CreateUser(req)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("Get existing user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// 创建用户
		createdUser, _, err := userService.CreateUser(CreateUserRequest{
			Email: "getuser@example.com",
		})
		require.NoError(t, err)

		// When
		user, err := userService.GetUserByID(createdUser.ID)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, createdUser.Email, user.Email)
	})

	t.Run("Get non-existent user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// When
		user, err := userService.GetUserByID(99999)

		// Then
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestGetUserByAPIKey(t *testing.T) {
	t.Run("Get user by valid API key", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		createdUser, _, err := userService.CreateUser(CreateUserRequest{
			Email: "apikey@example.com",
		})
		require.NoError(t, err)

		// When
		user, err := userService.GetUserByAPIKey(createdUser.APIKey)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, createdUser.Email, user.Email)
	})

	t.Run("Get user by invalid API key", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// When
		user, err := userService.GetUserByAPIKey("invalid-key-12345")

		// Then
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestRegenerateAPIKey(t *testing.T) {
	t.Run("Regenerate API key successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "regenerate@example.com",
		})
		require.NoError(t, err)

		oldAPIKey := user.APIKey
		oldAPISecret := user.APISecret

		// When
		newUser, _, err := userService.RegenerateAPIKey(user.ID)

		// Then
		require.NoError(t, err)
		assert.NotNil(t, newUser)
		assert.NotEqual(t, oldAPIKey, newUser.APIKey, "API Key should be different")
		assert.NotEqual(t, oldAPISecret, newUser.APISecret, "API Secret should be different")
		assert.Greater(t, len(newUser.APIKey), 20)
		assert.Greater(t, len(newUser.APISecret), 30)
	})

	t.Run("Regenerate API key for non-existent user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// When
		_, _, err := userService.RegenerateAPIKey(99999)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUpdateUserStatus(t *testing.T) {
	t.Run("Activate user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "status@example.com",
		})
		require.NoError(t, err)

		// 先设置为 inactive
		_, err = userService.UpdateUserStatus(user.ID, "inactive")
		require.NoError(t, err)

		// When: 激活用户
		updatedUser, err := userService.UpdateUserStatus(user.ID, "active")

		// Then
		require.NoError(t, err)
		assert.Equal(t, "active", updatedUser.Status)
	})

	t.Run("Suspend user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "suspend@example.com",
		})
		require.NoError(t, err)

		// When
		updatedUser, err := userService.UpdateUserStatus(user.ID, "suspended")

		// Then
		require.NoError(t, err)
		assert.Equal(t, "suspended", updatedUser.Status)
	})

	t.Run("Invalid status", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "invalidstatus@example.com",
		})
		require.NoError(t, err)

		// When
		_, err = userService.UpdateUserStatus(user.ID, "unknown")

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestGenerateAPICredentials(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()
	service := NewUserService(db, cfg, logger)

	t.Run("Generate unique credentials", func(t *testing.T) {
		// When: 生成多组凭证
		apiKey1, apiSecret1, err1 := service.generateAPICredentials()
		require.NoError(t, err1)
		apiKey2, apiSecret2, err2 := service.generateAPICredentials()
		require.NoError(t, err2)

		// Then: 应该是唯一的
		assert.NotEqual(t, apiKey1, apiKey2, "API Keys should be unique")
		assert.NotEqual(t, apiSecret1, apiSecret2, "API Secrets should be unique")
		assert.Greater(t, len(apiKey1), 20)
		assert.Greater(t, len(apiSecret1), 30)
	})
}
