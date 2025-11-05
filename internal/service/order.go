package service

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/talkincode/quicksilver/internal/config"
	"github.com/talkincode/quicksilver/internal/engine"
	"github.com/talkincode/quicksilver/internal/model"
)

// OrderService 订单管理服务
type OrderService struct {
	db             *gorm.DB
	cfg            *config.Config
	logger         *zap.Logger
	balanceService *BalanceService
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Symbol        string   `json:"symbol"`
	Side          string   `json:"side"`   // buy | sell
	Type          string   `json:"type"`   // market | limit
	Amount        float64  `json:"amount"` // 数量
	Price         *float64 `json:"price"`  // 价格（限价单必填）
	ClientOrderID string   `json:"client_order_id,omitempty"`
}

// NewOrderService 创建订单服务
func NewOrderService(db *gorm.DB, cfg *config.Config, logger *zap.Logger, balanceService *BalanceService) *OrderService {
	return &OrderService{
		db:             db,
		cfg:            cfg,
		logger:         logger,
		balanceService: balanceService,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID uint, req CreateOrderRequest) (*model.Order, error) {
	// 1. 参数验证
	if err := s.validateOrderRequest(req); err != nil {
		return nil, fmt.Errorf("invalid order request: %w", err)
	}

	// 2. 获取当前市场价格（用于市价单）
	var currentPrice float64
	if req.Type == "market" {
		var ticker model.Ticker
		if err := s.db.Where("symbol = ?", req.Symbol).First(&ticker).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("ticker not found for symbol %s", req.Symbol)
			}
			return nil, fmt.Errorf("failed to get ticker: %w", err)
		}
		currentPrice = ticker.LastPrice
	} else {
		// 限价单使用用户指定价格
		currentPrice = *req.Price
	}

	// 3. 计算需要冻结的资金
	frozenAmount, frozenAsset := s.calculateFrozenAmount(req, currentPrice)

	// 4. 检查余额并冻结资金
	if err := s.balanceService.FreezeBalance(userID, frozenAsset, frozenAmount); err != nil {
		return nil, fmt.Errorf("failed to freeze balance: %w", err)
	}

	// 5. 创建订单记录
	order := &model.Order{
		UserID:        userID,
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Amount:        req.Amount,
		Price:         req.Price,
		Status:        "new",
		Filled:        0,
	}

	if err := s.db.Create(order).Error; err != nil {
		// 创建失败，解冻资金
		s.balanceService.UnfreezeBalance(userID, frozenAsset, frozenAmount)
		s.logger.Error("Failed to create order",
			zap.Uint("user_id", userID),
			zap.String("symbol", req.Symbol),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Info("Order created",
		zap.Uint("order_id", order.ID),
		zap.Uint("user_id", userID),
		zap.String("symbol", order.Symbol),
		zap.String("side", order.Side),
		zap.String("type", order.Type),
		zap.Float64("amount", order.Amount),
	)

	// 触发撮合引擎（异步）
	go func() {
		// 延迟导入以避免循环依赖
		engine := s.createMatchingEngine()
		if err := engine.MatchOrder(order.ID); err != nil {
			s.logger.Error("Failed to match order",
				zap.Uint("order_id", order.ID),
				zap.Error(err),
			)
		}
	}()

	return order, nil
}

// GetOrderByID 根据 ID 获取订单
func (s *OrderService) GetOrderByID(orderID uint) (*model.Order, error) {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// GetUserOrders 获取用户订单列表（分页）
func (s *OrderService) GetUserOrders(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	if err := s.db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// 查询订单列表（按创建时间倒序）
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, total, nil
}

// GetOpenOrders 获取用户未完成订单
func (s *OrderService) GetOpenOrders(userID uint) ([]model.Order, error) {
	var orders []model.Order

	if err := s.db.Where("user_id = ? AND status = ?", userID, "new").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	return orders, nil
}

// CancelOrder 撤销订单
func (s *OrderService) CancelOrder(userID, orderID uint) error {
	// 1. 获取订单
	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// 2. 验证订单所有者
	if order.UserID != userID {
		return fmt.Errorf("order does not belong to user")
	}

	// 3. 检查订单状态
	if order.Status != "new" {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	// 4. 计算需要解冻的资金
	var frozenAmount float64
	var frozenAsset string

	if order.Side == "buy" {
		// 买单：解冻 USDT
		frozenAsset = s.getQuoteAsset(order.Symbol)
		if order.Type == "market" {
			// 市价单：需要重新计算当时冻结的金额
			// 这里简化处理，假设按当前价格计算
			var ticker model.Ticker
			if err := s.db.Where("symbol = ?", order.Symbol).First(&ticker).Error; err == nil {
				frozenAmount = order.Amount * ticker.LastPrice
			}
		} else {
			// 限价单：按限价计算
			frozenAmount = order.Amount * (*order.Price)
		}
	} else {
		// 卖单：解冻基础币
		frozenAsset = s.getBaseAsset(order.Symbol)
		frozenAmount = order.Amount
	}

	// 5. 更新订单状态
	order.Status = "cancelled"
	if err := s.db.Save(order).Error; err != nil {
		s.logger.Error("Failed to update order status",
			zap.Uint("order_id", orderID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// 6. 解冻资金
	if err := s.balanceService.UnfreezeBalance(userID, frozenAsset, frozenAmount); err != nil {
		s.logger.Error("Failed to unfreeze balance",
			zap.Uint("order_id", orderID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unfreeze balance: %w", err)
	}

	s.logger.Info("Order cancelled",
		zap.Uint("order_id", orderID),
		zap.Uint("user_id", userID),
	)

	return nil
}

// validateOrderRequest 验证订单请求参数
func (s *OrderService) validateOrderRequest(req CreateOrderRequest) error {
	// 1. 验证交易对
	if req.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	// 2. 验证方向
	if req.Side != "buy" && req.Side != "sell" {
		return fmt.Errorf("side must be buy or sell")
	}

	// 3. 验证订单类型
	if req.Type != "market" && req.Type != "limit" {
		return fmt.Errorf("type must be market or limit")
	}

	// 4. 验证数量
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if req.Amount < s.cfg.Trading.MinOrderAmount {
		return fmt.Errorf("amount is too small, minimum is %.8f", s.cfg.Trading.MinOrderAmount)
	}

	// 5. 限价单必须提供价格
	if req.Type == "limit" && req.Price == nil {
		return fmt.Errorf("price is required for limit orders")
	}

	// 6. 限价单价格必须为正
	if req.Type == "limit" && *req.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}

	return nil
}

// calculateFrozenAmount 计算需要冻结的资金数量和币种
func (s *OrderService) calculateFrozenAmount(req CreateOrderRequest, price float64) (amount float64, asset string) {
	if req.Side == "buy" {
		// 买单：冻结计价币（USDT）
		asset = s.getQuoteAsset(req.Symbol)
		amount = req.Amount * price
	} else {
		// 卖单：冻结基础币（BTC）
		asset = s.getBaseAsset(req.Symbol)
		amount = req.Amount
	}
	return amount, asset
}

// getBaseAsset 从交易对获取基础币种 (BTC/USDT -> BTC)
func (s *OrderService) getBaseAsset(symbol string) string {
	parts := strings.Split(symbol, "/")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// getQuoteAsset 从交易对获取计价币种 (BTC/USDT -> USDT)
func (s *OrderService) getQuoteAsset(symbol string) string {
	parts := strings.Split(symbol, "/")
	if len(parts) == 2 {
		return parts[1]
	}
	return "USDT" // 默认
}

// createMatchingEngine 创建撮合引擎实例
func (s *OrderService) createMatchingEngine() *engine.MatchingEngine {
	return engine.NewMatchingEngine(s.db, s.cfg, s.logger)
}

// CreateStopLossOrder 创建止损单
func (s *OrderService) CreateStopLossOrder(userID uint, symbol, side string, amount, stopPrice float64) (*model.Order, error) {
	s.logger.Debug("CreateStopLossOrder called",
		zap.Uint("user_id", userID),
		zap.String("symbol", symbol),
		zap.String("side", side),
		zap.Float64("amount", amount),
		zap.Float64("stop_price", stopPrice),
	)

	// 1. 参数验证
	if side != "sell" && side != "buy" {
		return nil, fmt.Errorf("invalid side: %s", side)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if stopPrice <= 0 {
		return nil, fmt.Errorf("stop price must be positive")
	}

	// 2. 检查余额（止损单需要冻结资产）
	var asset string
	var frozenAmount float64
	if side == "sell" {
		// 卖单止损：冻结基础币
		asset = s.getBaseAsset(symbol)
		frozenAmount = amount
	} else {
		// 买单止损：冻结计价币（需要估算成本）
		asset = s.getQuoteAsset(symbol)
		frozenAmount = amount * stopPrice
	}

	if err := s.balanceService.CheckBalance(userID, asset, frozenAmount); err != nil {
		return nil, fmt.Errorf("insufficient balance: %w", err)
	}

	// 3. 冻结资金
	if err := s.balanceService.FreezeBalance(userID, asset, frozenAmount); err != nil {
		return nil, fmt.Errorf("failed to freeze balance: %w", err)
	}

	// 4. 创建止损单
	// 止损条件：卖单价格 <= 止损价，买单价格 >= 止损价
	triggerCondition := "<=" // 默认卖单止损
	if side == "buy" {
		triggerCondition = ">="
	}

	order := &model.Order{
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		Type:             "stop_loss",
		Status:           "new",
		StopPrice:        &stopPrice,
		TriggerCondition: triggerCondition,
		Amount:           amount,
	}

	if err := s.db.Create(order).Error; err != nil {
		// 回滚冻结
		s.balanceService.UnfreezeBalance(userID, asset, frozenAmount)
		return nil, fmt.Errorf("failed to create stop loss order: %w", err)
	}

	s.logger.Info("Stop loss order created",
		zap.Uint("order_id", order.ID),
		zap.String("symbol", symbol),
		zap.Float64("stop_price", stopPrice),
	)

	return order, nil
}

// CreateTakeProfitOrder 创建止盈单
func (s *OrderService) CreateTakeProfitOrder(userID uint, symbol, side string, amount, takeProfitPrice float64) (*model.Order, error) {
	s.logger.Debug("CreateTakeProfitOrder called",
		zap.Uint("user_id", userID),
		zap.String("symbol", symbol),
		zap.String("side", side),
		zap.Float64("amount", amount),
		zap.Float64("take_profit_price", takeProfitPrice),
	)

	// 1. 参数验证
	if side != "sell" && side != "buy" {
		return nil, fmt.Errorf("invalid side: %s", side)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if takeProfitPrice <= 0 {
		return nil, fmt.Errorf("take profit price must be positive")
	}

	// 2. 检查余额
	var asset string
	var frozenAmount float64
	if side == "sell" {
		asset = s.getBaseAsset(symbol)
		frozenAmount = amount
	} else {
		asset = s.getQuoteAsset(symbol)
		frozenAmount = amount * takeProfitPrice
	}

	if err := s.balanceService.CheckBalance(userID, asset, frozenAmount); err != nil {
		return nil, fmt.Errorf("insufficient balance: %w", err)
	}

	// 3. 冻结资金
	if err := s.balanceService.FreezeBalance(userID, asset, frozenAmount); err != nil {
		return nil, fmt.Errorf("failed to freeze balance: %w", err)
	}

	// 4. 创建止盈单
	// 止盈条件：卖单价格 >= 止盈价，买单价格 <= 止盈价
	triggerCondition := ">=" // 默认卖单止盈
	if side == "buy" {
		triggerCondition = "<="
	}

	order := &model.Order{
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		Type:             "take_profit",
		Status:           "new",
		StopPrice:        &takeProfitPrice,
		TriggerCondition: triggerCondition,
		Amount:           amount,
	}

	if err := s.db.Create(order).Error; err != nil {
		s.balanceService.UnfreezeBalance(userID, asset, frozenAmount)
		return nil, fmt.Errorf("failed to create take profit order: %w", err)
	}

	s.logger.Info("Take profit order created",
		zap.Uint("order_id", order.ID),
		zap.String("symbol", symbol),
		zap.Float64("take_profit_price", takeProfitPrice),
	)

	return order, nil
}
