#!/bin/bash

# 应用数据库索引优化脚本
# 使用方法: ./scripts/apply-indexes.sh

echo "=== 应用数据库索引优化 ==="
echo ""

# 从配置文件读取数据库连接信息
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-xingyunpan}
DB_PASSWORD=${DB_PASSWORD:-xingyunpan123}
DB_NAME=${DB_NAME:-xingyunpan}

echo "数据库连接信息:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# 执行索引创建脚本
echo "执行索引创建脚本..."
mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME < scripts/add-indexes.sql

if [ $? -eq 0 ]; then
    echo ""
    echo "=== 索引优化完成 ==="
    echo ""
    echo "可以使用以下命令验证索引:"
    echo "  mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME -e 'SHOW INDEX FROM user_files;'"
else
    echo ""
    echo "=== 索引优化失败 ==="
    exit 1
fi
