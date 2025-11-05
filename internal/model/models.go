package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username  string     `gorm:"size:50" json:"username,omitempty"`
	APIKey    string     `gorm:"uniqueIndex;size:64;not null" json:"api_key"`
	APISecret string     `gorm:"size:128;not null" json:"-"`
	Status    string     `gorm:"size:20;default:active" json:"status"`
	Role      string     `gorm:"size:20;default:user" json:"role"` // user/admin
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// Balance 余额模型
type Balance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Asset     string    `gorm:"size:10;not null" json:"asset"`
	Available float64   `gorm:"type:decimal(20,8);default:0" json:"available"`
	Locked    float64   `gorm:"type:decimal(20,8);default:0" json:"locked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// Order 订单模型
type Order struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"not null;index" json:"user_id"`
	Symbol           string     `gorm:"size:20;not null;index" json:"symbol"`
	Side             string     `gorm:"size:4;not null" json:"side"`
	Type             string     `gorm:"size:10;not null;index:idx_status_type" json:"type"`               // market/limit/stop_loss/take_profit
	Status           string     `gorm:"size:20;not null;default:new;index:idx_status_type" json:"status"` // new/open/filled/canceled/triggered
	Price            *float64   `gorm:"type:decimal(20,8)" json:"price,omitempty"`
	StopPrice        *float64   `gorm:"type:decimal(20,8)" json:"stop_price,omitempty"` // 止盈止损触发价格
	TriggerCondition string     `gorm:"size:10" json:"trigger_condition,omitempty"`     // ">=" 或 "<="
	Amount           float64    `gorm:"type:decimal(20,8);not null" json:"amount"`
	Filled           float64    `gorm:"type:decimal(20,8);default:0" json:"filled"`
	AveragePrice     *float64   `gorm:"type:decimal(20,8)" json:"average_price,omitempty"`
	Fee              float64    `gorm:"type:decimal(20,8);default:0" json:"fee"`
	FeeAsset         string     `gorm:"size:10" json:"fee_asset,omitempty"`
	ClientOrderID    string     `gorm:"size:64;index" json:"client_order_id,omitempty"`
	ParentOrderID    *uint      `gorm:"index" json:"parent_order_id,omitempty"` // 关联的父订单ID（用于止盈止损）
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FilledAt         *time.Time `json:"filled_at,omitempty"`
	CanceledAt       *time.Time `json:"canceled_at,omitempty"`
	TriggeredAt      *time.Time `json:"triggered_at,omitempty"` // 止盈止损触发时间

	User   *User   `gorm:"foreignKey:UserID" json:"-"`
	Trades []Trade `gorm:"foreignKey:OrderID" json:"trades,omitempty"`
}

// Trade 成交模型
type Trade struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderID     uint      `gorm:"not null;index" json:"order_id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Symbol      string    `gorm:"size:20;not null" json:"symbol"`
	Side        string    `gorm:"size:4;not null" json:"side"`
	Price       float64   `gorm:"type:decimal(20,8);not null" json:"price"`
	Amount      float64   `gorm:"type:decimal(20,8);not null" json:"amount"`
	QuoteAmount float64   `gorm:"type:decimal(20,8);not null" json:"quote_amount"`
	Fee         float64   `gorm:"type:decimal(20,8);default:0" json:"fee"`
	FeeAsset    string    `gorm:"size:10" json:"fee_asset,omitempty"`
	IsMaker     bool      `gorm:"default:false" json:"is_maker"`
	CreatedAt   time.Time `json:"created_at"`

	Order *Order `gorm:"foreignKey:OrderID" json:"-"`
	User  *User  `gorm:"foreignKey:UserID" json:"-"`
}

// Ticker 行情模型
type Ticker struct {
	Symbol                string    `gorm:"primaryKey;size:20" json:"symbol"`
	LastPrice             float64   `gorm:"type:decimal(20,8);not null" json:"last_price"`
	BidPrice              *float64  `gorm:"type:decimal(20,8)" json:"bid_price,omitempty"`
	AskPrice              *float64  `gorm:"type:decimal(20,8)" json:"ask_price,omitempty"`
	High24h               *float64  `gorm:"type:decimal(20,8)" json:"high_24h,omitempty"`
	Low24h                *float64  `gorm:"type:decimal(20,8)" json:"low_24h,omitempty"`
	Volume24hBase         *float64  `gorm:"type:decimal(20,8)" json:"volume_24h_base,omitempty"`
	Volume24hQuote        *float64  `gorm:"type:decimal(20,8)" json:"volume_24h_quote,omitempty"`
	PriceChange24h        *float64  `gorm:"type:decimal(20,8)" json:"price_change_24h,omitempty"`
	PriceChangePercent24h *float64  `gorm:"type:decimal(10,4)" json:"price_change_percent_24h,omitempty"`
	UpdatedAt             time.Time `json:"updated_at"`
	Source                string    `gorm:"size:20;default:binance" json:"source"`
}

// Kline K线/蜡烛图数据模型
type Kline struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `gorm:"size:20;not null;uniqueIndex:idx_symbol_interval_time" json:"symbol"`
	Interval  string    `gorm:"size:10;not null;uniqueIndex:idx_symbol_interval_time" json:"interval"` // 1m, 5m, 15m, 1h, 4h, 1d
	OpenTime  time.Time `gorm:"not null;uniqueIndex:idx_symbol_interval_time" json:"open_time"`
	CloseTime time.Time `gorm:"not null" json:"close_time"`
	Open      float64   `gorm:"type:decimal(20,8);not null" json:"open"`
	High      float64   `gorm:"type:decimal(20,8);not null" json:"high"`
	Low       float64   `gorm:"type:decimal(20,8);not null" json:"low"`
	Close     float64   `gorm:"type:decimal(20,8);not null" json:"close"`
	Volume    float64   `gorm:"type:decimal(20,8);not null;default:0" json:"volume"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Balance) TableName() string {
	return "balances"
}

func (Order) TableName() string {
	return "orders"
}

func (Trade) TableName() string {
	return "trades"
}

func (Ticker) TableName() string {
	return "tickers"
}

func (Kline) TableName() string {
	return "klines"
}
