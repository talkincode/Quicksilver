# Hyper-Sim 数据库设计文档 (Database Schema)

> **版本**: v1.0  
> **数据库**: PostgreSQL 16+  
> **字符集**: UTF-8  
> **时区**: UTC  
> **更新日期**: 2024-11-01

---

## 目录

- [1. 数据库架构概览](#1-数据库架构概览)
- [2. 表结构详细设计](#2-表结构详细设计)
- [3. 索引设计](#3-索引设计)
- [4. 数据关系图](#4-数据关系图)
- [5. 字段命名规范](#5-字段命名规范)
- [6. 数据迁移策略](#6-数据迁移策略)
- [7. 性能优化建议](#7-性能优化建议)

---

## 1. 数据库架构概览

### 1.1 核心表结构

```
┌─────────────────────────────────────────────────────────────────┐
│                      Hyper-Sim Database                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐         ┌──────────┐         ┌──────────┐       │
│  │  users   │────────▶│ balances │         │ tickers  │       │
│  │  用户表  │  1:N    │  余额表  │         │ 行情表   │       │
│  └────┬─────┘         └──────────┘         └──────────┘       │
│       │                                                         │
│       │ 1:N                                                     │
│       ├────────────────────┐                                    │
│       │                    │                                    │
│       ▼                    ▼                                    │
│  ┌──────────┐         ┌──────────┐                            │
│  │  orders  │────────▶│  trades  │                            │
│  │  订单表  │  1:N    │  成交表  │                            │
│  └──────────┘         └──────────┘                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

表数量: 5 个核心表
关系类型: 一对多 (1:N)
总字段数: ~45 个字段
```

### 1.2 表统计信息

| 表名       | 中文名 | 预估行数   | 增长速度 | 主键类型 | 分区策略   |
| ---------- | ------ | ---------- | -------- | -------- | ---------- |
| `users`    | 用户表 | 1K - 10K   | 慢       | SERIAL   | 无         |
| `balances` | 余额表 | 10K - 100K | 中       | SERIAL   | 无         |
| `orders`   | 订单表 | 100K - 1M  | 快       | SERIAL   | 建议按时间 |
| `trades`   | 成交表 | 100K - 1M  | 快       | SERIAL   | 建议按时间 |
| `tickers`  | 行情表 | < 100      | 慢       | VARCHAR  | 无         |

### 1.3 存储空间预估

```
基于 10,000 用户、每天 10,000 笔交易的场景:

┌────────────────────────────────────────────────┐
│ 表名          │ 单行大小  │ 日增长  │ 月增长  │
├────────────────────────────────────────────────┤
│ users         │ ~200B     │ ~20KB   │ ~600KB  │
│ balances      │ ~100B     │ ~50KB   │ ~1.5MB  │
│ orders        │ ~300B     │ ~3MB    │ ~90MB   │
│ trades        │ ~250B     │ ~2.5MB  │ ~75MB   │
│ tickers       │ ~200B     │ 几乎无  │ 几乎无  │
├────────────────────────────────────────────────┤
│ 合计          │ -         │ ~5.6MB  │ ~168MB  │
└────────────────────────────────────────────────┘

年度存储需求: ~2GB (不含索引)
含索引预估: ~4-5GB
```

---

## 2. 表结构详细设计

### 2.1 users (用户表)

**用途**: 存储系统用户基本信息和 API 密钥

**数据量级**: 小 (< 10K)

**查询频率**: 中等

**表结构**:

```sql
CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    email       VARCHAR(255) NOT NULL UNIQUE,
    username    VARCHAR(50),
    api_key     VARCHAR(64) NOT NULL UNIQUE,
    api_secret  VARCHAR(128) NOT NULL,
    status      VARCHAR(20) DEFAULT 'active',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login  TIMESTAMP,
    metadata    JSONB
);

-- 注释
COMMENT ON TABLE users IS '用户基础信息表';
COMMENT ON COLUMN users.id IS '用户唯一标识';
COMMENT ON COLUMN users.email IS '用户邮箱，用于登录';
COMMENT ON COLUMN users.username IS '用户名（可选）';
COMMENT ON COLUMN users.api_key IS 'API访问密钥（公开）';
COMMENT ON COLUMN users.api_secret IS 'API签名密钥（加密存储）';
COMMENT ON COLUMN users.status IS '用户状态: active/suspended/deleted';
COMMENT ON COLUMN users.metadata IS '扩展字段，存储JSON格式的额外信息';
```

**字段详细说明**:

| 字段名       | 类型      | 长度 | 必填 | 默认值   | 说明                  | 示例值                |
| ------------ | --------- | ---- | ---- | -------- | --------------------- | --------------------- |
| `id`         | SERIAL    | -    | ✓    | AUTO     | 主键，自增 ID         | `1001`                |
| `email`      | VARCHAR   | 255  | ✓    | -        | 邮箱地址，唯一        | `user@example.com`    |
| `username`   | VARCHAR   | 50   | ✗    | NULL     | 用户名                | `alice`               |
| `api_key`    | VARCHAR   | 64   | ✓    | -        | API Key (32 字节 HEX) | `a1b2c3d4...`         |
| `api_secret` | VARCHAR   | 128  | ✓    | -        | API Secret (哈希后)   | `$2a$10$...`          |
| `status`     | VARCHAR   | 20   | ✓    | `active` | 账户状态              | `active`              |
| `created_at` | TIMESTAMP | -    | ✓    | NOW()    | 创建时间              | `2024-11-01 10:00:00` |
| `updated_at` | TIMESTAMP | -    | ✓    | NOW()    | 更新时间              | `2024-11-01 10:00:00` |
| `last_login` | TIMESTAMP | -    | ✗    | NULL     | 最后登录时间          | `2024-11-01 12:30:00` |
| `metadata`   | JSONB     | -    | ✗    | NULL     | 扩展数据              | `{"level":"vip"}`     |

**业务规则**:

- `email` 必须符合邮箱格式验证
- `api_key` 生成时使用 `crypto/rand` 生成 32 字节随机数
- `api_secret` 必须使用 bcrypt 加密存储（成本因子 ≥ 10）
- `status` 枚举值: `active`, `suspended`, `deleted`
- 软删除：`status = 'deleted'`，不物理删除记录

**触发器**:

```sql
-- 自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

### 2.2 balances (余额表)

**用途**: 存储用户各币种的可用余额和冻结余额

**数据量级**: 中等 (< 100K)

**查询频率**: 高

**表结构**:

```sql
CREATE TABLE balances (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset       VARCHAR(10) NOT NULL,
    available   DECIMAL(20, 8) DEFAULT 0 CHECK (available >= 0),
    locked      DECIMAL(20, 8) DEFAULT 0 CHECK (locked >= 0),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, asset)
);

-- 注释
COMMENT ON TABLE balances IS '用户资产余额表';
COMMENT ON COLUMN balances.user_id IS '关联用户ID';
COMMENT ON COLUMN balances.asset IS '资产代码，如 BTC, USDT';
COMMENT ON COLUMN balances.available IS '可用余额';
COMMENT ON COLUMN balances.locked IS '冻结余额（挂单中）';

-- 约束：总余额不能为负
ALTER TABLE balances ADD CONSTRAINT check_total_balance
    CHECK (available + locked >= 0);
```

**字段详细说明**:

| 字段名       | 类型      | 精度 | 必填 | 默认值 | 说明          | 示例值                |
| ------------ | --------- | ---- | ---- | ------ | ------------- | --------------------- |
| `id`         | SERIAL    | -    | ✓    | AUTO   | 主键          | `1001`                |
| `user_id`    | INTEGER   | -    | ✓    | -      | 用户 ID，外键 | `1001`                |
| `asset`      | VARCHAR   | 10   | ✓    | -      | 币种代码      | `BTC`, `USDT`         |
| `available`  | DECIMAL   | 20,8 | ✓    | `0`    | 可用余额      | `1.50000000`          |
| `locked`     | DECIMAL   | 20,8 | ✓    | `0`    | 冻结余额      | `0.25000000`          |
| `created_at` | TIMESTAMP | -    | ✓    | NOW()  | 创建时间      | `2024-11-01 10:00:00` |
| `updated_at` | TIMESTAMP | -    | ✓    | NOW()  | 更新时间      | `2024-11-01 10:00:00` |

**业务规则**:

- `available` 和 `locked` 必须 ≥ 0
- `available + locked` 表示总余额
- 每个用户每个币种只能有一条记录（`UNIQUE(user_id, asset)`）
- 支持的资产类型：`BTC`, `USDT`, `ETH` (v1.0 仅支持 BTC/USDT)
- 精度：统一使用 8 位小数

**余额操作类型**:

```
1. 冻结 (Freeze):
   available -= amount
   locked += amount

2. 解冻 (Unfreeze):
   locked -= amount
   available += amount

3. 扣除 (Deduct):
   locked -= amount

4. 增加 (Credit):
   available += amount
```

**触发器**:

```sql
CREATE TRIGGER update_balances_updated_at
    BEFORE UPDATE ON balances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

### 2.3 orders (订单表)

**用途**: 存储所有交易订单信息

**数据量级**: 大 (> 100K)

**查询频率**: 极高

**表结构**:

```sql
CREATE TABLE orders (
    id              BIGSERIAL PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES users(id),
    symbol          VARCHAR(20) NOT NULL,
    side            VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    type            VARCHAR(10) NOT NULL CHECK (type IN ('market', 'limit')),
    status          VARCHAR(20) NOT NULL DEFAULT 'new',
    price           DECIMAL(20, 8),
    amount          DECIMAL(20, 8) NOT NULL CHECK (amount > 0),
    filled          DECIMAL(20, 8) DEFAULT 0 CHECK (filled >= 0),
    average_price   DECIMAL(20, 8),
    fee             DECIMAL(20, 8) DEFAULT 0,
    fee_asset       VARCHAR(10),
    client_order_id VARCHAR(64),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    filled_at       TIMESTAMP,
    canceled_at     TIMESTAMP,
    metadata        JSONB
);

-- 注释
COMMENT ON TABLE orders IS '交易订单表';
COMMENT ON COLUMN orders.symbol IS '交易对，如 BTC/USDT';
COMMENT ON COLUMN orders.side IS '买卖方向: buy/sell';
COMMENT ON COLUMN orders.type IS '订单类型: market/limit';
COMMENT ON COLUMN orders.status IS '订单状态';
COMMENT ON COLUMN orders.filled IS '已成交数量';
COMMENT ON COLUMN orders.average_price IS '平均成交价';

-- 约束
ALTER TABLE orders ADD CONSTRAINT check_filled_amount
    CHECK (filled <= amount);

ALTER TABLE orders ADD CONSTRAINT check_limit_order_price
    CHECK (type != 'limit' OR price IS NOT NULL);
```

**字段详细说明**:

| 字段名            | 类型      | 精度 | 必填 | 默认值 | 说明          | 示例值                       |
| ----------------- | --------- | ---- | ---- | ------ | ------------- | ---------------------------- |
| `id`              | BIGSERIAL | -    | ✓    | AUTO   | 订单 ID       | `1000001`                    |
| `user_id`         | INTEGER   | -    | ✓    | -      | 用户 ID       | `1001`                       |
| `symbol`          | VARCHAR   | 20   | ✓    | -      | 交易对        | `BTC/USDT`                   |
| `side`            | VARCHAR   | 4    | ✓    | -      | 买卖方向      | `buy`, `sell`                |
| `type`            | VARCHAR   | 10   | ✓    | -      | 订单类型      | `market`, `limit`            |
| `status`          | VARCHAR   | 20   | ✓    | `new`  | 订单状态      | `open`, `filled`, `canceled` |
| `price`           | DECIMAL   | 20,8 | ✗    | NULL   | 委托价格      | `50000.00000000`             |
| `amount`          | DECIMAL   | 20,8 | ✓    | -      | 委托数量      | `0.01000000`                 |
| `filled`          | DECIMAL   | 20,8 | ✓    | `0`    | 已成交数量    | `0.01000000`                 |
| `average_price`   | DECIMAL   | 20,8 | ✗    | NULL   | 平均成交价    | `50000.00000000`             |
| `fee`             | DECIMAL   | 20,8 | ✓    | `0`    | 手续费        | `0.00001000`                 |
| `fee_asset`       | VARCHAR   | 10   | ✗    | NULL   | 手续费币种    | `BTC`                        |
| `client_order_id` | VARCHAR   | 64   | ✗    | NULL   | 客户端订单 ID | `my-order-123`               |
| `created_at`      | TIMESTAMP | -    | ✓    | NOW()  | 创建时间      | `2024-11-01 10:00:00`        |
| `updated_at`      | TIMESTAMP | -    | ✓    | NOW()  | 更新时间      | `2024-11-01 10:00:01`        |
| `filled_at`       | TIMESTAMP | -    | ✗    | NULL   | 完全成交时间  | `2024-11-01 10:00:01`        |
| `canceled_at`     | TIMESTAMP | -    | ✗    | NULL   | 撤销时间      | `2024-11-01 10:05:00`        |
| `metadata`        | JSONB     | -    | ✗    | NULL   | 扩展信息      | `{"ip":"1.2.3.4"}`           |

**订单状态流转**:

```
状态枚举: new, open, partially_filled, filled, canceled, rejected

状态流转图:
                     ┌──────┐
                     │ new  │ (订单创建)
                     └──┬───┘
                        │
           ┌────────────┼────────────┐
           │            │            │
           ▼            ▼            ▼
      ┌─────────┐  ┌──────┐    ┌──────────┐
      │rejected │  │ open │    │  filled  │ (市价单直接成交)
      └─────────┘  └──┬───┘    └──────────┘
                      │
           ┌──────────┼──────────┐
           │          │          │
           ▼          ▼          ▼
    ┌──────────┐ ┌────────┐ ┌──────────┐
    │ canceled │ │partially│ │  filled  │
    └──────────┘ │_filled  │ └──────────┘
                 └────┬────┘
                      │
                 ┌────┴────┐
                 ▼         ▼
            ┌──────────┐ ┌──────────┐
            │ canceled │ │  filled  │
            └──────────┘ └──────────┘

终态: filled, canceled, rejected
```

**业务规则**:

- 限价单 `price` 必填，市价单 `price` 为 NULL
- `filled` 不能超过 `amount`
- 订单创建后 `status = 'new'`
- 市价单立即撮合，成交后 `status = 'filled'`
- 限价单未立即成交则 `status = 'open'`
- 手续费率：0.1% (taker), 0.05% (maker) - 配置化

**触发器**:

```sql
CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

### 2.4 trades (成交表)

**用途**: 记录所有成交明细

**数据量级**: 大 (> 100K)

**查询频率**: 高

**表结构**:

```sql
CREATE TABLE trades (
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders(id),
    user_id     INTEGER NOT NULL REFERENCES users(id),
    symbol      VARCHAR(20) NOT NULL,
    side        VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    price       DECIMAL(20, 8) NOT NULL CHECK (price > 0),
    amount      DECIMAL(20, 8) NOT NULL CHECK (amount > 0),
    quote_amount DECIMAL(20, 8) NOT NULL,
    fee         DECIMAL(20, 8) DEFAULT 0,
    fee_asset   VARCHAR(10),
    is_maker    BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata    JSONB
);

-- 注释
COMMENT ON TABLE trades IS '成交记录表';
COMMENT ON COLUMN trades.order_id IS '关联订单ID';
COMMENT ON COLUMN trades.quote_amount IS '成交金额 (price * amount)';
COMMENT ON COLUMN trades.is_maker IS '是否为 Maker 订单';
```

**字段详细说明**:

| 字段名         | 类型      | 精度 | 必填 | 默认值  | 说明       | 示例值                |
| -------------- | --------- | ---- | ---- | ------- | ---------- | --------------------- |
| `id`           | BIGSERIAL | -    | ✓    | AUTO    | 成交 ID    | `2000001`             |
| `order_id`     | BIGINT    | -    | ✓    | -       | 订单 ID    | `1000001`             |
| `user_id`      | INTEGER   | -    | ✓    | -       | 用户 ID    | `1001`                |
| `symbol`       | VARCHAR   | 20   | ✓    | -       | 交易对     | `BTC/USDT`            |
| `side`         | VARCHAR   | 4    | ✓    | -       | 方向       | `buy`, `sell`         |
| `price`        | DECIMAL   | 20,8 | ✓    | -       | 成交价格   | `50000.00000000`      |
| `amount`       | DECIMAL   | 20,8 | ✓    | -       | 成交数量   | `0.01000000`          |
| `quote_amount` | DECIMAL   | 20,8 | ✓    | -       | 成交额     | `500.00000000`        |
| `fee`          | DECIMAL   | 20,8 | ✓    | `0`     | 手续费     | `0.50000000`          |
| `fee_asset`    | VARCHAR   | 10   | ✗    | NULL    | 手续费币种 | `USDT`                |
| `is_maker`     | BOOLEAN   | -    | ✓    | `FALSE` | 是否挂单方 | `true`, `false`       |
| `created_at`   | TIMESTAMP | -    | ✓    | NOW()   | 成交时间   | `2024-11-01 10:00:01` |
| `metadata`     | JSONB     | -    | ✗    | NULL    | 扩展信息   | `{}`                  |

**业务规则**:

- `quote_amount = price * amount`
- `fee` 根据 Maker/Taker 不同计费
- Maker fee: 0.05% (挂单方)
- Taker fee: 0.1% (吃单方)
- 成交记录为不可变数据，只增不改

**计算示例**:

```sql
-- 买单成交示例
INSERT INTO trades (
    order_id, user_id, symbol, side,
    price, amount, quote_amount,
    fee, fee_asset, is_maker
) VALUES (
    1000001, 1001, 'BTC/USDT', 'buy',
    50000.00000000, 0.01000000, 500.00000000,
    0.00001000, 'BTC', FALSE  -- Taker: 0.1% of amount
);

-- 卖单成交示例
INSERT INTO trades (
    order_id, user_id, symbol, side,
    price, amount, quote_amount,
    fee, fee_asset, is_maker
) VALUES (
    1000002, 1002, 'BTC/USDT', 'sell',
    50000.00000000, 0.01000000, 500.00000000,
    0.50000000, 'USDT', FALSE  -- Taker: 0.1% of quote_amount
);
```

---

### 2.5 tickers (行情表)

**用途**: 存储交易对的实时行情数据

**数据量级**: 极小 (< 100)

**查询频率**: 极高

**表结构**:

```sql
CREATE TABLE tickers (
    symbol          VARCHAR(20) PRIMARY KEY,
    last_price      DECIMAL(20, 8) NOT NULL,
    bid_price       DECIMAL(20, 8),
    ask_price       DECIMAL(20, 8),
    high_24h        DECIMAL(20, 8),
    low_24h         DECIMAL(20, 8),
    volume_24h_base DECIMAL(20, 8),
    volume_24h_quote DECIMAL(20, 8),
    price_change_24h DECIMAL(20, 8),
    price_change_percent_24h DECIMAL(10, 4),
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    source          VARCHAR(20) DEFAULT 'binance',
    metadata        JSONB
);

-- 注释
COMMENT ON TABLE tickers IS '交易对行情数据表';
COMMENT ON COLUMN tickers.last_price IS '最新成交价';
COMMENT ON COLUMN tickers.bid_price IS '买一价';
COMMENT ON COLUMN tickers.ask_price IS '卖一价';
COMMENT ON COLUMN tickers.volume_24h_base IS '24小时成交量(基础币)';
COMMENT ON COLUMN tickers.volume_24h_quote IS '24小时成交额(计价币)';
```

**字段详细说明**:

| 字段名                     | 类型      | 精度 | 必填 | 默认值    | 说明          | 示例值                |
| -------------------------- | --------- | ---- | ---- | --------- | ------------- | --------------------- |
| `symbol`                   | VARCHAR   | 20   | ✓    | -         | 交易对 (主键) | `BTC/USDT`            |
| `last_price`               | DECIMAL   | 20,8 | ✓    | -         | 最新价        | `50000.00000000`      |
| `bid_price`                | DECIMAL   | 20,8 | ✗    | NULL      | 买一价        | `49950.00000000`      |
| `ask_price`                | DECIMAL   | 20,8 | ✗    | NULL      | 卖一价        | `50050.00000000`      |
| `high_24h`                 | DECIMAL   | 20,8 | ✗    | NULL      | 24h 最高价    | `51000.00000000`      |
| `low_24h`                  | DECIMAL   | 20,8 | ✗    | NULL      | 24h 最低价    | `49000.00000000`      |
| `volume_24h_base`          | DECIMAL   | 20,8 | ✗    | NULL      | 24h 成交量    | `1234.56789000`       |
| `volume_24h_quote`         | DECIMAL   | 20,8 | ✗    | NULL      | 24h 成交额    | `61234567.89000000`   |
| `price_change_24h`         | DECIMAL   | 20,8 | ✗    | NULL      | 24h 价格变化  | `1000.00000000`       |
| `price_change_percent_24h` | DECIMAL   | 10,4 | ✗    | NULL      | 24h 涨跌幅(%) | `2.0400`              |
| `updated_at`               | TIMESTAMP | -    | ✓    | NOW()     | 更新时间      | `2024-11-01 10:00:00` |
| `source`                   | VARCHAR   | 20   | ✓    | `binance` | 数据源        | `binance`             |
| `metadata`                 | JSONB     | -    | ✗    | NULL      | 扩展信息      | `{"open":"49000"}`    |

**业务规则**:

- 数据源默认为 Binance
- 更新频率：每秒更新一次 (可配置)
- 使用 `UPSERT` 模式更新数据
- 建议搭配内存缓存使用

**UPSERT 示例**:

```sql
INSERT INTO tickers (
    symbol, last_price, bid_price, ask_price,
    high_24h, low_24h, volume_24h_base, volume_24h_quote,
    price_change_24h, price_change_percent_24h,
    updated_at, source
)
VALUES (
    'BTC/USDT', 50000.00000000, 49950.00000000, 50050.00000000,
    51000.00000000, 49000.00000000, 1234.56789000, 61234567.89000000,
    1000.00000000, 2.0400,
    CURRENT_TIMESTAMP, 'binance'
)
ON CONFLICT (symbol)
DO UPDATE SET
    last_price = EXCLUDED.last_price,
    bid_price = EXCLUDED.bid_price,
    ask_price = EXCLUDED.ask_price,
    high_24h = EXCLUDED.high_24h,
    low_24h = EXCLUDED.low_24h,
    volume_24h_base = EXCLUDED.volume_24h_base,
    volume_24h_quote = EXCLUDED.volume_24h_quote,
    price_change_24h = EXCLUDED.price_change_24h,
    price_change_percent_24h = EXCLUDED.price_change_percent_24h,
    updated_at = CURRENT_TIMESTAMP,
    source = EXCLUDED.source;
```

---

## 3. 索引设计

### 3.1 索引策略

```
索引原则:
  ✓ 为外键创建索引
  ✓ 为高频查询字段创建索引
  ✓ 为组合查询创建复合索引
  ✓ 避免过度索引影响写入性能
```

### 3.2 索引清单

#### 3.2.1 users 表索引

```sql
-- 主键索引（自动创建）
-- PRIMARY KEY (id)

-- 唯一索引
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE UNIQUE INDEX idx_users_api_key ON users(api_key);

-- 普通索引
CREATE INDEX idx_users_status ON users(status) WHERE status = 'active';
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

**索引说明**:

| 索引名                 | 类型    | 字段              | 用途             |
| ---------------------- | ------- | ----------------- | ---------------- |
| `users_pkey`           | PRIMARY | `id`              | 主键查询         |
| `idx_users_email`      | UNIQUE  | `email`           | 登录查询、防重复 |
| `idx_users_api_key`    | UNIQUE  | `api_key`         | API 认证         |
| `idx_users_status`     | PARTIAL | `status`          | 查询活跃用户     |
| `idx_users_created_at` | BTREE   | `created_at DESC` | 按时间排序       |

#### 3.2.2 balances 表索引

```sql
-- 主键索引（自动创建）
-- PRIMARY KEY (id)

-- 唯一索引（自动创建）
-- UNIQUE (user_id, asset)

-- 外键索引
CREATE INDEX idx_balances_user_id ON balances(user_id);

-- 复合索引
CREATE INDEX idx_balances_user_asset ON balances(user_id, asset);

-- 查询优化索引
CREATE INDEX idx_balances_updated_at ON balances(updated_at DESC);
```

**索引说明**:

| 索引名                    | 类型    | 字段             | 用途             |
| ------------------------- | ------- | ---------------- | ---------------- |
| `balances_pkey`           | PRIMARY | `id`             | 主键查询         |
| `idx_balances_user_id`    | BTREE   | `user_id`        | 查询用户余额     |
| `idx_balances_user_asset` | BTREE   | `user_id, asset` | 精确查询特定币种 |

#### 3.2.3 orders 表索引

```sql
-- 主键索引（自动创建）
-- PRIMARY KEY (id)

-- 外键索引
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- 高频查询索引
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);

-- 复合索引（重要！）
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_orders_symbol_status ON orders(symbol, status);
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);

-- 部分索引（性能优化）
CREATE INDEX idx_orders_open ON orders(symbol, side, price)
    WHERE status = 'open';

-- 客户端订单ID索引
CREATE INDEX idx_orders_client_order_id ON orders(client_order_id)
    WHERE client_order_id IS NOT NULL;

-- 时间范围查询索引
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_updated_at ON orders(updated_at DESC);
```

**索引说明**:

| 索引名                     | 类型    | 字段                  | 用途             |
| -------------------------- | ------- | --------------------- | ---------------- |
| `orders_pkey`              | PRIMARY | `id`                  | 主键查询         |
| `idx_orders_user_id`       | BTREE   | `user_id`             | 查询用户订单     |
| `idx_orders_user_status`   | BTREE   | `user_id, status`     | 查询特定状态订单 |
| `idx_orders_symbol_status` | BTREE   | `symbol, status`      | 查询交易对订单薄 |
| `idx_orders_open`          | PARTIAL | `symbol, side, price` | 撮合引擎查询     |
| `idx_orders_created_at`    | BTREE   | `created_at DESC`     | 历史订单查询     |

#### 3.2.4 trades 表索引

```sql
-- 主键索引（自动创建）
-- PRIMARY KEY (id)

-- 外键索引
CREATE INDEX idx_trades_order_id ON trades(order_id);
CREATE INDEX idx_trades_user_id ON trades(user_id);

-- 复合索引
CREATE INDEX idx_trades_user_created ON trades(user_id, created_at DESC);
CREATE INDEX idx_trades_symbol_created ON trades(symbol, created_at DESC);

-- 时间索引（重要！用于统计）
CREATE INDEX idx_trades_created_at ON trades(created_at DESC);

-- 统计查询索引
CREATE INDEX idx_trades_symbol_side_created
    ON trades(symbol, side, created_at DESC);
```

**索引说明**:

| 索引名                      | 类型    | 字段                       | 用途             |
| --------------------------- | ------- | -------------------------- | ---------------- |
| `trades_pkey`               | PRIMARY | `id`                       | 主键查询         |
| `idx_trades_order_id`       | BTREE   | `order_id`                 | 查询订单成交明细 |
| `idx_trades_user_created`   | BTREE   | `user_id, created_at DESC` | 用户成交历史     |
| `idx_trades_symbol_created` | BTREE   | `symbol, created_at DESC`  | 市场成交记录     |
| `idx_trades_created_at`     | BTREE   | `created_at DESC`          | 时序查询         |

#### 3.2.5 tickers 表索引

```sql
-- 主键索引（自动创建）
-- PRIMARY KEY (symbol)

-- 时间索引
CREATE INDEX idx_tickers_updated_at ON tickers(updated_at DESC);

-- 数据源索引
CREATE INDEX idx_tickers_source ON tickers(source);
```

### 3.3 索引维护

```sql
-- 查看索引大小
SELECT
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_indexes
JOIN pg_class ON pg_indexes.indexname = pg_class.relname
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;

-- 查看未使用的索引
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0
    AND indexrelname NOT LIKE 'pg_toast%'
ORDER BY pg_relation_size(indexrelid) DESC;

-- 重建索引（定期维护）
REINDEX TABLE orders;
REINDEX TABLE trades;
```

---

## 4. 数据关系图

### 4.1 ER 图（实体关系）

```
┌─────────────────────────────────────────────────────────────────┐
│                      实体关系图 (ERD)                           │
└─────────────────────────────────────────────────────────────────┘

        ┌──────────────────────┐
        │       users          │
        ├──────────────────────┤
        │ PK  id               │
        │     email     UNIQUE │
        │     api_key   UNIQUE │
        │     api_secret       │
        │     status           │
        │     created_at       │
        └──────────┬───────────┘
                   │
                   │ 1
                   │
                   │
        ┌──────────┼───────────────────────┐
        │          │                       │
        │ N        │ N                     │ N
        │          │                       │
┌───────▼──────┐   │              ┌────────▼──────┐
│   balances   │   │              │    orders     │
├──────────────┤   │              ├───────────────┤
│ PK  id       │   │              │ PK  id        │
│ FK  user_id  │   │              │ FK  user_id   │
│     asset    │   │              │     symbol    │
│     available│   │              │     side      │
│     locked   │   │              │     type      │
└──────────────┘   │              │     status    │
                   │              │     amount    │
                   │              │     filled    │
                   │              └────────┬──────┘
                   │                       │
                   │                       │ 1
                   │                       │
                   │                       │
                   │                       │ N
                   │                       │
                   │              ┌────────▼──────┐
                   │              │    trades     │
                   │              ├───────────────┤
                   │              │ PK  id        │
                   │              │ FK  order_id  │
                   └──────────────│ FK  user_id   │
                                  │     symbol    │
                                  │     price     │
                                  │     amount    │
                                  └───────────────┘

独立表:
┌──────────────────┐
│     tickers      │
├──────────────────┤
│ PK  symbol       │
│     last_price   │
│     bid_price    │
│     ask_price    │
│     volume_24h   │
└──────────────────┘

关系说明:
  • users 1:N balances  (一个用户有多个币种余额)
  • users 1:N orders    (一个用户有多个订单)
  • users 1:N trades    (一个用户有多个成交)
  • orders 1:N trades   (一个订单可能有多笔成交)
  • tickers 独立表      (不与其他表关联)
```

### 4.2 外键约束

```sql
-- balances 外键
ALTER TABLE balances
    ADD CONSTRAINT fk_balances_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

-- orders 外键
ALTER TABLE orders
    ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE RESTRICT;  -- 不允许删除有订单的用户

-- trades 外键
ALTER TABLE trades
    ADD CONSTRAINT fk_trades_order
    FOREIGN KEY (order_id)
    REFERENCES orders(id)
    ON DELETE RESTRICT;  -- 不允许删除有成交的订单

ALTER TABLE trades
    ADD CONSTRAINT fk_trades_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE RESTRICT;
```

---

## 5. 字段命名规范

### 5.1 命名约定

```
表名:
  • 全小写
  • 复数形式
  • 使用下划线分隔
  例: users, balances, orders

字段名:
  • 全小写
  • 使用下划线分隔
  • 避免缩写
  例: user_id, created_at, average_price

主键:
  • 统一使用 id
  • 类型: SERIAL (小表), BIGSERIAL (大表)

外键:
  • {表名单数}_id
  例: user_id, order_id

时间戳:
  • created_at: 创建时间
  • updated_at: 更新时间
  • {action}_at: 特定动作时间
  例: filled_at, canceled_at

布尔值:
  • is_{描述}
  例: is_maker, is_active

JSONB:
  • metadata: 扩展字段
```

### 5.2 数据类型选择

| 用途        | 推荐类型               | 说明                    |
| ----------- | ---------------------- | ----------------------- |
| 主键 (小表) | `SERIAL`               | 自增整数，范围 1 ~ 2^31 |
| 主键 (大表) | `BIGSERIAL`            | 自增长整数，范围更大    |
| 外键        | `INTEGER` / `BIGINT`   | 与主键类型保持一致      |
| 金额/价格   | `DECIMAL(20,8)`        | 高精度小数，8 位精度    |
| 百分比      | `DECIMAL(10,4)`        | 4 位精度足够            |
| 字符串 (短) | `VARCHAR(N)`           | N < 255                 |
| 字符串 (长) | `TEXT`                 | 无长度限制              |
| 时间戳      | `TIMESTAMP`            | 不带时区 (应用层处理)   |
| 日期        | `DATE`                 | 只需日期时              |
| 布尔        | `BOOLEAN`              | true/false              |
| JSON        | `JSONB`                | 二进制 JSON，支持索引   |
| 枚举        | `VARCHAR(N)` + `CHECK` | 灵活性更好              |

---

## 6. 数据迁移策略

### 6.1 版本管理

```bash
# 使用 golang-migrate 或 GORM AutoMigrate

# 初始化
migrate create -ext sql -dir db/migrations -seq init_schema

# 执行迁移
migrate -path db/migrations -database "postgres://localhost/hypersim" up

# 回滚
migrate -path db/migrations -database "postgres://localhost/hypersim" down 1
```

### 6.2 迁移文件示例

```sql
-- 000001_init_schema.up.sql
BEGIN;

CREATE TABLE users (...);
CREATE TABLE balances (...);
CREATE TABLE orders (...);
CREATE TABLE trades (...);
CREATE TABLE tickers (...);

-- 创建索引
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
-- ... 其他索引

COMMIT;

-- 000001_init_schema.down.sql
BEGIN;

DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS tickers;
DROP TABLE IF EXISTS users;

COMMIT;
```

### 6.3 数据备份

```bash
# 全量备份
pg_dump -U postgres -d hypersim -F c -f hypersim_backup.dump

# 仅结构
pg_dump -U postgres -d hypersim -s -f schema.sql

# 仅数据
pg_dump -U postgres -d hypersim -a -f data.sql

# 恢复
pg_restore -U postgres -d hypersim hypersim_backup.dump
```

---

## 7. 性能优化建议

### 7.1 查询优化

```sql
-- ❌ 不推荐: 全表扫描
SELECT * FROM orders WHERE user_id = 1001;

-- ✅ 推荐: 使用索引 + 指定字段
SELECT id, symbol, side, amount, status, created_at
FROM orders
WHERE user_id = 1001 AND status IN ('open', 'partially_filled')
ORDER BY created_at DESC
LIMIT 100;

-- ❌ 不推荐: 函数破坏索引
SELECT * FROM trades WHERE DATE(created_at) = '2024-11-01';

-- ✅ 推荐: 范围查询
SELECT * FROM trades
WHERE created_at >= '2024-11-01 00:00:00'
  AND created_at < '2024-11-02 00:00:00';
```

### 7.2 连接池配置

```go
// GORM 配置示例
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, _ := db.DB()

// 连接池设置
sqlDB.SetMaxIdleConns(10)           // 最小空闲连接
sqlDB.SetMaxOpenConns(50)           // 最大打开连接
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
```

### 7.3 分区表策略 (v2.0)

```sql
-- orders 按月分区
CREATE TABLE orders_2024_11 PARTITION OF orders
    FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');

CREATE TABLE orders_2024_12 PARTITION OF orders
    FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');

-- 自动创建分区 (使用 pg_partman 扩展)
```

### 7.4 慢查询监控

```sql
-- 开启慢查询日志
ALTER DATABASE hypersim SET log_min_duration_statement = 1000; -- 1秒

-- 查看慢查询
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
WHERE mean_exec_time > 1000
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### 7.5 数据归档

```sql
-- 归档3个月前的历史订单
INSERT INTO orders_archive
SELECT * FROM orders
WHERE created_at < NOW() - INTERVAL '3 months'
  AND status IN ('filled', 'canceled');

-- 删除已归档数据
DELETE FROM orders
WHERE created_at < NOW() - INTERVAL '3 months'
  AND status IN ('filled', 'canceled');

-- 定期执行 VACUUM
VACUUM ANALYZE orders;
VACUUM ANALYZE trades;
```

---

## 附录

### A. 完整建表脚本

参见: `db/schema.sql`

### B. 测试数据脚本

参见: `db/seed.sql`

### C. 性能测试脚本

参见: `db/benchmark.sql`

### D. 变更日志

| 版本 | 日期       | 变更内容             |
| ---- | ---------- | -------------------- |
| v1.0 | 2024-11-01 | 初始版本，5 个核心表 |

---

**维护者**: Quicksilver Team  
**联系方式**: dev@quicksilver.local  
**最后审核**: 2024-11-01
