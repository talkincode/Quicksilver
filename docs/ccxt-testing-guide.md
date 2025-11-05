# CCXT 客户端测试脚本使用指南

本文档说明如何使用 `scripts/test_ccxt_client.py` 测试 Quicksilver 与 CCXT 的兼容性。

## 快速开始

### 1. 安装依赖

```bash
# 安装 CCXT 库
pip install ccxt

# 或使用虚拟环境
python3 -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate  # Windows
pip install ccxt
```

### 2. 启动 Quicksilver 服务

```bash
# 启动数据库
docker-compose up -d postgres

# 启动服务
make run
# 或
make dev
```

### 3. 准备测试用户

```bash
# 创建测试用户并初始化余额
./scripts/init_test_user.sh
```

这将创建一个测试用户并显示：

```
API Key: qs_test_1234567890abcdef
API Secret: secret_1234567890abcdef1234567890abcdef
```

### 4. 运行测试

#### 仅测试公开 API（无需认证）

```bash
python scripts/test_ccxt_client.py
```

#### 测试完整功能（包括私有 API）

```bash
python scripts/test_ccxt_client.py \
  --api-key "qs_test_1234567890abcdef" \
  --api-secret "secret_1234567890abcdef1234567890abcdef"
```

#### 测试远程服务器

```bash
python scripts/test_ccxt_client.py \
  --url "https://your-quicksilver-instance.com" \
  --api-key "YOUR_API_KEY" \
  --api-secret "YOUR_API_SECRET"
```

## 测试覆盖范围

### 公开 API（无需认证）

| 端点                     | CCXT 方法       | 测试内容            |
| ------------------------ | --------------- | ------------------- |
| `GET /v1/time`           | `publicGetTime` | 服务器时间          |
| `GET /v1/markets`        | `fetchMarkets`  | 交易对列表          |
| `GET /v1/ticker/:symbol` | `fetchTicker`   | 行情数据（BTC/ETH） |
| `GET /v1/trades/:symbol` | `fetchTrades`   | 最近成交记录        |

### 私有 API（需要认证）

| 端点                   | CCXT 方法         | 测试内容              |
| ---------------------- | ----------------- | --------------------- |
| `GET /v1/balance`      | `fetchBalance`    | 账户余额              |
| `POST /v1/order`       | `createOrder`     | 创建订单（限价/市价） |
| `GET /v1/order/:id`    | `fetchOrder`      | 查询单个订单          |
| `DELETE /v1/order/:id` | `cancelOrder`     | 撤销订单              |
| `GET /v1/orders`       | `fetchOrders`     | 查询所有订单          |
| `GET /v1/orders/open`  | `fetchOpenOrders` | 查询未完成订单        |
| `GET /v1/myTrades`     | `fetchMyTrades`   | 查询我的成交记录      |

## 示例输出

### 成功示例

```
============================================================
  Quicksilver CCXT 兼容性测试
============================================================

📂 Public API Tests
------------------------------------------------------------
🔍 Testing: Server Time
✅ PASS | GET /v1/time
  Server Time: 2025-11-05T13:45:30.123Z

🔍 Testing: Fetch Markets
✅ PASS | GET /v1/markets
  Total Markets: 2
  Sample: BTC/USDT

🔍 Testing: Fetch Ticker (BTC/USDT)
✅ PASS | GET /v1/ticker/BTC/USDT
  Last Price: 109965.5
  24h Volume: 123.45

...

📂 Private API Tests (Authenticated)
------------------------------------------------------------
🔍 Testing: Fetch Balance
✅ PASS | GET /v1/balance
  Assets: 2

🔍 Testing: Create Order (buy limit)
✅ PASS | POST /v1/order
  Order ID: 123
  Status: new

...

============================================================
  测试结果汇总
============================================================
✅ Passed: 12
❌ Failed: 0
📊 Success Rate: 100.0%
============================================================
```

### 失败示例

```
🔍 Testing: Fetch Ticker (BTC/USDT)
❌ FAIL | GET /v1/ticker/BTC/USDT
  └─ Error: 缺少字段: baseVolume

============================================================
  测试结果汇总
============================================================
✅ Passed: 8
❌ Failed: 1
📊 Success Rate: 88.9%

❌ Failed Tests:
  - GET /v1/ticker/BTC/USDT: 缺少字段: baseVolume
============================================================
```

## 故障排查

### 问题 1: 连接失败

```
Error: [Errno 61] Connection refused
```

**解决方案**:

- 确认 Quicksilver 服务正在运行: `curl http://localhost:8080/health`
- 检查端口是否正确: 默认 8080

### 问题 2: 认证失败

```
Error: user not authenticated
```

**解决方案**:

- 确认 API Key/Secret 正确
- 检查测试用户是否存在: `psql -h localhost -U quicksilver -d quicksilver -c "SELECT * FROM users;"`
- 重新创建测试用户: `./scripts/init_test_user.sh`

### 问题 3: 数据为空

```
Total Trades: 0
```

**解决方案**:

- 确认行情数据已同步: `curl http://localhost:8080/v1/ticker/BTC-USDT`
- 检查撮合引擎是否运行
- 手动创建订单触发成交

### 问题 4: 格式不兼容

```
Error: 缺少字段: baseVolume
```

**解决方案**:

- 检查 `internal/ccxt/transformer.go` 的格式转换逻辑
- 运行单元测试: `make test-unit`
- 查看 CCXT 标准文档: https://docs.ccxt.com/

## 集成到 CI/CD

### GitHub Actions 示例

```yaml
name: CCXT Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: quicksilver
          POSTGRES_PASSWORD: quicksilver123
          POSTGRES_DB: quicksilver
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.24"

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: "3.11"

      - name: Install CCXT
        run: pip install ccxt

      - name: Build Quicksilver
        run: make build

      - name: Start Quicksilver
        run: |
          ./bin/quicksilver &
          sleep 5

      - name: Create test user
        run: ./scripts/init_test_user.sh

      - name: Run CCXT tests
        run: |
          python scripts/test_ccxt_client.py \
            --api-key "$TEST_API_KEY" \
            --api-secret "$TEST_API_SECRET"
```

## 扩展测试

### 添加新的测试用例

在 `QuicksilverTester` 类中添加新方法：

```python
def test_your_feature(self):
    """测试新功能"""
    print("🔍 Testing: Your Feature")
    try:
        response = self.exchange.yourApiMethod()

        # 验证逻辑
        assert 'field' in response, "缺少字段"

        self.log_test("Your Test", True)
        return True
    except Exception as e:
        self.log_test("Your Test", False, str(e))
        return False
```

然后在 `run_all_tests()` 中调用：

```python
def run_all_tests(self):
    # ...现有测试...
    self.test_your_feature()
```

## 相关文档

- CCXT 官方文档: https://docs.ccxt.com/
- Quicksilver API 文档: `docs/api-reference.md`
- 系统设计文档: `docs/system-design-mvp.md`
