# 开发计划 - 测试覆盖率提升与 K 线功能 (Week 45, 2025)

> **计划制定时间**: 2025-11-05  
> **计划执行周期**: 本周 (Week 45: 11 月 5 日 - 11 月 11 日)  
> **下次评审时间**: 2025-11-11 17:00  
> **规划依据**: `.github/prompts/planner.prompt.md`

---

## 📊 项目状态诊断

### 整体健康度 ⭐⭐⭐⭐⭐ (5/5)

**当前进度**: 97% 完成 ✅  
**测试覆盖率**: 62.0% (目标: 70%, 差距: 8%) ⚠️  
**关键阻塞**: 无阻塞性问题 ✅  
**编译状态**: 通过 ✅  
**代码质量**: 通过 go vet, go fmt ✅  
**待办事项**: 2 个 TODO (仅数据源扩展相关)

### 本周已完成 (2025-11-05) ✅

| 功能           | 状态 | 覆盖率 | 测试用例数 |
| -------------- | ---- | ------ | ---------- |
| K 线数据服务   | ✅   | ~85%   | 13 个      |
| K 线 API 端点  | ✅   | 已实现 | -          |
| 撮合引擎       | ✅   | 73.3%  | 8 个       |
| CCXT 兼容层    | ✅   | 100%   | 10 个      |
| 管理员 API     | ✅   | 82.2%  | 5 个       |
| 认证中间件     | ✅   | 95.7%  | 7 个       |
| 用户管理服务   | ✅   | 74.6%  | 6 个       |
| 余额管理服务   | ✅   | ~75%   | 12 个      |
| ListUsers 测试 | ✅   | 100%   | 8 个子测试 |

### 关键成就 🎉

1. **✅ K 线功能完整实现**

   - 新增 `internal/service/kline.go` 及完整测试
   - 支持 1m/5m/15m/1h/4h/1d/1w 多个时间周期
   - 实现数据聚合、历史查询、实时更新
   - API 文档: `docs/kline-api.md`

2. **✅ ListUsers 测试补充**

   - 从 0% → 100% 覆盖率
   - 8 个完整子测试场景
   - 测试通过率 100%

3. **✅ 数据模型优化**
   - 添加 Kline 表结构
   - 优化索引设计 (symbol, interval, start_time)

### 待解决问题 ⚠️

| 问题             | 当前值 | 目标值 | 优先级 | 预计工时 |
| ---------------- | ------ | ------ | ------ | -------- |
| 整体测试覆盖率   | 62.0%  | 72%+   | P1     | 3h       |
| Config 层覆盖率  | 6.7%   | 70%+   | P1     | 1.5h     |
| Streamlit 仪表盘 | 80%    | 100%   | P1     | 2h       |
| Binance 数据源   | 0%     | 100%   | P2     | 3h       |

---

## 🎯 本周目标 (SMART 原则)

### 主要里程碑

1. **Specific (具体)**:

   - 提升测试覆盖率至 72%+
   - 完成 Streamlit 管理仪表盘
   - 补充 Config 层测试
   - 运行 CCXT 客户端集成测试

2. **Measurable (可衡量)**:

   - 覆盖率: 62.0% → 72%+ (⬆️ 10%)
   - Config 层: 6.7% → 70%+ (⬆️ 63.3%)
   - 仪表盘完成度: 80% → 100% (⬆️ 20%)

3. **Achievable (可实现)**:

   - 预估总工时: 6.5 小时
   - 本周可用时间: 20 小时
   - 缓冲余量: 67.5% ✅

4. **Relevant (相关)**:

   - 满足 MVP 质量标准
   - 为生产部署做准备
   - 提升系统可靠性

5. **Time-bound (有时限)**:
   - 完成日期: 2025-11-11
   - 每日进度检查
   - 周五项目评审

### 成功指标

| 指标             | 当前值 | 目标值   | 差距     | 达成难度 |
| ---------------- | ------ | -------- | -------- | -------- |
| 整体覆盖率       | 62.0%  | **72%+** | ⬆️ 10%   | 🟢 简单  |
| Config 层覆盖率  | 6.7%   | **70%+** | ⬆️ 63.3% | 🟡 中等  |
| Streamlit 仪表盘 | 80%    | **100%** | ⬆️ 20%   | 🟢 简单  |
| CCXT 测试通过率  | 0%     | **100%** | ⬆️ 100%  | 🟢 简单  |
| 测试通过率       | 100%   | **100%** | ✅ 保持  | 🟢 简单  |

---

## 📝 任务列表 (优先级排序)

### P0 - 必须完成 (阻塞性)

✅ **无 P0 阻塞性问题** - 所有核心功能已实现且测试通过

---

### P1 - 高优先级 (本周内必须完成)

#### 任务 #1: 补充 Config 层测试用例

**优先级**: P1 - 高优先级  
**预估**: 1.5 小时  
**状态**: ⏳ 待开始

**目标**: Config 层覆盖率从 6.7% 提升至 70%+

**价值**:

- 验证配置加载和验证逻辑健壮性
- 防止配置错误导致生产环境故障
- 提升整体覆盖率约 3-5%
- 消除最大的测试缺口

**验收标准**:

- [ ] TestLoadConfig 包含至少 5 个子测试
- [ ] 测试 YAML 配置文件加载
- [ ] 测试环境变量覆盖机制 (QS_SERVER_PORT)
- [ ] 测试配置验证规则
- [ ] 测试缺失配置的默认值
- [ ] 测试无效配置的错误处理
- [ ] 测试 GetDSN() 等辅助方法
- [ ] Config 层覆盖率 ≥ 70%

**技术要点**:

- 使用临时文件测试配置加载
- 使用 `t.Setenv()` 测试环境变量
- 验证 Viper 自动环境变量映射
- 确保测试之间隔离 (cleanup)

**前置依赖**: 无

**测试策略**:

- 单元测试: 覆盖所有配置加载路径
- 集成测试: 验证完整配置解析流程

**执行步骤**:

1. [ ] 分析 config.go 代码结构 (15 分钟)
   - 识别未覆盖的函数和分支
   - 列出需要测试的场景
2. [ ] 编写配置加载测试 (30 分钟)
   - 创建临时 YAML 文件
   - 测试正常加载流程
   - 测试文件缺失场景
   - 测试无效 YAML 格式
3. [ ] 编写环境变量测试 (20 分钟)
   - 测试 QS\_ 前缀变量
   - 测试变量覆盖优先级
   - 验证自动类型转换
4. [ ] 编写验证和错误处理测试 (20 分钟)
   - 测试无效端口号
   - 测试无效数据库配置
   - 测试缺失必填字段
5. [ ] 运行测试并验证覆盖率 (15 分钟)
   - `make test-coverage`
   - 查看 coverage.html 确认 Config 层覆盖率

**测试用例模板**:

```go
func TestLoadConfig(t *testing.T) {
    t.Run("Load valid config file", func(t *testing.T) {
        // Given: 临时 config.yaml
        tmpFile := createTempConfig(t, validYAML)

        // When: 加载配置
        cfg, err := LoadConfigFrom(tmpFile)

        // Then: 加载成功
        require.NoError(t, err)
        assert.Equal(t, 8080, cfg.Server.Port)
    })

    t.Run("Load config with env override", func(t *testing.T) {
        // Given: 环境变量 QS_SERVER_PORT=9090
        t.Setenv("QS_SERVER_PORT", "9090")

        // When: 加载配置
        cfg, err := Load()

        // Then: 端口被环境变量覆盖
        assert.Equal(t, 9090, cfg.Server.Port)
    })

    t.Run("Return error on missing file", func(t *testing.T) {
        // When: 配置文件不存在
        _, err := LoadConfigFrom("/nonexistent/config.yaml")

        // Then: 返回错误
        require.Error(t, err)
        assert.Contains(t, err.Error(), "failed to read config")
    })

    t.Run("Return error on invalid YAML", func(t *testing.T) {
        // Given: 无效 YAML
        tmpFile := createTempConfig(t, "invalid: yaml: :")

        // When: 加载配置
        _, err := LoadConfigFrom(tmpFile)

        // Then: 返回解析错误
        require.Error(t, err)
    })
}

func TestDatabaseConfigGetDSN(t *testing.T) {
    t.Run("Generate correct DSN string", func(t *testing.T) {
        cfg := &DatabaseConfig{
            Host:     "localhost",
            Port:     5432,
            User:     "test_user",
            Password: "test_pass",
            Name:     "test_db",
            SSLMode:  "disable",
        }

        expected := "host=localhost port=5432 user=test_user password=test_pass dbname=test_db sslmode=disable"
        assert.Equal(t, expected, cfg.GetDSN())
    })
}
```

---

#### 任务 #2: 完善 Streamlit 管理仪表盘

**优先级**: P1 - 高优先级  
**预估**: 2 小时  
**状态**: ⏳ 待开始

**目标**: 完成 Streamlit 仪表盘剩余 20% 功能，达到生产可用状态

**价值**:

- 提供完整的管理工具
- 可视化监控系统状态
- 简化运维操作
- 无需手动调用 API 即可管理系统

**验收标准**:

- [ ] 余额管理模块实现 (查询、充值、提现)
- [ ] 系统监控模块实现 (API 状态、数据库连接、行情更新)
- [ ] UI 样式优化 (统一主题、响应式布局)
- [ ] 错误处理完善 (友好的错误提示)
- [ ] 所有功能通过手动测试
- [ ] 更新 README 添加仪表盘使用说明

**技术要点**:

- 使用 `dashboard/api/client.py` 封装 API 调用
- Streamlit 组件: `st.metric()`, `st.dataframe()`, `st.form()`
- 实时刷新: `st.rerun()` 或定时器
- 异常处理: try-except 包装所有 API 调用
- 布局优化: `st.columns()`, `st.container()`

**前置依赖**: 无 (管理员 API 已实现)

**测试策略**:

- 手动测试所有页面功能
- 验证错误场景 (网络失败、API 错误)
- 测试不同浏览器兼容性

**执行步骤**:

1. [ ] 实现余额管理页面 (45 分钟)
   - 创建 `dashboard/pages/balances.py`
   - 查询用户余额列表
   - 充值功能 (模拟)
   - 提现功能 (模拟)
   - 余额变动历史
2. [ ] 实现系统监控页面 (30 分钟)
   - 创建 `dashboard/pages/system.py`
   - API 健康检查 (/health)
   - 数据库连接状态
   - 行情更新状态
   - 服务器时间同步
3. [ ] UI 样式优化 (30 分钟)
   - 统一颜色主题
   - 优化表格显示
   - 添加图标和标签
   - 响应式布局调整
4. [ ] 测试和调试 (15 分钟)
   - 启动仪表盘: `cd dashboard && streamlit run app.py`
   - 测试所有页面
   - 修复发现的 bug

**余额管理页面示例**:

```python
# dashboard/pages/balances.py
import streamlit as st
from api.client import QuicksilverClient

st.title("💰 余额管理")

client = QuicksilverClient()

# 用户选择
user_id = st.number_input("用户 ID", min_value=1, value=1)

# 查询余额
if st.button("查询余额"):
    try:
        balances = client.get_balance(user_id)
        if balances:
            st.success(f"用户 {user_id} 的余额:")
            st.dataframe(balances)
        else:
            st.warning("余额为空")
    except Exception as e:
        st.error(f"查询失败: {e}")

# 充值功能
with st.expander("充值"):
    asset = st.text_input("资产代码", value="USDT")
    amount = st.number_input("充值金额", min_value=0.0, value=100.0)
    if st.button("确认充值"):
        # TODO: 调用充值 API
        st.success(f"充值 {amount} {asset} 成功")
```

---

#### 任务 #3: 运行 CCXT 客户端集成测试

**优先级**: P1 - 高优先级  
**预估**: 1 小时  
**状态**: ⏳ 待开始

**目标**: 使用 CCXT Python 客户端验证 API 完整兼容性

**价值**:

- 验证 CCXT 标准兼容性
- 发现潜在的格式问题
- 确保第三方客户端可正常使用
- 为用户提供使用示例

**验收标准**:

- [ ] 启动 Quicksilver 服务 (后台)
- [ ] 运行 `scripts/test_ccxt_client.py`
- [ ] 所有测试用例通过 (100%)
- [ ] 更新文档记录测试结果
- [ ] 修复发现的兼容性问题 (如有)

**技术要点**:

- 确保服务运行在 http://localhost:8080
- 确保测试用户已创建 (test_user, API Key)
- CCXT 版本要求: >= 4.0.0
- 测试覆盖:
  - fetchMarkets()
  - fetchTicker(symbol)
  - fetchBalance()
  - createOrder()
  - fetchOrder(id)
  - cancelOrder(id)

**前置依赖**:

- Quicksilver 服务正常运行
- 测试用户已创建 (可用 `scripts/init_test_user.sh`)

**测试策略**:

- 端到端集成测试
- 模拟真实用户使用场景
- 验证所有 CCXT 方法

**执行步骤**:

1. [ ] 启动 Quicksilver 服务 (10 分钟)

   ```bash
   # 后台启动服务
   make run &

   # 等待服务就绪
   sleep 5
   curl http://localhost:8080/health
   ```

2. [ ] 初始化测试用户 (5 分钟)

   ```bash
   ./scripts/init_test_user.sh
   ```

3. [ ] 运行 CCXT 测试脚本 (30 分钟)

   ```bash
   python3 scripts/test_ccxt_client.py
   ```

4. [ ] 分析测试结果 (10 分钟)

   - 记录通过/失败用例
   - 识别兼容性问题
   - 更新文档

5. [ ] 修复问题 (如有) (10 分钟)
   - 调整 CCXT 格式转换
   - 修复 API 响应格式
   - 重新运行测试

**预期测试输出**:

```
✅ 测试连接和时间
✅ 获取市场信息 (2 个交易对)
✅ 获取行情数据 (BTC/USDT)
✅ 查询余额
✅ 创建市价买单
✅ 查询订单状态
✅ 撤销订单
✅ 查询成交记录

总计: 8/8 通过 (100%) ✅
```

---

### P2 - 中优先级 (可选，本月内完成)

#### 任务 #4: 实现 Binance 数据源支持

**优先级**: P2 - 中优先级  
**预估**: 3 小时  
**状态**: 📅 计划中

**目标**: 添加 Binance 作为备用数据源，提高系统可靠性

**价值**:

- 数据源冗余，提高可用性
- 支持多数据源切换
- 为后续扩展打基础
- 解决 TODO 标记项

**验收标准**:

- [ ] 实现 `updateBinanceTickers()` 函数
- [ ] 支持配置文件切换数据源
- [ ] 添加单元测试 (覆盖率 > 80%)
- [ ] 集成测试验证
- [ ] 更新文档说明数据源配置

**技术要点**:

- Binance API: `https://api.binance.com/api/v3/ticker/price`
- 响应格式: `{"symbol":"BTCUSDT","price":"50000.50"}`
- 需要格式转换: `BTCUSDT` → `BTC/USDT`
- HTTP 超时处理: 5 秒
- 错误重试机制: 3 次

**前置依赖**: 无

**测试策略**:

- 单元测试: Mock HTTP 响应
- 集成测试: 调用真实 Binance API

**执行步骤**:

1. [ ] 研究 Binance API 文档 (30 分钟)
2. [ ] 实现 API 调用逻辑 (1 小时)
3. [ ] 编写单元测试 (1 小时)
4. [ ] 集成测试和文档更新 (30 分钟)

**实现示例**:

```go
// internal/service/market.go
func (s *MarketService) updateBinanceTickers() error {
    url := "https://api.binance.com/api/v3/ticker/price"

    resp, err := s.client.Get(url)
    if err != nil {
        return fmt.Errorf("failed to fetch from Binance: %w", err)
    }
    defer resp.Body.Close()

    var tickers []struct {
        Symbol string `json:"symbol"`
        Price  string `json:"price"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
        return err
    }

    for _, t := range tickers {
        // 转换格式: BTCUSDT -> BTC/USDT
        symbol := convertBinanceSymbol(t.Symbol)
        price, _ := strconv.ParseFloat(t.Price, 64)

        ticker := model.Ticker{
            Symbol:    symbol,
            LastPrice: price,
            Source:    "binance",
        }
        s.db.Save(&ticker)
    }

    return nil
}
```

---

#### 任务 #5: 添加管理员权限中间件

**优先级**: P2 - 中优先级  
**预估**: 1.5 小时  
**状态**: 📅 计划中

**目标**: 为管理员 API 添加权限验证，防止普通用户越权访问

**价值**:

- 提高系统安全性
- 防止数据泄露
- 符合最佳实践

**验收标准**:

- [ ] 实现 AdminAuth 中间件
- [ ] User 模型添加 role 字段 (user/admin)
- [ ] 测试覆盖率 ≥ 80%
- [ ] 更新文档说明权限控制

---

### P3 - 低优先级 (长期规划)

#### 任务 #6: 添加 API 速率限制

**优先级**: P3 - 低优先级  
**预估**: 1 小时  
**状态**: 📋 待定

**目标**: 添加限流中间件，防止 API 滥用

**价值**: 保护系统资源，防止恶意攻击

---

## ⚠️ 风险和依赖

### 技术风险

1. **风险**: CCXT 格式兼容性问题

   - **影响**: 可能需要多次调整格式
   - **缓解措施**: 参考 CCXT 官方文档，使用真实客户端测试
   - **概率**: 🟡 中等 (20%)
   - **应对**: 预留调试时间

2. **风险**: Config 层测试覆盖复杂
   - **影响**: 可能需要更多时间
   - **缓解措施**: 分批实现，先覆盖核心场景
   - **概率**: 🟢 低 (10%)
   - **应对**: 目标从 70% 调整为 50% (仍可接受)

### 外部依赖

1. **依赖**: Hyperliquid API 稳定性

   - **状态**: 目前稳定 ✅
   - **备选方案**: 实现 Binance 数据源 (任务 #4)
   - **风险评级**: 🟢 低

2. **依赖**: CCXT 库最新版本
   - **状态**: 已安装 (v4.x) ✅
   - **备选方案**: 无需备选，CCXT 成熟稳定
   - **风险评级**: 🟢 低

---

## 🔄 迭代策略

### 本周迭代计划

**Day 1 (周二 11 月 5 日)** - ✅ 已完成:

- [x] 创建本周开发计划
- [x] K 线功能实现和测试
- [x] ListUsers 测试补充

**Day 2 (周三 11 月 6 日)** - 计划:

- [ ] 任务 #1: Config 层测试 (1.5h)
- [ ] 任务 #2: Streamlit 仪表盘 (1h)

**Day 3 (周四 11 月 7 日)** - 计划:

- [ ] 任务 #2: 完成 Streamlit 仪表盘 (1h)
- [ ] 任务 #3: CCXT 客户端测试 (1h)

**Day 4 (周五 11 月 8 日)** - 计划:

- [ ] 代码审查和重构 (1h)
- [ ] 文档更新 (0.5h)
- [ ] 运行完整测试套件 (0.5h)

**Day 5 (周六 11 月 9 日)** - 可选:

- [ ] 任务 #4: Binance 数据源 (3h)

### 下周预览 (Week 46)

- 部署到测试环境
- 用户验收测试 (UAT)
- 性能优化迭代
- 准备生产发布

---

## 📈 进度跟踪表

| 任务                | 优先级 | 预估 | 实际 | 状态      | 完成度 | 负责人 |
| ------------------- | ------ | ---- | ---- | --------- | ------ | ------ |
| #1 Config 层测试    | P1     | 1.5h | -    | ⏳ 待开始 | 0%     | AI     |
| #2 Streamlit 仪表盘 | P1     | 2h   | -    | ⏳ 待开始 | 80%    | AI     |
| #3 CCXT 集成测试    | P1     | 1h   | -    | ⏳ 待开始 | 0%     | AI     |
| #4 Binance 数据源   | P2     | 3h   | -    | 📅 计划中 | 0%     | AI     |
| #5 管理员权限       | P2     | 1.5h | -    | 📅 计划中 | 0%     | AI     |

**总预估时间**: 9 小时 (P1: 4.5h, P2: 4.5h)  
**本周可用时间**: 20 小时  
**缓冲余量**: 55% ✅  
**预计完成率**: 100% (P1 任务)

---

## 📊 质量指标追踪

### 代码质量仪表盘

| 指标                 | 当前值 | 目标值 | 趋势 | 达成情况  |
| -------------------- | ------ | ------ | ---- | --------- |
| 测试覆盖率           | 62.0%  | 72%+   | ⬆️   | 🟡 接近   |
| Model 层覆盖率       | 100%   | 100%   | ➡️   | ✅ 完美   |
| CCXT 层覆盖率        | 100%   | 100%   | ➡️   | ✅ 完美   |
| Middleware 层覆盖率  | 95.7%  | > 80%  | ➡️   | ✅ 优秀   |
| API 层覆盖率         | 82.2%  | > 60%  | ➡️   | ✅ 优秀   |
| Service 层覆盖率     | ~73%   | > 60%  | ➡️   | ✅ 良好   |
| Config 层覆盖率      | 6.7%   | 70%+   | ⏳   | ⚠️ 需改进 |
| GolangCI-Lint 通过率 | ~85%   | > 95%  | ⬆️   | 🟡 良好   |
| 测试通过率           | 100%   | 100%   | ➡️   | ✅ 完美   |

### 性能指标

| 指标             | 当前值 | 目标值  | 达成情况  |
| ---------------- | ------ | ------- | --------- |
| API P99 响应时间 | 未测试 | < 100ms | ⏳ 待测试 |
| 订单处理 TPS     | 未测试 | > 100   | ⏳ 待测试 |
| 行情更新延迟     | < 1s   | < 2s    | ✅ 已达成 |
| K 线查询响应     | < 50ms | < 100ms | ✅ 预估   |

---

## 🛠️ 开发工具和命令

### 测试相关

```bash
# 运行所有测试
make test

# 仅单元测试 (快速)
make test-unit

# 查看覆盖率
make test-coverage
open coverage.html

# 性能基准测试
make bench

# 监听文件变化自动测试
make test-watch
```

### 代码质量

```bash
# 完整质量检查
make quality-check

# 格式检查
make fmt-check

# Lint 检查
make lint

# 静态分析
make vet

# 竞态检测
make race
```

### 服务运行

```bash
# 启动开发服务器 (热重载)
make dev

# 直接运行
make run

# 启动数据库
make db-start

# 后台运行 (用于测试)
make run &
```

### API 测试

```bash
# 健康检查
curl http://localhost:8080/health

# 获取行情
curl http://localhost:8080/v1/ticker/BTC/USDT

# 获取K线数据
curl "http://localhost:8080/v1/kline/BTC/USDT?interval=1m&limit=100"

# 使用 httpie (更友好)
http GET :8080/v1/ticker/BTC/USDT
http GET :8080/v1/kline/BTC/USDT interval==1m limit==100
```

---

## 🎉 成功标准

### 本周结束时应达到的状态

**必须 (MUST)**:

- ✅ Config 层测试覆盖率 > 70%
- ✅ Streamlit 仪表盘功能完整 (100%)
- ✅ CCXT 客户端测试通过率 100%
- ✅ 整体测试覆盖率 > 72%
- ✅ 所有测试用例通过

**应该 (SHOULD)**:

- ✅ GolangCI-Lint 通过率 > 95%
- ✅ 文档更新完整
- ✅ 代码审查通过

**可以 (COULD)**:

- Binance 数据源实现
- 管理员权限中间件

### 验收清单

- [ ] 运行 `make test` - 所有测试通过
- [ ] 运行 `make test-coverage` - 覆盖率 > 72%
- [ ] 运行 `make quality-check` - 无严重问题
- [ ] 运行 `python3 scripts/test_ccxt_client.py` - 100% 通过
- [ ] 启动 Streamlit 仪表盘 - 所有页面可用
- [ ] 检查 Git 状态 - 无未提交的重要修改
- [ ] 更新 PROJECT_STATUS.md - 反映最新进度

---

## 📝 每日任务清单

### Tuesday (11 月 5 日) ✅

- [x] 创建本周开发计划
- [x] 实现 K 线数据服务
- [x] 补充 ListUsers 测试
- [x] 更新项目文档

### Wednesday (11 月 6 日)

- [ ] 早上 9:00-10:30: 任务 #1 - Config 层测试 (1.5h)
- [ ] 下午 14:00-15:00: 任务 #2 - Streamlit 余额模块 (1h)
- [ ] 晚上 20:00: 提交代码和更新进度

### Thursday (11 月 7 日)

- [ ] 早上 9:00-10:00: 任务 #2 - Streamlit 监控模块 (1h)
- [ ] 下午 14:00-15:00: 任务 #3 - CCXT 集成测试 (1h)
- [ ] 晚上 20:00: 代码审查和测试

### Friday (11 月 8 日)

- [ ] 上午 9:00-10:00: 代码重构和优化 (1h)
- [ ] 下午 14:00-15:00: 文档更新和整理 (1h)
- [ ] 下午 17:00: 周总结和评审

---

## 🔍 每日检查清单

### 每日必做 ✅

- [ ] 查看今日任务列表
- [ ] 运行 `make test` 确保测试通过
- [ ] 至少 1 次代码提交
- [ ] 更新任务进度 (Markdown 文件)
- [ ] 记录遇到的问题和解决方案

### 每周必做 ✅

- [ ] 周一: 制定本周计划
- [ ] 周三: 中期进度检查
- [ ] 周五: 运行 `make quality-check`
- [ ] 周五: 更新项目文档
- [ ] 周五: Sprint 回顾会议 (17:00)

---

## 📚 参考资料

### 内部文档

- **系统设计**: `docs/system-design-mvp.md`
- **开发指南**: `.github/copilot-instructions.md`
- **项目状态**: `PROJECT_STATUS.md`
- **质量报告**: `QUALITY_REPORT.md`
- **数据库设计**: `docs/database.md`
- **K 线 API 文档**: `docs/kline-api.md`

### 外部资源

- **CCXT 文档**: https://docs.ccxt.com/
- **Echo 框架**: https://echo.labstack.com/
- **GORM 指南**: https://gorm.io/docs/
- **Streamlit 文档**: https://docs.streamlit.io/
- **Go 测试最佳实践**: https://go.dev/doc/tutorial/add-a-test

### 工具和库

- **GolangCI-Lint**: https://golangci-lint.run/
- **Testify**: https://github.com/stretchr/testify
- **Viper**: https://github.com/spf13/viper
- **Zap**: https://pkg.go.dev/go.uber.org/zap

---

## 🎯 下一步行动

### 立即开始 (现在)

1. **阅读本计划**: 理解本周目标和任务
2. **运行测试**: `make test` 确保环境正常
3. **选择任务**: 从 P1 任务 #1 开始
4. **遵循 TDD**: 先写测试，再实现功能

### 推荐工作流 (TDD)

```bash
# 1. 选择任务 (例如: Config层测试)
# 2. 创建功能分支
git checkout -b feature/config-tests

# 3. 先写测试 (红阶段)
vim internal/config/config_test.go

# 4. 运行测试 (应该失败)
make test-unit

# 5. 实现功能 (绿阶段)
vim internal/config/config.go

# 6. 运行测试 (应该通过)
make test-unit

# 7. 重构代码
# ... 优化代码质量

# 8. 再次测试
make test

# 9. 检查覆盖率
make test-coverage

# 10. 提交代码
git add .
git commit -m "test: add config layer tests (70% coverage)"
git push origin feature/config-tests
```

---

## 💡 提示和技巧

### 时间管理

- 使用番茄工作法 (25 分钟工作 + 5 分钟休息)
- 每完成一个任务立即提交代码
- 遇到困难超过 30 分钟，暂停并寻求帮助或换思路

### 测试编写

- 先写最简单的成功场景测试
- 再添加边界条件和异常场景
- 使用表驱动测试 (Table-Driven Tests) 提高效率
- 测试命名清晰: `Test<Function>_<Scenario>`

### 代码质量

- 每次提交前运行 `make fmt`
- 定期运行 `make lint` 检查问题
- 保持函数简短 (< 50 行)
- 添加必要的注释，尤其是复杂逻辑

### 调试技巧

- 使用结构化日志而非 fmt.Println
- 善用 VSCode 断点调试
- 编写单元测试隔离问题
- 查看测试覆盖率报告 (coverage.html) 找到未覆盖分支

---

**计划制定者**: AI Development Planner  
**审核者**: Quicksilver 开发团队  
**版本**: v2.0.0 (基于 planner.prompt.md v1.0.0)  
**最后更新**: 2025-11-05 23:15

---

## 📞 反馈和改进

如有任何问题或建议，请通过以下方式反馈:

- **GitHub Issues**: 创建 Issue 讨论
- **文档更新**: 提交 PR 改进本计划
- **即时沟通**: 在代码审查时提出

**记住**: 计划是指导，不是束缚。根据实际情况灵活调整，但要保持目标清晰！🎯
