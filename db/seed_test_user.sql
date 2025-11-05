-- ============================================================================
-- Quicksilver 测试用户初始化脚本
-- 
-- 功能:
-- 1. 创建一个测试用户 (test@quicksilver.local)
-- 2. 初始化 BTC 和 USDT 余额
-- 3. 提供固定的 API Key 和 Secret 用于测试
--
-- 使用方法:
-- psql -h localhost -U postgres -d quicksilver < db/seed_test_user.sql
-- 
-- 或在 Docker 容器中:
-- docker exec -i quicksilver-postgres psql -U postgres -d quicksilver < db/seed_test_user.sql
-- ============================================================================

-- 设置时区
SET TIME ZONE 'UTC';

-- 删除旧的测试用户数据 (如果存在)
DELETE FROM trades WHERE user_id IN (SELECT id FROM users WHERE email = 'test@quicksilver.local');
DELETE FROM orders WHERE user_id IN (SELECT id FROM users WHERE email = 'test@quicksilver.local');
DELETE FROM balances WHERE user_id IN (SELECT id FROM users WHERE email = 'test@quicksilver.local');
DELETE FROM users WHERE email = 'test@quicksilver.local';

-- 创建测试用户
-- API Key: qs-test-api-key-2024
-- API Secret: qs-test-api-secret-change-in-production
INSERT INTO users (email, username, api_key, api_secret, status, created_at, updated_at)
VALUES (
    'test@quicksilver.local',
    'test_user',
    'qs-test-api-key-2024',
    'qs-test-api-secret-change-in-production',
    'active',
    NOW(),
    NOW()
) ON CONFLICT (email) DO UPDATE SET
    api_key = EXCLUDED.api_key,
    api_secret = EXCLUDED.api_secret,
    status = EXCLUDED.status,
    updated_at = NOW();

-- 获取用户 ID
DO $$
DECLARE
    test_user_id INT;
BEGIN
    SELECT id INTO test_user_id FROM users WHERE email = 'test@quicksilver.local';

    -- 初始化测试余额
    -- USDT: 100,000 (用于买入测试)
    INSERT INTO balances (user_id, asset, available, locked, created_at, updated_at)
    VALUES (
        test_user_id,
        'USDT',
        100000.00000000,
        0.00000000,
        NOW(),
        NOW()
    ) ON CONFLICT (user_id, asset) DO UPDATE SET
        available = EXCLUDED.available,
        locked = 0.00000000,
        updated_at = NOW();

    -- BTC: 10 (用于卖出测试)
    INSERT INTO balances (user_id, asset, available, locked, created_at, updated_at)
    VALUES (
        test_user_id,
        'BTC',
        10.00000000,
        0.00000000,
        NOW(),
        NOW()
    ) ON CONFLICT (user_id, asset) DO UPDATE SET
        available = EXCLUDED.available,
        locked = 0.00000000,
        updated_at = NOW();

    -- ETH: 100 (用于卖出测试)
    INSERT INTO balances (user_id, asset, available, locked, created_at, updated_at)
    VALUES (
        test_user_id,
        'ETH',
        100.00000000,
        0.00000000,
        NOW(),
        NOW()
    ) ON CONFLICT (user_id, asset) DO UPDATE SET
        available = EXCLUDED.available,
        locked = 0.00000000,
        updated_at = NOW();

    -- 打印成功信息
    RAISE NOTICE '✅ 测试用户创建成功!';
    RAISE NOTICE '----------------------------------------';
    RAISE NOTICE 'Email: test@quicksilver.local';
    RAISE NOTICE 'API Key: qs-test-api-key-2024';
    RAISE NOTICE 'API Secret: qs-test-api-secret-change-in-production';
    RAISE NOTICE '----------------------------------------';
    RAISE NOTICE '初始余额:';
    RAISE NOTICE '  USDT: 100,000.00';
    RAISE NOTICE '  BTC: 10.00';
    RAISE NOTICE '  ETH: 100.00';
    RAISE NOTICE '----------------------------------------';
END $$;

-- 验证数据
SELECT 
    u.id,
    u.email,
    u.username,
    u.api_key,
    u.status,
    u.created_at
FROM users u
WHERE u.email = 'test@quicksilver.local';

SELECT 
    b.user_id,
    b.asset,
    b.available,
    b.locked
FROM balances b
JOIN users u ON b.user_id = u.id
WHERE u.email = 'test@quicksilver.local'
ORDER BY b.asset;
