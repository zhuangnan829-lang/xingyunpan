#!/bin/bash
# 路径: scripts/monitor-cache.sh
# Redis 缓存监控脚本

set -e

echo "========================================"
echo "Redis 缓存监控"
echo "========================================"
echo ""

# 检查 Redis 连接
echo "检查 Redis 连接..."
if ! redis-cli ping &> /dev/null; then
    echo "错误: Redis 连接失败"
    echo "请确保 Redis 正在运行"
    exit 1
fi

echo "Redis 连接正常"
echo ""

# 运行监控脚本
go run scripts/monitor-cache.go

echo ""
echo "========================================"
echo "监控完成"
echo "========================================"

exit 0
