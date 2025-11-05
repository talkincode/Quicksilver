package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/engine"
	"github.com/talkincode/quicksilver/internal/model"
)

// MarketService 市场数据服务
type MarketService struct {
	db                *gorm.DB
	cfg               *config.Config
	logger            *zap.Logger
	client            *http.Client
	matchingSemaphore *semaphore.Weighted // 并发控制信号量
}

// NewMarketService 创建市场数据服务
func NewMarketService(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MarketService {
	return &MarketService{
		db:     db,
		cfg:    cfg,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		matchingSemaphore: semaphore.NewWeighted(10), // 最多 10 个并发撮合
	}
}

// HyperliquidTicker Hyperliquid API 响应结构
type HyperliquidTicker struct {
	Coin        string `json:"coin"`
	LastPrice   string `json:"lastPx"`
	BidPrice    string `json:"bidPx"`
	AskPrice    string `json:"askPx"`
	High24h     string `json:"high24h"`
	Low24h      string `json:"low24h"`
	Volume24h   string `json:"volume24h"`
	PriceChange string `json:"priceChange24h"`
}

// HyperliquidMetaResponse Hyperliquid meta 响应
type HyperliquidMetaResponse struct {
	Universe []struct {
		Name       string `json:"name"`
		SzDecimals int    `json:"szDecimals"`
	} `json:"universe"`
}

// HyperliquidAllMidsResponse Hyperliquid allMids 响应
// HyperliquidAllMidsResponse Hyperliquid allMids API 响应
// 注意: allMids 返回的是扁平的键值对对象,不是 {"mids": {...}} 的嵌套结构
type HyperliquidAllMidsResponse map[string]string

// UpdateTickers 更新行情数据
func (s *MarketService) UpdateTickers() error {
	s.logger.Debug("UpdateTickers called", zap.String("source", s.cfg.Market.DataSource))

	var err error
	switch s.cfg.Market.DataSource {
	case "hyperliquid":
		err = s.updateHyperliquidTickers()
	case "binance":
		err = s.updateBinanceTickers()
	default:
		return fmt.Errorf("unsupported data source: %s", s.cfg.Market.DataSource)
	}

	if err != nil {
		return err
	}

	// 行情更新后，异步触发未成交限价单的撮合
	go func() {
		if matchErr := s.TriggerPendingOrdersMatching(); matchErr != nil {
			s.logger.Error("Failed to trigger pending orders matching", zap.Error(matchErr))
		}
	}()

	// 同时触发止盈止损订单检查
	go func() {
		if stopErr := s.TriggerStopOrders(); stopErr != nil {
			s.logger.Error("Failed to trigger stop orders", zap.Error(stopErr))
		}
	}()

	return nil
}

// updateHyperliquidTickers 从 Hyperliquid 更新行情
func (s *MarketService) updateHyperliquidTickers() error {
	s.logger.Debug("updateHyperliquidTickers called")

	// Hyperliquid API 请求体
	requestBody := map[string]interface{}{
		"type": "allMids",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	url := s.cfg.Market.APIURL + s.cfg.Market.Hyperliquid.InfoEndpoint
	s.logger.Debug("Requesting Hyperliquid API", zap.String("url", url))

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch from Hyperliquid", zap.Error(err))
		return fmt.Errorf("failed to fetch tickers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Unexpected status code from Hyperliquid", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 读取原始响应体用于调试
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body", zap.Error(err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 仅在 Debug 级别记录原始响应
	s.logger.Debug("Raw API response", zap.Int("body_length", len(bodyBytes)))

	// 解析响应
	var midsResp HyperliquidAllMidsResponse
	if err := json.Unmarshal(bodyBytes, &midsResp); err != nil {
		s.logger.Error("Failed to decode Hyperliquid response",
			zap.Error(err),
			zap.String("raw_body", string(bodyBytes)),
		)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	s.logger.Debug("Received Hyperliquid data",
		zap.Int("mids_count", len(midsResp)),
	)

	// 更新数据库
	updatedCount := 0
	for _, symbol := range s.cfg.Market.Symbols {
		// 转换交易对格式: BTC/USDT -> BTC
		coin := convertSymbolToCoin(symbol)
		s.logger.Debug("Processing symbol", zap.String("symbol", symbol), zap.String("coin", coin))

		if priceStr, ok := midsResp[coin]; ok {
			var price float64
			if _, err := fmt.Sscanf(priceStr, "%f", &price); err != nil {
				s.logger.Error("Failed to parse price",
					zap.String("coin", coin),
					zap.String("price", priceStr),
					zap.Error(err),
				)
				continue
			}

			// 使用 mid price 计算 bid/ask (简化模拟，实际应该从 API 获取)
			// 假设 0.05% 的买卖价差
			bidPrice := price * 0.9995
			askPrice := price * 1.0005

			ticker := model.Ticker{
				Symbol:    symbol,
				LastPrice: price,
				BidPrice:  &bidPrice,
				AskPrice:  &askPrice,
				UpdatedAt: time.Now(),
				Source:    "hyperliquid",
			}

			// UPSERT 操作
			if err := s.db.Save(&ticker).Error; err != nil {
				s.logger.Error("Failed to save ticker",
					zap.String("symbol", symbol),
					zap.Error(err),
				)
				continue
			}

			updatedCount++
			s.logger.Debug("Ticker updated",
				zap.String("symbol", symbol),
				zap.Float64("price", price),
			)
		}
	}

	// 仅在 Info 级别输出汇总信息
	if updatedCount > 0 {
		s.logger.Info("Tickers updated successfully",
			zap.Int("count", updatedCount),
			zap.String("source", "hyperliquid"),
		)
	}

	return nil
}

// updateBinanceTickers 从 Binance 更新行情 (保留原有逻辑)
func (s *MarketService) updateBinanceTickers() error {
	// TODO: 实现 Binance API 调用
	s.logger.Warn("Binance ticker update not implemented yet")
	return nil
}

// StartAutoUpdate 启动自动更新
func (s *MarketService) StartAutoUpdate() {
	interval, err := time.ParseDuration(s.cfg.Market.UpdateInterval)
	if err != nil {
		s.logger.Error("Invalid update interval", zap.Error(err))
		interval = 1 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		// 立即执行一次
		if err := s.UpdateTickers(); err != nil {
			s.logger.Error("Failed to update tickers", zap.Error(err))
		}

		for range ticker.C {
			if err := s.UpdateTickers(); err != nil {
				s.logger.Error("Failed to update tickers", zap.Error(err))
			}
		}
	}()

	s.logger.Info("Market data auto-update started",
		zap.String("source", s.cfg.Market.DataSource),
		zap.Duration("interval", interval),
	)
}

// convertSymbolToCoin 转换交易对格式
// BTC/USDT -> BTC
// ETH/USDT -> ETH
func convertSymbolToCoin(symbol string) string {
	parts := strings.Split(symbol, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return symbol
}

// TriggerPendingOrdersMatching 触发未成交限价单的撮合
func (s *MarketService) TriggerPendingOrdersMatching() error {
	// 查询所有未成交的限价单（添加索引优化）
	var pendingOrders []model.Order
	err := s.db.Where("status = ? AND type = ?", "new", "limit").
		Order("created_at ASC"). // 按创建时间排序，先进先出
		Find(&pendingOrders).Error
	if err != nil {
		s.logger.Error("Failed to query pending limit orders", zap.Error(err))
		return fmt.Errorf("failed to query pending orders: %w", err)
	}

	if len(pendingOrders) == 0 {
		s.logger.Debug("No pending limit orders to match")
		return nil
	}

	s.logger.Debug("Found pending limit orders", zap.Int("count", len(pendingOrders)))

	// 为每个订单触发撮合（使用信号量控制并发）
	for _, order := range pendingOrders {
		// 异步触发，但使用信号量限制并发数
		go s.matchPendingOrderWithLimit(order.ID)
	}

	return nil
}

// matchPendingOrderWithLimit 带并发限制的异步撮合
func (s *MarketService) matchPendingOrderWithLimit(orderID uint) {
	// 获取信号量（最多 10 个并发）
	ctx := context.Background()
	if err := s.matchingSemaphore.Acquire(ctx, 1); err != nil {
		s.logger.Error("Failed to acquire semaphore",
			zap.Uint("order_id", orderID),
			zap.Error(err),
		)
		return
	}
	defer s.matchingSemaphore.Release(1)

	// 执行撮合
	s.matchPendingOrder(orderID)
}

// matchPendingOrder 异步撮合单个订单
func (s *MarketService) matchPendingOrder(orderID uint) {
	// 导入撮合引擎
	matchEngine := s.createMatchingEngine()

	if err := matchEngine.MatchOrder(orderID); err != nil {
		s.logger.Error("Failed to match pending order",
			zap.Uint("order_id", orderID),
			zap.Error(err),
		)
	}
} // createMatchingEngine 创建撮合引擎实例
func (s *MarketService) createMatchingEngine() *engine.MatchingEngine {
	return engine.NewMatchingEngine(s.db, s.cfg, s.logger)
}

// TriggerStopOrders 触发止盈止损订单
func (s *MarketService) TriggerStopOrders() error {
	// 查询所有未触发的止盈止损单
	var stopOrders []model.Order
	err := s.db.Where("status = ? AND type IN (?)", "new", []string{"stop_loss", "take_profit"}).
		Order("created_at ASC").
		Find(&stopOrders).Error
	if err != nil {
		s.logger.Error("Failed to query stop orders", zap.Error(err))
		return fmt.Errorf("failed to query stop orders: %w", err)
	}

	if len(stopOrders) == 0 {
		s.logger.Debug("No stop orders to trigger")
		return nil
	}

	s.logger.Debug("Found stop orders", zap.Int("count", len(stopOrders)))

	// 为每个订单检查触发条件
	for _, order := range stopOrders {
		go s.checkAndTriggerStopOrder(order.ID)
	}

	return nil
}

// checkAndTriggerStopOrder 检查并触发单个止盈止损订单
func (s *MarketService) checkAndTriggerStopOrder(orderID uint) {
	// 获取订单
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		s.logger.Error("Failed to get stop order", zap.Uint("order_id", orderID), zap.Error(err))
		return
	}

	// 获取当前市场价格
	var ticker model.Ticker
	if err := s.db.Where("symbol = ?", order.Symbol).First(&ticker).Error; err != nil {
		s.logger.Error("Failed to get ticker",
			zap.String("symbol", order.Symbol),
			zap.Error(err))
		return
	}

	// 检查触发条件
	currentPrice := ticker.LastPrice
	if order.StopPrice == nil {
		s.logger.Error("Stop price is null", zap.Uint("order_id", orderID))
		return
	}

	triggered := false
	switch order.TriggerCondition {
	case ">=":
		triggered = currentPrice >= *order.StopPrice
	case "<=":
		triggered = currentPrice <= *order.StopPrice
	default:
		s.logger.Error("Invalid trigger condition",
			zap.Uint("order_id", orderID),
			zap.String("condition", order.TriggerCondition))
		return
	}

	if !triggered {
		s.logger.Debug("Stop order condition not met",
			zap.Uint("order_id", orderID),
			zap.Float64("current_price", currentPrice),
			zap.Float64("stop_price", *order.StopPrice),
			zap.String("condition", order.TriggerCondition))
		return
	}

	// 触发条件满足，创建市价单
	s.logger.Info("Stop order triggered",
		zap.Uint("order_id", orderID),
		zap.String("type", order.Type),
		zap.Float64("current_price", currentPrice),
		zap.Float64("stop_price", *order.StopPrice))

	// 使用事务确保原子性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新止盈止损单状态为 triggered
		now := time.Now()
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":       "triggered",
			"triggered_at": now,
			"updated_at":   now,
		}).Error; err != nil {
			return fmt.Errorf("failed to update stop order status: %w", err)
		}

		// 2. 创建市价单（继承止盈止损单的参数）
		marketOrder := &model.Order{
			UserID:        order.UserID,
			Symbol:        order.Symbol,
			Side:          order.Side,
			Type:          "market",
			Status:        "new",
			Amount:        order.Amount,
			ParentOrderID: &order.ID, // 关联父订单
		}

		if err := tx.Create(marketOrder).Error; err != nil {
			return fmt.Errorf("failed to create market order: %w", err)
		}

		s.logger.Info("Market order created from stop order",
			zap.Uint("parent_order_id", order.ID),
			zap.Uint("market_order_id", marketOrder.ID))

		// 3. 触发撮合引擎（在事务外异步执行）
		go func() {
			matchEngine := s.createMatchingEngine()
			if err := matchEngine.MatchOrder(marketOrder.ID); err != nil {
				s.logger.Error("Failed to match market order from stop order",
					zap.Uint("order_id", marketOrder.ID),
					zap.Error(err))
			}
		}()

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to trigger stop order",
			zap.Uint("order_id", orderID),
			zap.Error(err))
	}
}
