# K 线数据 API 文档

## 概述

K 线（蜡烛图）数据提供了市场价格在指定时间周期内的开盘价(Open)、最高价(High)、最低价(Low)、收盘价(Close)和成交量(Volume)信息，是技术分析的重要数据来源。

---

## API 端点

### 获取 OHLCV 数据

```
GET /v1/ohlcv/:symbol
```

**描述**: 获取指定交易对的 K 线数据（CCXT 标准格式）

**路径参数**:

- `symbol` (string, required): 交易对符号，支持格式：
  - `BTC/USDT`
  - `BTC-USDT` (自动转换为 `BTC/USDT`)

**查询参数**:

- `timeframe` (string, optional): 时间周期，默认 `1h`
  - 支持的值: `1m`, `5m`, `15m`, `1h`, `4h`, `1d`
- `limit` (integer, optional): 返回数量，默认 `100`
  - 范围: 1 - 1000
  - 超过 1000 会被限制为 1000
- `since` (integer, optional): 开始时间（Unix 毫秒时间戳）
  - 只返回此时间之后的 K 线数据

---

## 响应格式

### CCXT 标准 OHLCV 格式

返回一个二维数组，每个元素代表一根 K 线：

```json
[
  [
    1704096000000, // timestamp: Unix 毫秒时间戳
    50000.0, // open: 开盘价
    51000.0, // high: 最高价
    49500.0, // low: 最低价
    50500.0, // close: 收盘价
    123.456 // volume: 成交量
  ],
  [1704099600000, 50500.0, 52000.0, 50000.0, 51500.0, 156.789]
]
```

**字段说明**:

1. **timestamp**: K 线开始时间（Unix 毫秒时间戳）
2. **open**: 开盘价（该时间周期的第一笔交易价格）
3. **high**: 最高价（该时间周期内的最高交易价格）
4. **low**: 最低价（该时间周期内的最低交易价格）
5. **close**: 收盘价（该时间周期的最后一笔交易价格）
6. **volume**: 成交量（该时间周期内的总交易量）

---

## 使用示例

### cURL 示例

#### 1. 获取 BTC/USDT 1 小时 K 线数据

```bash
curl -X GET "http://localhost:8080/v1/ohlcv/BTC/USDT?timeframe=1h&limit=100"
```

#### 2. 获取最近 50 根 5 分钟 K 线

```bash
curl -X GET "http://localhost:8080/v1/ohlcv/BTC-USDT?timeframe=5m&limit=50"
```

#### 3. 获取指定时间之后的 K 线数据

```bash
# 获取 2024-01-01 00:00:00 之后的数据
curl -X GET "http://localhost:8080/v1/ohlcv/BTC/USDT?timeframe=1h&since=1704067200000&limit=100"
```

---

### Python (CCXT) 示例

```python
import ccxt

# 初始化交易所
exchange = ccxt.Exchange({
    'id': 'quicksilver',
    'urls': {
        'api': {
            'public': 'http://localhost:8080/v1',
        }
    },
    'has': {
        'fetchOHLCV': True,
    },
})

# 获取K线数据
ohlcv = exchange.fetch('/ohlcv/BTC/USDT', params={
    'timeframe': '1h',
    'limit': 100
})

# 处理数据
for candle in ohlcv:
    timestamp, open, high, low, close, volume = candle
    print(f"Time: {timestamp}, O: {open}, H: {high}, L: {low}, C: {close}, V: {volume}")
```

---

### JavaScript 示例

```javascript
// 使用 fetch API
async function fetchKlines(symbol, timeframe = "1h", limit = 100) {
  const url = `http://localhost:8080/v1/ohlcv/${symbol}?timeframe=${timeframe}&limit=${limit}`;

  const response = await fetch(url);
  const klines = await response.json();

  return klines.map(([timestamp, open, high, low, close, volume]) => ({
    timestamp: new Date(timestamp),
    open,
    high,
    low,
    close,
    volume,
  }));
}

// 使用示例
const klines = await fetchKlines("BTC/USDT", "1h", 50);
console.log(klines);
```

---

## 时间周期说明

| Timeframe | 中文名称 | 更新频率   | 适用场景                 |
| --------- | -------- | ---------- | ------------------------ |
| `1m`      | 1 分钟   | 每 1 分钟  | 超短线交易、快速波动监控 |
| `5m`      | 5 分钟   | 每 5 分钟  | 短线交易、日内趋势分析   |
| `15m`     | 15 分钟  | 每 15 分钟 | 短线交易、关键支撑阻力位 |
| `1h`      | 1 小时   | 每 1 小时  | 中短线交易、趋势确认     |
| `4h`      | 4 小时   | 每 4 小时  | 波段交易、中期趋势分析   |
| `1d`      | 1 天     | 每 24 小时 | 长线投资、宏观趋势判断   |

---

## 数据来源

当前 K 线数据来源于 **Hyperliquid** API，支持以下交易对：

- BTC/USDT
- ETH/USDT
- SOL/USDT
- (可通过配置文件添加更多交易对)

---

## 数据更新机制

K 线数据采用**定时自动更新**策略：

- **1 分钟周期**: 每 1 分钟更新一次
- **5 分钟周期**: 每 5 分钟更新一次
- **15 分钟周期**: 每 15 分钟更新一次
- **1 小时周期**: 每 1 小时更新一次
- **4 小时周期**: 每 4 小时更新一次
- **1 天周期**: 每 24 小时更新一次

系统会在后台自动同步最新数据到数据库，API 查询时直接从数据库读取。

---

## 数据存储

K 线数据存储在 `klines` 表中：

```sql
CREATE TABLE klines (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open_time TIMESTAMP NOT NULL,
    close_time TIMESTAMP NOT NULL,
    open DECIMAL(20,8) NOT NULL,
    high DECIMAL(20,8) NOT NULL,
    low DECIMAL(20,8) NOT NULL,
    close DECIMAL(20,8) NOT NULL,
    volume DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 复合索引：优化查询性能
CREATE INDEX idx_symbol_interval_time ON klines(symbol, interval, open_time);
```

---

## 错误处理

### 常见错误

#### 1. 交易对不存在

```json
{
  "error": "failed to fetch klines: no data found"
}
```

**原因**: 请求的交易对不在支持列表中  
**解决**: 检查 `/v1/markets` 端点获取支持的交易对列表

#### 2. 无效的时间周期

```json
{
  "error": "invalid timeframe"
}
```

**原因**: `timeframe` 参数值不在支持范围内  
**解决**: 使用支持的值：`1m`, `5m`, `15m`, `1h`, `4h`, `1d`

#### 3. 数量超限

自动限制为最大 1000 条，不会返回错误。

---

## 性能优化建议

1. **合理设置 limit**: 根据实际需求设置合理的数量限制

   - 图表展示通常 100-200 条足够
   - 技术分析计算可能需要更多历史数据

2. **使用 since 参数**: 增量获取新数据，避免重复传输

   ```bash
   # 获取最新数据
   curl "http://localhost:8080/v1/ohlcv/BTC/USDT?timeframe=1h&since=1704096000000"
   ```

3. **选择合适的时间周期**:

   - 短期分析: 使用 1m, 5m, 15m
   - 中期分析: 使用 1h, 4h
   - 长期分析: 使用 1d

4. **客户端缓存**: 对于历史 K 线数据，建议在客户端缓存

---

## 完整示例

### 构建交易图表

```python
import ccxt
import pandas as pd
import matplotlib.pyplot as plt

# 获取K线数据
exchange = ccxt.Exchange({
    'id': 'quicksilver',
    'urls': {'api': {'public': 'http://localhost:8080/v1'}},
})

ohlcv = exchange.fetch('/ohlcv/BTC/USDT', params={
    'timeframe': '1h',
    'limit': 100
})

# 转换为 DataFrame
df = pd.DataFrame(ohlcv, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
df['timestamp'] = pd.to_datetime(df['timestamp'], unit='ms')
df.set_index('timestamp', inplace=True)

# 绘制K线图
import mplfinance as mpf
mpf.plot(df, type='candle', volume=True, title='BTC/USDT 1H')
```

---

## 相关 API

- [获取市场列表](/v1/markets): 查看所有支持的交易对
- [获取实时行情](/v1/ticker/:symbol): 获取最新价格和 24h 统计
- [获取成交记录](/v1/trades/:symbol): 查看最近的成交历史

---

## 技术支持

如有问题或建议，请参考：

- GitHub Issues
- 技术文档: `/docs/`
- API 测试文件: `apitest_kline.http`
