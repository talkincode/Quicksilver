package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// MarketService 市场数据服务
type MarketService struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *zap.Logger
	client *http.Client
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

	switch s.cfg.Market.DataSource {
	case "hyperliquid":
		return s.updateHyperliquidTickers()
	case "binance":
		return s.updateBinanceTickers()
	default:
		return fmt.Errorf("unsupported data source: %s", s.cfg.Market.DataSource)
	}
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

			ticker := model.Ticker{
				Symbol:    symbol,
				LastPrice: price,
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
