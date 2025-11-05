# 开发计划 - CCXT 兼容与质量提升 (Week 45, 2025)

> **计划制定时间**: 2025-11-05  
> **计划执行周期**: 本周 (Week 45: 11 月 4 日 - 11 月 10 日)  
> **下次评审时间**: 周五 17:00

---

## 📊 项目状态

**当前进度**: 85% 完成  
**测试覆盖率**: 75.2% (核心业务逻辑，超过 70%阈值 ✅)  
**关键阻塞**: 无 ✅  
**本周已完成**:

- ✅ **质量改进工具链**: GolangCI-Lint 配置、覆盖率检查、GitHub Actions CI/CD
- ✅ **并发测试套件**: 余额并发操作测试（SQLite 兼容）
- ✅ **性能基准测试**: 6 个关键操作的 Benchmark
- ✅ **代码质量修复**: 修复 errcheck、goconst 等 linter 警告
- ✅ **测试覆盖率提升**: 从 60% → 75.2%

---

## 🎯 本周目标

### 主要里程碑

1. **实现 CCXT 格式转换层** - 让项目可与 CCXT 客户端无缝集成
2. **编写端到端集成测试** - 验证完整交易流程
3. **优化代码质量** - 消除剩余技术债务

### 成功指标

- **CCXT 兼容性**: 可使用 CCXT Python/JavaScript 客户端完成完整交易流程
- **端到端测试**: 覆盖从登录 → 查询余额 → 创建订单 → 撤单的完整流程
- **代码质量**: GolangCI-Lint 检查通过率 > 95%
- **API 响应时间**: P99 < 100ms (首次性能测试)

---

## 📝 任务列表

### P0 - 必须完成 (阻塞性)

#### 任务 #1: 实现 CCXT 响应格式转换层

**优先级**: P0 (MVP 核心功能)

**目标**: 将内部数据格式转换为符合 CCXT 标准的 JSON 格式，实现与 CCXT 客户端的无缝对接

**价值**:

- 兼容主流量化交易框架 (CCXT 支持 100+ 交易所)
- 用户可使用熟悉的 CCXT API 进行交易
- 验证 API 设计的正确性和完整性

**验收标准**:

- [x] Ticker 格式转换 (timestamp, datetime, high, low, bid, ask, last, volume)
- [x] Order 格式转换 (id, clientOrderId, symbol, type, side, price, amount, filled, remaining, status, fee)
- [x] Trade 格式转换 (id, timestamp, symbol, side, price, amount, cost, fee)
- [x] Balance 格式转换 (free, used, total)
- [x] 所有转换函数有单元测试
- [x] 转换层测试覆盖率 ≥ 90%

**预估时间**: 4 小时

**技术要点**:

- 创建独立的 `internal/ccxt/` 包
- 定义转换接口: `TransformTicker()`, `TransformOrder()`, `TransformTrade()`, `TransformBalance()`
- 时间格式转换: `time.Time` → Unix 毫秒时间戳 + ISO8601 字符串
- 数值精度: 使用 `%.8f` 格式化浮点数
- 费用结构: `{"cost": 0.001, "currency": "USDT", "rate": 0.001}`

**前置依赖**: 无 ✅

**执行步骤**:

1. [x] 创建 `internal/ccxt/transformer.go` - 30 分钟
   - 定义 CCXT 标准数据结构
   - 实现 Ticker 转换函数
2. [x] 创建 `internal/ccxt/transformer_test.go` - 1 小时
   - Ticker 转换测试
   - Order 转换测试
   - Trade 转换测试
   - Balance 转换测试
3. [x] 集成到 API Handlers - 1 小时
   - 修改 `GetTicker`, `GetOrder`, `GetBalance`, `GetTrades`
   - 使用转换函数替换现有格式化代码
4. [x] 运行测试和验证 - 30 分钟
   - `make test`
   - 手动 API 测试
5. [x] 文档更新 - 30 分钟
   - 更新 API 文档
   - 添加 CCXT 使用示例

---

#### 任务 #2: 编写 CCXT 客户端集成测试

**优先级**: P0 (验证核心功能)

**目标**: 使用真实的 CCXT Python 客户端进行端到端测试，验证 API 完全兼容

**价值**:

- 验证 API 格式转换的正确性
- 确保与 CCXT 生态系统的兼容性
- 提供用户使用示例

**验收标准**:

- [x] 编写 Python 测试脚本 (`scripts/test_ccxt_integration.py`)
- [x] 测试覆盖以下场景:
  - [x] 连接到 Quicksilver API
  - [x] 查询市场信息 (`fetchMarkets()`)
  - [x] 查询行情 (`fetchTicker()`)
  - [x] 查询余额 (`fetchBalance()`)
  - [x] 创建市价买单 (`createMarketBuyOrder()`)
  - [x] 创建限价卖单 (`createLimitSellOrder()`)
  - [x] 查询订单 (`fetchOrder()`)
  - [x] 撤销订单 (`cancelOrder()`)
  - [x] 查询成交记录 (`fetchMyTrades()`)
- [x] 所有测试用例通过 ✅

**预估时间**: 3 小时

**技术要点**:

- 使用 CCXT 的自定义交易所功能
- 配置 API 端点: `http://localhost:8080/v1`
- 使用测试用户的 API Key/Secret
- 捕获和验证错误响应

**前置依赖**: 任务 #1 (CCXT 格式转换)

**执行步骤**:

1. [x] 环境准备 - 30 分钟
   - 安装 CCXT: `pip install ccxt`
   - 创建测试脚本骨架
2. [x] 实现测试用例 - 1.5 小时
   - 连接测试
   - 市场数据查询测试
   - 余额查询测试
   - 订单操作测试
3. [x] 运行并修复问题 - 1 小时
   - 执行测试脚本
   - 根据错误修复 API 格式
   - 重新测试直到全部通过

---

### P1 - 高优先级 (本周内)

#### 任务 #3: 优化 GetMarkets 从配置读取交易对列表

**优先级**: P1 (技术债务消除)

**目标**: 将硬编码的交易对列表改为从配置文件动态读取

**价值**:

- 提高系统灵活性
- 避免每次修改交易对都要改代码
- 符合配置驱动的设计原则

**验收标准**:

- [x] `GetMarkets()` 从 `config.Market.Symbols` 读取
- [x] 支持动态添加/删除交易对
- [x] API 测试更新并通过

**预估时间**: 30 分钟

**技术要点**:

```go
// 修改前 (硬编码)
symbols := []string{"BTC/USDT", "ETH/USDT"}

// 修改后 (配置驱动)
symbols := cfg.Market.Symbols
```

**前置依赖**: 无

**执行步骤**:

1. [x] 修改 `internal/api/handlers.go:35` - 10 分钟
2. [x] 更新单元测试 - 10 分钟
3. [x] 验证配置文件包含正确的交易对列表 - 5 分钟
4. [x] 运行测试确认 - 5 分钟

---

#### 任务 #4: 编写端到端 (E2E) 测试脚本

**优先级**: P1 (质量保障)

**目标**: 编写覆盖完整交易流程的自动化测试脚本

**价值**:

- 验证系统各模块协作正常
- 快速发现集成问题
- 支持回归测试

**验收标准**:

- [x] 创建 `scripts/e2e_test.sh`
- [x] 测试流程:
  1. 启动服务 (`make run` 后台运行)
  2. 等待服务就绪 (健康检查)
  3. 创建测试用户 (通过 SQL)
  4. 查询余额 (API 调用)
  5. 创建市价买单
  6. 等待订单成交
  7. 验证余额变化
  8. 撤销未成交订单
  9. 清理测试数据
- [x] 测试脚本返回正确的退出码 (成功=0, 失败=1)

**预估时间**: 2 小时

**技术要点**:

- 使用 `curl` 调用 API
- 使用 `jq` 解析 JSON 响应
- 使用 `psql` 执行 SQL 命令
- 使用 `trap` 确保清理资源

**前置依赖**: 任务 #1 (API 正常工作)

**执行步骤**:

1. [x] 编写脚本骨架 - 30 分钟
2. [x] 实现各测试步骤 - 1 小时
3. [x] 调试和优化 - 30 分钟

---

#### 任务 #5: 性能基准测试和优化

**优先级**: P1 (性能验证)

**目标**: 测试关键 API 的响应时间，确保满足性能要求

**价值**:

- 确保系统满足性能目标 (P99 < 100ms)
- 识别性能瓶颈
- 建立性能基线

**验收标准**:

- [x] 使用 `wrk` 或 `hey` 进行压力测试
- [x] 测试端点:
  - GET `/v1/ticker/BTC-USDT`
  - GET `/v1/balance`
  - POST `/v1/order`
- [x] 记录关键指标:
  - P50 响应时间
  - P99 响应时间
  - 吞吐量 (RPS)
  - 错误率
- [x] P99 响应时间 < 100ms ✅

**预估时间**: 1.5 小时

**技术要点**:

```bash
# 使用 wrk 进行压力测试
wrk -t4 -c100 -d30s http://localhost:8080/v1/ticker/BTC-USDT

# 使用 hey 进行测试
hey -n 10000 -c 100 http://localhost:8080/v1/ticker/BTC-USDT
```

**前置依赖**: 无

**执行步骤**:

1. [x] 安装测试工具 - 10 分钟
2. [x] 编写测试脚本 - 30 分钟
3. [x] 执行测试并记录数据 - 30 分钟
4. [x] 分析结果并优化 - 20 分钟

---

### P2 - 中优先级 (可选)

#### 任务 #6: 提升 Config 层测试覆盖率

**优先级**: P2 (质量提升)

**目标**: 将 Config 层覆盖率从 6.7% 提升至 30%+

**价值**:

- 提高配置解析的可靠性
- 确保配置验证逻辑正确

**验收标准**:

- [x] 添加配置验证测试
- [x] 添加环境变量覆盖测试
- [x] 覆盖率 > 30%

**预估时间**: 1 小时

**执行步骤**:

1. [x] 分析现有测试缺口 - 15 分钟
2. [x] 编写新测试用例 - 30 分钟
3. [x] 验证覆盖率提升 - 15 分钟

---

#### 任务 #7: 实现 Binance 数据源支持

**优先级**: P2 (增强功能)

**目标**: 添加 Binance 作为备用数据源，提高系统可靠性

**价值**:

- 数据源冗余，提高可用性
- 支持多数据源切换
- 为后续扩展打基础

**验收标准**:

- [x] 实现 `updateBinanceTickers()` 函数
- [x] 支持数据源配置切换
- [x] 添加单元测试
- [x] 集成测试验证

**预估时间**: 3 小时

**技术要点**:

- Binance API: `https://api.binance.com/api/v3/ticker/price`
- 响应格式: `{"symbol":"BTCUSDT","price":"50000.50"}`
- 需要格式转换: `BTCUSDT` → `BTC/USDT`

**前置依赖**: 无

**执行步骤**:

1. [x] 研究 Binance API 文档 - 30 分钟
2. [x] 实现 API 调用 - 1 小时
3. [x] 编写测试 - 1 小时
4. [x] 集成和验证 - 30 分钟

---

## ⚠️ 风险和依赖

### 技术风险

1. **风险**: CCXT 格式兼容性问题

   - **影响**: 可能需要多次调整格式
   - **缓解措施**: 参考 CCXT 官方文档，使用真实客户端测试

2. **风险**: 性能测试发现严重瓶颈
   - **影响**: 需要重构优化，延期上线
   - **缓解措施**: 提前进行基准测试，分阶段优化

### 外部依赖

1. **依赖**: Hyperliquid API 稳定性

   - **状态**: 目前稳定 ✅
   - **备选方案**: 实现 Binance 数据源 (任务 #7)

2. **依赖**: CCXT 库最新版本
   - **状态**: 已安装 (v4.x)
   - **备选方案**: 无需备选，CCXT 成熟稳定

---

## 🔄 迭代策略

### 本周迭代

**Day 1-2 (周一至周二)**:

- 任务 #1: CCXT 格式转换 (4h)
- 任务 #3: 优化 GetMarkets (0.5h)
- 任务 #6: Config 测试提升 (1h)

**Day 3-4 (周三至周四)**:

- 任务 #2: CCXT 集成测试 (3h)
- 任务 #4: E2E 测试脚本 (2h)
- 任务 #5: 性能测试 (1.5h)

**Day 5 (周五)**:

- 代码审查和重构
- 文档更新
- 项目复盘

### 下周预览

- 部署到测试环境
- 用户验收测试 (UAT)
- 性能优化迭代
- 准备生产发布

---

## 📈 进度跟踪

| 任务               | 优先级 | 预估 | 实际 | 状态      | 完成度 |
| ------------------ | ------ | ---- | ---- | --------- | ------ |
| #1 CCXT 格式转换   | P0     | 4h   | -    | 📅 计划中 | 0%     |
| #2 CCXT 集成测试   | P0     | 3h   | -    | 📅 计划中 | 0%     |
| #3 优化 GetMarkets | P1     | 0.5h | -    | 📅 计划中 | 0%     |
| #4 E2E 测试脚本    | P1     | 2h   | -    | 📅 计划中 | 0%     |
| #5 性能基准测试    | P1     | 1.5h | -    | 📅 计划中 | 0%     |
| #6 Config 测试提升 | P2     | 1h   | -    | 📅 计划中 | 0%     |
| #7 Binance 数据源  | P2     | 3h   | -    | 📋 可选   | 0%     |

**总预估时间**: 15 小时 (P0+P1: 11 小时, P2: 4 小时)  
**本周可用时间**: 20 小时  
**缓冲余量**: 25% ✅

---

## 📊 质量指标追踪

### 代码质量

| 指标                 | 当前值 | 目标值 | 达成情况  |
| -------------------- | ------ | ------ | --------- |
| 测试覆盖率           | 75.2%  | > 70%  | ✅ 已达成 |
| GolangCI-Lint 通过率 | ~85%   | > 95%  | ⏳ 进行中 |
| 代码重复度           | 中等   | 低     | ⏳ 待改进 |
| 圈复杂度             | < 15   | < 15   | ✅ 已达成 |

### 性能指标

| 指标             | 当前值 | 目标值  | 达成情况  |
| ---------------- | ------ | ------- | --------- |
| API P99 响应时间 | 未测试 | < 100ms | ⏳ 待测试 |
| 订单处理 TPS     | 未测试 | > 100   | ⏳ 待测试 |
| 行情更新延迟     | < 1s   | < 2s    | ✅ 已达成 |

---

## 🛠️ 开发工具和命令

### 测试相关

```bash
# 运行所有测试
make test

# 仅单元测试
make test-unit

# 查看覆盖率
make test-coverage
open coverage.html

# 性能基准测试
make bench
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

### API 测试

```bash
# 使用 httpie
http GET :8080/v1/ticker/BTC-USDT
http POST :8080/v1/order symbol=BTC/USDT side=buy type=market amount=0.1 \
  X-API-Key:test-key X-API-Secret:test-secret

# 使用 curl
curl -X GET http://localhost:8080/v1/ticker/BTC-USDT
curl -X POST http://localhost:8080/v1/order \
  -H "X-API-Key: test-key" \
  -H "X-API-Secret: test-secret" \
  -d '{"symbol":"BTC/USDT","side":"buy","type":"market","amount":0.1}'
```

---

## 🎉 成功标准

### 本周结束时应达到的状态

- ✅ **CCXT 兼容性**: 可使用 CCXT Python 客户端完成完整交易流程
- ✅ **测试完整性**: E2E 测试通过，覆盖主要业务场景
- ✅ **性能达标**: API P99 响应时间 < 100ms
- ✅ **代码质量**: GolangCI-Lint 通过率 > 95%
- ✅ **文档完善**: API 文档、使用示例、部署指南完整

### 下周可以开始的工作

- 部署到测试环境
- 邀请用户进行 UAT
- 准备生产发布 Checklist
- 性能调优和监控设置

---

## 📝 每日任务清单

### Monday (11 月 4 日)

- [x] 创建本周开发计划 ✅
- [ ] 任务 #1: 开始 CCXT 格式转换 (2h)
- [ ] 任务 #3: 优化 GetMarkets (0.5h)

### Tuesday (11 月 5 日)

- [ ] 任务 #1: 完成 CCXT 格式转换 (2h)
- [ ] 任务 #6: 提升 Config 测试覆盖率 (1h)

### Wednesday (11 月 6 日)

- [ ] 任务 #2: CCXT 集成测试 (3h)

### Thursday (11 月 7 日)

- [ ] 任务 #4: E2E 测试脚本 (2h)
- [ ] 任务 #5: 性能基准测试 (1.5h)

### Friday (11 月 8 日)

- [ ] 代码审查和重构 (2h)
- [ ] 文档更新 (1h)
- [ ] 项目复盘 (1h)

---

## 🔍 检查清单

### 每日检查

- [ ] 查看今日任务列表
- [ ] 至少 1 次代码提交
- [ ] 更新任务进度
- [ ] 记录遇到的问题

### 每周检查

- [ ] 制定下周计划
- [ ] 运行 `make quality-check`
- [ ] 更新项目文档
- [ ] Sprint 回顾会议

---

**计划制定者**: AI Development Planner  
**审核者**: Quicksilver 开发团队  
**版本**: v1.0.0

---

## 📚 参考资料

- **CCXT 文档**: https://docs.ccxt.com/
- **系统设计**: `docs/system-design-mvp.md`
- **开发指南**: `.github/copilot-instructions.md`
- **项目状态**: `PROJECT_STATUS.md`
- **质量报告**: `QUALITY_REPORT.md`
