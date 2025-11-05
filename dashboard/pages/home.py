"""首页 - 概览、实时行情、用户权益曲线"""

import streamlit as st
import pandas as pd
from datetime import datetime


def show_home_page(api):
    """显示首页：数据概览 + 实时行情 + 权益曲线"""

    tab1, tab2 = st.tabs(["## 概览与行情", "## 权益分析"])

    with tab1:
        # ============================================================================
        # 快速统计
        # ============================================================================
        st.subheader("🚀 快速统计")
        try:
            users_result = api.get_users(page=1, limit=1)
            total_users = users_result.get("total", 0)

            col1, col2, col3, col4 = st.columns(4)

            with col1:
                st.metric("👥 用户总数", total_users)
            with col2:
                st.metric("📈 交易对", "2", delta="BTC/USDT, ETH/USDT")
            with col3:
                st.metric("📝 订单总数", "待实现")
            with col4:
                st.metric("💰 成交总数", "待实现")

        except Exception as e:
            st.error(f"❌ 加载统计数据失败: {str(e)}")

        # ============================================================================
        # 实时行情
        # ============================================================================
        st.markdown("---")
        st.subheader("💹 实时行情")

        col1, col2 = st.columns(2)

        try:
            with col1:
                ticker = api.get_ticker("BTC-USDT")
                price = ticker.get("last", 0)
                st.metric("BTC/USDT", f"${price:,.2f}" if price else "N/A")
                if price:
                    st.caption(
                        f"买一: ${ticker.get('bid', 0):,.2f} | 卖一: ${ticker.get('ask', 0):,.2f}"
                    )

            with col2:
                ticker = api.get_ticker("ETH-USDT")
                price = ticker.get("last", 0)
                st.metric("ETH/USDT", f"${price:,.2f}" if price else "N/A")
                if price:
                    st.caption(
                        f"买一: ${ticker.get('bid', 0):,.2f} | 卖一: ${ticker.get('ask', 0):,.2f}"
                    )

        except Exception as e:
            st.warning(f"⚠️ 加载行情失败: {str(e)}")

    with tab2:
        # ============================================================================
        # 用户权益曲线（模拟数据）
        # ============================================================================
        st.subheader("📈 用户权益曲线")

        # TODO: 从 API 获取真实数据
        dates = pd.date_range(end=datetime.now(), periods=30, freq="D")
        equity = pd.DataFrame(
            {"日期": dates, "权益": [10000 + i * 100 + (i % 5) * 50 for i in range(30)]}
        )

        st.line_chart(equity.set_index("日期"))
        st.caption("⚠️ 当前为模拟数据，待实现真实权益统计")

        st.markdown("---")
        st.write("待添加更多分析图表...")


def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_home_page(api)


main()
