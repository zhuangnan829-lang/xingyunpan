#!/bin/bash

# 星云盘 V2 数据库恢复脚本
# Database Restore Script for Xingyunpan V2

set -e

# 配置
BACKUP_DIR="${BACKUP_DIR:-/backup/xingyunpan}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-xingyunpan}"
DB_PASSWORD="${DB_PASSWORD}"
DB_NAME="${DB_NAME:-xingyunpan}"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# 显示用法
usage() {
    echo "用法: $0 <backup_file>"
    echo ""
    echo "示例:"
    echo "  $0 /backup/xingyunpan/xingyunpan_20240308_120000.sql.gz"
    echo "  $0 xingyunpan_20240308_120000.sql.gz  # 从默认备份目录"
    echo ""
    echo "列出可用备份:"
    echo "  $0 --list"
    exit 1
}

# 列出可用备份
list_backups() {
    log "=== 可用备份列表 ==="
    if [ ! -d "$BACKUP_DIR" ]; then
        log "备份目录不存在: $BACKUP_DIR"
        exit 1
    fi
    
    ls -lht "$BACKUP_DIR"/xingyunpan_*.sql.gz 2>/dev/null | while read line; do
        echo "$line"
    done
    exit 0
}

# 检查参数
if [ $# -eq 0 ]; then
    usage
fi

if [ "$1" = "--list" ] || [ "$1" = "-l" ]; then
    list_backups
fi

BACKUP_FILE="$1"

# 如果只提供文件名，从备份目录查找
if [ ! -f "$BACKUP_FILE" ]; then
    BACKUP_FILE="$BACKUP_DIR/$1"
fi

# 检查备份文件是否存在
if [ ! -f "$BACKUP_FILE" ]; then
    log "错误: 备份文件不存在: $BACKUP_FILE"
    exit 1
fi

log "=== 星云盘 V2 数据库恢复开始 ==="
log "备份文件: $BACKUP_FILE"
log "数据库: $DB_NAME@$DB_HOST:$DB_PORT"

# 检查必需的环境变量
if [ -z "$DB_PASSWORD" ]; then
    log "错误: DB_PASSWORD 环境变量未设置"
    exit 1
fi

# 确认恢复操作
echo ""
echo "警告: 此操作将覆盖当前数据库 $DB_NAME 的所有数据！"
echo "请确认是否继续？(yes/no)"
read -r CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    log "恢复操作已取消"
    exit 0
fi

# 检查 mysql 命令是否存在
if ! command -v mysql &> /dev/null; then
    log "错误: mysql 命令不存在，请安装 MySQL 客户端"
    exit 1
fi

# 创建临时目录
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# 解压备份文件（如果是压缩文件）
if [[ "$BACKUP_FILE" == *.gz ]]; then
    log "解压备份文件..."
    SQL_FILE="$TEMP_DIR/restore.sql"
    gunzip -c "$BACKUP_FILE" > "$SQL_FILE"
    
    # 验证解压后的文件
    if [ ! -s "$SQL_FILE" ]; then
        log "错误: 解压后的文件为空"
        exit 1
    fi
else
    SQL_FILE="$BACKUP_FILE"
fi

# 备份当前数据库（可选）
log "备份当前数据库..."
CURRENT_BACKUP="$BACKUP_DIR/xingyunpan_before_restore_$(date +%Y%m%d_%H%M%S).sql.gz"
mysqldump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --password="$DB_PASSWORD" \
    --single-transaction \
    --quick \
    "$DB_NAME" | gzip > "$CURRENT_BACKUP" 2>/dev/null || log "警告: 当前数据库备份失败"

if [ -f "$CURRENT_BACKUP" ]; then
    log "当前数据库已备份到: $CURRENT_BACKUP"
fi

# 恢复数据库
log "开始恢复数据库..."
mysql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --password="$DB_PASSWORD" \
    "$DB_NAME" < "$SQL_FILE"

if [ $? -ne 0 ]; then
    log "错误: 数据库恢复失败"
    log "可以从以下备份恢复: $CURRENT_BACKUP"
    exit 1
fi

log "数据库恢复成功"

# 验证恢复结果
log "验证恢复结果..."
TABLE_COUNT=$(mysql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --password="$DB_PASSWORD" \
    --skip-column-names \
    -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$DB_NAME'" 2>/dev/null)

log "数据库表数量: $TABLE_COUNT"

# 显示主要表的记录数
log "主要表记录数:"
for table in users user_files physical_files multipart_uploads; do
    COUNT=$(mysql \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        --password="$DB_PASSWORD" \
        --skip-column-names \
        -e "SELECT COUNT(*) FROM $DB_NAME.$table" 2>/dev/null || echo "0")
    log "  $table: $COUNT"
done

log "=== 数据库恢复完成 ==="
log "请重启应用以使更改生效"

exit 0
