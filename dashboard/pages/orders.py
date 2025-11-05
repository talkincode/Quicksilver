"""订单管理页面 - 简化版"""

import streamlit as st
import pandas as pd


def show_orders_page(api):
    """显示订单管理页面"""

    # 搜索框
    search = st.text_input(
        "搜索",
        placeholder="输入用户ID或交易对...",
        label_visibility="collapsed",
        key="orders_search_box",
    )

    st.info("⚠️ 订单管理功能待后端实现")


def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_orders_page(api)


main()
