package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/testutil"
)

// TestFreezeBalanceConcurrency 测试并发冻结余额的竞态条件
func TestFreezeBalanceConcurrency(t *testing.T) {
	t.Skip("Skipping due to SQLite memory database concurrency limitations")

	t.Run("Concurrent freeze operations with race protection", func(t *testing.T) {
		// Given: 独立数据库实例
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()

		// Given: 用户有 1000 USDT
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 1000.0, 0)

		service := NewBalanceService(db, cfg, logger)

		// When: 10 个并发请求，每次冻结 200
		var wg sync.WaitGroup
		successChan := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// 添加小延迟避免 SQLite 锁竞争
				time.Sleep(time.Millisecond * 5)
				err := service.FreezeBalance(user.ID, "USDT", 200.0)
				successChan <- (err == nil)
			}()
		}

		wg.Wait()
		close(successChan)

		// Then: 应该有成功和失败的请求（至少有 1 个成功，至多 5 个成功）
		successCount := 0
		for success := range successChan {
			if success {
				successCount++
			}
		}

		// SQLite 并发限制下，成功数量应该 >= 1 且 <= 5
		assert.GreaterOrEqual(t, successCount, 1, "至少应该有 1 个请求成功")
		assert.LessOrEqual(t, successCount, 5, "最多应该有 5 个请求成功 (1000 / 200)")

		// 验证最终余额状态：可用余额 + 冻结余额 = 1000
		balance, err := service.GetBalance(user.ID, "USDT")
		require.NoError(t, err)
		total := balance.Available + balance.Locked
		assert.Equal(t, 1000.0, total, "总余额应该保持 1000")

		// 验证冻结金额是成功操作的倍数
		assert.Equal(t, float64(successCount)*200.0, balance.Locked)
	})

	t.Run("Concurrent freeze and unfreeze operations", func(t *testing.T) {
		// Given: 独立数据库实例
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()

		// Given: 用户有 1000 USDT
		user := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user.ID, "USDT", 500.0, 500.0)

		service := NewBalanceService(db, cfg, logger)

		// When: 并发执行冻结和解冻
		var wg sync.WaitGroup
		operations := 100

		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				if index%2 == 0 {
					// 冻结 10
					service.FreezeBalance(user.ID, "USDT", 10.0)
				} else {
					// 解冻 10
					service.UnfreezeBalance(user.ID, "USDT", 10.0)
				}
			}(i)
		}

		wg.Wait()

		// Then: 总余额应该不变
		balance, err := service.GetBalance(user.ID, "USDT")
		require.NoError(t, err)
		total := balance.Available + balance.Locked
		assert.Equal(t, 1000.0, total, "总余额应该保持 1000")
	})

	t.Run("Concurrent transfer operations", func(t *testing.T) {
		// Given: 独立数据库实例
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()

		// Given: 两个用户各有 500 USDT
		user1 := testutil.SeedUser(t, db)
		user2 := testutil.SeedUser(t, db)
		testutil.SeedBalance(t, db, user1.ID, "USDT", 500.0, 0)
		testutil.SeedBalance(t, db, user2.ID, "USDT", 500.0, 0)

		service := NewBalanceService(db, cfg, logger)

		// When: 并发双向转账 (user1 -> user2 和 user2 -> user1)
		var wg sync.WaitGroup
		transfers := 50

		for i := 0; i < transfers; i++ {
			wg.Add(2)

			// user1 -> user2
			go func() {
				defer wg.Done()
				service.TransferBalance(user1.ID, user2.ID, "USDT", 5.0)
			}()

			// user2 -> user1
			go func() {
				defer wg.Done()
				service.TransferBalance(user2.ID, user1.ID, "USDT", 5.0)
			}()
		}

		wg.Wait()

		// Then: 两个用户的总余额应该保持 1000
		balance1, err := service.GetBalance(user1.ID, "USDT")
		require.NoError(t, err)
		balance2, err := service.GetBalance(user2.ID, "USDT")
		require.NoError(t, err)

		totalBalance := balance1.Available + balance2.Available
		assert.InDelta(t, 1000.0, totalBalance, 0.01, "总余额应该保持不变")
	})
}

// TestDeadlockPrevention 测试死锁预防
func TestDeadlockPrevention(t *testing.T) {
	t.Skip("Skipping due to SQLite memory database concurrency limitations")

	t.Run("No deadlock in circular operations", func(t *testing.T) {
		// Given: 独立数据库实例
		db := testutil.SetupTestDB(t)
		cfg := testutil.LoadTestConfig(t)
		logger := testutil.NewTestLogger()

		// Given: 三个用户各有 1000 USDT
		users := make([]*model.User, 3)
		for i := 0; i < 3; i++ {
			users[i] = testutil.SeedUser(t, db)
			testutil.SeedBalance(t, db, users[i].ID, "USDT", 1000.0, 0)
		}

		service := NewBalanceService(db, cfg, logger)

		// When: 循环转账 (user1 -> user2 -> user3 -> user1)
		var wg sync.WaitGroup
		iterations := 20

		for i := 0; i < iterations; i++ {
			wg.Add(3)

			go func() {
				defer wg.Done()
				service.TransferBalance(users[0].ID, users[1].ID, "USDT", 10.0)
			}()

			go func() {
				defer wg.Done()
				service.TransferBalance(users[1].ID, users[2].ID, "USDT", 10.0)
			}()

			go func() {
				defer wg.Done()
				service.TransferBalance(users[2].ID, users[0].ID, "USDT", 10.0)
			}()
		}

		// 使用超时机制检测死锁
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 成功完成，没有死锁
			assert.True(t, true, "操作成功完成，无死锁")
		case <-time.After(10 * time.Second):
			t.Fatal("❌ 检测到死锁：操作超时")
		}

		// Then: 总余额应该保持 3000
		total := 0.0
		for _, user := range users {
			balance, err := service.GetBalance(user.ID, "USDT")
			require.NoError(t, err)
			total += balance.Available
		}
		assert.InDelta(t, 3000.0, total, 0.01)
	})
}
