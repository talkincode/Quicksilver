"""Quicksilver API 客户端"""

from typing import Optional, Dict, Any, List
import requests


class QuicksilverAPI:
    """Quicksilver API 客户端"""

    def __init__(self, base_url: str, api_key: str, api_secret: str):
        """
        初始化 API 客户端

        Args:
            base_url: API 基础 URL
            api_key: API Key
            api_secret: API Secret
        """
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.api_secret = api_secret
        self.session = requests.Session()

    def _sign_request(self, method: str, path: str, body: str = "") -> Dict[str, str]:
        """
        生成请求头（使用简单的 API Key + Secret 认证）

        Args:
            method: HTTP 方法
            path: 请求路径
            body: 请求体（JSON 字符串）

        Returns:
            包含认证信息的请求头
        """
        return {
            "X-API-Key": self.api_key,
            "X-API-Secret": self.api_secret,
            "Content-Type": "application/json",
        }

    def _request(
        self,
        method: str,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        json: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """
        发送 HTTP 请求

        Args:
            method: HTTP 方法
            path: 请求路径
            params: 查询参数
            json: 请求体（自动转换为 JSON）

        Returns:
            响应数据

        Raises:
            requests.HTTPError: HTTP 错误
        """
        url = f"{self.base_url}{path}"
        body = ""
        if json:
            import json as json_lib

            body = json_lib.dumps(json)

        headers = self._sign_request(method, path, body)

        response = self.session.request(
            method=method,
            url=url,
            headers=headers,
            params=params,
            json=json,
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    # ========== 健康检查 ==========

    def health_check(self) -> Dict[str, str]:
        """健康检查"""
        return self._request("GET", "/health")

    # ========== 用户管理 ==========

    def get_users(
        self,
        page: int = 1,
        limit: int = 20,
        search: Optional[str] = None,
        status: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        获取用户列表

        Args:
            page: 页码（从 1 开始）
            limit: 每页数量
            search: 搜索关键字（邮箱或 API Key）
            status: 状态过滤（active/inactive/suspended）

        Returns:
            用户列表数据
        """
        params = {"page": page, "limit": limit}
        if search:
            params["search"] = search
        if status:
            params["status"] = status

        return self._request("GET", "/v1/admin/users", params=params)

    def create_user(self, email: str, username: Optional[str] = None) -> Dict[str, Any]:
        """
        创建新用户

        Args:
            email: 用户邮箱
            username: 用户名（可选）

        Returns:
            创建的用户信息（包含 API Key 和 Secret）
        """
        json = {"email": email}
        if username:
            json["username"] = username

        return self._request("POST", "/v1/admin/users", json=json)

    def get_user(self, user_id: int) -> Dict[str, Any]:
        """获取用户详情"""
        return self._request("GET", f"/v1/admin/users/{user_id}")

    def update_user(
        self,
        user_id: int,
        status: Optional[str] = None,
        regenerate_api_key: bool = False,
    ) -> Dict[str, Any]:
        """
        更新用户信息

        Args:
            user_id: 用户 ID
            status: 新状态
            regenerate_api_key: 是否重新生成 API Key

        Returns:
            更新后的用户信息
        """
        json = {}
        if status:
            json["status"] = status
        if regenerate_api_key:
            json["regenerate_api_key"] = True

        return self._request("PUT", f"/v1/admin/users/{user_id}", json=json)

    def delete_user(self, user_id: int) -> Dict[str, str]:
        """删除用户（软删除）"""
        return self._request("DELETE", f"/v1/admin/users/{user_id}")

    # ========== 行情数据 ==========

    def get_ticker(self, symbol: str) -> Dict[str, Any]:
        """
        获取行情数据

        Args:
            symbol: 交易对（如 BTC/USDT）

        Returns:
            行情数据
        """
        return self._request("GET", f"/v1/ticker/{symbol}")

    def get_markets(self) -> List[Dict[str, Any]]:
        """获取所有交易对"""
        return self._request("GET", "/v1/markets")

    # ========== 订单管理 ==========

    def get_orders(
        self,
        symbol: Optional[str] = None,
        status: Optional[str] = None,
        side: Optional[str] = None,
        type: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        """
        获取订单列表

        Args:
            symbol: 交易对过滤
            status: 状态过滤
            side: 方向过滤（buy/sell）
            type: 类型过滤（market/limit/stop_loss/take_profit）

        Returns:
            订单列表
        """
        params = {}
        if symbol:
            params["symbol"] = symbol
        if status:
            params["status"] = status
        if side:
            params["side"] = side
        if type:
            params["type"] = type

        return self._request("GET", "/v1/orders", params=params)

    def get_order(self, order_id: int) -> Dict[str, Any]:
        """获取订单详情"""
        return self._request("GET", f"/v1/order/{order_id}")

    def cancel_order(self, order_id: int) -> Dict[str, str]:
        """撤销订单"""
        return self._request("DELETE", f"/v1/order/{order_id}")

    # ========== 成交记录 ==========

    def get_trades(self, symbol: str) -> List[Dict[str, Any]]:
        """获取成交记录"""
        return self._request("GET", f"/v1/trades/{symbol}")

    def get_my_trades(self) -> List[Dict[str, Any]]:
        """获取我的成交记录"""
        return self._request("GET", "/v1/myTrades")

    # ========== 余额管理 ==========

    def get_balance(self) -> Dict[str, Any]:
        """获取账户余额"""
        return self._request("GET", "/v1/balance")
