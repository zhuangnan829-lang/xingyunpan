#!/bin/bash
# 路径: scripts/verify.sh
# Phase 1 快速验证脚本

echo "=== 星云盘 V2 Phase 1 验证脚本 ==="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_service() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1"
        return 1
    fi
}

# 1. 检查 Docker 服务
echo "1. 检查 Docker 服务..."
docker-compose ps > /dev/null 2>&1
check_service "Docker Compose 可用"

# 2. 检查 MySQL
echo ""
echo "2. 检查 MySQL..."
docker exec xingyunpan-mysql mysql -uroot -ppassword -e "SELECT 1" > /dev/null 2>&1
check_service "MySQL 连接正常"

# 3. 检查 Redis
echo ""
echo "3. 检查 Redis..."
docker exec xingyunpan-redis redis-cli PING > /dev/null 2>&1
check_service "Redis 连接正常"

# 4. 运行单元测试
echo ""
echo "4. 运行单元测试..."
echo "   - 测试日志工具..."
go test ./pkg/logger/ -run TestInit > /dev/null 2>&1
check_service "日志工具测试"

echo "   - 测试响应格式..."
go test ./pkg/response/ -run TestSuccess > /dev/null 2>&1
check_service "响应格式测试"

echo "   - 测试配置管理..."
go test ./internal/config/ -run TestLoadConfig > /dev/null 2>&1
check_service "配置管理测试"

echo "   - 测试数据模型..."
go test ./internal/model/ -run TestBaseModel > /dev/null 2>&1
check_service "数据模型测试"

# 5. 检查数据库表
echo ""
echo "5. 检查数据库表..."
TABLES=$(docker exec xingyunpan-mysql mysql -uroot -ppassword xingyunpan -e "SHOW TABLES;" 2>/dev/null | grep -E "users|physical_files|user_files" | wc -l)
if [ "$TABLES" -eq 3 ]; then
    check_service "数据库表已创建 (3/3)"
else
    echo -e "${YELLOW}⚠${NC} 数据库表未完全创建 ($TABLES/3)"
    echo "   运行迁移: go run scripts/migrate.go"
fi

# 6. 检查服务器（如果正在运行）
echo ""
echo "6. 检查服务器..."
curl -s http://localhost:8080/ping > /dev/null 2>&1
if [ $? -eq 0 ]; then
    check_service "服务器正在运行"
    
    # 测试健康检查
    HEALTH=$(curl -s http://localhost:8080/health | grep -o '"code":200' | wc -l)
    if [ "$HEALTH" -eq 1 ]; then
        check_service "健康检查通过"
    else
        echo -e "${RED}✗${NC} 健康检查失败"
    fi
else
    echo -e "${YELLOW}⚠${NC} 服务器未运行"
    echo "   启动服务器: go run cmd/server/main.go"
fi

echo ""
echo "=== 验证完成 ==="
echo ""
echo "详细测试文档: docs/phase1-integration-test.md"
