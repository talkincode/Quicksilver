"""Quicksilver 管理仪表盘"""

import streamlit as st
from config import config
from api import QuicksilverAPI


def ensure_api_client() -> None:
    """Initialize shared API client in session state if missing."""
    if "api" not in st.session_state:
        st.session_state.api = QuicksilverAPI(
            base_url=config.API_URL,
            api_key=config.ADMIN_API_KEY,
            api_secret=config.ADMIN_API_SECRET,
        )


def setup_page() -> None:
    """Apply global page settings and styles."""
    st.set_page_config(
        page_title="Quicksilver",
        page_icon="⚡",
        layout="wide",
        initial_sidebar_state="expanded",
    )
    st.config.set_option("client.showSidebarNavigation", True)
    st.markdown(
        """
<style>
    .stMetric {
        background-color: #f0f2f6;
        padding: 1rem;
        border-radius: 0.5rem;
    }
</style>
""",
        unsafe_allow_html=True,
    )


def main() -> None:
    """Entry point for Streamlit app."""
    setup_page()
    ensure_api_client()

    pages = [
        st.Page("pages/home.py", title="概览", icon="🏠", default=True),
        st.Page("pages/users.py", title="用户管理", icon="👥"),
        st.Page("pages/balances.py", title="余额管理", icon="💰"),
        st.Page("pages/orders.py", title="订单管理", icon="📝"),
        st.Page("pages/trades.py", title="成交记录", icon="📊"),
    ]

    st.navigation(pages, position="top", expanded=True).run()


main()
