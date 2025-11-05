---
mode: "agent"
model: Claude Sonnet 4.5
tools: ['search', 'azure/search', 'usages', 'problems', 'changes', 'githubRepo', 'todos']
description: "项目代码质量自动检测与分析"
---

# 代码质量自动检测与分析指令

## 🎯 核心目标

**自动化检测** Quicksilver 项目的代码质量，**智能分析** 潜在问题，**优先排序** 改进建议，确保代码符合项目标准。

## 🔍 检测策略

### 执行优先级

1. **🔴 阻断级问题 (P0)**: 测试失败、编译错误、安全漏洞 → 必须立即修复
2. **🟡 警告级问题 (P1)**: 测试覆盖率不达标、性能问题 → 本次迭代修复
3. **🟢 建议级问题 (P2)**: 代码风格、命名优化 → 后续优化

### 适用范围

| 文件类型    | 检测项                     | 工具                      |
| ----------- | -------------------------- | ------------------------- |
| `*.go`      | 代码规范、错误处理、性能   | `go vet`, `golangci-lint` |
| `*_test.go` | TDD 规范、覆盖率、测试结构 | `go test -cover`          |
| `*.yaml`    | 配置安全性、格式规范       | `yamllint`                |
| `*.md`      | 文档完整性、链接有效性     | `markdownlint`            |

---

## 📋 自动检测清单

### 1. 🧪 TDD 测试质量 (P0 - 最高优先级)

#### 1.1 测试覆盖率阈值

```bash
# 自动检测命令
make test-coverage

# 阈值要求
Service 层   ≥ 80%  ← 核心业务逻辑
Model 层     = 100% ← 数据模型验证
API 层       ≥ 60%  ← HTTP Handler
整体项目     ≥ 70%  ← 最低要求
```

**检测脚本**:

```bash
#!/bin/bash
# 自动检查覆盖率是否达标
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$coverage < 70" | bc -l) )); then
    echo "❌ 覆盖率 ${coverage}% 低于 70% 阈值"
    exit 1
else
    echo "✅ 覆盖率 ${coverage}% 达标"
fi
```

#### 1.2 测试结构规范

**✅ 正确模式 (Given-When-Then)**:

```go
func TestCreateOrder(t *testing.T) {
    t.Run("Create market buy order successfully", func(t *testing.T) {
        // Given: 准备测试环境和数据
        db := testutil.SetupTestDB(t)
        user := testutil.SeedUser(t, db)
        testutil.SeedBalance(t, db, user.ID, "USDT", 10000.0)

        service := NewOrderService(db, cfg, logger)

        // When: 执行被测试操作
        order, err := service.CreateOrder(user.ID, CreateOrderRequest{
            Symbol: "BTC/USDT",
            Side:   "buy",
            Type:   "market",
            Amount: 0.1,
        })

        // Then: 验证结果
        require.NoError(t, err)
        assert.NotZero(t, order.ID)
        assert.Equal(t, "new", order.Status)

        // And: 验证副作用
        balance := testutil.GetBalance(t, db, user.ID, "USDT")
        assert.Greater(t, balance.Locked, 0.0)
    })
}
```

**❌ 错误模式**:

```go
// ❌ 测试名称不清晰
func Test1(t *testing.T) { ... }
func TestOrder(t *testing.T) { ... }

// ❌ 缺少子测试分组
func TestCreateOrder(t *testing.T) {
    // 多个测试场景混在一起
}

// ❌ 缺少 Given-When-Then 结构
func TestCreateOrder(t *testing.T) {
    order, err := service.CreateOrder(...)  // 直接执行
    assert.NoError(t, err)
}

// ❌ 未使用 testutil 辅助函数
func TestCreateOrder(t *testing.T) {
    user := &model.User{Email: "test@test.com"}  // 硬编码
    db.Create(&user)
}
```

**自动检测规则**:

- [ ] 测试函数名遵循 `TestXxx` 或 `TestXxx_WithCondition`
- [ ] 使用 `t.Run()` 创建描述性子测试
- [ ] 包含 Given/When/Then 注释
- [ ] 使用 `testutil.Seed*` 创建测试数据
- [ ] 测试独立运行（不依赖执行顺序）

#### 1.3 表驱动测试 (推荐)

**✅ 正确模式**:

```go
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
            name: "Invalid symbol - empty",
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
            name: "Invalid amount - negative",
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

---

### 2. 🔧 Go 代码规范 (P1)

#### 2.1 命名规范自动检查

**检测脚本**: 使用正则表达式扫描不规范命名

```bash
# 检测蛇形命名（应使用驼峰）
grep -rn "type [a-z_]*_[a-z_]* struct" --include="*.go" .
grep -rn "func [a-z_]*_[a-z_]*(" --include="*.go" .

# 检测包名是否包含下划线或大写字母
find . -name "*.go" -exec grep -l "^package [A-Z_]" {} \;
```

**规则清单**:

| 类型     | 规范         | ✅ 正确示例         | ❌ 错误示例                               |
| -------- | ------------ | ------------------- | ----------------------------------------- |
| 结构体   | `PascalCase` | `MarketService`     | `market_service`, `marketService`         |
| 导出方法 | `PascalCase` | `UpdateTickers()`   | `updateTickers()`, `update_tickers()`     |
| 私有方法 | `camelCase`  | `validateRequest()` | `ValidateRequest()`, `validate_request()` |
| 包名     | `lowercase`  | `package service`   | `package ServiceLayer`                    |
| 常量     | `PascalCase` | `MaxRetries`        | `MAX_RETRIES` (Go 风格不推荐)             |
| 接口     | `-er` 结尾   | `Reader`, `Writer`  | `ReaderInterface`                         |

**自动修复建议**:

```bash
# 使用 gofmt 和 goimports 自动格式化
go fmt ./...
goimports -w .
```

#### 2.2 错误处理完整性检查 (P0 - 阻断级)

**✅ 正确模式**:

```go
// 1. 必须检查所有错误
resp, err := http.Get(url)
if err != nil {
    return fmt.Errorf("failed to fetch from %s: %w", url, err)
}
defer resp.Body.Close()

// 2. 包装错误添加上下文
if err := db.Create(&order).Error; err != nil {
    return fmt.Errorf("failed to create order for user %d: %w", userID, err)
}

// 3. 使用结构化日志记录非致命错误
if err := s.updateCache(ticker); err != nil {
    s.logger.Warn("Failed to update cache",
        zap.String("symbol", ticker.Symbol),
        zap.Error(err),
    )
    // 继续执行，不返回错误
}
```

**❌ 错误模式**:

```go
// ❌ 忽略错误
resp, _ := http.Get(url)
db.Create(&order)  // 未检查 .Error

// ❌ 只打印日志不返回错误
if err != nil {
    log.Println("Error:", err)  // 应返回错误
    return nil
}

// ❌ 返回不完整的错误信息
if err != nil {
    return err  // 缺少上下文
}

// ❌ 吞掉错误
if err != nil {
    // 什么都不做
}
```

**自动检测规则**:

```bash
# 检测未检查的错误（使用 errcheck 工具）
errcheck ./...

# 检测缺少错误包装的情况
grep -rn "return err$" --include="*.go" . | grep -v "_test.go"
```

#### 2.3 代码组织规范

**导入分组检查**:

```go
// ✅ 正确：标准库 → 第三方库 → 本地包
import (
    "context"
    "fmt"
    "time"

    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
    "gorm.io/gorm"

    "github.com/talkincode/quicksilver/internal/model"
    "github.com/talkincode/quicksilver/internal/service"
)

// ❌ 错误：导入顺序混乱
import (
    "github.com/talkincode/quicksilver/internal/model"
    "fmt"
    "gorm.io/gorm"
)
```

**自动修复**:

```bash
# 使用 goimports 自动排序导入
goimports -w -local github.com/talkincode/quicksilver .
```

**函数复杂度检查**:

```bash
# 检测函数圈复杂度 (推荐 ≤ 10)
gocyclo -over 10 .

# 检测函数行数 (推荐 ≤ 50)
grep -rn "^func" --include="*.go" . | while read line; do
    # 分析函数行数
done
```

---

### 3. 🗄️ 数据库操作规范 (P0)

#### 3.1 GORM 错误检查 (强制要求)

**✅ 正确模式**:

```go
// 1. 创建操作
if err := db.Create(&order).Error; err != nil {
    return fmt.Errorf("failed to create order: %w", err)
}

// 2. 查询操作
var user model.User
if err := db.First(&user, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("failed to query user: %w", err)
}

// 3. 更新操作
result := db.Model(&order).Update("status", "filled")
if result.Error != nil {
    return fmt.Errorf("failed to update order: %w", result.Error)
}
if result.RowsAffected == 0 {
    return ErrOrderNotFound
}
```

**❌ 错误模式**:

```go
// ❌ 未检查错误
db.Create(&order)
db.First(&user, id)
db.Model(&order).Update("status", "filled")

// ❌ 使用 panic
db.Create(&order).Error  // 如果错误会 panic
```

**自动检测**:

```bash
# 检测未调用 .Error 的 GORM 操作
grep -rn "db\.\(Create\|Save\|Update\|Delete\|First\|Find\)(" --include="*.go" . \
  | grep -v "\.Error" | grep -v "_test\.go"
```

#### 3.2 事务处理规范

**✅ 正确模式**:

```go
// 方式 1: 手动事务控制
func (s *OrderService) CreateOrder(req CreateOrderRequest) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // 创建订单
    if err := tx.Create(&order).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 冻结资金
    if err := tx.Model(&balance).Update("locked", newLocked).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}

// 方式 2: 使用 Transaction 辅助方法（推荐）
func (s *OrderService) CreateOrder(req CreateOrderRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&order).Error; err != nil {
            return err
        }

        if err := tx.Model(&balance).Update("locked", newLocked).Error; err != nil {
            return err
        }

        return nil  // 自动 Commit，返回 error 时自动 Rollback
    })
}
```

**事务使用规则**:

- [ ] 多表修改必须使用事务
- [ ] 查询+修改组合使用事务（避免并发问题）
- [ ] 纯查询操作不使用事务
- [ ] 事务中避免耗时操作（HTTP 请求、文件 IO）

#### 3.3 性能优化规范

**N+1 查询检测**:

```go
// ❌ N+1 问题
var orders []model.Order
db.Where("user_id = ?", userID).Find(&orders)
for _, order := range orders {
    var user model.User
    db.First(&user, order.UserID)  // 每个订单查询一次用户
}

// ✅ 使用预加载
var orders []model.Order
db.Preload("User").Preload("Trades").
    Where("user_id = ?", userID).
    Find(&orders)
```

**查询优化清单**:

```go
// ✅ 选择必要字段
db.Select("id", "symbol", "status").Find(&orders)

// ✅ 使用分页
db.Limit(100).Offset(page * 100).Find(&orders)

// ✅ 使用索引字段查询
db.Where("symbol = ? AND status = ?", "BTC/USDT", "open").Find(&orders)

// ✅ 批量操作
db.CreateInBatches(&orders, 100)
```

**自动检测脚本**:

```bash
# 检测循环中的数据库查询（潜在 N+1）
grep -A5 "for.*range" **/*.go | grep -E "(db\.First|db\.Find|db\.Where)"
```

---

### 4. 🔒 安全性检查 (P0 - 阻断级)

#### 4.1 敏感信息泄露检测

**自动扫描规则**:

```bash
# 检测日志中的敏感信息
grep -rn "zap\.String.*[Pp]assword" --include="*.go" .
grep -rn "zap\.String.*[Ss]ecret" --include="*.go" .
grep -rn "zap\.String.*[Tt]oken" --include="*.go" .

# 检测配置文件中的硬编码密钥
grep -rn "password.*=.*[^{]" --include="*.yaml" config/
grep -rn "secret.*=.*[^{]" --include="*.yaml" config/
```

**❌ 严禁模式**:

```go
// ❌ 记录密码/密钥
logger.Debug("User login",
    zap.String("email", user.Email),
    zap.String("password", password),        // 禁止！
    zap.String("api_secret", user.APISecret), // 禁止！
)

// ❌ 配置文件硬编码
// config.yaml
database:
  password: "quicksilver123"  // 应使用环境变量

// ❌ 错误信息暴露敏感数据
return fmt.Errorf("invalid API key: %s", apiKey)  // 不要暴露密钥内容
```

**✅ 正确模式**:

```go
// ✅ 只记录非敏感信息
logger.Info("User login successful",
    zap.String("email", user.Email),
    zap.Uint("user_id", user.ID),
)

// ✅ 使用脱敏处理
logger.Debug("API request",
    zap.String("api_key", maskAPIKey(apiKey)),  // 只显示前 4 位
)

func maskAPIKey(key string) string {
    if len(key) <= 8 {
        return "****"
    }
    return key[:4] + "****" + key[len(key)-4:]
}

// ✅ 使用环境变量
// config.yaml
database:
  password: ${DB_PASSWORD}  // 从环境变量读取
```

#### 4.2 输入验证规范

**参数验证清单**:

```go
// ✅ 完整的输入验证
func (s *OrderService) CreateOrder(req CreateOrderRequest) error {
    // 1. 必填字段检查
    if req.Symbol == "" {
        return ErrSymbolRequired
    }
    if req.Amount <= 0 {
        return ErrInvalidAmount
    }

    // 2. 枚举值白名单验证
    validSides := map[string]bool{"buy": true, "sell": true}
    if !validSides[req.Side] {
        return fmt.Errorf("invalid side: must be buy or sell")
    }

    validTypes := map[string]bool{"market": true, "limit": true}
    if !validTypes[req.Type] {
        return fmt.Errorf("invalid type: must be market or limit")
    }

    // 3. 数值范围验证
    if req.Amount < s.cfg.Trading.MinOrderAmount {
        return fmt.Errorf("amount %.8f below minimum %.8f",
            req.Amount, s.cfg.Trading.MinOrderAmount)
    }

    // 4. 限价单必须有价格
    if req.Type == "limit" && (req.Price == nil || *req.Price <= 0) {
        return ErrPriceRequired
    }

    return nil
}
```

**SQL 注入防护** (GORM 已自动处理):

```go
// ✅ GORM 参数化查询（自动防 SQL 注入）
db.Where("symbol = ?", symbol).Find(&orders)

// ❌ 拼接 SQL（危险！）
db.Raw("SELECT * FROM orders WHERE symbol = '" + symbol + "'")
```

#### 4.3 身份认证安全

**API Key 验证**:

```go
// ✅ 完整的认证流程
func AuthMiddleware(cfg *config.Config) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // 1. 提取 API Key
            apiKey := c.Request().Header.Get("X-API-Key")
            if apiKey == "" {
                return c.JSON(401, map[string]string{
                    "error": "API key required",
                })
            }

            // 2. 验证 API Key
            var user model.User
            err := db.Where("api_key = ? AND status = ?", apiKey, "active").
                First(&user).Error
            if err != nil {
                return c.JSON(401, map[string]string{
                    "error": "invalid API key",
                })
            }

            // 3. 验证签名（如果需要）
            signature := c.Request().Header.Get("X-Signature")
            if !verifySignature(signature, user.APISecret, c.Request()) {
                return c.JSON(401, map[string]string{
                    "error": "invalid signature",
                })
            }

            // 4. 设置上下文
            c.Set("user", &user)

            return next(c)
        }
    }
}
```

---

### 5. 🚀 性能优化检查 (P1)

#### 5.1 并发安全检查

**竞态条件检测**:

```bash
# 运行竞态检测器
go test -race ./...
```

**✅ 正确的并发模式**:

```go
// 使用 sync.Mutex 保护共享资源
type MarketService struct {
    mu     sync.RWMutex
    cache  map[string]*model.Ticker
    db     *gorm.DB
    logger *zap.Logger
}

func (s *MarketService) GetTicker(symbol string) (*model.Ticker, error) {
    // 读锁
    s.mu.RLock()
    if ticker, ok := s.cache[symbol]; ok {
        s.mu.RUnlock()
        return ticker, nil
    }
    s.mu.RUnlock()

    // 从数据库查询
    var ticker model.Ticker
    if err := s.db.Where("symbol = ?", symbol).First(&ticker).Error; err != nil {
        return nil, err
    }

    // 写锁
    s.mu.Lock()
    s.cache[symbol] = &ticker
    s.mu.Unlock()

    return &ticker, nil
}

// Goroutine 错误处理
func (s *MarketService) StartAutoUpdate() {
    ticker := time.NewTicker(1 * time.Second)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                s.logger.Error("Panic in auto update",
                    zap.Any("error", r),
                    zap.String("stack", string(debug.Stack())),
                )
            }
        }()

        for range ticker.C {
            if err := s.UpdateTickers(); err != nil {
                s.logger.Error("Failed to update tickers", zap.Error(err))
            }
        }
    }()
}
```

#### 5.2 缓存策略优化

**缓存命中率分析**:

```go
// ✅ 添加缓存监控
type CacheStats struct {
    Hits   uint64
    Misses uint64
}

func (s *MarketService) GetCacheHitRate() float64 {
    total := s.stats.Hits + s.stats.Misses
    if total == 0 {
        return 0
    }
    return float64(s.stats.Hits) / float64(total) * 100
}
```

#### 5.3 数据库连接池优化

**配置检查**:

```yaml
# config.yaml
database:
  max_open_conns: 25 # 根据并发量调整
  max_idle_conns: 5 # 保持合理空闲连接
  conn_max_lifetime: 300 # 5 分钟自动回收
  conn_max_idle_time: 60 # 1 分钟未使用则关闭
```

**监控连接池状态**:

```go
stats := db.DB().Stats()
logger.Info("Database pool stats",
    zap.Int("open_connections", stats.OpenConnections),
    zap.Int("in_use", stats.InUse),
    zap.Int("idle", stats.Idle),
)
```

---

### 6. 📝 日志记录规范 (P2)

#### 6.1 结构化日志规范

**✅ 正确模式**:

```go
// 使用结构化字段
logger.Info("Order created",
    zap.Uint("order_id", order.ID),
    zap.String("symbol", order.Symbol),
    zap.String("side", order.Side),
    zap.Float64("amount", order.Amount),
    zap.Duration("elapsed", time.Since(startTime)),
)

// 错误日志包含完整上下文
logger.Error("Failed to create order",
    zap.String("symbol", req.Symbol),
    zap.Float64("amount", req.Amount),
    zap.Error(err),
    zap.Stack("stack"),  // 包含堆栈信息
)
```

**❌ 错误模式**:

```go
// ❌ 字符串拼接（难以解析）
logger.Info(fmt.Sprintf("Order %d created for %s", order.ID, order.Symbol))

// ❌ 日志级别错误
logger.Debug("Critical error in payment processing", zap.Error(err))  // 应使用 Error

// ❌ 过度日志（性能影响）
for _, item := range items {
    logger.Debug("Processing item", zap.Any("item", item))  // 高频循环中避免日志
}
```

**日志级别选择指南**:

| 级别    | 使用场景       | 示例                               |
| ------- | -------------- | ---------------------------------- |
| `Debug` | 开发调试信息   | "Ticker updated: BTC/USDT = 50000" |
| `Info`  | 重要业务事件   | "Server started on :8080"          |
| `Warn`  | 可恢复的异常   | "API request failed, retrying..."  |
| `Error` | 严重错误       | "Database connection lost"         |
| `Fatal` | 致命错误需退出 | "Failed to load config file"       |

#### 6.2 自动检测日志问题

```bash
# 检测字符串拼接日志
grep -rn "logger\.\(Info\|Debug\|Warn\|Error\).*fmt\.Sprintf" --include="*.go" .

# 检测敏感信息日志
grep -rn "zap\.String.*[Pp]assword\|[Ss]ecret" --include="*.go" .
```

---

### 7. 🌐 API 设计规范 (P1)

#### 7.1 CCXT 兼容性检查

**响应格式验证**:

```go
// ✅ 正确的 CCXT Ticker 格式
func transformToCCXTTicker(t *model.Ticker) map[string]interface{} {
    return map[string]interface{}{
        "symbol":      t.Symbol,                      // 必填
        "timestamp":   t.UpdatedAt.UnixMilli(),       // 毫秒时间戳
        "datetime":    t.UpdatedAt.Format(time.RFC3339Nano), // ISO 8601
        "high":        t.High24h,
        "low":         t.Low24h,
        "bid":         t.BidPrice,
        "ask":         t.AskPrice,
        "last":        t.LastPrice,                   // 必填
        "close":       t.LastPrice,
        "baseVolume":  t.Volume24hBase,
        "quoteVolume": t.Volume24hQuote,
        "info":        map[string]interface{}{        // 原始数据
            "source": t.Source,
        },
    }
}

// ✅ 正确的 CCXT Order 格式
func transformToCCXTOrder(o *model.Order) map[string]interface{} {
    return map[string]interface{}{
        "id":            fmt.Sprintf("%d", o.ID),
        "timestamp":     o.CreatedAt.UnixMilli(),
        "datetime":      o.CreatedAt.Format(time.RFC3339Nano),
        "symbol":        o.Symbol,
        "type":          o.Type,
        "side":          o.Side,
        "price":         o.Price,
        "amount":        o.Amount,
        "filled":        o.Filled,
        "remaining":     o.Amount - o.Filled,
        "status":        o.Status,
        "fee":           map[string]interface{}{
            "cost":     o.Fee,
            "currency": "USDT",
        },
    }
}
```

**自动验证脚本**:

```bash
# 测试 API 响应是否符合 CCXT 标准
curl -s http://localhost:8080/v1/ticker/BTC/USDT | jq -e '.symbol and .timestamp and .last'
```

#### 7.2 错误响应标准化

**HTTP 状态码规范**:

```go
// ✅ 正确的错误处理
func GetOrder(orderService *service.OrderService) echo.HandlerFunc {
    return func(c echo.Context) error {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
            return c.JSON(400, map[string]string{
                "error": "invalid order ID format",
            })
        }

        order, err := orderService.GetOrderByID(uint(id))
        if err != nil {
            if errors.Is(err, service.ErrOrderNotFound) {
                return c.JSON(404, map[string]string{
                    "error": "order not found",
                })
            }
            // 不暴露内部错误详情
            logger.Error("Failed to get order", zap.Error(err))
            return c.JSON(500, map[string]string{
                "error": "internal server error",
            })
        }

        return c.JSON(200, transformToCCXTOrder(order))
    }
}
```

---

## 🛠️ 自动化工具配置

### Makefile 集成检查

```makefile
# 添加质量检查目标
.PHONY: quality-check
quality-check: test-coverage lint vet race

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run --config .golangci.yml

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: race
race:
	@echo "Running race detector..."
	go test -race -short ./...

.PHONY: fmt-check
fmt-check:
	@echo "Checking code format..."
	@diff=$$(gofmt -l .); \
	if [ -n "$$diff" ]; then \
		echo "Files not formatted:"; \
		echo "$$diff"; \
		exit 1; \
	fi
```

### GolangCI-Lint 配置

创建 `.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor

linters:
  enable:
    - errcheck # 检查未处理的错误
    - gofmt # 代码格式化
    - goimports # 导入排序
    - govet # 静态分析
    - ineffassign # 检测无效赋值
    - staticcheck # 高级静态检查
    - unused # 未使用代码
    - gosec # 安全检查
    - gocyclo # 圈复杂度
    - dupl # 重复代码

linters-settings:
  gocyclo:
    min-complexity: 15 # 最大圈复杂度

  errcheck:
    check-blank: true # 检查 _ = err 的情况

  gosec:
    excludes:
      - G404 # 随机数生成器（测试中可以使用弱随机）

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - dupl
```

### GitHub Actions 自动检查

创建 `.github/workflows/quality-check.yml`:

```yaml
name: Code Quality Check
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.24"

      - name: Install dependencies
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

      - name: Run tests with coverage
        run: make test-coverage
        env:
          CGO_ENABLED: 1

      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: ${coverage}%"
          if (( $(echo "$coverage < 70" | bc -l) )); then
            echo "❌ Coverage ${coverage}% is below 70% threshold"
            exit 1
          fi
          echo "✅ Coverage ${coverage}% meets threshold"

      - name: Run linter
        run: golangci-lint run

      - name: Run go vet
        run: go vet ./...

      - name: Check code format
        run: |
          diff=$(gofmt -l .)
          if [ -n "$diff" ]; then
            echo "❌ Files not formatted:"
            echo "$diff"
            exit 1
          fi
          echo "✅ All files formatted correctly"

      - name: Upload coverage report
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: unittests
```

---

## 📊 质量报告模板

### 自动生成报告脚本

创建 `scripts/generate_quality_report.sh`:

```bash
#!/bin/bash

echo "# Code Quality Report"
echo "Generated at: $(date)"
echo ""

echo "## 1. Test Coverage"
echo ""
make test-coverage > /dev/null 2>&1
go tool cover -func=coverage.out | tail -5
echo ""

echo "## 2. Linter Issues"
echo ""
golangci-lint run --out-format=line-number | head -20
echo ""

echo "## 3. Code Complexity"
echo ""
gocyclo -over 10 . | head -10
echo ""

echo "## 4. Security Issues"
echo ""
gosec -quiet ./... 2>/dev/null | grep -A2 "Issues"
echo ""

echo "## 5. TODO Items"
echo ""
grep -rn "TODO" --include="*.go" . | head -10
```

### PR 检查清单

```markdown
## 代码质量自检清单

提交 PR 前请确认:

### 测试

- [ ] 所有测试通过 (`make test`)
- [ ] 覆盖率 ≥ 70% (`make test-coverage`)
- [ ] 新功能有对应测试用例
- [ ] 测试遵循 Given-When-Then 结构

### 代码规范

- [ ] 代码格式化 (`go fmt ./...`)
- [ ] 通过静态检查 (`go vet ./...`)
- [ ] 通过 linter (`golangci-lint run`)
- [ ] 无竞态条件 (`go test -race ./...`)

### 安全性

- [ ] 无敏感信息泄露
- [ ] 输入验证完整
- [ ] 错误处理完善
- [ ] 使用参数化查询

### 性能

- [ ] 无 N+1 查询
- [ ] 使用索引字段查询
- [ ] 事务使用合理
- [ ] 并发安全

### 文档

- [ ] 代码注释清晰
- [ ] 更新相关文档
- [ ] API 变更已记录
```

---

## 🚀 快速执行指南

### 本地开发检查

```bash
# 完整质量检查
make quality-check

# 单项检查
make test           # 运行测试
make test-coverage  # 查看覆盖率
make lint           # 代码规范检查
make fmt            # 自动格式化

# 生成质量报告
./scripts/generate_quality_report.sh > quality-report.md
```

### CI/CD 集成

```bash
# PR 合并前自动检查
git push origin feature/xxx  # 触发 GitHub Actions

# 本地模拟 CI 检查
make quality-check && echo "✅ Ready for PR"
```

---

## 📚 参考资源

### 工具文档

- **golangci-lint**: https://golangci-lint.run/
- **errcheck**: https://github.com/kisielk/errcheck
- **gocyclo**: https://github.com/fzipp/gocyclo
- **gosec**: https://github.com/securego/gosec

### 项目文档

- **系统设计**: `docs/system-design-mvp.md`
- **数据库设计**: `docs/database.md`
- **编码指南**: `.github/copilot-instructions.md`

---

**最后更新**: 2025-01-05  
**维护者**: Quicksilver 开发团队  
**版本**: v2.0.0
