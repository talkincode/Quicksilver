#!/bin/bash

###############################################################################
# Quicksilver 测试用户初始化脚本
#
# 功能:
# - 创建测试用户并初始化余额
# - 支持本地 PostgreSQL 和 Docker 容器
#
# 使用方法:
#   ./scripts/init_test_user.sh          # 本地 PostgreSQL
#   ./scripts/init_test_user.sh docker   # Docker 容器
###############################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-quicksilver}"
DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-pgdb}"

API_KEY="qs-test-api-key-2024"
API_SECRET="qs-test-api-secret-change-in-production"
TEST_EMAIL="test@quicksilver.local"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Quicksilver Test User Initialization${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检测运行模式
if [ "$1" = "docker" ]; then
    echo -e "${YELLOW}Mode: Docker Container${NC}"
    DOCKER_MODE=true
else
    echo -e "${YELLOW}Mode: Local PostgreSQL${NC}"
    DOCKER_MODE=false
fi

# 检查依赖
if [ "$DOCKER_MODE" = true ]; then
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Error: Docker is not installed${NC}"
        exit 1
    fi
    
    # 检查容器是否运行
    if ! docker ps | grep -q quicksilver-postgres; then
        echo -e "${RED}Error: PostgreSQL container 'quicksilver-postgres' is not running${NC}"
        echo -e "${YELLOW}Please start it with: docker-compose up -d postgres${NC}"
        exit 1
    fi
else
    if ! command -v psql &> /dev/null; then
        echo -e "${RED}Error: psql is not installed${NC}"
        echo -e "${YELLOW}Install it with: brew install postgresql${NC}"
        exit 1
    fi
fi

# 执行 SQL
echo -e "${YELLOW}Creating test user...${NC}"
echo ""

SQL="
-- 删除旧数据
DELETE FROM trades WHERE user_id IN (SELECT id FROM users WHERE email = '$TEST_EMAIL');
DELETE FROM orders WHERE user_id IN (SELECT id FROM users WHERE email = '$TEST_EMAIL');
DELETE FROM balances WHERE user_id IN (SELECT id FROM users WHERE email = '$TEST_EMAIL');
DELETE FROM users WHERE email = '$TEST_EMAIL';

-- 创建测试用户（管理员权限）
INSERT INTO users (email, username, api_key, api_secret, status, role, created_at, updated_at)
VALUES (
    '$TEST_EMAIL',
    'test_user',
    '$API_KEY',
    '$API_SECRET',
    'active',
    'admin',
    NOW(),
    NOW()
) RETURNING id;

-- 获取用户 ID 并创建余额
DO \$\$
DECLARE
test_user_id INT;
BEGIN
SELECT id INTO test_user_id FROM users WHERE email = '$TEST_EMAIL';

-- USDT: 100,000
INSERT INTO balances (user_id, asset, available, locked, created_at, updated_at)
VALUES (test_user_id, 'USDT', 100000.00000000, 0.00000000, NOW(), NOW());

-- BTC: 10
INSERT INTO balances (user_id, asset, available, locked, created_at, updated_at)
VALUES (test_user_id, 'BTC', 10.00000000, 0.00000000, NOW(), NOW());

-- ETH: 100
INSERT INTO balances (user_id, asset, available, locked, created_at, updated_at)
VALUES (test_user_id, 'ETH', 100.00000000, 0.00000000, NOW(), NOW());
END \$\$;

-- 查询结果
SELECT
u.id,
u.email,
u.username,
u.status,
b.asset,
b.available,
b.locked
FROM users u
LEFT JOIN balances b ON u.id = b.user_id
WHERE u.email = '$TEST_EMAIL'
ORDER BY b.asset;
"

# 执行 SQL
if [ "$DOCKER_MODE" = true ]; then
    echo "$SQL" | docker exec -i quicksilver-postgres psql -U postgres -d quicksilver
else
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$SQL"
fi

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ Test User Created Successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}Credentials:${NC}"
    echo -e "  Email:      ${YELLOW}$TEST_EMAIL${NC}"
    echo -e "  API Key:    ${YELLOW}$API_KEY${NC}"
    echo -e "  API Secret: ${YELLOW}$API_SECRET${NC}"
    echo ""
    echo -e "${BLUE}Initial Balances:${NC}"
    echo -e "  USDT: ${YELLOW}100,000.00${NC}"
    echo -e "  BTC:  ${YELLOW}10.00${NC}"
    echo -e "  ETH:  ${YELLOW}100.00${NC}"
    echo ""
    echo -e "${GREEN}These credentials are already configured in apitest.http${NC}"
    echo -e "${GREEN}You can now run API tests!${NC}"
    echo ""
else
    echo -e "${RED}Failed to create test user${NC}"
    exit 1
fi
