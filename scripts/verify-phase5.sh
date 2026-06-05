#!/bin/bash
# Phase 5 部署检查脚本
# 验证 Phase 5 所有组件是否正确部署

set -e

echo "=========================================="
echo "Phase 5 部署检查开始"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查计数器
PASSED=0
FAILED=0

# 检查函数
check_pass() {
    echo -e "${GREEN}✓ $1${NC}"
    ((PASSED++))
}

check_fail() {
    echo -e "${RED}✗ $1${NC}"
    ((FAILED++))
}

check_warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 1. 检查数据库表是否存在
echo "1. 检查数据库表..."
echo "-------------------------------------------"

# 读取数据库配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USERNAME:-xingyunpan}
DB_PASS=${DB_PASSWORD}
DB_NAME=${DB_NAME:-xingyunpan}

if [ -z "$DB_PASS" ]; then
    check_fail "数据库密码未设置（DB_PASSWORD 环境变量）"
    exit 1
fi

# 检查表是否存在
TABLES=("shares" "share_files" "recycle_bin" "file_versions" "collaborations")
for table in "${TABLES[@]}"; do
    result=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sse "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$DB_NAME' AND table_name='$table'")
    if [ "$result" -eq 1 ]; then
        check_pass "表 $table 存在"
    else
        check_fail "表 $table 不存在"
    fi
done
echo ""

# 2. 检查索引是否存在
echo "2. 检查数据库索引..."
echo "-------------------------------------------"

# shares 表索引
INDEXES=(
    "shares:idx_share_token"
    "shares:idx_user_id"
    "shares:idx_expires_at"
    "recycle_bin:idx_user_expires"
    "file_versions:idx_file_versions"
    "file_versions:idx_current_version"
    "collaborations:unique_collaboration"
    "collaborations:idx_file_owner"
    "collaborations:idx_collaborator"
    "user_files:idx_file_name"
    "user_files:idx_modified_at"
    "user_files:idx_file_size"
)

for index_info in "${INDEXES[@]}"; do
    IFS=':' read -r table index <<< "$index_info"
    result=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sse "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema='$DB_NAME' AND table_name='$table' AND index_name='$index'")
    if [ "$result" -gt 0 ]; then
        check_pass "索引 $table.$index 存在"
    else
        check_fail "索引 $table.$index 不存在"
    fi
done
echo ""

# 3. 检查 API 端点是否可访问
echo "3. 检查 API 端点..."
echo "-------------------------------------------"

# 检查服务是否运行
BASE_URL=${BASE_URL:-http://localhost:8080}
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || echo "000")

if [ "$HEALTH_CHECK" -eq 200 ]; then
    check_pass "服务健康检查通过"
else
    check_fail "服务健康检查失败（HTTP $HEALTH_CHECK）"
    check_warn "请确保服务正在运行"
    exit 1
fi

# 检查 Swagger 文档
SWAGGER_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/swagger/index.html" || echo "000")
if [ "$SWAGGER_CHECK" -eq 200 ]; then
    check_pass "Swagger 文档可访问"
else
    check_warn "Swagger 文档不可访问（HTTP $SWAGGER_CHECK）"
fi

# 检查 Prometheus metrics
METRICS_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/metrics" || echo "000")
if [ "$METRICS_CHECK" -eq 200 ]; then
    check_pass "Prometheus metrics 端点可访问"
else
    check_warn "Prometheus metrics 端点不可访问（HTTP $METRICS_CHECK）"
fi
echo ""

# 4. 检查配置文件
echo "4. 检查配置文件..."
echo "-------------------------------------------"

if [ -f "configs/config.yaml" ] || [ -f "configs/config.prod.yaml" ]; then
    check_pass "配置文件存在"
else
    check_fail "配置文件不存在"
fi

# 检查必需的环境变量
REQUIRED_VARS=("DB_PASSWORD" "JWT_SECRET")
for var in "${REQUIRED_VARS[@]}"; do
    if [ -n "${!var}" ]; then
        check_pass "环境变量 $var 已设置"
    else
        check_fail "环境变量 $var 未设置"
    fi
done
echo ""

# 5. 检查 Phase 5 特定配置
echo "5. 检查 Phase 5 配置..."
echo "-------------------------------------------"

# 检查 bcrypt cost 配置
if [ -n "$BCRYPT_COST" ]; then
    if [ "$BCRYPT_COST" -ge 10 ] && [ "$BCRYPT_COST" -le 14 ]; then
        check_pass "BCRYPT_COST 配置合理（$BCRYPT_COST）"
    else
        check_warn "BCRYPT_COST 配置不在推荐范围（10-14）: $BCRYPT_COST"
    fi
else
    check_pass "BCRYPT_COST 使用默认值（12）"
fi

# 检查限流配置
RATE_LIMIT_VARS=("RATE_LIMIT_SEARCH" "RATE_LIMIT_FILE_OPS" "RATE_LIMIT_SHARE_VERIFY")
for var in "${RATE_LIMIT_VARS[@]}"; do
    if [ -n "${!var}" ]; then
        check_pass "限流配置 $var 已设置"
    else
        check_pass "$var 使用默认值"
    fi
done

# 检查回收站配置
if [ -n "$RECYCLE_BIN_RETENTION_DAYS" ]; then
    check_pass "回收站保留期配置: $RECYCLE_BIN_RETENTION_DAYS 天"
else
    check_pass "回收站保留期使用默认值（30 天）"
fi

# 检查版本限制配置
if [ -n "$MAX_FILE_VERSIONS" ]; then
    check_pass "文件版本限制配置: $MAX_FILE_VERSIONS"
else
    check_pass "文件版本限制使用默认值（10）"
fi
echo ""

# 6. 检查 Worker 进程
echo "6. 检查 Worker 进程..."
echo "-------------------------------------------"

# 检查 worker 是否运行（通过进程名或 systemd）
if command -v systemctl &> /dev/null; then
    if systemctl is-active --quiet xingyunpan-worker; then
        check_pass "Worker 进程运行中（systemd）"
    else
        check_warn "Worker 进程未运行（systemd）"
    fi
elif pgrep -f "xingyunpan.*worker" > /dev/null; then
    check_pass "Worker 进程运行中"
else
    check_warn "Worker 进程未运行"
fi
echo ""

# 7. 检查日志目录
echo "7. 检查日志和存储目录..."
echo "-------------------------------------------"

LOG_DIR=${LOG_DIR:-/data/xingyunpan/logs}
STORAGE_DIR=${STORAGE_PATH:-/data/xingyunpan/storage}

if [ -d "$LOG_DIR" ]; then
    check_pass "日志目录存在: $LOG_DIR"
else
    check_fail "日志目录不存在: $LOG_DIR"
fi

if [ -d "$STORAGE_DIR" ]; then
    check_pass "存储目录存在: $STORAGE_DIR"
else
    check_fail "存储目录不存在: $STORAGE_DIR"
fi
echo ""

# 8. 运行简单的 API 测试
echo "8. 运行 API 功能测试..."
echo "-------------------------------------------"

# 注意：这里只做基本的端点检查，不做完整的集成测试
# 完整测试应该在测试环境运行

# 测试登录（获取 token）
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"test_user","password":"test_password"}' || echo "")

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    check_pass "登录 API 响应正常"
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
else
    check_warn "登录 API 测试跳过（需要测试用户）"
    TOKEN=""
fi

# 如果有 token，测试 Phase 5 端点
if [ -n "$TOKEN" ]; then
    # 测试搜索 API
    SEARCH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/search/files?keyword=test" \
        -H "Authorization: Bearer $TOKEN" || echo "000")
    if [ "$SEARCH_CHECK" -eq 200 ]; then
        check_pass "搜索 API 可访问"
    else
        check_warn "搜索 API 响应异常（HTTP $SEARCH_CHECK）"
    fi
    
    # 测试回收站 API
    RECYCLE_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/recycle" \
        -H "Authorization: Bearer $TOKEN" || echo "000")
    if [ "$RECYCLE_CHECK" -eq 200 ]; then
        check_pass "回收站 API 可访问"
    else
        check_warn "回收站 API 响应异常（HTTP $RECYCLE_CHECK）"
    fi
    
    # 测试分享列表 API
    SHARE_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/shares/me" \
        -H "Authorization: Bearer $TOKEN" || echo "000")
    if [ "$SHARE_CHECK" -eq 200 ]; then
        check_pass "分享 API 可访问"
    else
        check_warn "分享 API 响应异常（HTTP $SHARE_CHECK）"
    fi
else
    check_warn "跳过 API 功能测试（需要有效的认证 token）"
fi
echo ""

# 9. 总结
echo "=========================================="
echo "检查完成"
echo "=========================================="
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ Phase 5 部署检查全部通过！${NC}"
    exit 0
else
    echo -e "${RED}✗ Phase 5 部署检查发现 $FAILED 个问题，请修复后重试${NC}"
    exit 1
fi
