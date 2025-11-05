"""首页 - 概览、实时行情、用户权益曲线"""

import streamlit as st
import pandas as pd
from datetime import datetime


def show_home_page(api):
    """显示首页：数据概览 + 实时行情 + 权益曲线"""

    tab1, tab2, tab3 = st.tabs(["## 概览与行情", "## 权益分析", "## 系统监控"])

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
                # 尝试获取订单统计
                try:
                    orders = api.get_orders()
                    if isinstance(orders, list):
                        order_count = len(orders)
                        st.metric("📝 订单总数", order_count)
                    else:
                        st.metric("📝 订单总数", "N/A")
                except:
                    st.metric("📝 订单总数", "N/A")
            with col4:
                # 尝试获取成交统计
                try:
                    trades = api.get_my_trades()
                    if isinstance(trades, list):
                        trade_count = len(trades)
                        st.metric("💰 成交总数", trade_count)
                    else:
                        st.metric("💰 成交总数", "N/A")
                except:
                    st.metric("💰 成交总数", "N/A")

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

    with tab3:
        # ============================================================================
        # 系统监控
        # ============================================================================
        st.subheader("🖥️ 系统状态监控")

        # API 健康检查
        col1, col2, col3 = st.columns(3)

        with col1:
            try:
                health = api.health_check()
                status = health.get("status", "unknown")
                if status == "ok":
                    st.success("✅ API 服务")
                    st.caption(f"状态: {status}")
                else:
                    st.warning(f"⚠️ API 服务")
                    st.caption(f"状态: {status}")
            except Exception as e:
                st.error("❌ API 服务")
                st.caption(f"错误: {str(e)[:30]}")

        with col2:
            # 数据库连接检查（通过尝试查询用户来间接检测）
            try:
                api.get_users(page=1, limit=1)
                st.success("✅ 数据库连接")
                st.caption("状态: 正常")
            except Exception as e:
                st.error("❌ 数据库连接")
                st.caption(f"错误: {str(e)[:30]}")

        with col3:
            # 市场数据服务检查
            try:
                api.get_ticker("BTC-USDT")
                st.success("✅ 市场数据")
                st.caption("状态: 正常")
            except Exception as e:
                st.error("❌ 市场数据")
                st.caption(f"错误: {str(e)[:30]}")

        st.markdown("---")

        # 实时数据统计
        st.subheader("📊 实时数据统计")

        try:
            col1, col2, col3, col4 = st.columns(4)

            with col1:
                users_result = api.get_users(page=1, limit=1)
                total_users = users_result.get("total", 0)
                st.metric("总用户数", total_users)

            with col2:
                try:
                    balances = api.get_all_balances(page=1, limit=1000)
                    total_balances = balances.get("total", 0)
                    st.metric("余额记录", total_balances)
                except:
                    st.metric("余额记录", "N/A")

            with col3:
                try:
                    orders = api.get_orders()
                    if isinstance(orders, list):
                        order_count = len(orders)
                        st.metric("活跃订单", order_count)
                    else:
                        st.metric("活跃订单", "N/A")
                except:
                    st.metric("活跃订单", "N/A")

            with col4:
                st.metric("交易对数量", 2)

        except Exception as e:
            st.error(f"加载统计失败: {str(e)}")

        st.markdown("---")

        # 系统信息
        st.subheader("ℹ️ 系统信息")

        info_cols = st.columns(2)

        with info_cols[0]:
            st.markdown("**服务配置**")
            try:
                import os
                from config import config

                st.code(
                    f"""
API URL: {config.API_URL}
环境: {'生产' if 'production' in config.API_URL.lower() else '开发'}
API Key: {config.ADMIN_API_KEY[:10]}...
                """.strip()
                )
            except Exception as e:
                st.error(f"无法加载配置: {str(e)}")

        with info_cols[1]:
            st.markdown("**版本信息**")
            st.code(
                f"""
Quicksilver: v0.1.0
更新时间: 2025-11-05
Dashboard: Streamlit
            """.strip()
            )


def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_home_page(api)


main()
