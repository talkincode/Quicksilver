package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/quicksilver/internal/model"
	"github.com/talkincode/quicksilver/internal/testutil"
)

func TestNewKlineService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	service := NewKlineService(db, cfg, logger)
	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.cfg)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.client)
}

func TestGetKlines(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	// 自动迁移 Kline 表
	require.NoError(t, db.AutoMigrate(&model.Kline{}))

	service := NewKlineService(db, cfg, logger)

	t.Run("Get klines successfully", func(t *testing.T) {
		// 准备测试数据
		now := time.Now().Truncate(time.Minute)
		klines := []model.Kline{
			{
				Symbol:    "BTC/USDT",
				Interval:  "1h",
				OpenTime:  now.Add(-3 * time.Hour),
				CloseTime: now.Add(-2 * time.Hour),
				Open:      50000,
				High:      51000,
				Low:       49500,
				Close:     50500,
				Volume:    100,
			},
			{
				Symbol:    "BTC/USDT",
				Interval:  "1h",
				OpenTime:  now.Add(-2 * time.Hour),
				CloseTime: now.Add(-1 * time.Hour),
				Open:      50500,
				High:      52000,
				Low:       50000,
				Close:     51500,
				Volume:    150,
			},
			{
				Symbol:    "BTC/USDT",
				Interval:  "1h",
				OpenTime:  now.Add(-1 * time.Hour),
				CloseTime: now,
				Open:      51500,
				High:      53000,
				Low:       51000,
				Close:     52500,
				Volume:    200,
			},
		}

		for _, kline := range klines {
			require.NoError(t, db.Create(&kline).Error)
		}

		// 查询K线数据
		result, err := service.GetKlines("BTC/USDT", "1h", 10, nil)
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// 验证数据按时间正序排列
		assert.Equal(t, 50000.0, result[0].Open)
		assert.Equal(t, 50500.0, result[1].Open)
		assert.Equal(t, 51500.0, result[2].Open)
	})

	t.Run("Get klines with limit", func(t *testing.T) {
		result, err := service.GetKlines("BTC/USDT", "1h", 2, nil)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})

	t.Run("Get klines with since parameter", func(t *testing.T) {
		since := time.Now().Add(-90 * time.Minute)
		result, err := service.GetKlines("BTC/USDT", "1h", 10, &since)
		require.NoError(t, err)

		// 所有返回的K线都应该在 since 之后
		for _, kline := range result {
			assert.True(t, kline.OpenTime.After(since) || kline.OpenTime.Equal(since))
		}
	})

	t.Run("Get klines for non-existent symbol", func(t *testing.T) {
		result, err := service.GetKlines("XXX/USDT", "1h", 10, nil)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestConvertIntervalToHyperliquid(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	service := NewKlineService(db, cfg, logger)

	tests := []struct {
		name     string
		interval string
		expected string
	}{
		{"1 minute", "1m", "1m"},
		{"5 minutes", "5m", "5m"},
		{"15 minutes", "15m", "15m"},
		{"1 hour", "1h", "1h"},
		{"4 hours", "4h", "4h"},
		{"1 day", "1d", "1d"},
		{"unknown", "unknown", "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.convertIntervalToHyperliquid(tt.interval)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCloseTime(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	service := NewKlineService(db, cfg, logger)

	openTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		interval      string
		expectedDelta time.Duration
	}{
		{"1 minute", "1m", 1 * time.Minute},
		{"5 minutes", "5m", 5 * time.Minute},
		{"15 minutes", "15m", 15 * time.Minute},
		{"1 hour", "1h", 1 * time.Hour},
		{"4 hours", "4h", 4 * time.Hour},
		{"1 day", "1d", 24 * time.Hour},
		{"unknown", "unknown", 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closeTime := service.calculateCloseTime(openTime, tt.interval)
			assert.Equal(t, openTime.Add(tt.expectedDelta), closeTime)
		})
	}
}

func TestGetUpdateInterval(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.LoadTestConfig(t)
	logger := testutil.NewTestLogger()

	service := NewKlineService(db, cfg, logger)

	tests := []struct {
		name     string
		interval string
		expected time.Duration
	}{
		{"1 minute", "1m", 1 * time.Minute},
		{"5 minutes", "5m", 5 * time.Minute},
		{"15 minutes", "15m", 15 * time.Minute},
		{"1 hour", "1h", 1 * time.Hour},
		{"4 hours", "4h", 4 * time.Hour},
		{"1 day", "1d", 24 * time.Hour},
		{"unknown", "unknown", 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getUpdateInterval(tt.interval)
			assert.Equal(t, tt.expected, result)
		})
	}
}
