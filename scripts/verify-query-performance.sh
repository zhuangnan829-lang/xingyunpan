#!/bin/bash

# 查询性能验证脚本 - Linux/Mac 版本
# 用于验证索引优化后的查询性能提升

set -e

echo "========================================"
echo "查询性能验证"
echo "========================================"
echo ""

# 检查数据库连接信息
if [ -z "$DB_PASSWORD" ]; then
    echo "错误: 请设置 DB_PASSWORD 环境变量"
    echo ""
    echo "使用方法:"
    echo "  export DB_HOST=localhost"
    echo "  export DB_PORT=3306"
    echo "  export DB_USER=xingyunpan"
    echo "  export DB_PASSWORD=your_password"
    echo "  export DB_NAME=xingyunpan"
    echo "  ./scripts/verify-query-performance.sh"
    echo ""
    exit 1
fi

# 设置默认值
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-xingyunpan}
DB_NAME=${DB_NAME:-xingyunpan}

echo "数据库连接信息:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# 运行 Go 程序
echo "正在执行查询性能测试..."
echo ""

go run scripts/verify-query-performance.go

echo ""
echo "========================================"
echo "验证完成"
echo "========================================"
echo ""
echo "报告已生成，请查看上方输出"
echo ""
echo "下一步:"
echo "  1. 如果所有查询都使用了索引，说明优化成功"
echo "  2. 如果查询时间都在 10ms 以内，性能优秀"
echo "  3. 如果有查询未使用索引，请检查索引是否创建成功"
echo "  4. 运行 ANALYZE TABLE 更新统计信息"
echo ""
