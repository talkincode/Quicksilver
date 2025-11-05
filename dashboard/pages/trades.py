"""成交记录页面 - 简化版"""

import streamlit as st
import pandas as pd


def show_trades_page(api):
    """显示成交记录页面"""
    st.title("💰 成交记录")

    # 搜索框
    search = st.text_input(
        "搜索",
        placeholder="输入用户ID或交易对...",
        label_visibility="collapsed",
    )

    st.info("⚠️ 成交记录功能待后端实现")
