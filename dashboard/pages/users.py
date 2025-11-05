"""用户管理页面"""

import streamlit as st
from datetime import datetime


def show_users_page(api):
    """显示用户管理页面"""

    # ============================================================================
    # 辅助函数
    # ============================================================================
    def init_session_state():
        """初始化会话状态"""
        if "selected_user_ids" not in st.session_state:
            st.session_state.selected_user_ids = set()

    def is_user_selected(user_id):
        """检查用户是否被选中"""
        return user_id in st.session_state.selected_user_ids

  
    
    def clear_all_selections():
        """清除所有选择"""
        st.session_state.selected_user_ids.clear()

    def format_datetime(dt_str):
        """格式化日期时间"""
        try:
            dt = datetime.fromisoformat(dt_str.replace('Z', '+00:00'))
            return dt.strftime('%Y-%m-%d %H:%M')
        except:
            return dt_str[:16] if dt_str else "N/A"

    def get_status_emoji(status):
        """获取状态图标"""
        status_map = {
            'active': '🟢',
            'inactive': '🔴',
            'suspended': '🟡'
        }
        return status_map.get(status, '⚪')

    # ============================================================================
    # 删除确认弹窗
    # ============================================================================
    @st.dialog("删除用户确认")
    def delete_user_dialog(users_to_delete):
        st.warning(f"确定要删除以下 {len(users_to_delete)} 个用户吗？")

        # 显示要删除的用户列表
        for user in users_to_delete:
            st.code(f"ID: {user['id']} | 邮箱: {user['email']}")

        st.error("⚠️ 注意：此操作将删除用户及其所有相关数据（订单、余额、交易记录等），无法恢复！")

        col1, col2 = st.columns(2)
        with col1:
            if st.button("取消", width="stretch"):
                st.rerun()
        with col2:
            if st.button("确认删除", type="primary", width="stretch"):
                try:
                    success_count = 0
                    error_count = 0
                    error_messages = []

                    for user in users_to_delete:
                        try:
                            api.delete_user(user['id'])
                            success_count += 1
                        except Exception as e:
                            error_count += 1
                            error_messages.append(f"用户 {user['email']} 删除失败: {str(e)}")

                    if success_count > 0:
                        st.success(f"✅ 成功删除 {success_count} 个用户")
                    if error_count > 0:
                        st.error(f"❌ 删除失败 {error_count} 个用户:")
                        for msg in error_messages:
                            st.error(msg)

                    # 清除选择状态并刷新页面
                    clear_all_selections()
                    st.rerun()

                except Exception as e:
                    st.error(f"删除操作失败: {str(e)}")

    # ============================================================================
    # 创建用户弹窗
    # ============================================================================
    @st.dialog("创建新用户")
    def create_user_dialog():
        email = st.text_input(
            "邮箱",
            placeholder="user@example.com",
            key="dialog_email_input",
        )
        col1, col2 = st.columns(2)
        msgbox = st.empty()
        with col1:
            if st.button("取消", width="stretch"):
                st.rerun()
        with col2:
            if st.button("创建", type="primary", width="stretch"):
                if not email:
                    st.error("邮箱不能为空")
                    return

                try:
                    result = api.create_user(email, None)
                    msgbox.success("✅ 创建成功！")
                    msgbox.warning("⚠️ 请保存以下凭证（仅显示一次）")
                    msgbox.code(
                        f"用户 ID: {result['id']}\n"
                        f"邮箱: {result['email']}\n"
                        f"API Key: {result['api_key']}\n"
                        f"API Secret: {result['api_secret']}"
                    )
                except Exception as e:
                    st.error(f"创建失败: {str(e)}")

    # ============================================================================
    # 初始化会话状态
    # ============================================================================
    init_session_state()

    # ============================================================================
    # 搜索栏和操作按钮
    # ============================================================================
    col1, col2, col3, col4 = st.columns([3, 1, 1, 1])

    with col1:
        search = st.text_input(
            "搜索用户",
            placeholder="输入邮箱搜索...",
            label_visibility="collapsed",
            key="users_page_search_input"
        )

    with col2:
        if st.button("🔍 查询", width="stretch"):
            # 触发搜索操作
            if "search_trigger" not in st.session_state:
                st.session_state.search_trigger = 0
            st.session_state.search_trigger += 1
            st.session_state.current_search = search
            st.rerun()

    with col3:
        if st.button("➕ 创建用户", width="stretch"):
            create_user_dialog()

    with col4:
        selected_count = len(st.session_state.selected_user_ids)
        if selected_count > 0 and st.button(f"🗑️ 删除选中 ({selected_count})", width="stretch", type="secondary"):
            # 获取选中的用户详情
            try:
                result = api.get_users(page=1, limit=1000)  # 获取更多用户用于查找
                all_users = result.get("data", [])
                selected_users = [
                    user for user in all_users
                    if user['id'] in st.session_state.selected_user_ids
                ]
                delete_user_dialog(selected_users)
            except Exception as e:
                st.error(f"获取用户信息失败: {str(e)}")

    # ============================================================================
    # 用户列表
    # ============================================================================
    try:
        result = api.get_users(page=1, limit=100)
        users = result.get("data", [])

        if not users:
            st.info("暂无用户数据")
        else:
            # 搜索过滤（使用保存的搜索关键词）
            current_search = st.session_state.get("current_search", "")
            if current_search:
                users = [user for user in users if current_search.lower() in user.get('email', '').lower()]

            st.subheader("用户列表")

            # 构建表格数据
            table_data = []
            for user in users:
                table_data.append({
                    "选择": is_user_selected(user['id']),
                    "ID": user['id'],
                    "邮箱": user['email'],
                    "API Key": user['api_key'][:10] + "..." if len(user['api_key']) > 10 else user['api_key'],
                    "状态": f"{get_status_emoji(user['status'])} {user['status']}",
                    "创建时间": format_datetime(user['created_at'])
                })

            # 使用可编辑数据表格显示用户列表
            edited_data = st.data_editor(
                table_data,
                column_config={
                    "选择": st.column_config.CheckboxColumn(
                        "选择",
                        help="选择要删除的用户",
                        default=False,
                        width="small"
                    ),
                    "ID": st.column_config.NumberColumn(
                        "ID",
                        help="用户ID",
                        width="small",
                        format="%d"
                    ),
                    "邮箱": st.column_config.TextColumn(
                        "邮箱",
                        help="用户邮箱地址",
                        width="large"
                    ),
                    "API Key": st.column_config.TextColumn(
                        "API Key",
                        help="用户API密钥（前10位）",
                        width="medium"
                    ),
                    "状态": st.column_config.TextColumn(
                        "状态",
                        help="用户状态",
                        width="small"
                    ),
                    "创建时间": st.column_config.TextColumn(
                        "创建时间",
                        help="账户创建时间",
                        width="medium"
                    )
                },
                hide_index=True,
                use_container_width=True,
                height=min(500, max(200, len(users) * 43 + 50)),
                num_rows="fixed",
                key="users_table"
            )

            # 同步表格选择状态到会话状态
            current_selections = set()
            for row in edited_data:
                if row["选择"]:
                    current_selections.add(row["ID"])

            # 只在表格状态变化时更新会话状态
            if current_selections != st.session_state.selected_user_ids:
                st.session_state.selected_user_ids = current_selections
                st.rerun()

            # 显示选择状态
            selected_count = len(st.session_state.selected_user_ids)
            if selected_count > 0:
                st.info(f"已选择 {selected_count} 个用户")

            st.caption(f"共 {len(users)} 个用户")

    except Exception as e:
        st.error(f"加载用户列表失败: {str(e)}")


def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_users_page(api)


main()


