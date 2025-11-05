# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
本文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 项目概述

**Quicksilver** 是一个 CCXT 兼容的模拟加密货币交易所，专为量化交易策略测试而构建。它提供来自 Hyperliquid/Binance 的真实市场数据，以及完整的交易引擎，包括订单撮合、余额管理和账户管理。

**技术栈**: Go 1.24+ + Echo v4.13.4 + GORM v1.31.0 + PostgreSQL 16+
**架构**: 单体分层架构 (MVP)，设计为可演进为微服务
**开发**: 测试驱动开发 (TDD)，具有全面的测试要求

## 核心命令

### 开发工作流

```bash
# 启动开发服务器（热重载）
make dev                    # 使用 air 热重载
go run cmd/server/main.go  # 直接运行

# 数据库操作
make db-migrate            # 运行迁移
make db-seed               # 种子测试数据
make db-seed-test-user     # 为 API 测试创建测试用户
```

### 测试命令（TDD 必需）

```bash
# 核心测试（频繁运行）
make test                  # 运行所有测试和覆盖率
make test-unit            # 仅单元测试（快速）
make test-coverage        # 生成 coverage.html 报告
make test-watch           # 监听文件变化自动测试

# CCXT 兼容性测试（API 验证的关键）
make test-ccxt            # Python CCXT 测试
make test-ccxt-js         # Node.js CCXT 测试
make test-ccxt-setup-python  # 安装 Python 测试依赖
make test-ccxt-setup-nodejs  # 安装 Node.js 测试依赖

# 质量检查
make fmt                  # 格式化代码
make lint                 # Golangci-lint 代码分析
make quality-check        # 完整质量管道
```

### 构建和部署

```bash
# 构建应用
make build                # 构建到 bin/quicksilver
make all                  # 完整构建管道

# Docker 操作
make docker-build         # 构建 Docker 镜像
make docker-up            # 使用 docker-compose 启动
make docker-down          # 停止服务

# 仪表盘（Python Streamlit）
cd dashboard && ./start.sh  # 启动管理仪表盘
```

## 架构与分层设计

### 层级结构

```
API 层 (internal/api)      ← HTTP 处理器，CCXT 格式转换
Service 层 (internal/service) ← 业务逻辑，撮合引擎，市场数据同步
Repository 层 (internal/repository) ← 数据访问抽象
Model 层 (internal/model) ← GORM 模型，数据库映射
```

**关键规则**:

- **禁止跨层调用**: API 层不能直接访问 Model，必须通过 Service 层
- **依赖注入**: 所有服务通过构造函数接收依赖
- **错误传播**: 使用 `fmt.Errorf("context: %w", err)` 包装错误并提供上下文

### 核心服务

- **MarketService**: 每秒从 Hyperliquid API 同步行情数据
- **OrderService**: 订单创建、验证、余额管理（未完全实现）
- **MatchingService**: 订单执行和成交生成（计划中）

### 数据库与模型

- **自动迁移**: 启动时运行以创建/更新表
- **GORM**: 与 PostgreSQL 一起使用，支持 SQLite 用于测试
- **关键模型**: User, Balance, Order, Trade, Ticker
- **事务支持**: 订单操作的关键

## TDD 开发流程（强制要求）

本项目强制执行测试驱动开发：

1. **红阶段**: 先编写失败的测试
2. **绿阶段**: 编写最少代码使测试通过
3. **重构阶段**: 提高代码质量

### 测试要求

- **Service 层**: ≥ 80% 覆盖率（核心业务逻辑）
- **Model 层**: 100% 覆盖率（数据验证）
- **API 层**: ≥ 60% 覆盖率（HTTP 处理器）
- **整体项目**: ≥ 70% 覆盖率

### 测试结构

```go
// 使用 Given-When-Then 结构
func TestCreateOrder(t *testing.T) {
    t.Run("Create market buy order", func(t *testing.T) {
        // Given: 用户有足够余额
        db := testutil.SetupTestDB(t)
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
        assert.Equal(t, "new", order.Status)
    })
}
```

## CCXT 兼容性要求

### API 响应格式

所有 API 响应必须符合 CCXT 标准：

```go
// Ticker 格式示例
{
    "symbol": "BTC/USDT",
    "timestamp": 1703001234567,
    "datetime": "2023-12-20T12:34:56.789Z",
    "last": 109965.50,
    "bid": 109960.00,
    "ask": 109970.00,
    // ... 其他 CCXT 标准字段
}
```

### 关键 API 端点

| CCXT 方法 | HTTP 端点 | 状态 |
|-------------|---------------|--------|
| `fetchTicker(symbol)` | `GET /v1/ticker/:symbol` | ✅ 已实现 |
| `fetchBalance()` | `GET /v1/balance` | ⏳ 待实现 |
| `createOrder(...)` | `POST /v1/order` | ⏳ 待实现 |
| `cancelOrder(id)` | `DELETE /v1/order/:id` | ⏳ 待实现 |
| `fetchOrders()` | `GET /v1/orders` | ⏳ 待实现 |

## 配置管理

配置使用 Viper 支持 YAML 文件和环境变量：

```yaml
# config/config.yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: quicksilver

market:
  update_interval: 1s
  data_source: hyperliquid
  symbols:
    - "BTC/USDT"
    - "ETH/USDT"
```

**优先级**: 环境变量 > 配置文件 > 默认值

## 开发环境设置

### 前置要求

- Go 1.24+
- PostgreSQL 16+
- Docker & Docker Compose（可选）
- Python 3.8+（用于仪表盘）
- Node.js 18+（用于 CCXT 测试）

### 快速开始

```bash
# 1. 克隆并安装依赖
git clone https://github.com/talkincode/quicksilver.git
cd quicksilver
go mod download

# 2. 配置环境
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 配置数据库连接

# 3. 启动数据库（Docker）
docker-compose up -d postgres

# 4. 运行迁移和种子数据
make db-migrate
make db-seed-test-user

# 5. 启动开发服务器
make dev

# 6. 测试 API（在另一个终端）
make test-api
```

## 代码标准与模式

### Go 约定

- **命名**: 导出用 PascalCase，私有用 camelCase
- **错误处理**: 总是用上下文包装错误
- **日志记录**: 使用 zap 结构化日志
- **指针**: 状态修改使用指针接收器

### GORM 模式

```go
// 带正确标签的模型定义
type Order struct {
    ID       uint      `gorm:"primaryKey" json:"id"`
    UserID   uint      `gorm:"not null;index" json:"user_id"`
    Symbol   string    `gorm:"size:20;not null;index" json:"symbol"`
    // 可选字段使用指针
    Price    *float64  `gorm:"type:decimal(20,8)" json:"price,omitempty"`
}

// UPSERT 模式
ticker := model.Ticker{Symbol: "BTC/USDT", LastPrice: 50000.0}
db.Save(&ticker)  // 存在则更新，不存在则插入
```

### Echo 处理器模式

```go
func GetTicker(db *gorm.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        symbol := c.Param("symbol")
        var ticker model.Ticker
        if err := db.Where("symbol = ?", symbol).First(&ticker).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return c.JSON(404, map[string]string{"error": "ticker not found"})
            }
            return c.JSON(500, map[string]string{"error": "internal server error"})
        }
        return c.JSON(200, transformToCCXTTicker(&ticker))
    }
}
```

## 测试策略

### 单元测试

- 使用 SQLite 内存数据库通过 `testutil.SetupTestDB()`
- 模拟外部依赖（HTTP 客户端、API）
- 多场景的表驱动测试
- 目标：快速执行，隔离测试

### 集成测试

- 使用真实 PostgreSQL 数据库测试
- 测试外部 API 集成（Hyperliquid）
- 端到端工作流验证
- 使用 `testing.Short()` 在 CI 中跳过

### CCXT 兼容性测试

验证 API 合规性的关键：

```bash
# 开发期间定期运行 CCXT 测试
make test-ccxt          # Python 版本
make test-ccxt-js       # Node.js 版本

# 这些测试验证：
# - 公开 API 访问（行情、成交）
# - 私有 API 访问（余额、订单）
# - 认证机制
# - 响应格式合规性
```

## 仪表盘（管理界面）

项目包含一个 Python Streamlit 仪表盘用于管理：

```bash
cd dashboard
./start.sh  # 在 http://localhost:8501 启动
```

功能：

- 用户管理
- 订单监督
- 成交历史
- 余额监控

## 常见问题与解决方案

### 数据库连接

```bash
# 检查 PostgreSQL 状态
docker ps | grep postgres

# 重置数据库
make db-reset  # 危险：删除所有表
```

### 测试失败

```bash
# 覆盖率阈值失败
make test-coverage && open coverage.html  # 查看详细报告

# 集成测试失败
make db-seed-test-user  # 确保测试数据存在
```

### CCXT 测试问题

```bash
# 安装依赖
make test-ccxt-setup-python
make test-ccxt-setup-nodejs

# 检查 API 凭证
psql -h localhost -U postgres -d quicksilver
SELECT * FROM users WHERE api_key = 'qs-test-api-key-2024';
```

## 需要理解的关键文件

- `cmd/server/main.go` - 应用程序入口点
- `internal/service/market.go` - 市场数据同步
- `internal/config/config.go` - 配置管理
- `Makefile` - 构建和开发命令
- `docs/system-design-mvp.md` - 完整架构文档
- `.github/copilot-instructions.md` - 详细开发指南

## 优先实现的功能

1. **认证中间件** - API Key/Secret 验证
2. **订单创建** - 带余额检查的完整业务逻辑
3. **撮合引擎** - 简化的市价单执行
4. **余额管理** - 冻结/解冻/更新余额
5. **CCXT 格式转换** - 所有 API 响应

实现新功能时，始终遵循 TDD 方法：先写测试，实现最小功能，然后重构提高质量。