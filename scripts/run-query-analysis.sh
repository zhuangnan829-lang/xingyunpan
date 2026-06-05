#!/bin/bash
# 路径: scripts/run-query-analysis.sh
# 查询性能分析脚本

set -e

echo "========================================"
echo "Phase 5 查询性能分析"
echo "========================================"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境"
    exit 1
fi

# 检查 MySQL 连接
echo "检查 MySQL 连接..."
if ! go run scripts/test-db-connection.go &> /dev/null; then
    echo "错误: MySQL 连接失败"
    echo "请确保 MySQL 正在运行并且配置正确"
    exit 1
fi

echo ""
echo "========================================"
echo "运行 EXPLAIN 分析"
echo "========================================"
echo ""

# 运行查询分析
go run scripts/analyze-queries.go

echo ""
echo "========================================"
echo "分析完成"
echo "========================================"
echo ""

echo "优化建议:"
echo "  1. 确保所有索引都已创建（运行 scripts/migrations/phase5_schema.sql）"
echo "  2. 检查 EXPLAIN 输出中的 \"key\" 列，确认使用了正确的索引"
echo "  3. 如果 \"type\" 列显示 \"ALL\"，说明进行了全表扫描，需要优化"
echo "  4. 如果 \"rows\" 列数值很大，考虑添加更多过滤条件或优化索引"
echo ""

exit 0
