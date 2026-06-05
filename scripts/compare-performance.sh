#!/bin/bash

# 性能对比脚本
# 使用方法: ./scripts/compare-performance.sh <before_file> <after_file>

if [ $# -ne 2 ]; then
    echo "使用方法: ./scripts/compare-performance.sh <before_file> <after_file>"
    echo "示例: ./scripts/compare-performance.sh benchmark-before.txt benchmark-after.txt"
    exit 1
fi

BEFORE_FILE=$1
AFTER_FILE=$2

echo "=== 性能对比分析 ==="
echo ""

if [ ! -f "$BEFORE_FILE" ]; then
    echo "错误: 文件不存在: $BEFORE_FILE"
    exit 1
fi

if [ ! -f "$AFTER_FILE" ]; then
    echo "错误: 文件不存在: $AFTER_FILE"
    exit 1
fi

echo "优化前: $BEFORE_FILE"
echo "优化后: $AFTER_FILE"
echo ""

echo "正在分析性能数据..."
echo ""

# 提取关键指标
extract_metric() {
    local file=$1
    local pattern=$2
    grep "$pattern" "$file" | awk '{print $NF}'
}

# 提取响应时间
before_time=$(extract_metric "$BEFORE_FILE" "Time per request.*mean")
after_time=$(extract_metric "$AFTER_FILE" "Time per request.*mean")

# 提取吞吐量
before_rps=$(extract_metric "$BEFORE_FILE" "Requests per second")
after_rps=$(extract_metric "$AFTER_FILE" "Requests per second")

echo "=== 对比结果 ==="
echo ""
echo "响应时间:"
echo "  优化前: $before_time"
echo "  优化后: $after_time"
echo ""
echo "吞吐量:"
echo "  优化前: $before_rps"
echo "  优化后: $after_rps"
echo ""
echo "预期改善:"
echo "  - 响应时间减少 30%+"
echo "  - 吞吐量提升 40%+"
echo "  - 数据库查询时间减少 50%+"
echo ""
