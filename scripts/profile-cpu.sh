#!/bin/bash

# CPU 性能分析脚本
# 使用方法: ./scripts/profile-cpu.sh

echo "=== CPU 性能分析 ==="
echo ""

# 检查 pprof 端点是否可用
if ! curl -s http://localhost:6060/debug/pprof/ > /dev/null; then
    echo "错误: pprof 端点不可用，请确保服务器运行在开发模式"
    exit 1
fi

echo "1. 采集 CPU profile (30秒)..."
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30

echo ""
echo "2. 分析 CPU profile..."
go tool pprof -http=:8081 cpu.prof

echo ""
echo "=== CPU 性能分析完成 ==="
echo "Profile 文件: cpu.prof"
echo "可以使用以下命令查看:"
echo "  go tool pprof cpu.prof"
echo "  go tool pprof -http=:8081 cpu.prof"
