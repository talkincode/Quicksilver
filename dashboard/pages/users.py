"""用户管理页面"""

import streamlit as st
import pandas as pd


def show_users_page(api):
    """显示用户管理页面"""
    st.title("👥 用户管理")

    # ============================================================================
    # 创建用户弹窗
    # ============================================================================
    @st.dialog("创建新用户")
    def create_user_dialog():
        email = st.text_input("邮箱", placeholder="user@example.com")

        col1, col2 = st.columns(2)
        with col1:
            if st.button("取消", use_container_width=True):
                st.rerun()
        with col2:
            if st.button("创建", type="primary", use_container_width=True):
                if not email:
                    st.error("邮箱不能为空")
                    return

                try:
                    result = api.create_user(email, None)
                    st.success("✅ 创建成功！")
                    st.warning("⚠️ 请保存以下凭证（仅显示一次）")
                    st.code(
                        f"用户 ID: {result['id']}\n"
                        f"邮箱: {result['email']}\n"
                        f"API Key: {result['api_key']}\n"
                        f"API Secret: {result['api_secret']}"
                    )
                except Exception as e:
                    st.error(f"创建失败: {str(e)}")

    # ============================================================================
    # 搜索栏和创建按钮
    # ============================================================================
    col1, col2 = st.columns([4, 1])

    with col1:
        search = st.text_input(
            "搜索用户", placeholder="输入邮箱搜索...", label_visibility="collapsed"
        )

    with col2:
        if st.button("➕ 创建用户", use_container_width=True):
            create_user_dialog()

    # ============================================================================
    # 用户列表
    # ============================================================================
    try:
        result = api.get_users(page=1, limit=100)
        users = result.get("data", [])

        if not users:
            st.info("暂无用户数据")
        else:
            df = pd.DataFrame(users)

            # 搜索过滤
            if search:
                df = df[df["email"].str.contains(search, case=False, na=False)]

            # 显示表格
            st.dataframe(
                df[["id", "email", "api_key", "status", "created_at"]],
                use_container_width=True,
                hide_index=True,
                column_config={
                    "id": st.column_config.NumberColumn("ID", width="small"),
                    "email": st.column_config.TextColumn("邮箱", width="medium"),
                    "api_key": st.column_config.TextColumn("API Key", width="large"),
                    "status": st.column_config.TextColumn("状态", width="small"),
                    "created_at": st.column_config.TextColumn(
                        "创建时间", width="medium"
                    ),
                },
            )

            st.caption(f"共 {len(df)} 个用户")

    except Exception as e:
        st.error(f"加载用户列表失败: {str(e)}")
