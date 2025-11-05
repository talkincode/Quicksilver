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
        page_title="Quicksilver Admin",
        page_icon="⚡",
        layout="wide",
        initial_sidebar_state="expanded",
    )
    st.config.set_option("client.showSidebarNavigation", True)

    # 增强的全局样式
    st.markdown(
        """
<style>
    /* 指标卡片样式 */
    .stMetric {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        padding: 1.2rem;
        border-radius: 0.8rem;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    }
    
    .stMetric label {
        color: white !important;
        font-weight: 600;
    }
    
    .stMetric [data-testid="stMetricValue"] {
        color: white !important;
        font-size: 1.8rem;
    }
    
    /* 按钮样式优化 */
    .stButton > button {
        border-radius: 0.5rem;
        font-weight: 500;
        transition: all 0.3s ease;
    }
    
    .stButton > button:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
    
    /* 数据表格样式 */
    .stDataFrame {
        border-radius: 0.5rem;
        overflow: hidden;
    }
    
    /* 标签页样式 */
    .stTabs [data-baseweb="tab-list"] {
        gap: 8px;
    }
    
    .stTabs [data-baseweb="tab"] {
        border-radius: 0.5rem 0.5rem 0 0;
        padding: 0.8rem 1.5rem;
        font-weight: 500;
    }
    
    /* 输入框样式 */
    .stTextInput > div > div > input {
        border-radius: 0.5rem;
    }
    
    .stSelectbox > div > div > div {
        border-radius: 0.5rem;
    }
    
    /* 成功/错误/警告消息样式 */
    .stSuccess, .stError, .stWarning, .stInfo {
        border-radius: 0.5rem;
        padding: 1rem;
    }
    
    /* 页面标题样式 */
    h1 {
        color: #1e293b;
        font-weight: 700;
    }
    
    h2 {
        color: #334155;
        font-weight: 600;
    }
    
    /* 加载动画 */
    @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
    }
    
    .loading {
        animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
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
