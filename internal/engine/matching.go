package engine

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/model"
)

// MatchingEngine 撮合引擎
type MatchingEngine struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *zap.Logger
}

// NewMatchingEngine 创建撮合引擎实例
func NewMatchingEngine(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *MatchingEngine {
	return &MatchingEngine{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

// MatchOrder 撮合订单
func (m *MatchingEngine) MatchOrder(orderID uint) error {
	// 1. 查询订单
	var order model.Order
	if err := m.db.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// 2. 检查订单状态
	if order.Status != "new" {
		return fmt.Errorf("order status is not new: %s", order.Status)
	}

	// 3. 根据订单类型进行撮合
	if order.Type == "market" {
		return m.matchMarketOrder(&order)
	} else if order.Type == "limit" {
		return m.matchLimitOrder(&order)
	}

	return fmt.Errorf("unsupported order type: %s", order.Type)
}

// matchMarketOrder 撮合市价单
func (m *MatchingEngine) matchMarketOrder(order *model.Order) error {
	// 1. 获取市场价格
	var ticker model.Ticker
	if err := m.db.Where("symbol = ?", order.Symbol).First(&ticker).Error; err != nil {
		return fmt.Errorf("ticker not found for %s: %w", order.Symbol, err)
	}

	// 2. 确定成交价格
	var price float64
	if order.Side == "buy" {
		// 买单使用 ask 价格
		if ticker.AskPrice == nil {
			return fmt.Errorf("ask price not available for %s", order.Symbol)
		}
		price = *ticker.AskPrice
	} else if order.Side == "sell" {
		// 卖单使用 bid 价格
		if ticker.BidPrice == nil {
			return fmt.Errorf("bid price not available for %s", order.Symbol)
		}
		price = *ticker.BidPrice
	} else {
		return fmt.Errorf("invalid order side: %s", order.Side)
	}

	// 3. 创建成交记录
	if err := m.createTradeRecord(order, price); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	// 4. 更新订单状态
	if err := m.updateOrderStatus(order, order.Amount); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	m.logger.Info("Market order matched",
		zap.Uint("order_id", order.ID),
		zap.String("symbol", order.Symbol),
		zap.String("side", order.Side),
		zap.Float64("price", price),
		zap.Float64("amount", order.Amount),
	)

	return nil
}

// matchLimitOrder 撮合限价单
func (m *MatchingEngine) matchLimitOrder(order *model.Order) error {
	// 1. 获取市场价格
	var ticker model.Ticker
	if err := m.db.Where("symbol = ?", order.Symbol).First(&ticker).Error; err != nil {
		return fmt.Errorf("ticker not found for %s: %w", order.Symbol, err)
	}

	// 2. 检查限价单是否可以成交
	if order.Price == nil {
		return fmt.Errorf("limit order must have price")
	}

	limitPrice := *order.Price
	canMatch := false
	var matchPrice float64

	if order.Side == "buy" {
		// 买单：限价 >= ask 时可以成交
		if ticker.AskPrice != nil && limitPrice >= *ticker.AskPrice {
			canMatch = true
			matchPrice = *ticker.AskPrice // 实际成交价使用市场价
		}
	} else if order.Side == "sell" {
		// 卖单：限价 <= bid 时可以成交
		if ticker.BidPrice != nil && limitPrice <= *ticker.BidPrice {
			canMatch = true
			matchPrice = *ticker.BidPrice
		}
	}

	// 3. 如果不能成交，保持订单状态不变
	if !canMatch {
		m.logger.Debug("Limit order cannot be matched yet",
			zap.Uint("order_id", order.ID),
			zap.Float64("limit_price", limitPrice),
			zap.Float64p("ask_price", ticker.AskPrice),
			zap.Float64p("bid_price", ticker.BidPrice),
		)
		return nil // 不是错误，只是暂时无法成交
	}

	// 4. 创建成交记录
	if err := m.createTradeRecord(order, matchPrice); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	// 5. 更新订单状态
	if err := m.updateOrderStatus(order, order.Amount); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	m.logger.Info("Limit order matched",
		zap.Uint("order_id", order.ID),
		zap.String("symbol", order.Symbol),
		zap.String("side", order.Side),
		zap.Float64("limit_price", limitPrice),
		zap.Float64("match_price", matchPrice),
		zap.Float64("amount", order.Amount),
	)

	return nil
}

// createTradeRecord 创建成交记录并结算余额
func (m *MatchingEngine) createTradeRecord(order *model.Order, price float64) error {
	// 1. 计算手续费
	var feeRate float64
	if order.Type == "market" {
		feeRate = m.cfg.Trading.TakerFeeRate
	} else {
		feeRate = m.cfg.Trading.MakerFeeRate
	}

	fee := m.calculateFee(order.Amount, feeRate)

	// 2. 在事务中创建成交记录和结算余额
	return m.db.Transaction(func(tx *gorm.DB) error {
		// 创建成交记录
		trade := &model.Trade{
			OrderID:  order.ID,
			UserID:   order.UserID,
			Symbol:   order.Symbol,
			Side:     order.Side,
			Price:    price,
			Amount:   order.Amount,
			Fee:      fee,
			FeeAsset: m.getFeeAsset(order),
		}

		if err := tx.Create(trade).Error; err != nil {
			return fmt.Errorf("failed to create trade: %w", err)
		}

		// 结算余额
		if err := m.settleBalance(tx, order, trade); err != nil {
			return fmt.Errorf("failed to settle balance: %w", err)
		}

		return nil
	})
}

// settleBalance 结算余额
func (m *MatchingEngine) settleBalance(tx *gorm.DB, order *model.Order, trade *model.Trade) error {
	baseCoin, quoteCoin := m.splitSymbol(order.Symbol)

	if order.Side == "buy" {
		// 买单：扣除冻结的 USDT，增加 BTC
		cost := trade.Amount * trade.Price * (1 + m.cfg.Trading.TakerFeeRate)

		// 扣除冻结的 USDT
		var quoteBalance model.Balance
		if err := tx.Where("user_id = ? AND asset = ?", order.UserID, quoteCoin).
			First(&quoteBalance).Error; err != nil {
			return fmt.Errorf("quote balance not found: %w", err)
		}

		quoteBalance.Locked -= cost
		if quoteBalance.Locked < 0 {
			quoteBalance.Locked = 0
		}

		if err := tx.Save(&quoteBalance).Error; err != nil {
			return fmt.Errorf("failed to update quote balance: %w", err)
		}

		// 增加 BTC (扣除手续费)
		receivedAmount := trade.Amount - trade.Fee
		var baseBalance model.Balance
		err := tx.Where("user_id = ? AND asset = ?", order.UserID, baseCoin).
			First(&baseBalance).Error
		if err != nil {
			// 如果不存在则创建
			baseBalance = model.Balance{
				UserID:    order.UserID,
				Asset:     baseCoin,
				Available: receivedAmount,
				Locked:    0,
			}
			if err := tx.Create(&baseBalance).Error; err != nil {
				return fmt.Errorf("failed to create base balance: %w", err)
			}
		} else {
			baseBalance.Available += receivedAmount
			if err := tx.Save(&baseBalance).Error; err != nil {
				return fmt.Errorf("failed to update base balance: %w", err)
			}
		}

	} else if order.Side == "sell" {
		// 卖单：扣除冻结的 BTC，增加 USDT (扣除手续费)
		// 扣除冻结的 BTC
		var baseBalance model.Balance
		if err := tx.Where("user_id = ? AND asset = ?", order.UserID, baseCoin).
			First(&baseBalance).Error; err != nil {
			return fmt.Errorf("base balance not found: %w", err)
		}

		baseBalance.Locked -= trade.Amount
		if baseBalance.Locked < 0 {
			baseBalance.Locked = 0
		}

		if err := tx.Save(&baseBalance).Error; err != nil {
			return fmt.Errorf("failed to update base balance: %w", err)
		}

		// 增加 USDT (扣除手续费)
		receivedUSDT := trade.Amount * trade.Price * (1 - m.cfg.Trading.TakerFeeRate)
		var quoteBalance model.Balance
		err := tx.Where("user_id = ? AND asset = ?", order.UserID, quoteCoin).
			First(&quoteBalance).Error
		if err != nil {
			// 如果不存在则创建
			quoteBalance = model.Balance{
				UserID:    order.UserID,
				Asset:     quoteCoin,
				Available: receivedUSDT,
				Locked:    0,
			}
			if err := tx.Create(&quoteBalance).Error; err != nil {
				return fmt.Errorf("failed to create quote balance: %w", err)
			}
		} else {
			quoteBalance.Available += receivedUSDT
			if err := tx.Save(&quoteBalance).Error; err != nil {
				return fmt.Errorf("failed to update quote balance: %w", err)
			}
		}
	}

	return nil
}

// updateOrderStatus 更新订单状态
func (m *MatchingEngine) updateOrderStatus(order *model.Order, filledAmount float64) error {
	order.Filled = filledAmount
	order.Status = "filled"

	if err := m.db.Save(order).Error; err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// calculateFee 计算手续费
func (m *MatchingEngine) calculateFee(amount float64, feeRate float64) float64 {
	return amount * feeRate
}

// getFeeAsset 获取手续费资产
func (m *MatchingEngine) getFeeAsset(order *model.Order) string {
	baseCoin, _ := m.splitSymbol(order.Symbol)
	// 手续费使用交易的基础币种
	return baseCoin
}

// splitSymbol 分割交易对符号
func (m *MatchingEngine) splitSymbol(symbol string) (base string, quote string) {
	// BTC/USDT -> BTC, USDT
	for i := 0; i < len(symbol); i++ {
		if symbol[i] == '/' {
			return symbol[:i], symbol[i+1:]
		}
	}
	return symbol, "USDT" // 默认计价币种
}
