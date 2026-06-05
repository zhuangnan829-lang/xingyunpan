#!/bin/bash

# 星云盘 V2 磁盘空间检查脚本
# Disk Space Check Script for Xingyunpan V2

set -e

# 配置
THRESHOLD=80  # 告警阈值（百分比）
CHECK_PATHS=(
    "/data/xingyunpan"
    "/backup/xingyunpan"
    "/var/log"
)

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log "=== 磁盘空间检查 ==="

# 检查每个路径
for path in "${CHECK_PATHS[@]}"; do
    if [ ! -d "$path" ]; then
        log "警告: 路径不存在: $path"
        continue
    fi
    
    # 获取磁盘使用率
    USAGE=$(df -h "$path" | awk 'NR==2 {print $5}' | sed 's/%//')
    AVAILABLE=$(df -h "$path" | awk 'NR==2 {print $4}')
    
    log "路径: $path"
    log "  使用率: ${USAGE}%"
    log "  可用空间: $AVAILABLE"
    
    # 检查是否超过阈值
    if [ "$USAGE" -gt "$THRESHOLD" ]; then
        log "  ⚠️  警告: 磁盘使用率超过 ${THRESHOLD}%"
        
        # 可选：发送告警通知
        if [ -n "$WEBHOOK_URL" ]; then
            curl -X POST "$WEBHOOK_URL" \
                -H "Content-Type: application/json" \
                -d "{\"text\":\"磁盘空间告警: $path 使用率 ${USAGE}%, 可用空间 $AVAILABLE\"}" \
                2>/dev/null || log "发送告警失败"
        fi
    else
        log "  ✓ 正常"
    fi
    log ""
done

log "=== 检查完成 ==="
