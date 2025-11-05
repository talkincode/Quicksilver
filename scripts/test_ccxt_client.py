#!/usr/bin/env python3
"""
CCXT 客户端集成测试脚本

验证 Quicksilver 与 CCXT 库的兼容性
测试所有公开和私有 API 端点的格式转换

运行方式:
    python scripts/test_ccxt_client.py
"""

import ccxt
import json
import time
from datetime import datetime


class QuicksilverTester:
    """Quicksilver CCXT 兼容性测试器"""

    def __init__(self, base_url="http://localhost:8080", api_key=None, api_secret=None):
        """
        初始化测试器

        Args:
            base_url: Quicksilver API 地址
            api_key: 用户 API Key (测试私有接口时必需)
            api_secret: 用户 API Secret (测试私有接口时必需)
        """
        self.exchange = ccxt.Exchange(
            {
                "id": "quicksilver",
                "name": "Quicksilver",
                "urls": {
                    "api": {
                        "public": base_url + "/v1",
                        "private": base_url + "/v1",
                    },
                },
                "has": {
                    "fetchMarkets": True,
                    "fetchTicker": True,
                    "fetchTrades": True,
                    "fetchBalance": True,
                    "createOrder": True,
                    "cancelOrder": True,
                    "fetchOrder": True,
                    "fetchOrders": True,
                    "fetchOpenOrders": True,
                    "fetchMyTrades": True,
                },
                "apiKey": api_key,
                "secret": api_secret,
                "enableRateLimit": False,
            }
        )

        self.results = {"passed": 0, "failed": 0, "errors": []}

    def log_test(self, test_name, success, message=""):
        """记录测试结果"""
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} | {test_name}")
        if not success:
            print(f"  └─ Error: {message}")
            self.results["errors"].append({"test": test_name, "error": message})
        if success:
            self.results["passed"] += 1
        else:
            self.results["failed"] += 1
        print()

    def test_server_time(self):
        """测试 GET /v1/time"""
        print("🔍 Testing: Server Time")
        try:
            response = self.exchange.publicGetTime()

            # 验证响应格式
            assert "timestamp" in response, "缺少 timestamp 字段"
            assert "datetime" in response, "缺少 datetime 字段"
            assert isinstance(response["timestamp"], int), "timestamp 类型错误"

            self.log_test("GET /v1/time", True)
            print(f"  Server Time: {response['datetime']}")
            return True
        except Exception as e:
            self.log_test("GET /v1/time", False, str(e))
            return False

    def test_fetch_markets(self):
        """测试 GET /v1/markets (fetchMarkets)"""
        print("🔍 Testing: Fetch Markets")
        try:
            response = self.exchange.publicGetMarkets()

            # 验证响应格式
            assert isinstance(response, list), "markets 应该是数组"
            assert len(response) > 0, "markets 不应为空"

            market = response[0]
            required_fields = ["id", "symbol", "base", "quote", "active", "limits"]
            for field in required_fields:
                assert field in market, f"缺少字段: {field}"

            self.log_test("GET /v1/markets", True)
            print(f"  Total Markets: {len(response)}")
            print(f"  Sample: {market['symbol']}")
            return True
        except Exception as e:
            self.log_test("GET /v1/markets", False, str(e))
            return False

    def test_fetch_ticker(self, symbol="BTC/USDT"):
        """测试 GET /v1/ticker/:symbol (fetchTicker)"""
        print(f"🔍 Testing: Fetch Ticker ({symbol})")
        try:
            # CCXT 格式: BTC/USDT, API 路径需要转换为 BTC-USDT
            url_symbol = symbol.replace("/", "-")
            response = self.exchange.publicGetTickerSymbol({"symbol": url_symbol})

            # 验证 CCXT 标准 Ticker 格式
            required_fields = [
                "symbol",
                "timestamp",
                "datetime",
                "high",
                "low",
                "bid",
                "ask",
                "last",
                "baseVolume",
                "quoteVolume",
            ]
            for field in required_fields:
                assert field in response, f"缺少字段: {field}"

            assert response["symbol"] == symbol, f"symbol 不匹配: {response['symbol']}"
            assert isinstance(response["timestamp"], int), "timestamp 类型错误"
            assert isinstance(response["last"], (int, float)), "last price 类型错误"

            self.log_test(f"GET /v1/ticker/{symbol}", True)
            print(f"  Last Price: {response['last']}")
            print(f"  24h Volume: {response['baseVolume']}")
            return True
        except Exception as e:
            self.log_test(f"GET /v1/ticker/{symbol}", False, str(e))
            return False

    def test_fetch_trades(self, symbol="BTC/USDT"):
        """测试 GET /v1/trades/:symbol (fetchTrades)"""
        print(f"🔍 Testing: Fetch Trades ({symbol})")
        try:
            url_symbol = symbol.replace("/", "-")
            response = self.exchange.publicGetTradesSymbol({"symbol": url_symbol})

            # 验证响应格式
            assert isinstance(response, list), "trades 应该是数组"

            if len(response) > 0:
                trade = response[0]
                required_fields = [
                    "id",
                    "timestamp",
                    "datetime",
                    "symbol",
                    "side",
                    "price",
                    "amount",
                ]
                for field in required_fields:
                    assert field in trade, f"缺少字段: {field}"

                assert trade["symbol"] == symbol, "symbol 不匹配"
                assert trade["side"] in ["buy", "sell"], "side 值错误"

            self.log_test(f"GET /v1/trades/{symbol}", True)
            print(f"  Total Trades: {len(response)}")
            return True
        except Exception as e:
            self.log_test(f"GET /v1/trades/{symbol}", False, str(e))
            return False

    def test_fetch_balance(self):
        """测试 GET /v1/balance (fetchBalance) - 需要认证"""
        print("🔍 Testing: Fetch Balance (Private)")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("GET /v1/balance", False, "缺少 API Key/Secret")
            return False

        try:
            response = self.exchange.privateGetBalance()

            # 验证 CCXT 标准 Balance 格式
            assert isinstance(response, dict), "balance 应该是对象"

            # CCXT 格式应包含 'free', 'used', 'total' 等字段
            for asset in response:
                if asset not in [
                    "info",
                    "free",
                    "used",
                    "total",
                    "timestamp",
                    "datetime",
                ]:
                    balance = response[asset]
                    assert "free" in balance, f"{asset} 缺少 free 字段"
                    assert "used" in balance, f"{asset} 缺少 used 字段"
                    assert "total" in balance, f"{asset} 缺少 total 字段"

            self.log_test("GET /v1/balance", True)
            print(
                f"  Assets: {len([k for k in response.keys() if k not in ['info', 'free', 'used', 'total', 'timestamp', 'datetime']])}"
            )
            return True
        except Exception as e:
            self.log_test("GET /v1/balance", False, str(e))
            return False

    def test_create_order(
        self,
        symbol="BTC/USDT",
        side="buy",
        order_type="limit",
        amount=0.001,
        price=50000,
    ):
        """测试 POST /v1/order (createOrder) - 需要认证"""
        print(f"🔍 Testing: Create Order ({side} {order_type})")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("POST /v1/order", False, "缺少 API Key/Secret")
            return False

        try:
            params = {
                "symbol": symbol,
                "side": side,
                "type": order_type,
                "amount": amount,
            }

            if order_type == "limit":
                params["price"] = price

            response = self.exchange.privatePostOrder(params)

            # 验证 CCXT 标准 Order 格式
            required_fields = [
                "id",
                "timestamp",
                "datetime",
                "symbol",
                "type",
                "side",
                "price",
                "amount",
                "status",
            ]
            for field in required_fields:
                assert field in response, f"缺少字段: {field}"

            assert response["symbol"] == symbol, "symbol 不匹配"
            assert response["side"] == side, "side 不匹配"
            assert response["type"] == order_type, "type 不匹配"

            self.log_test("POST /v1/order", True)
            print(f"  Order ID: {response['id']}")
            print(f"  Status: {response['status']}")
            return response["id"]
        except Exception as e:
            self.log_test("POST /v1/order", False, str(e))
            return None

    def test_fetch_order(self, order_id):
        """测试 GET /v1/order/:id (fetchOrder) - 需要认证"""
        print(f"🔍 Testing: Fetch Order (ID: {order_id})")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("GET /v1/order/:id", False, "缺少 API Key/Secret")
            return False

        try:
            response = self.exchange.privateGetOrderId({"id": order_id})

            # 验证格式
            assert "id" in response, "缺少 id 字段"
            assert str(response["id"]) == str(order_id), "订单 ID 不匹配"

            self.log_test(f"GET /v1/order/{order_id}", True)
            print(f"  Status: {response.get('status', 'N/A')}")
            return True
        except Exception as e:
            self.log_test(f"GET /v1/order/{order_id}", False, str(e))
            return False

    def test_cancel_order(self, order_id):
        """测试 DELETE /v1/order/:id (cancelOrder) - 需要认证"""
        print(f"🔍 Testing: Cancel Order (ID: {order_id})")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("DELETE /v1/order/:id", False, "缺少 API Key/Secret")
            return False

        try:
            response = self.exchange.privateDeleteOrderId({"id": order_id})

            # 验证格式
            assert "id" in response, "缺少 id 字段"
            assert response.get("status") in [
                "cancelled",
                "canceled",
            ], "状态应为 cancelled"

            self.log_test(f"DELETE /v1/order/{order_id}", True)
            return True
        except Exception as e:
            self.log_test(f"DELETE /v1/order/{order_id}", False, str(e))
            return False

    def test_fetch_orders(self, symbol="BTC/USDT"):
        """测试 GET /v1/orders (fetchOrders) - 需要认证"""
        print(f"🔍 Testing: Fetch Orders ({symbol})")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("GET /v1/orders", False, "缺少 API Key/Secret")
            return False

        try:
            response = self.exchange.privateGetOrders({"symbol": symbol})

            # 验证格式
            assert isinstance(response, list), "orders 应该是数组"

            self.log_test("GET /v1/orders", True)
            print(f"  Total Orders: {len(response)}")
            return True
        except Exception as e:
            self.log_test("GET /v1/orders", False, str(e))
            return False

    def test_fetch_open_orders(self, symbol="BTC/USDT"):
        """测试 GET /v1/orders/open (fetchOpenOrders) - 需要认证"""
        print(f"🔍 Testing: Fetch Open Orders ({symbol})")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("GET /v1/orders/open", False, "缺少 API Key/Secret")
            return False

        try:
            response = self.exchange.privateGetOrdersOpen({"symbol": symbol})

            # 验证格式
            assert isinstance(response, list), "orders 应该是数组"

            self.log_test("GET /v1/orders/open", True)
            print(f"  Open Orders: {len(response)}")
            return True
        except Exception as e:
            self.log_test("GET /v1/orders/open", False, str(e))
            return False

    def test_fetch_my_trades(self, symbol="BTC/USDT"):
        """测试 GET /v1/myTrades (fetchMyTrades) - 需要认证"""
        print(f"🔍 Testing: Fetch My Trades ({symbol})")

        if not self.exchange.apiKey or not self.exchange.secret:
            self.log_test("GET /v1/myTrades", False, "缺少 API Key/Secret")
            return False

        try:
            response = self.exchange.privateGetMyTrades({"symbol": symbol})

            # 验证格式
            assert isinstance(response, list), "trades 应该是数组"

            self.log_test("GET /v1/myTrades", True)
            print(f"  My Trades: {len(response)}")
            return True
        except Exception as e:
            self.log_test("GET /v1/myTrades", False, str(e))
            return False

    def run_all_tests(self):
        """运行所有测试"""
        print("=" * 60)
        print("  Quicksilver CCXT 兼容性测试")
        print("=" * 60)
        print()

        # 公开 API 测试
        print("📂 Public API Tests")
        print("-" * 60)
        self.test_server_time()
        self.test_fetch_markets()
        self.test_fetch_ticker("BTC/USDT")
        self.test_fetch_ticker("ETH/USDT")
        self.test_fetch_trades("BTC/USDT")
        print()

        # 私有 API 测试
        if self.exchange.apiKey and self.exchange.secret:
            print("📂 Private API Tests (Authenticated)")
            print("-" * 60)
            self.test_fetch_balance()

            # 创建订单 -> 查询 -> 撤销流程
            order_id = self.test_create_order(
                symbol="BTC/USDT",
                side="buy",
                order_type="limit",
                amount=0.001,
                price=50000,
            )

            if order_id:
                time.sleep(0.5)  # 等待订单创建
                self.test_fetch_order(order_id)
                self.test_cancel_order(order_id)

            self.test_fetch_orders("BTC/USDT")
            self.test_fetch_open_orders("BTC/USDT")
            self.test_fetch_my_trades("BTC/USDT")
        else:
            print("⚠️  Skipping Private API Tests (No API Key provided)")

        print()
        print("=" * 60)
        print("  测试结果汇总")
        print("=" * 60)
        print(f"✅ Passed: {self.results['passed']}")
        print(f"❌ Failed: {self.results['failed']}")
        print(
            f"📊 Success Rate: {self.results['passed'] / (self.results['passed'] + self.results['failed']) * 100:.1f}%"
        )

        if self.results["errors"]:
            print()
            print("❌ Failed Tests:")
            for error in self.results["errors"]:
                print(f"  - {error['test']}: {error['error']}")

        print("=" * 60)

        return self.results["failed"] == 0


def main():
    """主函数"""
    import argparse

    parser = argparse.ArgumentParser(description="Quicksilver CCXT 兼容性测试")
    parser.add_argument(
        "--url", default="http://localhost:8080", help="Quicksilver API 地址"
    )
    parser.add_argument("--api-key", help="API Key (测试私有接口)")
    parser.add_argument("--api-secret", help="API Secret (测试私有接口)")

    args = parser.parse_args()

    # 创建测试器
    tester = QuicksilverTester(
        base_url=args.url, api_key=args.api_key, api_secret=args.api_secret
    )

    # 运行测试
    success = tester.run_all_tests()

    # 退出码
    exit(0 if success else 1)


if __name__ == "__main__":
    main()
