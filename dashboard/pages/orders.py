"""订单管理页面 - 简化版"""

import streamlit as st
import pandas as pd


def show_orders_page(api):
    """显示订单管理页面"""
    st.title("📝 订单管理")

    # 搜索框
    search = st.text_input(
        "搜索",
        placeholder="输入用户ID或交易对...",
        label_visibility="collapsed",
    )

    st.info("⚠️ 订单管理功能待后端实现")
