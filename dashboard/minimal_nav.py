"""最小化导航测试 - 完全按照文档示例"""

import streamlit as st

st.set_page_config(page_title="Minimal Nav Test")


def page1():
    st.title("Page 1")
    st.write("This is page 1 content")


def page2():
    st.title("Page 2")
    st.write("This is page 2 content")


pg = st.navigation(
    [
        st.Page(page1, title="Home", icon="🏠", default=True),
        st.Page(page2, title="About", icon="ℹ️"),
    ]
)
pg.run()
