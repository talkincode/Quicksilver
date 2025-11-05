"""成交记录页面"""

import streamlit as st
import pandas as pd
from datetime import datetime


def show_trades_page(api):
    """显示成交记录页面"""

    # ============================================================================
    # 筛选器
    # ============================================================================
    col1, col2 = st.columns([2, 1])

    with col1:
        symbol_filter = st.selectbox(
            "**选择交易对**",
            ["BTC/USDT", "ETH/USDT"],
            key="trades_symbol_filter",
            label_visibility="visible",
        )

    with col2:
        st.markdown("&nbsp;")  # 空行对齐
        if st.button("🔄 刷新数据", use_container_width=True):
            st.rerun()

    # ============================================================================
    # 成交记录列表
    # ============================================================================
    with st.spinner("🔄 正在加载成交记录..."):
        try:
            # 获取成交数据
            # 转换交易对格式: BTC/USDT -> BTC-USDT
            api_symbol = symbol_filter.replace("/", "-")
            trades = api.get_trades(api_symbol)

            # 检查返回数据类型
            if isinstance(trades, str):
                st.error(f"❌ API 返回错误: {trades}")
                return

            if not trades or not isinstance(trades, list) or len(trades) == 0:
                st.info(f"📭 暂无 {symbol_filter} 的成交记录")
            else:
                st.subheader(f"{symbol_filter} 成交记录 ({len(trades)} 条)")

                # 格式化成交数据
                def format_trade(trade):
                    """格式化单个成交"""
                    return {
                        "ID": trade.get("id", "N/A"),
                        "订单ID": trade.get("order_id", "N/A"),
                        "用户ID": trade.get("user_id", "N/A"),
                        "交易对": trade.get("symbol", "N/A"),
                        "方向": "🟢 买入" if trade.get("side") == "buy" else "🔴 卖出",
                        "价格": f"${trade.get('price', 0):,.2f}",
                        "数量": f"{trade.get('amount', 0):.8f}",
                        "成交额": f"${(trade.get('price', 0) * trade.get('amount', 0)):,.2f}",
                        "手续费": f"{trade.get('fee', 0):.8f} {trade.get('fee_currency', '')}",
                        "时间": format_datetime(trade.get("timestamp", "")),
                    }

                def format_datetime(dt_str):
                    """格式化日期时间"""
                    if not dt_str:
                        return "N/A"
                    try:
                        # 处理时间戳（毫秒）
                        if isinstance(dt_str, (int, float)):
                            dt = datetime.fromtimestamp(dt_str / 1000.0)
                            return dt.strftime("%Y-%m-%d %H:%M:%S")
                        # 处理 ISO 格式
                        dt = datetime.fromisoformat(dt_str.replace("Z", "+00:00"))
                        return dt.strftime("%Y-%m-%d %H:%M:%S")
                    except:
                        return (
                            str(dt_str)[:19] if len(str(dt_str)) >= 19 else str(dt_str)
                        )

                # 转换为 DataFrame
                trades_data = [format_trade(trade) for trade in trades]
                df = pd.DataFrame(trades_data)

                # 显示表格
                st.dataframe(
                    df,
                    use_container_width=True,
                    height=min(600, max(200, len(trades) * 43 + 50)),
                    hide_index=True,
                )

                # 统计信息
                st.markdown("---")
                st.subheader("📊 成交统计")

                col1, col2, col3, col4 = st.columns(4)

                with col1:
                    total_trades = len(trades)
                    st.metric("总成交笔数", total_trades)

                with col2:
                    buy_trades = sum(1 for t in trades if t.get("side") == "buy")
                    st.metric("买入笔数", buy_trades)

                with col3:
                    sell_trades = sum(1 for t in trades if t.get("side") == "sell")
                    st.metric("卖出笔数", sell_trades)

                with col4:
                    total_volume = sum(
                        t.get("price", 0) * t.get("amount", 0) for t in trades
                    )
                    st.metric("总成交额", f"${total_volume:,.2f}")

        except Exception as e:
            st.error(f"❌ 加载成交记录失败: {str(e)}")
            st.caption("提示: 确保后端服务正在运行且 API 端点已实现")

    # ============================================================================
    # 我的成交记录（如果已实现）
    # ============================================================================
    st.markdown("---")
    st.subheader("📝 我的成交记录")

    try:
        my_trades = api.get_my_trades()

        if not my_trades or len(my_trades) == 0:
            st.info("暂无我的成交记录")
        else:
            st.write(f"共 {len(my_trades)} 条记录")

            # 简化展示
            my_trades_data = []
            for trade in my_trades[:50]:  # 最多显示 50 条
                my_trades_data.append(
                    {
                        "ID": trade.get("id", "N/A"),
                        "交易对": trade.get("symbol", "N/A"),
                        "方向": "买" if trade.get("side") == "buy" else "卖",
                        "价格": f"${trade.get('price', 0):,.2f}",
                        "数量": f"{trade.get('amount', 0):.8f}",
                    }
                )

            my_df = pd.DataFrame(my_trades_data)
            st.dataframe(my_df, use_container_width=True, hide_index=True)

    except Exception as e:
        st.warning(f"⚠️ 无法加载我的成交记录: {str(e)}")


def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_trades_page(api)


main()
