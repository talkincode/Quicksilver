"""余额管理页面"""

import streamlit as st
import pandas as pd
from datetime import datetime


def show_balances_page(api):

    # 创建标签页
    tab1, tab2 = st.tabs(["👤 用户余额", "⚙️ 余额调整"])

    # ============================================================================
    # Tab 1: 用户余额查询
    # ============================================================================
    with tab1:
        show_user_balances(api)

    # ============================================================================
    # Tab 2: 余额调整
    # ============================================================================
    with tab2:
        show_balance_adjustment(api)


def show_user_balances(api):
    """显示用户余额查询"""
    st.subheader("用户余额查询")

    # 获取用户列表用于选择
    try:
        response = api.get_users()
        users = response.get("data", [])
        if not users:
            st.warning("系统中暂无用户")
            return

        user_options = {
            f"{user['id']} - {user['email']}": user["id"] for user in users
        }

        selected_user = st.selectbox(
            "选择用户",
            options=list(user_options.keys()),
            help="从列表中选择要查询余额的用户",
        )
        user_id = user_options[selected_user]

        # 获取余额
        balances = api.get_user_balances(user_id)

        if not balances:
            st.warning("该用户暂无余额记录")
            return

        # 显示余额表格
        st.markdown("#### 账户余额")
        df = pd.DataFrame(balances)
        df["total"] = df["available"] + df["locked"]

        # 创建带颜色的表格
        st.dataframe(
            df[["asset", "available", "locked", "total"]],
            width="stretch",
            hide_index=True,
            column_config={
                "asset": st.column_config.TextColumn("资产", width="medium"),
                "available": st.column_config.NumberColumn(
                    "可用余额",
                    width="medium",
                    format="%.8f",
                ),
                "locked": st.column_config.NumberColumn(
                    "冻结余额",
                    width="medium",
                    format="%.8f",
                ),
                "total": st.column_config.NumberColumn(
                    "总计",
                    width="medium",
                    format="%.8f",
                ),
            },
        )

        # 显示总价值（假设 USDT 为基准）
        usdt_value = (
            df[df["asset"] == "USDT"]["total"].sum()
            if "USDT" in df["asset"].values
            else 0
        )
        st.metric("USDT 总价值", f"{usdt_value:,.2f}")

    except Exception as e:
        st.error(f"查询失败: {str(e)}")


def show_balance_adjustment(api):
    """显示余额调整功能"""
    form_tab, history_tab = st.tabs(["调整操作", "最近调整记录"])

    with form_tab:
        with st.form("overview_balance_adjustment_form"):
            # 获取用户列表用于下拉选择
            try:
                response = api.get_users()
                users = response.get("data", [])
                if not users:
                    st.warning("系统中暂无用户")
                    st.form_submit_button("提交", disabled=True)
                    return

                user_options = {
                    f"{user['id']} - {user['email']}": user["id"] for user in users
                }
            except Exception as e:
                st.error(f"获取用户列表失败: {str(e)}")
                st.form_submit_button("提交", disabled=True)
                return

            left_col, right_col = st.columns([3, 2], gap="large")

            with left_col:
                selected_user = st.selectbox(
                    "选择用户 *",
                    options=list(user_options.keys()),
                    help="从列表中选择要调整余额的用户",
                )
                user_id = user_options[selected_user]

                asset = st.selectbox(
                    "资产类型 *",
                    ["USDT", "BTC", "ETH", "SOL"],
                    help="选择要调整的资产",
                )

                amount = st.number_input(
                    "调整金额 *",
                    min_value=0.00000001,
                    value=0.00000001,
                    step=0.01,
                    format="%.8f",
                    help="输入调整金额（正数）",
                )

                operation = st.radio(
                    "操作类型 *",
                    ["add", "deduct"],
                    format_func=lambda x: "➕ 增加" if x == "add" else "➖ 扣除",
                    help="选择增加或扣除余额",
                    horizontal=True,
                )

                note = st.text_area(
                    "备注说明 *",
                    placeholder="请填写调整原因...",
                    help="记录此次调整的原因，便于审计",
                    height=120,
                )

            with right_col:
                st.markdown("#### 操作确认")
                confirm_top_cols = st.columns(2, gap="small")
                with confirm_top_cols[0]:
                    st.metric("用户 ID", user_id)
                with confirm_top_cols[1]:
                    st.metric("资产", asset)

                confirm_bottom_cols = st.columns(2, gap="small")
                with confirm_bottom_cols[0]:
                    st.metric("操作", "增加" if operation == "add" else "扣除")
                with confirm_bottom_cols[1]:
                    st.metric("金额", f"{amount:.8f} {asset}")

                st.markdown("#### 提交")
                submitted = st.form_submit_button(
                    "✅ 确认调整",
                    type="primary",
                    use_container_width=True,
                )

            if submitted:
                if not note.strip():
                    st.error("❌ 请填写备注说明")
                elif amount <= 0:
                    st.error("❌ 调整金额必须大于 0")
                else:
                    try:
                        result = api.adjust_balance(
                            user_id=user_id,
                            asset=asset,
                            amount=amount,
                            operation=operation,
                            note=note,
                        )

                        st.success("✅ 余额调整成功！")
                        st.json(result)

                        # 记录操作日志
                        if "adjustment_logs" not in st.session_state:
                            st.session_state["adjustment_logs"] = []

                        st.session_state["adjustment_logs"].insert(
                            0,
                            {
                                "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                                "user_id": user_id,
                                "asset": asset,
                                "amount": amount,
                                "operation": operation,
                                "note": note,
                            },
                        )

                    except Exception as e:
                        st.error(f"❌ 调整失败: {str(e)}")

    with history_tab:
        if st.session_state.get("adjustment_logs"):
            st.markdown("### 📜 最近调整记录")

            logs_df = pd.DataFrame(st.session_state["adjustment_logs"])
            st.dataframe(
                logs_df,
                width="stretch",
                hide_index=True,
                column_config={
                    "time": st.column_config.TextColumn("时间", width="medium"),
                    "user_id": st.column_config.NumberColumn("用户 ID", width="small"),
                    "asset": st.column_config.TextColumn("资产", width="small"),
                    "amount": st.column_config.NumberColumn(
                        "金额", width="medium", format="%.8f"
                    ),
                    "operation": st.column_config.TextColumn("操作", width="small"),
                    "note": st.column_config.TextColumn("备注", width="large"),
                },
            )

            st.caption(f"共 {len(logs_df)} 条记录（仅显示当前会话）")
        else:
            st.info("当前会话暂无调整记录。")



def main() -> None:
    api = st.session_state.get("api")
    if api is None:
        st.error("API 客户端未初始化")
        return
    show_balances_page(api)


main()
