package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// jsonReader 用于读取 JSON 数据(替代 bytes.Reader)
type jsonReader struct {
	data   []byte
	offset int
}

func (r *jsonReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// KlineService K线数据服务
type KlineService struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *zap.Logger
	client *http.Client

	ensureIndexesOnce sync.Once
}

// NewKlineService 创建K线服务
func NewKlineService(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *KlineService {
	return &KlineService{
		db:     db,
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetKlines 获取K线数据
// symbol: 交易对 (如 "BTC/USDT")
// interval: 时间周期 (1m, 5m, 15m, 1h, 4h, 1d)
// limit: 返回数量 (默认100, 最大1000)
// since: 开始时间 (Unix毫秒时间戳, 可选)
func (s *KlineService) GetKlines(symbol, interval string, limit int, since *time.Time) ([]model.Kline, error) {
	query := s.db.Where("symbol = ? AND interval = ?", symbol, interval)

	if since != nil {
		query = query.Where("open_time >= ?", since)
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	var klines []model.Kline
	err := query.Order("open_time ASC").Limit(limit).Find(&klines).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query klines: %w", err)
	}

	return klines, nil
}

// hyperliquidCandle 适配 Hyperliquid candleSnapshot 响应
type hyperliquidCandle struct {
	OpenTime  int64  `json:"t"`
	CloseTime int64  `json:"T"`
	Symbol    string `json:"s"`
	Interval  string `json:"i"`
	Open      string `json:"o"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Volume    string `json:"v"`
}

// updateHyperliquidKlines 从 Hyperliquid API 更新 K 线数据
func (s *KlineService) updateHyperliquidKlines() error {
	s.ensureKlineIndexes()

	// 使用配置中的 symbols
	intervals := []string{"1m", "5m", "15m", "1h", "4h", "1d"}

	for _, symbol := range s.cfg.Market.Symbols {
		for _, interval := range intervals {
			coin := s.convertSymbolToCoin(symbol)

			// Hyperliquid 需要毫秒时间戳,获取最近24小时的数据
			startTime := time.Now().Add(-24 * time.Hour).UnixMilli()

			requestBody := map[string]interface{}{
				"type": "candleSnapshot",
				"req": map[string]interface{}{
					"coin":      coin,
					"interval":  s.convertIntervalToHyperliquid(interval),
					"startTime": startTime,
				},
			}

			jsonData, err := json.Marshal(requestBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			url := s.cfg.Market.APIURL + s.cfg.Market.Hyperliquid.InfoEndpoint
			req, err := http.NewRequest("POST", url, &jsonReader{data: jsonData})
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to fetch klines from %s: %w", url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
			}

			// Hyperliquid 返回对象数组: [{t, T, s, i, o, c, h, l, v, n}, ...]
			var klineData []hyperliquidCandle
			if err := json.NewDecoder(resp.Body).Decode(&klineData); err != nil {
				return fmt.Errorf("failed to decode klines response: %w", err)
			}

			// 解析并保存每条 K 线数据
			for _, kline := range klineData {
				open, err := strconv.ParseFloat(kline.Open, 64)
				if err != nil {
					s.logger.Warn("Skipping kline due to invalid open price",
						zap.String("symbol", symbol),
						zap.String("interval", interval),
						zap.String("value", kline.Open),
						zap.Error(err),
					)
					continue
				}

				high, err := strconv.ParseFloat(kline.High, 64)
				if err != nil {
					s.logger.Warn("Skipping kline due to invalid high price",
						zap.String("symbol", symbol),
						zap.String("interval", interval),
						zap.String("value", kline.High),
						zap.Error(err),
					)
					continue
				}

				low, err := strconv.ParseFloat(kline.Low, 64)
				if err != nil {
					s.logger.Warn("Skipping kline due to invalid low price",
						zap.String("symbol", symbol),
						zap.String("interval", interval),
						zap.String("value", kline.Low),
						zap.Error(err),
					)
					continue
				}

				closePrice, err := strconv.ParseFloat(kline.Close, 64)
				if err != nil {
					s.logger.Warn("Skipping kline due to invalid close price",
						zap.String("symbol", symbol),
						zap.String("interval", interval),
						zap.String("value", kline.Close),
						zap.Error(err),
					)
					continue
				}

				volume, err := strconv.ParseFloat(kline.Volume, 64)
				if err != nil {
					s.logger.Warn("Skipping kline due to invalid volume",
						zap.String("symbol", symbol),
						zap.String("interval", interval),
						zap.String("value", kline.Volume),
						zap.Error(err),
					)
					continue
				}

				openTime := time.UnixMilli(kline.OpenTime)
				closeTime := s.calculateCloseTime(openTime, interval)
				if kline.CloseTime > 0 {
					closeTime = time.UnixMilli(kline.CloseTime)
				}

				klineModel := model.Kline{
					Symbol:    symbol,
					Interval:  interval,
					OpenTime:  openTime,
					CloseTime: closeTime,
					Open:      open,
					High:      high,
					Low:       low,
					Close:     closePrice,
					Volume:    volume,
				}

				// UPSERT: 如果存在则更新,否则插入
				if err := s.db.
					Clauses(clause.OnConflict{
						Columns: []clause.Column{
							{Name: "symbol"},
							{Name: "interval"},
							{Name: "open_time"},
						},
						DoUpdates: clause.AssignmentColumns([]string{
							"close_time",
							"open",
							"high",
							"low",
							"close",
							"volume",
							"updated_at",
						}),
					}).
					Create(&klineModel).Error; err != nil {
					s.logger.Error("Failed to save kline",
						zap.String("symbol", symbol),
						zap.String("interval", interval),
						zap.Error(err),
					)
				}
			}
		}
	}

	return nil
}

// StartAutoUpdate 启动自动更新（定时任务）
func (s *KlineService) StartAutoUpdate() {
	intervals := []string{"1m", "5m", "15m", "1h", "4h", "1d"}

	for _, interval := range intervals {
		go s.updateLoop(interval)
	}
}

func (s *KlineService) updateLoop(interval string) {
	ticker := time.NewTicker(s.getUpdateInterval(interval))
	defer ticker.Stop()

	// 立即执行一次
	if err := s.updateHyperliquidKlines(); err != nil {
		s.logger.Error("Failed to update klines", zap.String("interval", interval), zap.Error(err))
	}

	// 定时更新
	for range ticker.C {
		if err := s.updateHyperliquidKlines(); err != nil {
			s.logger.Error("Failed to update klines", zap.String("interval", interval), zap.Error(err))
		}
	}
}

func (s *KlineService) getUpdateInterval(interval string) time.Duration {
	switch interval {
	case "1m":
		return 1 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 1 * time.Hour
	}
}

// convertSymbolToCoin 将交易对转换为币种 (BTC/USDT -> BTC)
func (s *KlineService) convertSymbolToCoin(symbol string) string {
	parts := strings.Split(symbol, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return symbol
}

// convertIntervalToHyperliquid 将标准周期转换为 Hyperliquid 格式
func (s *KlineService) convertIntervalToHyperliquid(interval string) string {
	// Hyperliquid 使用相同格式: 1m, 5m, 15m, 1h, 4h, 1d
	// 对于未知间隔，默认返回 1h
	switch interval {
	case "1m", "5m", "15m", "1h", "4h", "1d":
		return interval
	default:
		return "1h"
	}
}

// calculateCloseTime 计算K线关闭时间
func (s *KlineService) calculateCloseTime(openTime time.Time, interval string) time.Time {
	return openTime.Add(s.getUpdateInterval(interval))
}

// ensureKlineIndexes 确保存储 K 线所需索引
func (s *KlineService) ensureKlineIndexes() {
	s.ensureIndexesOnce.Do(func() {
		if err := s.db.AutoMigrate(&model.Kline{}); err != nil {
			s.logger.Error("Failed to auto-migrate kline model", zap.Error(err))
			return
		}

		if err := s.removeDuplicateKlines(); err != nil {
			s.logger.Error("Failed to remove duplicate klines", zap.Error(err))
		}

		const idxSQL = `CREATE UNIQUE INDEX IF NOT EXISTS idx_symbol_interval_time ON klines (symbol, interval, open_time)`
		if err := s.db.Exec(idxSQL).Error; err != nil {
			s.logger.Error("Failed to ensure kline unique index", zap.Error(err))
		}
	})
}

func (s *KlineService) removeDuplicateKlines() error {
	dialect := strings.ToLower(s.db.Dialector.Name())

	switch dialect {
	case "postgres":
		const query = `
DELETE FROM klines a
USING klines b
WHERE a.id < b.id
  AND a.symbol = b.symbol
  AND a.interval = b.interval
  AND a.open_time = b.open_time`
		return s.db.Exec(query).Error
	case "sqlite":
		const query = `
DELETE FROM klines
WHERE id NOT IN (
	SELECT MAX(id)
	FROM klines
	GROUP BY symbol, interval, open_time
)`
		return s.db.Exec(query).Error
	default:
		s.logger.Warn("Skipping duplicate kline cleanup for unsupported dialect", zap.String("dialect", dialect))
		return nil
	}
}
