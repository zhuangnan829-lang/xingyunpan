#!/bin/bash

# 星云盘 V2 备份验证脚本
# Backup Verification Script for Xingyunpan V2

set -e

# 配置
BACKUP_DIR="${BACKUP_DIR:-/backup/xingyunpan}"
MAX_AGE_HOURS=26  # 超过 26 小时没有新备份则告警

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log "=== 备份验证开始 ==="
log "备份目录: $BACKUP_DIR"

# 检查备份目录是否存在
if [ ! -d "$BACKUP_DIR" ]; then
    log "错误: 备份目录不存在: $BACKUP_DIR"
    exit 1
fi

# 查找最新备份
LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/xingyunpan_*.sql.gz 2>/dev/null | head -n 1)

if [ -z "$LATEST_BACKUP" ]; then
    log "错误: 没有找到备份文件"
    
    # 发送告警
    if [ -n "$WEBHOOK_URL" ]; then
        curl -X POST "$WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "{\"text\":\"备份告警: 没有找到备份文件\"}" \
            2>/dev/null
    fi
    
    exit 1
fi

log "最新备份: $(basename $LATEST_BACKUP)"

# 检查备份年龄
BACKUP_TIME=$(stat -c %Y "$LATEST_BACKUP" 2>/dev/null || stat -f %m "$LATEST_BACKUP")
CURRENT_TIME=$(date +%s)
BACKUP_AGE=$(( ($CURRENT_TIME - $BACKUP_TIME) / 3600 ))

log "备份年龄: $BACKUP_AGE 小时"

if [ $BACKUP_AGE -gt $MAX_AGE_HOURS ]; then
    log "错误: 最新备份已超过 $MAX_AGE_HOURS 小时"
    
    # 发送告警
    if [ -n "$WEBHOOK_URL" ]; then
        curl -X POST "$WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "{\"text\":\"备份告警: 最新备份已超过 $BACKUP_AGE 小时\"}" \
            2>/dev/null
    fi
    
    exit 1
fi

# 检查备份文件大小
FILE_SIZE=$(stat -c %s "$LATEST_BACKUP" 2>/dev/null || stat -f %z "$LATEST_BACKUP")
FILE_SIZE_MB=$(( $FILE_SIZE / 1048576 ))

log "备份大小: ${FILE_SIZE_MB}MB"

if [ $FILE_SIZE -lt 1024 ]; then
    log "错误: 备份文件太小，可能损坏"
    exit 1
fi

# 验证 gzip 完整性
log "验证压缩文件完整性..."
if gzip -t "$LATEST_BACKUP" 2>/dev/null; then
    log "✓ 压缩文件完整性验证通过"
else
    log "错误: 压缩文件损坏"
    
    # 发送告警
    if [ -n "$WEBHOOK_URL" ]; then
        curl -X POST "$WEBHOOK_URL" \
            -H "Content-Type: application/json" \
            -d "{\"text\":\"备份告警: 备份文件损坏\"}" \
            2>/dev/null
    fi
    
    exit 1
fi

# 检查 SQL 内容
log "验证 SQL 内容..."
if gunzip -c "$LATEST_BACKUP" | head -n 20 | grep -q "CREATE TABLE"; then
    log "✓ SQL 内容验证通过"
else
    log "错误: SQL 内容异常"
    exit 1
fi

# 统计备份信息
BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/xingyunpan_*.sql.gz 2>/dev/null | wc -l)
TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)

log "备份统计:"
log "  备份文件数量: $BACKUP_COUNT"
log "  备份目录总大小: $TOTAL_SIZE"

log "=== 备份验证完成 ==="
log "✓ 所有检查通过"

exit 0
