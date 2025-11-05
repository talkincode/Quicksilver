"""配置管理模块"""

import os
from dotenv import load_dotenv

# 加载环境变量
load_dotenv()


class Config:
    """配置类"""

    # API 配置
    API_URL = os.getenv("API_URL", "http://localhost:8080")
    ADMIN_API_KEY = os.getenv("ADMIN_API_KEY", "")
    ADMIN_API_SECRET = os.getenv("ADMIN_API_SECRET", "")

    # Streamlit 配置
    STREAMLIT_SERVER_PORT = int(os.getenv("STREAMLIT_SERVER_PORT", "8501"))
    STREAMLIT_SERVER_ADDRESS = os.getenv("STREAMLIT_SERVER_ADDRESS", "0.0.0.0")

    # 数据刷新间隔（秒）
    REFRESH_INTERVAL = 5

    # 分页配置
    DEFAULT_PAGE_SIZE = 20
    MAX_PAGE_SIZE = 100


config = Config()
