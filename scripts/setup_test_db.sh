#!/bin/bash

###############################################################################
# Quicksilver 测试数据库设置脚本
#
# 功能:
# - 创建 PostgreSQL 测试数据库
# - 为并发测试提供支持
#
# 使用方法:
#   ./scripts/setup_test_db.sh          # 创建测试数据库
#   ./scripts/setup_test_db.sh drop     # 删除测试数据库
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
DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-pgdb}"
TEST_DB_NAME="quicksilver_test"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Quicksilver Test Database Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查依赖
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql is not installed${NC}"
    echo -e "${YELLOW}Install it with: brew install postgresql${NC}"
    exit 1
fi

# 删除数据库
if [ "$1" = "drop" ]; then
    echo -e "${YELLOW}Dropping test database...${NC}"
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" postgres
    echo -e "${GREEN}✅ Test database dropped${NC}"
    exit 0
fi

# 创建数据库
echo -e "${YELLOW}Creating test database...${NC}"
PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" postgres 2>/dev/null || true
PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $TEST_DB_NAME;" postgres

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ Test Database Created Successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}Database Details:${NC}"
    echo -e "  Host:     ${YELLOW}$DB_HOST${NC}"
    echo -e "  Port:     ${YELLOW}$DB_PORT${NC}"
    echo -e "  Database: ${YELLOW}$TEST_DB_NAME${NC}"
    echo -e "  User:     ${YELLOW}$DB_USER${NC}"
    echo ""
    echo -e "${BLUE}Run tests with:${NC}"
    echo -e "  ${YELLOW}make test-pg${NC}"
    echo ""
    echo -e "${BLUE}Or manually:${NC}"
    echo -e "  ${YELLOW}TEST_DB=postgres make test${NC}"
    echo ""
else
    echo -e "${RED}Failed to create test database${NC}"
    exit 1
fi
