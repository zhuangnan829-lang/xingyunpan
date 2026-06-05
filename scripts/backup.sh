#!/bin/bash

# 星云盘 V2 数据库备份脚本
# Database Backup Script for Xingyunpan V2

set -e

# 配置
BACKUP_DIR="${BACKUP_DIR:-/backup/xingyunpan}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-xingyunpan}"
DB_PASSWORD="${DB_PASSWORD}"
DB_NAME="${DB_NAME:-xingyunpan}"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/xingyunpan_$DATE.sql"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log "=== 星云盘 V2 数据库备份开始 ==="
log "备份目录: $BACKUP_DIR"
log "数据库: $DB_NAME@$DB_HOST:$DB_PORT"
log "备份文件: $BACKUP_FILE"

# 检查必需的环境变量
if [ -z "$DB_PASSWORD" ]; then
    log "错误: DB_PASSWORD 环境变量未设置"
    exit 1
fi

# 创建备份目录
if [ ! -d "$BACKUP_DIR" ]; then
    log "创建备份目录: $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
fi

# 检查 mysqldump 是否存在
if ! command -v mysqldump &> /dev/null; then
    log "错误: mysqldump 命令不存在，请安装 MySQL 客户端"
    exit 1
fi

# 执行数据库备份
log "开始备份数据库..."
mysqldump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --password="$DB_PASSWORD" \
    --single-transaction \
    --quick \
    --lock-tables=false \
    --routines \
    --triggers \
    --events \
    "$DB_NAME" > "$BACKUP_FILE" 2>&1

# 检查备份是否成功
if [ $? -ne 0 ]; then
    log "错误: 数据库备份失败"
    rm -f "$BACKUP_FILE"
    exit 1
fi

# 验证备份文件不为空
if [ ! -s "$BACKUP_FILE" ]; then
    log "错误: 备份文件为空"
    rm -f "$BACKUP_FILE"
    exit 1
fi

BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log "备份完成，文件大小: $BACKUP_SIZE"

# 压缩备份文件
log "压缩备份文件..."
gzip "$BACKUP_FILE"
COMPRESSED_FILE="${BACKUP_FILE}.gz"
COMPRESSED_SIZE=$(du -h "$COMPRESSED_FILE" | cut -f1)
log "压缩完成，压缩后大小: $COMPRESSED_SIZE"

# 验证压缩文件完整性
log "验证压缩文件完整性..."
if gzip -t "$COMPRESSED_FILE" 2>/dev/null; then
    log "压缩文件完整性验证通过"
else
    log "错误: 压缩文件损坏"
    exit 1
fi

# 清理旧备份
log "清理超过 $RETENTION_DAYS 天的旧备份..."
DELETED_COUNT=0
for file in "$BACKUP_DIR"/xingyunpan_*.sql.gz; do
    if [ -f "$file" ]; then
        # 检查文件是否超过保留天数
        if [ $(find "$file" -mtime +$RETENTION_DAYS -print | wc -l) -gt 0 ]; then
            log "删除旧备份: $(basename $file)"
            rm -f "$file"
            DELETED_COUNT=$((DELETED_COUNT + 1))
        fi
    fi
done

if [ $DELETED_COUNT -eq 0 ]; then
    log "没有需要删除的旧备份"
else
    log "共删除 $DELETED_COUNT 个旧备份文件"
fi

# 备份统计
log "=== 备份统计 ==="
BACKUP_COUNT=$(ls -1 "$BACKUP_DIR"/xingyunpan_*.sql.gz 2>/dev/null | wc -l)
TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
log "当前备份文件数量: $BACKUP_COUNT"
log "备份目录总大小: $TOTAL_SIZE"

# 列出最近的备份
log "最近的 5 个备份:"
ls -lht "$BACKUP_DIR"/xingyunpan_*.sql.gz 2>/dev/null | head -n 5 | while read line; do
    log "  $line"
done

log "=== 数据库备份完成 ==="

# 可选：发送备份通知
# 如果设置了 WEBHOOK_URL，发送备份成功通知
if [ -n "$WEBHOOK_URL" ]; then
    log "发送备份通知..."
    curl -X POST "$WEBHOOK_URL" \
        -H "Content-Type: application/json" \
        -d "{\"text\":\"数据库备份成功: $COMPRESSED_FILE, 大小: $COMPRESSED_SIZE\"}" \
        2>/dev/null || log "警告: 发送通知失败"
fi

exit 0
