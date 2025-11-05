package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/model"
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

// TestDeleteUser 测试软删除用户
func TestDeleteUser(t *testing.T) {
	t.Run("Soft delete user successfully", func(t *testing.T) {
		// Given: 用户存在
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "delete@example.com",
		})
		require.NoError(t, err)
		assert.Equal(t, "active", user.Status)

		// When: 删除用户
		err = userService.DeleteUser(user.ID)

		// Then: 成功且状态变为 inactive
		require.NoError(t, err)

		// And: 用户仍然存在但状态为 inactive
		var deletedUser model.User
		err = db.First(&deletedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "inactive", deletedUser.Status)
		assert.Equal(t, user.Email, deletedUser.Email)
		assert.Equal(t, user.APIKey, deletedUser.APIKey)
	})

	t.Run("Delete non-existent user", func(t *testing.T) {
		// Given: 用户不存在
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// When: 尝试删除不存在的用户
		err := userService.DeleteUser(99999)

		// Then: 返回错误
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("Delete already inactive user", func(t *testing.T) {
		// Given: 用户已经是 inactive 状态
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "alreadyinactive@example.com",
		})
		require.NoError(t, err)

		// 先设置为 inactive
		_, err = userService.UpdateUserStatus(user.ID, "inactive")
		require.NoError(t, err)

		// When: 再次删除
		err = userService.DeleteUser(user.ID)

		// Then: 仍然成功
		require.NoError(t, err)

		// And: 状态保持 inactive
		var deletedUser model.User
		db.First(&deletedUser, user.ID)
		assert.Equal(t, "inactive", deletedUser.Status)
	})
}

// TestHardDeleteUser 测试物理删除用户
func TestHardDeleteUser(t *testing.T) {
	t.Run("Hard delete user and all related data", func(t *testing.T) {
		// Given: 用户存在并有关联数据
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)
		balanceService := NewBalanceService(db, cfg, logger)

		user, _, err := userService.CreateUser(CreateUserRequest{
			Email: "harddelete@example.com",
		})
		require.NoError(t, err)

		// 创建余额数据
		err = balanceService.AddBalance(user.ID, "USDT", 1000.0)
		require.NoError(t, err)

		// When: 彻底删除用户
		err = userService.HardDeleteUser(user.ID)

		// Then: 成功
		require.NoError(t, err)

		// And: 用户记录不存在
		var deletedUser model.User
		err = db.First(&deletedUser, user.ID).Error
		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		// And: 余额记录也被删除
		var balances []model.Balance
		err = db.Where("user_id = ?", user.ID).Find(&balances).Error
		require.NoError(t, err)
		assert.Empty(t, balances)
	})

	t.Run("Hard delete non-existent user", func(t *testing.T) {
		// Given: 用户不存在
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		userService := NewUserService(db, cfg, logger)

		// When: 尝试删除
		err := userService.HardDeleteUser(99999)

		// Then: 不报错（Delete 不存在的记录不报错）
		require.NoError(t, err)
	})
}
