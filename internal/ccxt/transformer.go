package ccxt

import (
	"strconv"
	"strings"
	"time"

	"github.com/talkincode/quicksilver/internal/model"
)

// TransformTicker 将内部 Ticker 模型转换为 CCXT 标准格式
func TransformTicker(ticker *model.Ticker) map[string]interface{} {
	result := map[string]interface{}{
		"symbol":    ticker.Symbol,
		"timestamp": ticker.UpdatedAt.UnixMilli(),
		"datetime":  ticker.UpdatedAt.Format(time.RFC3339Nano),
		"last":      ticker.LastPrice,
		"close":     ticker.LastPrice,
		"info":      map[string]interface{}{}, // 原始数据占位符
	}

	// CCXT 要求的必填字段，使用默认值 0 if nil
	if ticker.High24h != nil {
		result["high"] = *ticker.High24h
	} else {
		result["high"] = 0
	}

	if ticker.Low24h != nil {
		result["low"] = *ticker.Low24h
	} else {
		result["low"] = 0
	}

	if ticker.BidPrice != nil {
		result["bid"] = *ticker.BidPrice
	} else {
		result["bid"] = ticker.LastPrice // 使用 last price 作为后备
	}

	if ticker.AskPrice != nil {
		result["ask"] = *ticker.AskPrice
	} else {
		result["ask"] = ticker.LastPrice // 使用 last price 作为后备
	}

	if ticker.Volume24hBase != nil {
		result["baseVolume"] = *ticker.Volume24hBase
	} else {
		result["baseVolume"] = 0
	}

	if ticker.Volume24hQuote != nil {
		result["quoteVolume"] = *ticker.Volume24hQuote
	} else {
		result["quoteVolume"] = 0
	}

	return result
}

// TransformOrder 将内部 Order 模型转换为 CCXT 标准格式
func TransformOrder(order *model.Order) map[string]interface{} {
	var price interface{}
	if order.Price != nil {
		price = *order.Price
	}

	remaining := order.Amount - order.Filled

	result := map[string]interface{}{
		"id":            strconv.FormatUint(uint64(order.ID), 10),
		"clientOrderId": order.ClientOrderID,
		"timestamp":     order.CreatedAt.UnixMilli(),
		"datetime":      order.CreatedAt.Format(time.RFC3339Nano),
		"symbol":        order.Symbol,
		"type":          order.Type,
		"side":          order.Side,
		"price":         price,
		"amount":        order.Amount,
		"filled":        order.Filled,
		"remaining":     remaining,
		"status":        order.Status,
		"fee": map[string]interface{}{
			"cost":     order.Fee,
			"currency": order.FeeAsset,
		},
	}

	return result
}

// TransformTrade 将内部 Trade 模型转换为 CCXT 标准格式
func TransformTrade(trade *model.Trade) map[string]interface{} {
	cost := trade.Price * trade.Amount

	return map[string]interface{}{
		"id":        strconv.FormatUint(uint64(trade.ID), 10),
		"order":     strconv.FormatUint(uint64(trade.OrderID), 10),
		"symbol":    trade.Symbol,
		"side":      trade.Side,
		"price":     trade.Price,
		"amount":    trade.Amount,
		"cost":      cost,
		"timestamp": trade.CreatedAt.UnixMilli(),
		"datetime":  trade.CreatedAt.Format(time.RFC3339Nano),
		"fee": map[string]interface{}{
			"cost":     trade.Fee,
			"currency": "USDT", // 简化：假设所有费用用 USDT 支付
		},
	}
}

// TransformBalance 将内部 Balance 模型转换为 CCXT 标准格式
func TransformBalance(balance *model.Balance) map[string]interface{} {
	total := balance.Available + balance.Locked

	return map[string]interface{}{
		"currency": balance.Asset,
		"free":     balance.Available,
		"used":     balance.Locked,
		"total":    total,
	}
}

// TransformBalances 将多个余额转换为 CCXT fetchBalance 响应格式
func TransformBalances(balances []*model.Balance) map[string]interface{} {
	result := make(map[string]interface{})

	for _, balance := range balances {
		result[balance.Asset] = map[string]interface{}{
			"free":  balance.Available,
			"used":  balance.Locked,
			"total": balance.Available + balance.Locked,
		}
	}

	return result
}

// TransformMarket 将交易对信息转换为 CCXT market 格式
func TransformMarket(symbol string, minAmount float64) map[string]interface{} {
	parts := strings.Split(symbol, "/")
	base := parts[0]
	quote := "USDT"
	if len(parts) > 1 {
		quote = parts[1]
	}

	return map[string]interface{}{
		"symbol": symbol,
		"id":     symbol,
		"base":   base,
		"quote":  quote,
		"active": true,
		"limits": map[string]interface{}{
			"amount": map[string]interface{}{
				"min": minAmount,
			},
		},
	}
}
