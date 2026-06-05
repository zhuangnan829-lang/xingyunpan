#!/bin/bash

# 星云盘 V2 日志轮转脚本
# Log Rotation Script for Xingyunpan V2

set -e

# 配置
LOG_DIR="/data/xingyunpan/logs"
APP_LOG="$LOG_DIR/app.log"
MAX_SIZE_MB=100
MAX_AGE_DAYS=7
COMPRESS=true

echo "=== 星云盘 V2 日志轮转 ==="
echo "日志目录: $LOG_DIR"
echo "应用日志: $APP_LOG"
echo ""

# 检查日志目录是否存在
if [ ! -d "$LOG_DIR" ]; then
    echo "错误: 日志目录不存在: $LOG_DIR"
    exit 1
fi

# 检查日志文件是否存在
if [ ! -f "$APP_LOG" ]; then
    echo "警告: 日志文件不存在: $APP_LOG"
    echo "日志文件将在应用启动时自动创建"
    exit 0
fi

# 获取当前日志文件大小（MB）
CURRENT_SIZE=$(du -m "$APP_LOG" | cut -f1)
echo "当前日志文件大小: ${CURRENT_SIZE}MB"

# 如果日志文件超过最大大小，执行轮转
if [ $CURRENT_SIZE -ge $MAX_SIZE_MB ]; then
    echo "日志文件超过 ${MAX_SIZE_MB}MB，执行轮转..."
    
    # 生成时间戳
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    BACKUP_FILE="$LOG_DIR/app-$TIMESTAMP.log"
    
    # 复制当前日志文件
    cp "$APP_LOG" "$BACKUP_FILE"
    echo "已备份日志文件: $BACKUP_FILE"
    
    # 清空当前日志文件
    > "$APP_LOG"
    echo "已清空当前日志文件"
    
    # 压缩备份文件
    if [ "$COMPRESS" = true ]; then
        gzip "$BACKUP_FILE"
        echo "已压缩备份文件: ${BACKUP_FILE}.gz"
    fi
    
    # 发送 USR1 信号给应用，触发日志重新打开
    # 注意：需要应用支持 USR1 信号处理
    PID=$(pgrep -f "xingyunpan.*server" || echo "")
    if [ -n "$PID" ]; then
        kill -USR1 $PID
        echo "已发送 USR1 信号给进程 $PID"
    fi
else
    echo "日志文件大小正常，无需轮转"
fi

echo ""
echo "=== 清理旧日志文件 ==="

# 删除超过指定天数的日志文件
DELETED_COUNT=0
for file in "$LOG_DIR"/app-*.log.gz "$LOG_DIR"/app-*.log; do
    if [ -f "$file" ]; then
        # 获取文件修改时间（天数）
        FILE_AGE=$(find "$file" -mtime +$MAX_AGE_DAYS -print)
        if [ -n "$FILE_AGE" ]; then
            rm -f "$file"
            echo "已删除旧日志文件: $file"
            DELETED_COUNT=$((DELETED_COUNT + 1))
        fi
    fi
done

if [ $DELETED_COUNT -eq 0 ]; then
    echo "没有需要删除的旧日志文件"
else
    echo "共删除 $DELETED_COUNT 个旧日志文件"
fi

echo ""
echo "=== 日志统计 ==="
echo "当前日志文件数量: $(ls -1 "$LOG_DIR"/app*.log* 2>/dev/null | wc -l)"
echo "日志目录总大小: $(du -sh "$LOG_DIR" | cut -f1)"

echo ""
echo "=== 日志轮转完成 ==="
