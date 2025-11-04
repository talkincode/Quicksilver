package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/testutil"
)

func TestNewBalanceService(t *testing.T) {
	// Given
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	// When
	balanceService := NewBalanceService(db, cfg, logger)

	// Then
	assert.NotNil(t, balanceService)
	assert.NotNil(t, balanceService.db)
	assert.NotNil(t, balanceService.cfg)
	assert.NotNil(t, balanceService.logger)
}

func TestGetBalance(t *testing.T) {
	t.Run("Get existing balance", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		// 创建测试用户和余额
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When
		balance, err := balanceService.GetBalance(user.ID, "USDT")

		// Then
		require.NoError(t, err)
		assert.NotNil(t, balance)
		assert.Equal(t, user.ID, balance.UserID)
		assert.Equal(t, "USDT", balance.Asset)
		assert.Equal(t, 10000.0, balance.Available)
		assert.Equal(t, 0.0, balance.Locked)
	})

	t.Run("Get non-existent balance", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)

		// When: 查询不存在的资产
		balance, err := balanceService.GetBalance(user.ID, "BTC")

		// Then: 应该返回错误
		require.Error(t, err)
		assert.Nil(t, balance)
		assert.Contains(t, err.Error(), "balance not found")
	})

	t.Run("Get balance for non-existent user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		// When: 查询不存在的用户
		balance, err := balanceService.GetBalance(99999, "USDT")

		// Then
		require.Error(t, err)
		assert.Nil(t, balance)
	})
}

func TestGetAllBalances(t *testing.T) {
	t.Run("Get all balances for user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)
		testutil.SeedBalance(t, db, user.ID, "BTC", 0.5)
		testutil.SeedBalance(t, db, user.ID, "ETH", 5.0)

		// When
		balances, err := balanceService.GetAllBalances(user.ID)

		// Then
		require.NoError(t, err)
		assert.Len(t, balances, 3)

		// 验证包含所有资产
		assets := make(map[string]bool)
		for _, b := range balances {
			assets[b.Asset] = true
		}
		assert.True(t, assets["USDT"])
		assert.True(t, assets["BTC"])
		assert.True(t, assets["ETH"])
	})

	t.Run("Get empty balances for new user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)

		// When
		balances, err := balanceService.GetAllBalances(user.ID)

		// Then
		require.NoError(t, err)
		assert.Empty(t, balances)
	})
}

func TestFreezeBalance(t *testing.T) {
	t.Run("Freeze balance successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When: 冻结 500 USDT
		err := balanceService.FreezeBalance(user.ID, "USDT", 500.0)

		// Then
		require.NoError(t, err)

		// 验证余额状态
		balance, _ := balanceService.GetBalance(user.ID, "USDT")
		assert.Equal(t, 9500.0, balance.Available)
		assert.Equal(t, 500.0, balance.Locked)
	})

	t.Run("Freeze balance with insufficient funds", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 100.0)

		// When: 尝试冻结超过可用余额
		err := balanceService.FreezeBalance(user.ID, "USDT", 500.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")

		// 验证余额未改变
		balance, _ := balanceService.GetBalance(user.ID, "USDT")
		assert.Equal(t, 100.0, balance.Available)
		assert.Equal(t, 0.0, balance.Locked)
	})

	t.Run("Freeze negative amount", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When: 尝试冻结负数金额
		err := balanceService.FreezeBalance(user.ID, "USDT", -100.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("Freeze zero amount", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When
		err := balanceService.FreezeBalance(user.ID, "USDT", 0.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

func TestUnfreezeBalance(t *testing.T) {
	t.Run("Unfreeze balance successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 9500.0)

		// 先冻结一些资金
		balanceService.FreezeBalance(user.ID, "USDT", 500.0)

		// When: 解冻 300 USDT
		err := balanceService.UnfreezeBalance(user.ID, "USDT", 300.0)

		// Then
		require.NoError(t, err)

		// 验证余额状态
		balance, _ := balanceService.GetBalance(user.ID, "USDT")
		assert.Equal(t, 9300.0, balance.Available)
		assert.Equal(t, 200.0, balance.Locked)
	})

	t.Run("Unfreeze more than locked", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 9500.0)
		balanceService.FreezeBalance(user.ID, "USDT", 500.0)

		// When: 尝试解冻超过已冻结金额
		err := balanceService.UnfreezeBalance(user.ID, "USDT", 1000.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient locked balance")

		// 验证余额未改变
		balance, _ := balanceService.GetBalance(user.ID, "USDT")
		assert.Equal(t, 9000.0, balance.Available)
		assert.Equal(t, 500.0, balance.Locked)
	})

	t.Run("Unfreeze negative amount", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When
		err := balanceService.UnfreezeBalance(user.ID, "USDT", -100.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

func TestDeductBalance(t *testing.T) {
	t.Run("Deduct from locked balance successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 9500.0)
		balanceService.FreezeBalance(user.ID, "USDT", 500.0)

		// When: 从冻结余额扣除 300
		err := balanceService.DeductBalance(user.ID, "USDT", 300.0)

		// Then
		require.NoError(t, err)

		// 验证余额状态
		balance, _ := balanceService.GetBalance(user.ID, "USDT")
		assert.Equal(t, 9000.0, balance.Available)
		assert.Equal(t, 200.0, balance.Locked)
	})

	t.Run("Deduct more than locked", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 9500.0)
		balanceService.FreezeBalance(user.ID, "USDT", 500.0)

		// When: 尝试扣除超过冻结金额
		err := balanceService.DeductBalance(user.ID, "USDT", 1000.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient locked balance")
	})

	t.Run("Deduct negative amount", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When
		err := balanceService.DeductBalance(user.ID, "USDT", -100.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

func TestAddBalance(t *testing.T) {
	t.Run("Add to available balance successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When: 增加 5000 USDT
		err := balanceService.AddBalance(user.ID, "USDT", 5000.0)

		// Then
		require.NoError(t, err)

		// 验证余额状态
		balance, _ := balanceService.GetBalance(user.ID, "USDT")
		assert.Equal(t, 15000.0, balance.Available)
		assert.Equal(t, 0.0, balance.Locked)
	})

	t.Run("Add balance creates new record if not exists", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)

		// When: 为新资产添加余额
		err := balanceService.AddBalance(user.ID, "BTC", 0.5)

		// Then
		require.NoError(t, err)

		// 验证余额已创建
		balance, err := balanceService.GetBalance(user.ID, "BTC")
		require.NoError(t, err)
		assert.Equal(t, 0.5, balance.Available)
		assert.Equal(t, 0.0, balance.Locked)
	})

	t.Run("Add negative amount", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When
		err := balanceService.AddBalance(user.ID, "USDT", -100.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

func TestTransferBalance(t *testing.T) {
	t.Run("Transfer balance successfully", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		// 创建两个用户 - 使用不同邮箱
		user1 := testutil.CreateTestUser(t, db)
		testutil.SeedBalance(t, db, user1.ID, "USDT", 10000.0)

		// 创建第二个用户时修改邮箱
		user2 := &model.User{
			Email:     "user2@example.com",
			Username:  "testuser2",
			APIKey:    "test-api-key-2",
			APISecret: "test-api-secret-2",
			Status:    "active",
		}
		require.NoError(t, db.Create(user2).Error)
		testutil.SeedBalance(t, db, user2.ID, "USDT", 5000.0)

		// When: 从 user1 转账 1000 给 user2
		err := balanceService.TransferBalance(user1.ID, user2.ID, "USDT", 1000.0)

		// Then
		require.NoError(t, err)

		// 验证余额
		balance1, _ := balanceService.GetBalance(user1.ID, "USDT")
		assert.Equal(t, 9000.0, balance1.Available)

		balance2, _ := balanceService.GetBalance(user2.ID, "USDT")
		assert.Equal(t, 6000.0, balance2.Available)
	})

	t.Run("Transfer with insufficient balance", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user1 := testutil.CreateTestUser(t, db)
		testutil.SeedBalance(t, db, user1.ID, "USDT", 100.0)

		user2 := &model.User{
			Email:     "user3@example.com",
			Username:  "testuser3",
			APIKey:    "test-api-key-3",
			APISecret: "test-api-secret-3",
			Status:    "active",
		}
		require.NoError(t, db.Create(user2).Error)
		testutil.SeedBalance(t, db, user2.ID, "USDT", 0.0)

		// When: 尝试转账超过余额
		err := balanceService.TransferBalance(user1.ID, user2.ID, "USDT", 500.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")

		// 验证余额未改变
		balance1, _ := balanceService.GetBalance(user1.ID, "USDT")
		assert.Equal(t, 100.0, balance1.Available)
	})

	t.Run("Transfer to same user", func(t *testing.T) {
		// Given
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()
		balanceService := NewBalanceService(db, cfg, logger)

		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

		// When: 尝试转账给自己
		err := balanceService.TransferBalance(user.ID, user.ID, "USDT", 1000.0)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot transfer to yourself")
	})
}
