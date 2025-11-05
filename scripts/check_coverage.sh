#!/bin/bash
# 测试覆盖率阈值检查脚本（排除 testutil 包）
# 项目: Quicksilver
# 最后更新: 2025-11-05

set -e

COVERAGE_FILE="${1:-coverage.out}"
THRESHOLD="${2:-70}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Please run: make test-coverage"
    exit 1
fi

# 计算覆盖率（排除 testutil 包）
coverage_output=$(go tool cover -func="$COVERAGE_FILE" | grep -v "internal/testutil")
total_line=$(echo "$coverage_output" | grep "total:")

if [ -z "$total_line" ]; then
    # 如果没有 total 行，手动计算
    coverage=$(echo "$coverage_output" | awk '
        {
            if ($3 != "") {
                gsub(/%/, "", $3)
                sum += $3
                count++
            }
        }
        END {
            if (count > 0) print sum/count
            else print 0
        }
    ')
else
    coverage=$(echo "$total_line" | awk '{print $3}' | sed 's/%//')
fi

if [ -z "$coverage" ]; then
    echo "❌ Failed to parse coverage data"
    exit 1
fi

echo ""
echo "=========================================="
echo "  Test Coverage Report"
echo "=========================================="
echo "Coverage File: $COVERAGE_FILE"
echo "Total Coverage: ${coverage}% (excluding testutil)"
echo "Threshold: ${THRESHOLD}%"
echo "=========================================="
echo ""

# 使用 bc 进行浮点数比较
if command -v bc &> /dev/null; then
    result=$(echo "$coverage < $THRESHOLD" | bc -l)
    if [ "$result" -eq 1 ]; then
        echo "❌ FAILED: Coverage ${coverage}% is below ${THRESHOLD}% threshold"
        echo ""
        echo "Top uncovered packages (excluding testutil):"
        go tool cover -func="$COVERAGE_FILE" | \
        grep -v "internal/testutil" | \
        grep -v "100.0%" | \
        tail -10
        echo ""
        exit 1
    fi
else
    # 回退到整数比较
    coverage_int=${coverage%.*}
    if [ "$coverage_int" -lt "$THRESHOLD" ]; then
        echo "❌ FAILED: Coverage ${coverage}% is below ${THRESHOLD}% threshold"
        exit 1
    fi
fi

echo "✅ PASSED: Coverage ${coverage}% meets threshold"
echo ""

# 按包显示覆盖率摘要（排除 testutil）
echo "Coverage by package (excluding testutil):"
go tool cover -func="$COVERAGE_FILE" | \
grep -v "internal/testutil" | \
grep -E "^github.com" | \
awk '{printf "  %-60s %s\n", $1, $3}' | \
sort -u

exit 0
