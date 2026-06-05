#!/bin/bash

# 星云盘 V2 生产环境配置验证脚本
# Production Setup Verification Script for Xingyunpan V2

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 计数器
PASSED=0
FAILED=0
WARNINGS=0

# 日志函数
log_pass() {
    echo -e "${GREEN}✓${NC} $1"
    PASSED=$((PASSED + 1))
}

log_fail() {
    echo -e "${RED}✗${NC} $1"
    FAILED=$((FAILED + 1))
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    WARNINGS=$((WARNINGS + 1))
}

log_info() {
    echo -e "ℹ $1"
}

echo "========================================"
echo "星云盘 V2 生产环境配置验证"
echo "========================================"
echo ""

# 1. 检查目录结构
echo "1. 检查目录结构"
echo "----------------------------------------"

if [ -d "/data/xingyunpan/logs" ]; then
    log_pass "日志目录存在: /data/xingyunpan/logs"
else
    log_fail "日志目录不存在: /data/xingyunpan/logs"
fi

if [ -d "/data/xingyunpan/storage" ]; then
    log_pass "存储目录存在: /data/xingyunpan/storage"
else
    log_fail "存储目录不存在: /data/xingyunpan/storage"
fi

if [ -d "/data/xingyunpan/temp" ]; then
    log_pass "临时目录存在: /data/xingyunpan/temp"
else
    log_fail "临时目录不存在: /data/xingyunpan/temp"
fi

if [ -d "/backup/xingyunpan" ]; then
    log_pass "备份目录存在: /backup/xingyunpan"
else
    log_fail "备份目录不存在: /backup/xingyunpan"
fi

echo ""

# 2. 检查配置文件
echo "2. 检查配置文件"
echo "----------------------------------------"

if [ -f "configs/config.prod.yaml" ]; then
    log_pass "生产配置文件存在"
else
    log_fail "生产配置文件不存在: configs/config.prod.yaml"
fi

if [ -f "configs/.env.production" ]; then
    log_pass "环境变量文件存在"
    
    # 检查文件权限
    PERM=$(stat -c %a "configs/.env.production" 2>/dev/null || stat -f %A "configs/.env.production")
    if [ "$PERM" = "600" ]; then
        log_pass "环境变量文件权限正确 (600)"
    else
        log_warn "环境变量文件权限不正确: $PERM (应为 600)"
    fi
else
    log_fail "环境变量文件不存在: configs/.env.production"
fi

echo ""

# 3. 检查环境变量
echo "3. 检查环境变量"
echo "----------------------------------------"

if [ -f "configs/.env.production" ]; then
    source configs/.env.production
    
    if [ -n "$DB_PASSWORD" ]; then
        log_pass "DB_PASSWORD 已设置"
    else
        log_fail "DB_PASSWORD 未设置"
    fi
    
    if [ -n "$JWT_SECRET" ]; then
        JWT_LEN=${#JWT_SECRET}
        if [ $JWT_LEN -ge 32 ]; then
            log_pass "JWT_SECRET 已设置且长度足够 ($JWT_LEN 字符)"
        else
            log_warn "JWT_SECRET 长度不足: $JWT_LEN 字符 (建议至少 32 字符)"
        fi
    else
        log_fail "JWT_SECRET 未设置"
    fi
    
    if [ -n "$REDIS_PASSWORD" ]; then
        log_pass "REDIS_PASSWORD 已设置"
    else
        log_warn "REDIS_PASSWORD 未设置（建议设置）"
    fi
fi

echo ""

# 4. 检查脚本权限
echo "4. 检查脚本权限"
echo "----------------------------------------"

SCRIPTS=(
    "scripts/backup.sh"
    "scripts/restore.sh"
    "scripts/rotate-logs.sh"
    "scripts/check-disk-space.sh"
    "scripts/verify-backup.sh"
)

for script in "${SCRIPTS[@]}"; do
    if [ -f "$script" ]; then
        if [ -x "$script" ]; then
            log_pass "$(basename $script) 可执行"
        else
            log_fail "$(basename $script) 不可执行"
        fi
    else
        log_warn "$(basename $script) 不存在"
    fi
done

echo ""

# 5. 检查数据库连接
echo "5. 检查数据库连接"
echo "----------------------------------------"

if [ -n "$DB_HOST" ] && [ -n "$DB_USER" ] && [ -n "$DB_PASSWORD" ]; then
    if command -v mysql &> /dev/null; then
        if mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" &> /dev/null; then
            log_pass "数据库连接成功"
        else
            log_fail "数据库连接失败"
        fi
    else
        log_warn "mysql 命令不存在，跳过数据库连接测试"
    fi
else
    log_warn "数据库配置不完整，跳过连接测试"
fi

echo ""

# 6. 检查 Redis 连接
echo "6. 检查 Redis 连接"
echo "----------------------------------------"

if [ -n "$REDIS_HOST" ]; then
    if command -v redis-cli &> /dev/null; then
        if [ -n "$REDIS_PASSWORD" ]; then
            if redis-cli -h "$REDIS_HOST" -p "${REDIS_PORT:-6379}" -a "$REDIS_PASSWORD" ping &> /dev/null; then
                log_pass "Redis 连接成功"
            else
                log_fail "Redis 连接失败"
            fi
        else
            if redis-cli -h "$REDIS_HOST" -p "${REDIS_PORT:-6379}" ping &> /dev/null; then
                log_pass "Redis 连接成功"
            else
                log_fail "Redis 连接失败"
            fi
        fi
    else
        log_warn "redis-cli 命令不存在，跳过 Redis 连接测试"
    fi
else
    log_warn "Redis 配置不完整，跳过连接测试"
fi

echo ""

# 7. 检查定时任务
echo "7. 检查定时任务"
echo "----------------------------------------"

# 检查 systemd timer
if command -v systemctl &> /dev/null; then
    if systemctl list-timers | grep -q "xingyunpan-backup.timer"; then
        log_pass "Systemd backup timer 已配置"
    else
        log_warn "Systemd backup timer 未配置"
    fi
    
    if systemctl list-timers | grep -q "xingyunpan-monitoring.timer"; then
        log_pass "Systemd monitoring timer 已配置"
    else
        log_warn "Systemd monitoring timer 未配置"
    fi
    
    if systemctl list-timers | grep -q "xingyunpan-cleanup.timer"; then
        log_pass "Systemd cleanup timer 已配置"
    else
        log_warn "Systemd cleanup timer 未配置"
    fi
fi

# 检查 crontab
if command -v crontab &> /dev/null; then
    if crontab -l -u xingyunpan 2>/dev/null | grep -q "backup.sh"; then
        log_pass "Crontab 备份任务已配置"
    else
        log_warn "Crontab 备份任务未配置"
    fi
fi

echo ""

# 8. 检查磁盘空间
echo "8. 检查磁盘空间"
echo "----------------------------------------"

check_disk_usage() {
    local path=$1
    local threshold=80
    
    if [ -d "$path" ]; then
        local usage=$(df -h "$path" | awk 'NR==2 {print $5}' | sed 's/%//')
        local available=$(df -h "$path" | awk 'NR==2 {print $4}')
        
        if [ "$usage" -lt "$threshold" ]; then
            log_pass "$path: ${usage}% 使用，${available} 可用"
        else
            log_warn "$path: ${usage}% 使用，${available} 可用 (超过 ${threshold}%)"
        fi
    fi
}

check_disk_usage "/data/xingyunpan"
check_disk_usage "/backup/xingyunpan"

echo ""

# 9. 检查应用状态
echo "9. 检查应用状态"
echo "----------------------------------------"

if command -v systemctl &> /dev/null; then
    if systemctl is-active --quiet xingyunpan-server; then
        log_pass "Server 服务运行中"
    else
        log_warn "Server 服务未运行"
    fi
    
    if systemctl is-active --quiet xingyunpan-worker; then
        log_pass "Worker 服务运行中"
    else
        log_warn "Worker 服务未运行"
    fi
fi

# 检查健康端点
if command -v curl &> /dev/null; then
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_pass "应用健康检查通过"
    else
        log_warn "应用健康检查失败（应用可能未启动）"
    fi
fi

echo ""

# 10. 检查备份
echo "10. 检查备份"
echo "----------------------------------------"

if [ -d "/backup/xingyunpan" ]; then
    BACKUP_COUNT=$(ls -1 /backup/xingyunpan/xingyunpan_*.sql.gz 2>/dev/null | wc -l)
    
    if [ $BACKUP_COUNT -gt 0 ]; then
        log_pass "找到 $BACKUP_COUNT 个备份文件"
        
        # 检查最新备份
        LATEST_BACKUP=$(ls -t /backup/xingyunpan/xingyunpan_*.sql.gz 2>/dev/null | head -n 1)
        if [ -n "$LATEST_BACKUP" ]; then
            BACKUP_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_BACKUP" 2>/dev/null || stat -f %m "$LATEST_BACKUP")) / 3600 ))
            
            if [ $BACKUP_AGE -lt 26 ]; then
                log_pass "最新备份: $(basename $LATEST_BACKUP) ($BACKUP_AGE 小时前)"
            else
                log_warn "最新备份已超过 $BACKUP_AGE 小时"
            fi
        fi
    else
        log_warn "未找到备份文件（可能尚未执行首次备份）"
    fi
fi

echo ""

# 总结
echo "========================================"
echo "验证总结"
echo "========================================"
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${YELLOW}警告: $WARNINGS${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}✓ 所有检查通过！生产环境配置正确。${NC}"
        exit 0
    else
        echo -e "${YELLOW}⚠ 存在一些警告，建议检查并修复。${NC}"
        exit 0
    fi
else
    echo -e "${RED}✗ 存在配置问题，请修复后重新验证。${NC}"
    exit 1
fi
