#!/bin/bash
# 为 Quicksilver Dashboard 设置环境变量

set -e

echo "🔧 配置 Quicksilver Dashboard 环境..."

# 创建 .env 文件
cat > dashboard/.env << 'EOF'
# Quicksilver API 配置
API_URL=http://localhost:8080
ADMIN_API_KEY=qs-test-api-key-2024
ADMIN_API_SECRET=qs-test-api-secret-change-in-production

# Streamlit 配置
STREAMLIT_SERVER_PORT=8501
STREAMLIT_SERVER_ADDRESS=0.0.0.0
EOF

echo "✅ 环境变量已配置到 dashboard/.env"
echo ""
echo "📝 使用的凭证（测试账户）："
echo "   Email:      test@quicksilver.local"
echo "   API Key:    qs-test-api-key-2024"
echo "   API Secret: qs-test-api-secret-change-in-production"
echo ""
echo "⚠️  注意：这些是测试凭证，生产环境请使用安全的凭证！"
echo ""
echo "🚀 现在可以启动仪表盘了："
echo "   cd dashboard"
echo "   ./start.sh"
