# GitHub Copilot 编码指南：Quicksilver 项目

> **项目概述**: CCXT 兼容的精简模拟加密货币交易所，用于量化交易策略测试  
> **技术栈**: Go 1.24.0 + Echo v4.13.4 + GORM v1.31.0 + PostgreSQL 16+  
> **架构模式**: 单体分层架构 (MVP)，后期可演进为微服务  
> **开发模式**: 测试驱动开发 (TDD) - 测试先行、重构优先、质量保障  
> **设计原则**: 快速开发、基础功能优先、接口 CCXT 兼容

## ⚠️ 重要提示

**❌ 禁止自动生成文档**

- 不要在每次代码修改后自动创建或更新 Markdown 文档
- 不要创建 CHANGELOG.md、变更摘要或操作日志
- 只在用户明确要求时才生成文档
- 专注于代码实现和测试，而非文档编写

**✅ 响应规范**

- 简洁回复：完成任务后简短确认即可
- 直接执行：使用工具直接修改代码，不要展示代码块
- 测试优先：关注测试通过和功能实现
- 错误处理：遇到问题时提供清晰的错误信息和解决方案

---

## 0. TDD 开发模式 (Test-Driven Development)

### 0.1 核心理念

**⚠️ 强制要求：所有新功能和 Bug 修复必须遵循 TDD 流程**

```
红 → 绿 → 重构
Red → Green → Refactor

1. 🔴 Red:   先写失败的测试 (定义预期行为)
2. 🟢 Green: 写最简单的代码让测试通过 (实现功能)
3. 🔵 Refactor: 重构代码提升质量 (优化设计)
```

### 0.2 TDD 工作流程

#### 步骤 1: 红阶段 - 编写失败的测试

```go
// ✅ 正确示例：为新功能先写测试
// File: internal/service/order_test.go

func TestCreateOrder(t *testing.T) {
    db := testutil.SetupTestDB(t)
    cfg := testutil.LoadTestConfig(t)
    logger := testutil.NewTestLogger()

    orderService := NewOrderService(db, cfg, logger)

    t.Run("Create market buy order", func(t *testing.T) {
        // Given: 用户有足够余额
        userID := uint(1)
        testutil.SeedBalance(t, db, userID, "USDT", 10000.0)

        // When: 创建市价买单
        order, err := orderService.CreateOrder(userID, CreateOrderRequest{
            Symbol: "BTC/USDT",
            Side:   "buy",
            Type:   "market",
            Amount: 0.1,
        })

        // Then: 订单创建成功
        require.NoError(t, err)
        assert.NotZero(t, order.ID)
        assert.Equal(t, "new", order.Status)
        assert.Equal(t, 0.1, order.Amount)

        // And: 资金被正确冻结
        balance := testutil.GetBalance(t, db, userID, "USDT")
        assert.Greater(t, balance.Locked, 0.0)
    })
}
```

**此时运行测试应该失败** ❌，因为 `CreateOrder` 方法还未实现。

#### 步骤 2: 绿阶段 - 让测试通过

```go
// File: internal/service/order.go

func (s *OrderService) CreateOrder(userID uint, req CreateOrderRequest) (*model.Order, error) {
    // 最简单的实现让测试通过
    order := &model.Order{
        UserID: userID,
        Symbol: req.Symbol,
        Side:   req.Side,
        Type:   req.Type,
        Amount: req.Amount,
        Status: "new",
    }

    if err := s.db.Create(order).Error; err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    // 冻结资金 (简化实现)
    if err := s.freezeBalance(userID, req.Symbol, req.Side, req.Amount); err != nil {
        return nil, err
    }

    return order, nil
}
```

**运行测试应该通过** ✅

#### 步骤 3: 重构阶段 - 优化代码

```go
// ✅ 优化版本：添加事务、验证、错误处理

func (s *OrderService) CreateOrder(userID uint, req CreateOrderRequest) (*model.Order, error) {
    // 1. 参数验证
    if err := s.validateOrderRequest(req); err != nil {
        return nil, fmt.Errorf("invalid order request: %w", err)
    }

    // 2. 余额检查
    if err := s.checkBalance(userID, req); err != nil {
        return nil, fmt.Errorf("insufficient balance: %w", err)
    }

    // 3. 使用事务确保原子性
    var order *model.Order
    err := s.db.Transaction(func(tx *gorm.DB) error {
        order = &model.Order{
            UserID: userID,
            Symbol: req.Symbol,
            Side:   req.Side,
            Type:   req.Type,
            Amount: req.Amount,
            Status: "new",
        }

        if err := tx.Create(order).Error; err != nil {
            return err
        }

        // 冻结资金
        return s.freezeBalanceInTx(tx, userID, req)
    })

    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    s.logger.Info("Order created",
        zap.Uint("order_id", order.ID),
        zap.String("symbol", order.Symbol),
    )

    return order, nil
}
```

**再次运行测试确保重构后仍然通过** ✅

### 0.3 测试编写规范

#### 测试命名规范

```go
// ✅ 正确：测试函数名清晰描述测试内容
func TestCreateOrder(t *testing.T) { ... }
func TestCreateOrder_WithInsufficientBalance(t *testing.T) { ... }
func TestUpdateTicker_WhenAPIReturnsError(t *testing.T) { ... }

// ❌ 错误：名称过于简短或模糊
func TestOrder(t *testing.T) { ... }
func Test1(t *testing.T) { ... }
```

#### 测试结构：Given-When-Then

```go
// ✅ 推荐：使用 Given-When-Then 结构
func TestCancelOrder(t *testing.T) {
    t.Run("Cancel open order successfully", func(t *testing.T) {
        // Given: 存在一个未成交订单
        db := testutil.SetupTestDB(t)
        order := testutil.SeedOrder(t, db, OrderParams{
            Status: "new",
            Amount: 1.0,
        })

        // When: 用户取消订单
        err := orderService.CancelOrder(order.UserID, order.ID)

        // Then: 订单状态变为已取消
        require.NoError(t, err)

        var updated model.Order
        db.First(&updated, order.ID)
        assert.Equal(t, "cancelled", updated.Status)

        // And: 冻结资金被释放
        balance := testutil.GetBalance(t, db, order.UserID, "USDT")
        assert.Equal(t, 0.0, balance.Locked)
    })
}
```

#### 测试覆盖率要求

```bash
# ✅ 目标：核心业务逻辑覆盖率 > 80%
make test-coverage

# 查看覆盖率报告
open coverage.html
```

**覆盖率指标**:

- **Service 层**: 必须 ≥ 80% (核心业务逻辑)
- **Model 层**: 必须 = 100% (数据模型验证)
- **API 层**: 推荐 ≥ 60% (HTTP Handler)
- **整体项目**: 推荐 ≥ 70%

### 0.4 测试分类与策略

#### 单元测试 (Unit Tests)

```go
// ✅ 单元测试：测试单个函数/方法，使用 Mock/Stub
func TestConvertSymbolToCoin(t *testing.T) {
    tests := []struct {
        name     string
        symbol   string
        expected string
    }{
        {"BTC/USDT", "BTC/USDT", "BTC"},
        {"ETH/USDT", "ETH/USDT", "ETH"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := convertSymbolToCoin(tt.symbol)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**运行单元测试**:

```bash
make test-unit
# 或
CGO_ENABLED=1 go test -v -short ./...
```

#### 集成测试 (Integration Tests)

```go
// ✅ 集成测试：测试多个组件协作，使用真实依赖
func TestMarketServiceIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // 使用真实的 HTTP 服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "mids": map[string]string{
                "BTC": "50000.5",
            },
        })
    }))
    defer server.Close()

    cfg := testutil.LoadTestConfig(t)
    cfg.Market.APIURL = server.URL

    db := testutil.SetupTestDB(t)
    marketService := NewMarketService(db, cfg, testutil.NewTestLogger())

    err := marketService.UpdateTickers()
    require.NoError(t, err)

    // 验证数据库已更新
    var ticker model.Ticker
    err = db.Where("symbol = ?", "BTC/USDT").First(&ticker).Error
    require.NoError(t, err)
    assert.Equal(t, 50000.5, ticker.LastPrice)
}
```

**运行集成测试**:

```bash
make test-integration
# 或
CGO_ENABLED=1 go test -v -run Integration ./...
```

#### 表驱动测试 (Table-Driven Tests)

```go
// ✅ 表驱动测试：测试多个场景
func TestValidateOrderRequest(t *testing.T) {
    tests := []struct {
        name    string
        req     CreateOrderRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "Valid market order",
            req: CreateOrderRequest{
                Symbol: "BTC/USDT",
                Side:   "buy",
                Type:   "market",
                Amount: 0.1,
            },
            wantErr: false,
        },
        {
            name: "Invalid symbol",
            req: CreateOrderRequest{
                Symbol: "",
                Side:   "buy",
                Type:   "market",
                Amount: 0.1,
            },
            wantErr: true,
            errMsg:  "symbol is required",
        },
        {
            name: "Invalid amount",
            req: CreateOrderRequest{
                Symbol: "BTC/USDT",
                Side:   "buy",
                Type:   "market",
                Amount: -0.1,
            },
            wantErr: true,
            errMsg:  "amount must be positive",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateOrderRequest(tt.req)
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### 0.5 测试工具与辅助函数

#### 测试数据库设置

```go
// File: internal/testutil/testutil.go

// ✅ 使用内存 SQLite 数据库进行测试
func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper()

    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err, "failed to create test database")

    // 自动迁移所有模型
    err = db.AutoMigrate(
        &model.User{},
        &model.Balance{},
        &model.Order{},
        &model.Trade{},
        &model.Ticker{},
    )
    require.NoError(t, err)

    return db
}
```

#### 测试数据种子函数

```go
// ✅ 提供便捷的测试数据创建函数
func SeedUser(t *testing.T, db *gorm.DB) *model.User {
    t.Helper()

    user := &model.User{
        Email:     fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
        APIKey:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
        APISecret: "secret-123",
        Status:    "active",
    }

    err := db.Create(user).Error
    require.NoError(t, err)

    return user
}

func SeedBalance(t *testing.T, db *gorm.DB, userID uint, asset string, amount float64) *model.Balance {
    t.Helper()

    balance := &model.Balance{
        UserID:    userID,
        Asset:     asset,
        Available: amount,
        Locked:    0,
    }

    err := db.Create(balance).Error
    require.NoError(t, err)

    return balance
}
```

### 0.6 Mock 与 Stub 策略

```go
// ✅ 使用接口实现依赖注入，便于 Mock
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type MarketService struct {
    db     *gorm.DB
    cfg    *config.Config
    logger *zap.Logger
    client HTTPClient  // 接口类型，可以 Mock
}

// 测试中使用 Mock Client
type mockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

func TestUpdateTickers_WithMockClient(t *testing.T) {
    mockClient := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            // 返回模拟响应
            resp := &http.Response{
                StatusCode: 200,
                Body: io.NopCloser(strings.NewReader(`{"mids":{"BTC":"50000"}}`)),
            }
            return resp, nil
        },
    }

    service := &MarketService{
        client: mockClient,
        // ... 其他依赖
    }

    err := service.UpdateTickers()
    assert.NoError(t, err)
}
```

### 0.7 持续集成中的测试

```yaml
# ✅ GitHub Actions 配置示例
# File: .github/workflows/test.yml

name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.24"

      - name: Run tests
        run: |
          make test
        env:
          CGO_ENABLED: 1

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### 0.8 TDD 实践清单

**每次开发新功能时，必须遵循以下检查清单**:

- [ ] **第一步**: 编写测试用例描述预期行为
- [ ] **第二步**: 运行测试确认失败（红阶段）
- [ ] **第三步**: 编写最简单的代码让测试通过（绿阶段）
- [ ] **第四步**: 重构代码提升质量
- [ ] **第五步**: 运行测试确认仍然通过
- [ ] **第六步**: 提交代码前运行完整测试套件
- [ ] **第七步**: 检查测试覆盖率是否达标

**测试命令快捷方式**:

```bash
# 运行所有测试
make test

# 运行单元测试（快速）
make test-unit

# 查看覆盖率报告
make test-coverage

# 监听文件变化自动测试
make test-watch
```

---

## 1. 核心架构模式

### 1.1 分层架构 (Layered Architecture)

```
┌─────────────────────────────────────────────┐
│ API Layer (internal/api)                    │ ← HTTP 处理器、CCXT 格式转换
├─────────────────────────────────────────────┤
│ Service Layer (internal/service)            │ ← 业务逻辑、撮合引擎、市场数据同步
├─────────────────────────────────────────────┤
│ Repository Layer (internal/repository)      │ ← 数据访问抽象 (待实现)
├─────────────────────────────────────────────┤
│ Model Layer (internal/model)                │ ← GORM 数据模型、数据库映射
└─────────────────────────────────────────────┘
```

**关键规则**:

- ❌ **禁止跨层调用**: API 层不能直接访问 Model，必须通过 Service 层
- ✅ **依赖注入**: 所有服务通过构造函数传入依赖 (DB、Logger、Config)
- ✅ **错误向上传播**: 底层错误使用 `fmt.Errorf("context: %w", err)` 包装后向上抛
- ✅ **职责分离**:
  - **API 层**: 仅负责参数验证、数据格式转换 (内部格式 ↔ CCXT 格式)
  - **Service 层**: 业务逻辑、事务管理、数据缓存
  - **Model 层**: 数据持久化、关系映射

**示例: 正确的分层调用**

```go
// ❌ 错误：API 直接访问数据库
func GetOrder(db *gorm.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        var order model.Order
        db.First(&order, c.Param("id"))  // 不应直接调用 DB
        return c.JSON(200, order)
    }
}

// ✅ 正确：通过 Service 层
func GetOrder(orderService *service.OrderService) echo.HandlerFunc {
    return func(c echo.Context) error {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
        order, err := orderService.GetOrderByID(uint(id))
        if err != nil {
            return c.JSON(404, map[string]string{"error": "order not found"})
        }
        return c.JSON(200, transformToCCXTOrder(order))
    }
}
```

### 1.2 关键技术决策

| 决策点         | 选择                | 理由                       | 代码表现                    |
| -------------- | ------------------- | -------------------------- | --------------------------- |
| **缓存策略**   | 内存缓存 (sync.Map) | MVP 避免 Redis 复杂度      | `MarketService` 内置缓存    |
| **行情数据源** | Hyperliquid API     | 实时市场数据，备选 Binance | `config.market.data_source` |
| **撮合引擎**   | 简化价格匹配        | 不实现完整订单簿           | `MatchingService` (待实现)  |
| **认证机制**   | API Key/Secret      | 兼容 CCXT 客户端           | `AuthConfig.jwt_secret`     |
| **数据库连接** | GORM AutoMigrate    | 开发阶段自动建表           | `database.AutoMigrate()`    |

---

## 2. 代码规范与惯例

### 2.1 Go 语言风格

#### 命名规范

```go
// ✅ 正确：结构体使用 PascalCase
type MarketService struct {
    db     *gorm.DB      // 私有字段使用 camelCase
    logger *zap.Logger
}

// ✅ 正确：导出方法使用 PascalCase，私有方法使用 camelCase
func (s *MarketService) UpdateTickers() error { ... }
func (s *MarketService) updateHyperliquidTickers() error { ... }

// ❌ 错误：不使用蛇形命名
func update_tickers() error { ... }  // 不符合 Go 规范
```

#### 错误处理

```go
// ✅ 正确：包装错误并添加上下文
func (s *MarketService) UpdateTickers() error {
    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to fetch tickers from %s: %w", s.cfg.Market.APIURL, err)
    }
    // ... 继续处理
}

// ❌ 错误：吞掉错误或返回 nil 错误
if err != nil {
    log.Println("Error:", err)  // 不应只记录日志
    return nil                   // 不应返回 nil
}

// ✅ 正确：使用结构化日志记录非致命错误
s.logger.Error("Failed to save ticker",
    zap.String("symbol", symbol),
    zap.Error(err),
)
```

#### 指针接收器规则

```go
// ✅ 正确：修改状态或大型结构体使用指针接收器
func (s *MarketService) StartAutoUpdate() { ... }  // 需要访问字段

// ✅ 正确：小型不可变结构体可使用值接收器
func (r jsonReader) Read(p []byte) (n int, err error) { ... }

// ⚠️ 警告：同一类型的所有方法应保持一致（全部指针或全部值）
```

### 2.2 GORM 使用规范

#### 模型定义

```go
// ✅ 正确：使用完整的 GORM 标签
type Order struct {
    ID       uint      `gorm:"primaryKey" json:"id"`
    UserID   uint      `gorm:"not null;index" json:"user_id"`
    Symbol   string    `gorm:"size:20;not null;index" json:"symbol"`
    Side     string    `gorm:"size:4;not null" json:"side"`  // buy/sell
    Type     string    `gorm:"size:10;not null" json:"type"` // market/limit
    Status   string    `gorm:"size:20;not null;default:new" json:"status"`
    Price    *float64  `gorm:"type:decimal(20,8)" json:"price,omitempty"`  // 限价单必填
    Amount   float64   `gorm:"type:decimal(20,8);not null" json:"amount"`

    // 关联字段
    User   *User   `gorm:"foreignKey:UserID" json:"-"`  // 不序列化到 JSON
    Trades []Trade `gorm:"foreignKey:OrderID" json:"trades,omitempty"`
}

// ✅ 正确：指定表名
func (Order) TableName() string {
    return "orders"
}
```

#### 数据库操作模式

```go
// ✅ 正确：使用 Save() 进行 UPSERT (插入或更新)
ticker := model.Ticker{
    Symbol:    "BTC/USDT",
    LastPrice: 109965.50,
    Source:    "hyperliquid",
}
s.db.Save(&ticker)  // 如果 Symbol 存在则更新，否则插入

// ✅ 正确：使用事务处理多步操作
tx := db.Begin()
if err := tx.Create(&order).Error; err != nil {
    tx.Rollback()
    return err
}
if err := tx.Model(&balance).Update("available", newBalance).Error; err != nil {
    tx.Rollback()
    return err
}
tx.Commit()

// ❌ 错误：忘记检查错误
db.Create(&user)  // 应该检查 .Error
```

### 2.3 Echo 路由规范

#### 路由注册模式

```go
// ✅ 正确：使用闭包传递依赖
func SetupRoutes(e *echo.Echo, db *gorm.DB, cfg *config.Config, logger *zap.Logger) {
    // 依赖注入
    orderService := service.NewOrderService(db, cfg, logger)

    // 路由分组
    v1 := e.Group("/v1")
    public := v1.Group("")
    {
        public.GET("/ticker/:symbol", api.GetTicker(db))
    }

    private := v1.Group("")
    private.Use(middleware.Auth(cfg))  // 认证中间件
    {
        private.POST("/order", api.CreateOrder(orderService))
    }
}

// ❌ 错误：在 Handler 内部创建服务（性能差）
func CreateOrder(db *gorm.DB, cfg *config.Config) echo.HandlerFunc {
    return func(c echo.Context) error {
        svc := service.NewOrderService(db, cfg, nil)  // 每次请求都创建新实例
        // ...
    }
}
```

#### Handler 编写模式

```go
// ✅ 正确：标准 Handler 模式
func GetTicker(db *gorm.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        symbol := c.Param("symbol")

        // 参数验证
        if symbol == "" {
            return c.JSON(400, map[string]string{"error": "symbol is required"})
        }

        // 业务逻辑
        var ticker model.Ticker
        if err := db.Where("symbol = ?", symbol).First(&ticker).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return c.JSON(404, map[string]string{"error": "ticker not found"})
            }
            return c.JSON(500, map[string]string{"error": "internal server error"})
        }

        // 返回 CCXT 格式
        return c.JSON(200, map[string]interface{}{
            "symbol":    ticker.Symbol,
            "last":      ticker.LastPrice,
            "timestamp": ticker.UpdatedAt.UnixMilli(),
        })
    }
}
```

---

## 3. 关键业务逻辑模式

### 3.1 市场数据同步 (MarketService)

**当前实现**: `internal/service/market.go`

#### 核心流程

```
定时器 → UpdateTickers() → Hyperliquid API (POST /info)
                          → 解析 JSON (allMids)
                          → GORM Save (UPSERT)
                          → 日志记录
```

#### 关键代码模式

```go
// ✅ 当前模式：后台 Goroutine + Ticker
func (s *MarketService) StartAutoUpdate() {
    ticker := time.NewTicker(1 * time.Second)
    go func() {
        // 立即执行一次
        if err := s.UpdateTickers(); err != nil {
            s.logger.Error("Failed to update tickers", zap.Error(err))
        }

        // 定时循环
        for range ticker.C {
            if err := s.UpdateTickers(); err != nil {
                s.logger.Error("Failed to update tickers", zap.Error(err))
            }
        }
    }()
}

// ✅ Hyperliquid API 请求模式
func (s *MarketService) updateHyperliquidTickers() error {
    // 1. 构造 JSON 请求体
    requestBody := map[string]interface{}{"type": "allMids"}
    jsonData, _ := json.Marshal(requestBody)

    // 2. POST 请求 (注意：不是 GET)
    url := s.cfg.Market.APIURL + s.cfg.Market.Hyperliquid.InfoEndpoint
    req, _ := http.NewRequest("POST", url, &jsonReader{data: jsonData})
    req.Header.Set("Content-Type", "application/json")

    // 3. 解析响应
    var midsResp HyperliquidAllMidsResponse
    json.NewDecoder(resp.Body).Decode(&midsResp)

    // 4. 批量更新数据库
    for _, symbol := range s.cfg.Market.Symbols {
        coin := convertSymbolToCoin(symbol)  // BTC/USDT -> BTC
        if priceStr, ok := midsResp.Mids[coin]; ok {
            ticker := model.Ticker{Symbol: symbol, LastPrice: price}
            s.db.Save(&ticker)  // UPSERT
        }
    }
}
```

**扩展指南**:

- 添加新数据源时，参考 `updateHyperliquidTickers()` 模式
- 添加新交易对时，修改 `config.yaml` 的 `market.symbols` 列表
- 需要 WebSocket 实时推送时，在 `StartAutoUpdate()` 中启动新 Goroutine

### 3.2 订单处理流程 (待实现)

**设计模式**:

```go
// ✅ 推荐：使用 Service 封装复杂业务逻辑
type OrderService struct {
    db     *gorm.DB
    cfg    *config.Config
    logger *zap.Logger
}

func (s *OrderService) CreateOrder(userID uint, req CreateOrderRequest) (*model.Order, error) {
    // 1. 参数验证
    if req.Amount < s.cfg.Trading.MinOrderAmount {
        return nil, fmt.Errorf("amount too small: minimum is %.8f", s.cfg.Trading.MinOrderAmount)
    }

    // 2. 余额检查
    if err := s.checkBalance(userID, req.Symbol, req.Side, req.Amount); err != nil {
        return nil, fmt.Errorf("insufficient balance: %w", err)
    }

    // 3. 创建订单 + 冻结资金 (事务)
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    order := &model.Order{
        UserID: userID,
        Symbol: req.Symbol,
        Side:   req.Side,
        Type:   req.Type,
        Amount: req.Amount,
        Status: "new",
    }

    if err := tx.Create(order).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 冻结资金
    if err := s.freezeBalance(tx, userID, req.Symbol, req.Side, req.Amount); err != nil {
        tx.Rollback()
        return nil, err
    }

    tx.Commit()

    // 4. 触发撮合引擎 (异步)
    go s.matchOrder(order.ID)

    return order, nil
}
```

### 3.3 撮合引擎设计 (待实现)

**简化策略** (MVP 阶段):

```go
// ✅ MVP 版本：直接成交，不维护订单簿
func (s *MatchingService) MatchOrder(orderID uint) error {
    var order model.Order
    s.db.First(&order, orderID)

    // 市价单：直接以当前市场价格成交
    if order.Type == "market" {
        var ticker model.Ticker
        s.db.Where("symbol = ?", order.Symbol).First(&ticker)

        // 创建成交记录
        trade := model.Trade{
            OrderID: order.ID,
            UserID:  order.UserID,
            Symbol:  order.Symbol,
            Side:    order.Side,
            Price:   ticker.LastPrice,
            Amount:  order.Amount,
        }
        s.db.Create(&trade)

        // 更新订单状态
        order.Status = "filled"
        order.Filled = order.Amount
        s.db.Save(&order)

        // 解冻并扣除资金
        s.settleBalance(&order, &trade)
    }

    // 限价单：检查价格是否满足条件
    // TODO: 实现限价单匹配逻辑

    return nil
}
```

---

## 4. 配置管理模式

### 4.1 Viper 配置加载

**当前模式**: `internal/config/config.go`

```go
// ✅ 正确：支持多种配置来源
func Load() (*Config, error) {
    v := viper.New()

    // 1. 配置文件
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AddConfigPath(".")  // 备用路径

    // 2. 环境变量 (优先级高于配置文件)
    v.SetEnvPrefix("QS")                          // QS_SERVER_PORT
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))  // server.port -> SERVER_PORT
    v.AutomaticEnv()

    // 3. 读取并解析
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var config Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &config, nil
}
```

**配置优先级**: 环境变量 > 配置文件 > 默认值

#### 配置结构设计

```go
// ✅ 正确：使用嵌套结构体组织配置
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Market   MarketConfig   `mapstructure:"market"`
    Trading  TradingConfig  `mapstructure:"trading"`
}

// ✅ 正确：为配置添加辅助方法
func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
        c.Host, c.Port, c.User, c.Password, c.Name)
}
```

### 4.2 配置文件示例

**路径**: `config/config.yaml` (从 `config.example.yaml` 复制)

```yaml
server:
  port: 8080
  mode: debug # debug | release
  version: "0.1.0"

database:
  host: localhost
  port: 5432
  name: quicksilver
  user: quicksilver
  password: quicksilver123
  sslmode: disable

market:
  update_interval: "1s"
  data_source: "hyperliquid" # hyperliquid | binance
  api_url: "https://api.hyperliquid.xyz"
  symbols:
    - "BTC/USDT"
    - "ETH/USDT"
  hyperliquid:
    info_endpoint: "/info"
    ws_endpoint: "/ws"

trading:
  default_fee_rate: 0.001
  maker_fee_rate: 0.0005
  taker_fee_rate: 0.001
  min_order_amount: 0.0001

logging:
  level: "debug" # debug | info | warn | error
  format: "console" # console | json
```

---

## 5. 日志记录规范

### 5.1 Zap 结构化日志

**初始化** (在 `cmd/server/main.go`):

```go
// ✅ 正确：根据环境选择日志格式
func initLogger(cfg *config.Config) (*zap.Logger, error) {
    if cfg.Logging.Format == "json" {
        return zap.NewProduction()  // 生产环境：JSON 格式
    }
    return zap.NewDevelopment()     // 开发环境：彩色 Console 格式
}
```

**使用规范**:

```go
// ✅ 正确：使用结构化字段
logger.Info("Ticker updated",
    zap.String("symbol", "BTC/USDT"),
    zap.Float64("price", 109965.50),
    zap.String("source", "hyperliquid"),
)

logger.Error("Failed to save ticker",
    zap.String("symbol", symbol),
    zap.Error(err),
)

// ❌ 错误：使用字符串拼接
logger.Info(fmt.Sprintf("Ticker updated: %s - %.2f", symbol, price))  // 难以解析

// ⚠️ 警告：避免记录敏感信息
logger.Debug("User authenticated",
    zap.String("api_key", user.APIKey),    // ❌ 不应记录密钥
    zap.String("api_secret", user.APISecret),  // ❌ 严禁记录
)
```

**日志级别选择**:

- `Debug`: 开发调试信息 (如每次 Ticker 更新)
- `Info`: 重要事件 (如服务启动、用户登录)
- `Warn`: 异常但可恢复 (如 API 请求失败但会重试)
- `Error`: 严重错误 (如数据库连接失败)
- `Fatal`: 致命错误需立即退出

---

## 6. 数据库操作最佳实践

### 6.1 自动迁移

**初始化阶段** (在 `cmd/server/main.go`):

```go
// ✅ 正确：启动时自动迁移所有模型
func main() {
    db, _ := database.NewDatabase(cfg)
    database.AutoMigrate(db)  // 创建或更新表结构
    // ...
}

// internal/database/database.go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &model.User{},
        &model.Balance{},
        &model.Order{},
        &model.Trade{},
        &model.Ticker{},
    )
}
```

**⚠️ 注意事项**:

- AutoMigrate 只添加新字段/表，不会删除已有字段
- 生产环境应使用正式迁移工具 (如 `golang-migrate`)
- 修改字段类型需手动执行 SQL

### 6.2 索引策略

**参考**: `docs/database.md` 的索引设计

```go
// ✅ 正确：为高频查询字段添加索引
type Order struct {
    UserID uint   `gorm:"not null;index"`           // WHERE user_id = ?
    Symbol string `gorm:"size:20;not null;index"`   // WHERE symbol = ?
    Status string `gorm:"size:20;index"`            // WHERE status IN (...)
}

// ✅ 正确：复合索引 (需手动创建)
// 在 AutoMigrate 后执行:
db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_user_symbol
         ON orders(user_id, symbol)`)  // WHERE user_id = ? AND symbol = ?
```

### 6.3 查询优化

```go
// ✅ 正确：使用预加载避免 N+1 查询
var orders []model.Order
db.Preload("User").Preload("Trades").Where("user_id = ?", userID).Find(&orders)

// ❌ 错误：循环中查询 (N+1 问题)
var orders []model.Order
db.Where("user_id = ?", userID).Find(&orders)
for _, order := range orders {
    var user model.User
    db.First(&user, order.UserID)  // 每个订单查询一次
}

// ✅ 正确：使用选择字段减少数据传输
db.Select("id", "symbol", "status").Where("user_id = ?", userID).Find(&orders)

// ✅ 正确：使用分页避免大数据集
db.Limit(100).Offset(page * 100).Find(&orders)
```

---

## 7. CCXT 兼容性要求

### 7.1 响应格式转换

**核心原则**: 内部使用 Go 结构体，API 返回 CCXT 标准格式

#### Ticker 格式

```go
// ✅ CCXT 标准格式
{
    "symbol": "BTC/USDT",
    "timestamp": 1703001234567,
    "datetime": "2023-12-20T12:34:56.789Z",
    "high": 109965.50,
    "low": 105000.00,
    "bid": 109960.00,
    "ask": 109970.00,
    "last": 109965.50,
    "close": 109965.50,
    "baseVolume": 123.45,
    "quoteVolume": 13567890.12,
    "info": {}  // 原始数据
}

// 转换函数
func transformToCCXTTicker(t *model.Ticker) map[string]interface{} {
    return map[string]interface{}{
        "symbol":      t.Symbol,
        "timestamp":   t.UpdatedAt.UnixMilli(),
        "datetime":    t.UpdatedAt.Format(time.RFC3339Nano),
        "last":        t.LastPrice,
        "bid":         t.BidPrice,
        "ask":         t.AskPrice,
        "high":        t.High24h,
        "low":         t.Low24h,
        "baseVolume":  t.Volume24hBase,
        "quoteVolume": t.Volume24hQuote,
    }
}
```

#### Order 格式

```go
// ✅ CCXT 标准格式
{
    "id": "12345",
    "clientOrderId": "user_order_001",
    "timestamp": 1703001234567,
    "datetime": "2023-12-20T12:34:56.789Z",
    "symbol": "BTC/USDT",
    "type": "limit",
    "side": "buy",
    "price": 109000.00,
    "amount": 0.5,
    "filled": 0.3,
    "remaining": 0.2,
    "status": "open",
    "fee": {"cost": 0.0005, "currency": "USDT"}
}
```

### 7.2 API 端点映射

| CCXT 方法             | HTTP 端点                | 实现状态  |
| --------------------- | ------------------------ | --------- |
| `fetchTicker(symbol)` | `GET /v1/ticker/:symbol` | ✅ 已实现 |
| `fetchBalance()`      | `GET /v1/balance`        | ⏳ 待实现 |
| `createOrder(...)`    | `POST /v1/order`         | ⏳ 待实现 |
| `cancelOrder(id)`     | `DELETE /v1/order/:id`   | ⏳ 待实现 |
| `fetchOrders()`       | `GET /v1/orders`         | ⏳ 待实现 |
| `fetchOpenOrders()`   | `GET /v1/orders/open`    | ⏳ 待实现 |
| `fetchMyTrades()`     | `GET /v1/myTrades`       | ⏳ 待实现 |

---

## 8. 开发工作流

### 8.1 TDD 开发循环（推荐工作流）

**⚠️ 强烈推荐：每次开发新功能都遵循此流程**

```bash
# 1. 🔴 红阶段：编写失败的测试
vim internal/service/order_test.go  # 先写测试

# 2. 运行测试，确认失败
make test-unit
# 预期输出：FAIL (因为功能还未实现)

# 3. 🟢 绿阶段：实现最简单的代码让测试通过
vim internal/service/order.go  # 编写实现代码

# 4. 运行测试，确认通过
make test-unit
# 预期输出：PASS

# 5. 🔵 重构阶段：优化代码质量
vim internal/service/order.go  # 重构代码

# 6. 再次运行测试，确保仍然通过
make test-unit
# 预期输出：PASS

# 7. 检查测试覆盖率
make test-coverage
open coverage.html  # 查看覆盖率报告

# 8. 提交代码
git add .
git commit -m "feat: implement order creation with TDD"
```

**VS Code 中的 TDD 工作流**:

1. **安装 Go 测试插件**: 已配置在 `.vscode/settings.json` 中自动启用 CGO
2. **运行单个测试**: 点击测试函数上方的 `▶ run test` 按钮
3. **运行文件所有测试**: 点击文件顶部的 `▶ run file tests`
4. **调试测试**: 点击 `🐛 debug test` 设置断点调试
5. **查看覆盖率**: 运行测试后会在编辑器中高亮显示覆盖情况

### 8.2 本地开发

**启动服务** (推荐使用 Air 热重载):

```bash
# 方式 1: 使用 Makefile
make dev          # 启动热重载服务
make db-start     # 启动 PostgreSQL 容器
make db-migrate   # 执行数据库迁移

# 方式 2: 直接运行
go run cmd/server/main.go

# 方式 3: 使用 Air (需安装: go install github.com/cosmtrek/air@latest)
air
```

**运行测试** (CGO 已配置):

```bash
# 运行所有测试（推荐）
make test

# 仅运行单元测试（快速验证）
make test-unit

# 运行集成测试
make test-integration

# 生成覆盖率报告
make test-coverage

# 监听文件变化自动运行测试（开发时推荐）
make test-watch
```

**测试 API**:

```bash
# 健康检查
curl http://localhost:8080/health

# 获取行情
curl http://localhost:8080/v1/ticker/BTC/USDT

# 测试 Hyperliquid 连接
./scripts/test_hyperliquid.sh
```

### 8.3 Docker Compose 部署

**启动完整环境**:

```bash
docker-compose up -d      # 启动 PostgreSQL + Quicksilver
docker-compose logs -f    # 查看日志
docker-compose down       # 停止并删除容器
```

**仅启动数据库**:

```bash
docker-compose up -d postgres
```

### 8.4 调试技巧

#### VSCode 调试配置

创建 `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Quicksilver",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server",
      "env": {
        "QS_SERVER_MODE": "debug",
        "QS_LOGGING_LEVEL": "debug"
      },
      "args": []
    }
  ]
}
```

#### 常用调试命令

```bash
# 查看编译错误详情
go build -v ./cmd/server

# 检查依赖版本
go mod graph | grep -E "github.com/labstack/echo|gorm.io/gorm"

# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 运行测试 (待实现)
go test -v ./...
```

---

## 9. 常见问题与解决方案

### 9.1 编译错误

**问题**: `duplicate package` 声明

```go
package main
package main  // ❌ 重复声明
```

**解决**: 每个文件只保留一个 `package` 声明

---

**问题**: 依赖版本冲突

```
github.com/labstack/echo/v4 v4.12.0
  requires github.com/golang-jwt/jwt v3.2.2+incompatible
  conflicts with github.com/golang-jwt/jwt/v5 v5.0.0
```

**解决**: 升级所有依赖到最新稳定版

```bash
go get -u github.com/labstack/echo/v4@latest
go mod tidy
```

---

**问题**: `io.Reader` 接口实现错误

```go
func (r jsonReader) Read(p []byte) (n int, err error) { ... }
// 错误：值接收器无法修改 offset 字段
```

**解决**: 使用指针接收器

```go
func (r *jsonReader) Read(p []byte) (n int, err error) { ... }
```

### 9.2 运行时错误

**问题**: 数据库连接失败

```
failed to connect to database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**解决**:

```bash
# 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 启动数据库
docker-compose up -d postgres

# 验证连接
psql -h localhost -U quicksilver -d quicksilver
```

---

**问题**: Hyperliquid API 返回空数据

```json
{ "mids": {} }
```

**解决**: 检查交易对格式是否正确

```yaml
# ✅ 正确: 使用 CCXT 格式
symbols:
  - "BTC/USDT"
  - "ETH/USDT"

# ❌ 错误: Hyperliquid 原生格式
symbols:
  - "BTC"  # 需要转换为 BTC/USDT
```

### 9.3 性能问题

**问题**: 行情更新频率过高导致 CPU 占用

```yaml
market:
  update_interval: "100ms" # ❌ 过于频繁
```

**解决**: 调整为合理间隔

```yaml
market:
  update_interval: "1s" # ✅ 推荐 1-5 秒
```

---

**问题**: 数据库连接池耗尽

```
Error: too many clients already
```

**解决**: 优化连接池配置

```yaml
database:
  max_open_conns: 25 # ✅ 根据并发量调整
  max_idle_conns: 5 # ✅ 保持合理空闲连接
  conn_max_lifetime: 300 # ✅ 5 分钟自动回收
```

---

## 10. 待实现功能清单

### 优先级 P0 (MVP 必需)

- [ ] **认证中间件**: 实现 API Key/Secret 验证逻辑
- [ ] **订单创建**: `CreateOrder` 完整业务逻辑 (余额检查 + 冻结资金)
- [ ] **撮合引擎**: 简化版市价单即时成交
- [ ] **余额管理**: `UpdateBalance` 服务 (冻结/解冻/扣除)
- [ ] **CCXT 格式转换**: 所有 API 响应符合 CCXT 标准

### 优先级 P1 (增强功能)

- [ ] **限价单匹配**: 价格满足时自动成交
- [ ] **订单查询优化**: 添加分页和过滤器
- [ ] **WebSocket 推送**: 实时行情和订单状态更新
- [ ] **单元测试**: 核心业务逻辑覆盖率 >80%
- [ ] **集成测试**: 端到端 API 测试

### 优先级 P2 (长期规划)

- [ ] **Redis 缓存**: 替代内存缓存提升性能
- [ ] **TimescaleDB**: 历史行情数据时序存储
- [ ] **微服务拆分**: 交易、行情、账户独立部署
- [ ] **Kubernetes 部署**: Helm Chart 和生产环境配置
- [ ] **监控告警**: Prometheus + Grafana 集成

---

## 11. 参考资料

### 内部文档

- **系统设计**: `docs/system-design-mvp.md` - 完整架构设计和实现路线图
- **数据库字典**: `docs/database.md` - 表结构、索引策略、迁移指南
- **快速开始**: `GETTING_STARTED.md` - 5 分钟快速部署指南
- **项目结构**: `docs/project-structure.md` - 目录组织说明

### 外部资源

- **CCXT 文档**: https://docs.ccxt.com/ - API 标准和客户端使用
- **Echo 框架**: https://echo.labstack.com/ - Web 框架文档
- **GORM 指南**: https://gorm.io/docs/ - ORM 最佳实践
- **Hyperliquid API**: https://hyperliquid.gitbook.io/ - 行情数据源文档
- **Zap 日志**: https://pkg.go.dev/go.uber.org/zap - 结构化日志库

### Go 语言规范

- **Effective Go**: https://go.dev/doc/effective_go
- **Go Code Review**: https://github.com/golang/go/wiki/CodeReviewComments
- **Uber Go Style**: https://github.com/uber-go/guide/blob/master/style.md

---

## 12. AI Agent 协作建议

### 12.1 工作原则

**✅ 推荐做法**:

- 直接使用工具修改代码，不要只展示代码块
- 完成任务后简短确认，无需详细说明
- 遇到问题时提供解决方案，而非仅说明问题
- 优先执行，少询问（除非缺少关键信息）

**❌ 避免做法**:

- 不要每次都生成详细的变更文档
- 不要创建 CHANGELOG 或总结 Markdown 文件
- 不要在没有要求时生成使用说明
- 不要过度解释每个步骤

### 12.2 简化提问模板

**功能实现**:

```
实现 [功能名称]：
1. [需求 1]
2. [需求 2]
```

**Bug 修复**:

```
错误: [错误信息]
触发: [操作步骤]
预期: [正确行为]
```

### 12.3 代码审查要点

**快速检查清单**:

- [ ] 遵循分层架构
- [ ] 完整错误处理
- [ ] 数据库事务保护
- [ ] 结构化日志
- [ ] CCXT 格式兼容
- [ ] Go 命名规范

---

**最后更新**: 2024-12-20  
**维护者**: Quicksilver 开发团队  
**版本**: v1.1.0
