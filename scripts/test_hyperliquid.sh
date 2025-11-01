#!/bin/bash
# 测试 Hyperliquid API 连接

echo "🧪 测试 Hyperliquid API 连接..."
echo ""

# 测试 allMids 端点
echo "📊 获取 BTC 和 ETH 价格..."
curl -X POST https://api.hyperliquid.xyz/info \
-H "Content-Type: application/json" \
-d '{"type":"allMids"}' \
2>/dev/null | python3 -m json.tool | grep -E '"BTC"|"ETH"' | head -5

echo ""
echo "✅ 测试完成"
