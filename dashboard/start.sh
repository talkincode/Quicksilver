#!/bin/bash
# Quicksilver Dashboard 快速启动脚本 (使用 uv)

set -e

echo "🚀 启动 Quicksilver 管理仪表盘..."

# 检查是否在 dashboard 目录
if [ ! -f "pyproject.toml" ]; then
    echo "❌ 错误: 请在 dashboard 目录下运行此脚本"
    exit 1
fi

# 检查 uv 是否安装
if ! command -v uv &> /dev/null; then
    echo "❌ uv 未安装"
    echo "📥 请运行以下命令安装 uv:"
    echo "   curl -LsSf https://astral.sh/uv/install.sh | sh"
    exit 1
fi

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo "⚠️  警告: .env 文件不存在，将使用 .env.example"
    if [ -f ".env.example" ]; then
        cp .env.example .env
    fi
    echo "📝 请编辑 .env 文件配置 API 凭证："
    echo "   - ADMIN_API_KEY"
    echo "   - ADMIN_API_SECRET"
    echo ""
    read -p "按 Enter 继续..."
fi

# 同步依赖
echo "📦 同步依赖..."
uv sync

# 启动 Streamlit
echo "✨ 启动仪表盘..."
echo "🌐 访问地址: http://localhost:8501"
echo ""
uv run streamlit run app.py
