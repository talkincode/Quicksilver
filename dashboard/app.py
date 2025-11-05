"""Quicksilver 管理仪表盘"""

import streamlit as st
from config import config
from api import QuicksilverAPI

# ============================================================================
# 页面配置 - 必须是第一个 Streamlit 命令
# ============================================================================
st.set_page_config(
    page_title="Quicksilver",
    page_icon="⚡",
    layout="wide",
    initial_sidebar_state="expanded",
)

# ============================================================================
# 自定义样式
# ============================================================================
st.markdown(
    """
<style>
    .stMetric {
        background-color: #f0f2f6;
        padding: 1rem;
        border-radius: 0.5rem;
    }
    section[data-testid="stSidebar"] > div {
        padding-top: 2rem;
    }
    .sidebar-link {
        display: block;
        padding: 0.75rem 1rem;
        margin: 0.25rem 0;
        border-radius: 0.5rem;
        text-decoration: none;
        color: inherit;
        transition: background-color 0.2s;
    }
    .sidebar-link:hover {
        background-color: #f0f2f6;
    }
</style>
""",
    unsafe_allow_html=True,
)


# ============================================================================
# API 客户端
# ============================================================================
@st.cache_resource
def get_api_client():
    return QuicksilverAPI(
        base_url=config.API_URL,
        api_key=config.ADMIN_API_KEY,
        api_secret=config.ADMIN_API_SECRET,
    )


# ============================================================================
# 侧边栏导航
# ============================================================================
with st.sidebar:
    st.title("⚡ Quicksilver")
    st.markdown("---")

    # 使用 link_button 导航
    if st.button("🏠 概览", key="nav_home", use_container_width=True):
        st.session_state.page = "home"

    if st.button("👥 用户管理", key="nav_users", use_container_width=True):
        st.session_state.page = "users"

    if st.button("📝 订单管理", key="nav_orders", use_container_width=True):
        st.session_state.page = "orders"

    if st.button("💰 成交记录", key="nav_trades", use_container_width=True):
        st.session_state.page = "trades"

    st.markdown("---")
    st.caption(f"**API**: {config.API_URL}")
    st.caption("© 2025 Quicksilver v0.1.0")

# ============================================================================
# 页面路由
# ============================================================================
if "page" not in st.session_state:
    st.session_state.page = "home"

api = get_api_client()

if st.session_state.page == "home":
    from pages.home import show_home_page

    show_home_page(api)

elif st.session_state.page == "users":
    from pages.users import show_users_page

    show_users_page(api)

elif st.session_state.page == "orders":
    from pages.orders import show_orders_page

    show_orders_page(api)

elif st.session_state.page == "trades":
    from pages.trades import show_trades_page

    show_trades_page(api)
