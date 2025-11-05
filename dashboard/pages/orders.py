"""订单管理页面"""

import streamlit as st
import pandas as pd
from datetime import datetime


def show_orders_page(api):
    """显示订单管理页面"""

    st.title("📝 订单管理")

    # ============================================================================
    # 筛选器
    # ============================================================================
    col1, col2, col3, col4 = st.columns(4)

    with col1:
        symbol_filter = st.selectbox(
            "交易对", ["全部", "BTC/USDT", "ETH/USDT"], key="orders_symbol_filter"
        )

    with col2:
        status_filter = st.selectbox(
            "状态",
            ["全部", "new", "filled", "cancelled", "partial"],
            key="orders_status_filter",
        )

    with col3:
        side_filter = st.selectbox(
            "方向", ["全部", "buy", "sell"], key="orders_side_filter"
        )

    with col4:
        type_filter = st.selectbox(
            "类型",
            ["全部", "market", "limit", "stop_loss", "take_profit"],
            key="orders_type_filter",
        )

    # ============================================================================
    # 订单列表
    # ============================================================================
    try:
        # 构建查询参数
        params = {}
        if symbol_filter != "全部":
            params["symbol"] = symbol_filter
        if status_filter != "全部":
            params["status"] = status_filter
        if side_filter != "全部":
            params["side"] = side_filter
        if type_filter != "全部":
            params["type"] = type_filter

        # 获取订单数据
        orders = api.get_orders(**params)

        # 检查返回数据类型
        if isinstance(orders, str):
            st.error(f"❌ API 返回错误: {orders}")
            return
        
        if not orders or not isinstance(orders, list) or len(orders) == 0:
            st.info("📭 暂无订单数据")
        else:
            st.subheader(f"订单列表 ({len(orders)} 条)")

            # 格式化订单数据
            def format_order(order):
                """格式化单个订单"""
                return {
                    "ID": order.get("id", "N/A"),
                    "用户ID": order.get("user_id", "N/A"),
                    "交易对": order.get("symbol", "N/A"),
                    "方向": "🟢 买入" if order.get("side") == "buy" else "🔴 卖出",
                    "类型": order.get("type", "N/A"),
                    "价格": (
                        f"${order.get('price', 0):,.2f}"
                        if order.get("price")
                        else "市价"
                    ),
                    "数量": f"{order.get('amount', 0):.8f}",
                    "已成交": f"{order.get('filled', 0):.8f}",
                    "状态": get_status_badge(order.get("status", "unknown")),
                    "创建时间": format_datetime(order.get("created_at", "")),
                }

            def get_status_badge(status):
                """获取状态徽章"""
                status_map = {
                    "new": "🆕 新建",
                    "filled": "✅ 完全成交",
                    "partial": "⏳ 部分成交",
                    "cancelled": "❌ 已取消",
                }
                return status_map.get(status, f"⚪ {status}")

            def format_datetime(dt_str):
                """格式化日期时间"""
                if not dt_str:
                    return "N/A"
                try:
                    dt = datetime.fromisoformat(dt_str.replace("Z", "+00:00"))
                    return dt.strftime("%Y-%m-%d %H:%M:%S")
                except:
                    return dt_str[:19] if len(dt_str) >= 19 else dt_str

            # 转换为 DataFrame
            orders_data = [format_order(order) for order in orders]
            df = pd.DataFrame(orders_data)

            # 显示表格
            st.dataframe(
                df,
                use_container_width=True,
                height=min(600, max(200, len(orders) * 43 + 50)),
                hide_index=True,
            )

            # 详细信息展开
            st.markdown("---")
            st.subheader("订单详情")

            selected_order_id = st.number_input(
                "输入订单 ID 查看详情", min_value=1, step=1, key="order_detail_id_input"
            )

            if st.button("查询订单详情", key="query_order_detail"):
                try:
                    order_detail = api.get_order(int(selected_order_id))

                    st.success("✅ 订单详情")

                    col1, col2 = st.columns(2)

                    with col1:
                        st.markdown("**基本信息**")
                        st.json(
                            {
                                "ID": order_detail.get("id"),
                                "用户ID": order_detail.get("user_id"),
                                "交易对": order_detail.get("symbol"),
                                "方向": order_detail.get("side"),
                                "类型": order_detail.get("type"),
                            }
                        )

                    with col2:
                        st.markdown("**交易信息**")
                        st.json(
                            {
                                "价格": order_detail.get("price"),
                                "数量": order_detail.get("amount"),
                                "已成交": order_detail.get("filled"),
                                "剩余": order_detail.get("remaining"),
                                "状态": order_detail.get("status"),
                            }
                        )

                    # 取消订单按钮
                    if order_detail.get("status") in ["new", "partial"]:
                        if st.button(
                            "❌ 取消此订单", type="secondary", key="cancel_order_btn"
                        ):
                            try:
                                result = api.cancel_order(int(selected_order_id))
                                st.success(
                                    f"✅ 订单已取消: {result.get('message', '')}"
                                )
                                st.rerun()
                            except Exception as e:
                                st.error(f"取消失败: {str(e)}")

                except Exception as e:
                    st.error(f"查询失败: {str(e)}")

    except Exception as e:
        st.error(f"❌ 加载订单失败: {str(e)}")
        st.caption("提示: 确保后端服务正在运行且 API 端点已实现")


def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_orders_page(api)


main()
