#!/bin/bash

# 灾备方案验证脚本
# 验证灾备文档和备份脚本是否就绪

echo "=========================================="
echo "  星云盘 V2 灾备方案验证"
echo "=========================================="
echo ""

PASS=0
FAIL=0

# 检查灾备文档
echo "1. 检查灾备文档..."
if [ -f "docs/disaster-recovery.md" ]; then
    echo "   ✓ 灾备文档存在"
    ((PASS++))
else
    echo "   ✗ 灾备文档不存在"
    ((FAIL++))
fi

# 检查备份脚本
echo "2. 检查备份脚本..."
if [ -f "scripts/backup.sh" ]; then
    echo "   ✓ 备份脚本存在"
    ((PASS++))
else
    echo "   ✗ 备份脚本不存在"
    ((FAIL++))
fi

# 检查恢复脚本
echo "3. 检查恢复脚本..."
if [ -f "scripts/restore.sh" ]; then
    echo "   ✓ 恢复脚本存在"
    ((PASS++))
else
    echo "   ✗ 恢复脚本不存在"
    ((FAIL++))
fi

# 检查备份目录
echo "4. 检查备份目录配置..."
if grep -q "BACKUP_DIR" scripts/backup.sh 2>/dev/null; then
    echo "   ✓ 备份目录已配置"
    ((PASS++))
else
    echo "   ✗ 备份目录未配置"
    ((FAIL++))
fi

# 检查 RTO/RPO 定义
echo "5. 检查 RTO/RPO 定义..."
if grep -q "RTO" docs/disaster-recovery.md && grep -q "RPO" docs/disaster-recovery.md; then
    echo "   ✓ RTO/RPO 已定义"
    ((PASS++))
else
    echo "   ✗ RTO/RPO 未定义"
    ((FAIL++))
fi

# 检查恢复流程文档
echo "6. 检查恢复流程..."
if grep -q "恢复步骤" docs/disaster-recovery.md || grep -q "Recovery Procedures" docs/disaster-recovery.md; then
    echo "   ✓ 恢复流程已文档化"
    ((PASS++))
else
    echo "   ✗ 恢复流程未文档化"
    ((FAIL++))
fi

# 检查灾难场景
echo "7. 检查灾难场景覆盖..."
SCENARIOS=0
grep -q "数据库损坏" docs/disaster-recovery.md && ((SCENARIOS++))
grep -q "服务器故障" docs/disaster-recovery.md && ((SCENARIOS++))
grep -q "文件丢失" docs/disaster-recovery.md && ((SCENARIOS++))

if [ $SCENARIOS -ge 2 ]; then
    echo "   ✓ 已覆盖 $SCENARIOS 个灾难场景"
    ((PASS++))
else
    echo "   ✗ 灾难场景