#!/bin/bash

# 验证业务监控指标脚本

set -e

echo "=========================================="
echo "验证星云盘 V2 业务监控指标"
echo "=========================================="
echo ""

METRICS_URL="http://localhost:8080/metrics"
PROMETHEUS_URL="http://localhost:9090"

# 检查应用是否运行
echo "1. 检查应用状态..."
if ! curl -s "$METRICS_URL" > /dev/null; then
    echo "❌ 应用未运行或 /metrics 端点不可访问"
    echo "   请确保应用已启动"
    exit 1
fi
echo "✅ 应用正在运行"
echo ""

# 检查业务指标是否暴露
echo "2. 检查业务指标..."

METRICS=$(curl -s "$METRICS_URL")

check_metric() {
    local metric_name=$1
    local description=$2
    
    if echo "$METRICS" | grep -q "^$metric_name"; then
        echo "  ✅ $description ($metric_name)"
        # 显示当前值
        local value=$(echo "$METRICS" | grep "^$metric_name" | grep -v "^#" | head -1 | awk '{print $2}')
        echo "     当前值: $value"
    else
        echo "  ❌ $description ($metric_name) - 未找到"
        return 1
    fi
}

check_metric "xingyunpan_total_users" "总用户数"
check_metric "xingyunpan_total_files" "总文件数"
check_metric "xingyunpan_total_storage_bytes" "总存储使用量"
check_metric "xingyunpan_file_upload_success_total" "上传成功总数"
check_metric "xingyunpan_file_upload_failure_total" "上传失败总数"
check_metric "xingyunpan_api_requests_by_endpoint" "API 请求分布"

echo ""

# 检查 Prometheus 是否采集到数据
echo "3. 检查 Prometheus 数据采集..."
if ! curl -s "$PROMETHEUS_URL/api/v1/query?query=xingyunpan_total_users" > /dev/null; then
    echo "⚠️  Prometheus 未运行或不可访问"
    echo "   跳过 Prometheus 检查"
else
    PROM_RESULT=$(curl -s "$PROMETHEUS_URL/api/v1/query?query=xingyunpan_total_users")
    if echo "$PROM_RESULT" | grep -q '"status":"success"'; then
        echo "✅ Prometheus 已采集到业务指标"
        # 提取并显示值
        VALUE=$(echo "$PROM_RESULT" | grep -o '"value":\[[^]]*\]' | grep -o '\[[^]]*\]' | grep -o '[0-9.]*' | tail -1)
        echo "   总用户数: $VALUE"
    else
        echo "❌ Prometheus 未采集到数据"
        echo "   响应: $PROM_RESULT"
    fi
fi

echo ""

# 检查指标收集器日志
echo "4. 检查指标收集器日志..."
if [ -f "/data/xingyunpan/logs/app.log" ]; then
    if grep -q "Metrics collected" /data/xingyunpan/logs/app.log; then
        echo "✅ 指标收集器正在运行"
        echo "   最近一次收集："
        grep "Metrics collected" /data/xingyunpan/logs/app.log | tail -1
    else
        echo "⚠️  未找到指标收集日志"
        echo "   请确认指标收集器已启动"
    fi
else
    echo "⚠️  日志文件不存在，跳过日志检查"
fi

echo ""

# 生成测试报告
echo "=========================================="
echo "验证报告"
echo "=========================================="
echo ""
echo "指标端点: $METRICS_URL"
echo "Prometheus: $PROMETHEUS_URL"
echo ""
echo "业务指标状态："
echo "  - 总用户数: $(echo "$METRICS" | grep "^xingyunpan_total_users" | grep -v "^#" | awk '{print $2}')"
echo "  - 总文件数: $(echo "$METRICS" | grep "^xingyunpan_total_files" | grep -v "^#" | awk '{print $2}')"
echo "  - 总存储: $(echo "$METRICS" | grep "^xingyunpan_total_storage_bytes" | grep -v "^#" | awk '{print $2}') bytes"
echo ""
echo "下一步："
echo "1. 访问 Grafana 查看业务监控仪表盘"
echo "   http://localhost:3000"
echo "2. 如果仪表盘未创建，运行："
echo "   ./scripts/setup-business-dashboard.sh"
echo ""
