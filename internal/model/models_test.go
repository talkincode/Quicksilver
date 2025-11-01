package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&User{}, &Balance{}, &Order{}, &Trade{}, &Ticker{})
	require.NoError(t, err)

	return db
}

func TestUserModel(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Create user", func(t *testing.T) {
		user := &User{
			Email:     "test@example.com",
			Username:  "testuser",
			APIKey:    "api-key-123",
			APISecret: "api-secret-456",
			Status:    "active",
		}

		err := db.Create(user).Error
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("User email uniqueness", func(t *testing.T) {
		user1 := &User{
			Email:     "unique@example.com",
			APIKey:    "key1",
			APISecret: "secret1",
		}
		err := db.Create(user1).Error
		require.NoError(t, err)

		user2 := &User{
			Email:     "unique@example.com",
			APIKey:    "key2",
			APISecret: "secret2",
		}
		err = db.Create(user2).Error
		assert.Error(t, err, "should not allow duplicate email")
	})

	t.Run("User API key uniqueness", func(t *testing.T) {
		user1 := &User{
			Email:     "user1@example.com",
			APIKey:    "duplicate-key",
			APISecret: "secret1",
		}
		err := db.Create(user1).Error
		require.NoError(t, err)

		user2 := &User{
			Email:     "user2@example.com",
			APIKey:    "duplicate-key",
			APISecret: "secret2",
		}
		err = db.Create(user2).Error
		assert.Error(t, err, "should not allow duplicate API key")
	})
}

func TestBalanceModel(t *testing.T) {
	db := setupTestDB(t)

	// 创建测试用户
	user := &User{
		Email:     "balance@example.com",
		APIKey:    "balance-key",
		APISecret: "balance-secret",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("Create balance", func(t *testing.T) {
		balance := &Balance{
			UserID:    user.ID,
			Asset:     "BTC",
			Available: 1.5,
			Locked:    0.5,
		}

		err := db.Create(balance).Error
		require.NoError(t, err)
		assert.NotZero(t, balance.ID)
		assert.Equal(t, user.ID, balance.UserID)
		assert.Equal(t, 1.5, balance.Available)
		assert.Equal(t, 0.5, balance.Locked)
	})

	t.Run("Multiple assets for same user", func(t *testing.T) {
		btcBalance := &Balance{
			UserID:    user.ID,
			Asset:     "BTC",
			Available: 1.0,
			Locked:    0.0,
		}
		err := db.Create(btcBalance).Error
		require.NoError(t, err)

		ethBalance := &Balance{
			UserID:    user.ID,
			Asset:     "ETH",
			Available: 10.0,
			Locked:    0.0,
		}
		err = db.Create(ethBalance).Error
		require.NoError(t, err)

		var balances []Balance
		db.Where("user_id = ?", user.ID).Find(&balances)
		assert.GreaterOrEqual(t, len(balances), 2)
	})
}

func TestOrderModel(t *testing.T) {
	db := setupTestDB(t)

	user := &User{
		Email:     "order@example.com",
		APIKey:    "order-key",
		APISecret: "order-secret",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("Create market order", func(t *testing.T) {
		order := &Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "market",
			Status: "new",
			Amount: 0.5,
		}

		err := db.Create(order).Error
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.Equal(t, "new", order.Status)
		assert.Nil(t, order.Price)
	})

	t.Run("Create limit order", func(t *testing.T) {
		price := 50000.0
		order := &Order{
			UserID: user.ID,
			Symbol: "BTC/USDT",
			Side:   "buy",
			Type:   "limit",
			Status: "new",
			Amount: 0.5,
			Price:  &price,
		}

		err := db.Create(order).Error
		require.NoError(t, err)
		assert.NotZero(t, order.ID)
		assert.NotNil(t, order.Price)
		assert.Equal(t, 50000.0, *order.Price)
	})

	t.Run("Order with client order ID", func(t *testing.T) {
		order := &Order{
			UserID:        user.ID,
			Symbol:        "BTC/USDT",
			Side:          "sell",
			Type:          "market",
			Status:        "new",
			Amount:        0.1,
			ClientOrderID: "client-order-123",
		}

		err := db.Create(order).Error
		require.NoError(t, err)

		var found Order
		err = db.Where("client_order_id = ?", "client-order-123").First(&found).Error
		require.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
	})
}

func TestTradeModel(t *testing.T) {
	db := setupTestDB(t)

	user := &User{
		Email:     "trade@example.com",
		APIKey:    "trade-key",
		APISecret: "trade-secret",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	order := &Order{
		UserID: user.ID,
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "market",
		Status: "new",
		Amount: 1.0,
	}
	err = db.Create(order).Error
	require.NoError(t, err)

	t.Run("Create trade", func(t *testing.T) {
		trade := &Trade{
			OrderID:     order.ID,
			UserID:      user.ID,
			Symbol:      "BTC/USDT",
			Side:        "buy",
			Price:       50000.0,
			Amount:      1.0,
			QuoteAmount: 50000.0,
			Fee:         50.0,
			FeeAsset:    "USDT",
			IsMaker:     false,
		}

		err := db.Create(trade).Error
		require.NoError(t, err)
		assert.NotZero(t, trade.ID)
		assert.Equal(t, order.ID, trade.OrderID)
		assert.Equal(t, 50000.0, trade.Price)
	})

	t.Run("Query trades by order", func(t *testing.T) {
		// 创建多个成交记录
		for i := 0; i < 3; i++ {
			trade := &Trade{
				OrderID:     order.ID,
				UserID:      user.ID,
				Symbol:      "BTC/USDT",
				Side:        "buy",
				Price:       50000.0,
				Amount:      0.1,
				QuoteAmount: 5000.0,
				Fee:         5.0,
				FeeAsset:    "USDT",
			}
			db.Create(trade)
		}

		var trades []Trade
		db.Where("order_id = ?", order.ID).Find(&trades)
		assert.GreaterOrEqual(t, len(trades), 3)
	})
}

func TestTickerModel(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Create ticker", func(t *testing.T) {
		ticker := &Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50000.0,
			Source:    "hyperliquid",
		}

		err := db.Save(ticker).Error
		require.NoError(t, err)
		assert.Equal(t, "BTC/USDT", ticker.Symbol)
		assert.Equal(t, 50000.0, ticker.LastPrice)
	})

	t.Run("Update ticker (UPSERT)", func(t *testing.T) {
		// 第一次插入
		ticker1 := &Ticker{
			Symbol:    "ETH/USDT",
			LastPrice: 3000.0,
			Source:    "test",
		}
		db.Save(ticker1)

		// 更新同一交易对
		ticker2 := &Ticker{
			Symbol:    "ETH/USDT",
			LastPrice: 3100.0,
			Source:    "test",
		}
		db.Save(ticker2)

		// 验证只有一条记录
		var found Ticker
		err := db.Where("symbol = ?", "ETH/USDT").First(&found).Error
		require.NoError(t, err)
		assert.Equal(t, 3100.0, found.LastPrice)
	})

	t.Run("Ticker with additional fields", func(t *testing.T) {
		bid := 49900.0
		ask := 50100.0
		high := 51000.0
		low := 49000.0
		volume := 1000.0

		ticker := &Ticker{
			Symbol:        "BTC/USDT",
			LastPrice:     50000.0,
			BidPrice:      &bid,
			AskPrice:      &ask,
			High24h:       &high,
			Low24h:        &low,
			Volume24hBase: &volume,
			Source:        "test",
		}

		err := db.Save(ticker).Error
		require.NoError(t, err)

		var found Ticker
		db.Where("symbol = ?", "BTC/USDT").First(&found)
		assert.NotNil(t, found.BidPrice)
		assert.Equal(t, 49900.0, *found.BidPrice)
		assert.NotNil(t, found.High24h)
		assert.Equal(t, 51000.0, *found.High24h)
	})
}

func TestTableNames(t *testing.T) {
	t.Run("User table name", func(t *testing.T) {
		var user User
		assert.Equal(t, "users", user.TableName())
	})

	t.Run("Balance table name", func(t *testing.T) {
		var balance Balance
		assert.Equal(t, "balances", balance.TableName())
	})

	t.Run("Order table name", func(t *testing.T) {
		var order Order
		assert.Equal(t, "orders", order.TableName())
	})

	t.Run("Trade table name", func(t *testing.T) {
		var trade Trade
		assert.Equal(t, "trades", trade.TableName())
	})

	t.Run("Ticker table name", func(t *testing.T) {
		var ticker Ticker
		assert.Equal(t, "tickers", ticker.TableName())
	})
}

func TestTimestamps(t *testing.T) {
	db := setupTestDB(t)

	t.Run("User timestamps", func(t *testing.T) {
		user := &User{
			Email:     "timestamp@example.com",
			APIKey:    "timestamp-key",
			APISecret: "timestamp-secret",
		}

		err := db.Create(user).Error
		require.NoError(t, err)

		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())

		// 更新用户
		time.Sleep(time.Millisecond * 10)
		user.Username = "updated"
		db.Save(user)

		assert.True(t, user.UpdatedAt.After(user.CreatedAt))
	})

	t.Run("Ticker updated_at", func(t *testing.T) {
		ticker := &Ticker{
			Symbol:    "BTC/USDT",
			LastPrice: 50000.0,
			Source:    "test",
		}

		db.Save(ticker)
		firstUpdate := ticker.UpdatedAt

		time.Sleep(time.Millisecond * 10)
		ticker.LastPrice = 51000.0
		db.Save(ticker)

		assert.True(t, ticker.UpdatedAt.After(firstUpdate))
	})
}
