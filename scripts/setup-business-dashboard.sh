#!/bin/bash

# 星云盘 V2 业务监控仪表盘快速设置脚本

set -e

echo "=========================================="
echo "星云盘 V2 业务监控仪表盘设置"
echo "=========================================="
echo ""

# 检查 Grafana 是否运行
echo "检查 Grafana 状态..."
if ! curl -s http://localhost:3000/api/health > /dev/null; then
    echo "❌ Grafana 未运行，请先启动监控系统："
    echo "   docker-compose -f docker-compose.monitoring.yml up -d"
    exit 1
fi
echo "✅ Grafana 正在运行"
echo ""

# 检查 Prometheus 数据源
echo "检查 Prometheus 数据源..."
GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASS="admin"

# 创建业务监控仪表盘 JSON
cat > /tmp/business-dashboard.json << 'EOF'
{
  "dashboard": {
    "title": "星云盘业务监控",
    "tags": ["xingyunpan", "business"],
    "timezone": "browser",
    "schemaVersion": 36,
    "version": 1,
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "关键业务指标",
        "type": "stat",
        "gridPos": {"h": 4, "w": 24, "x": 0, "y": 0},
        "targets": [
          {
            "expr": "xingyunpan_total_users",
            "legendFormat": "总用户数",
            "refId": "A"
          },
          {
            "expr": "xingyunpan_total_files",
            "legendFormat": "总文件数",
            "refId": "B"
          },
          {
            "expr": "xingyunpan_total_storage_bytes / 1024 / 1024 / 1024",
            "legendFormat": "总存储 (GB)",
            "refId": "C"
          },
          {
            "expr": "increase(xingyunpan_file_upload_success_total[24h])",
            "legendFormat": "今日上传数",
            "refId": "D"
          }
        ],
        "options": {
          "graphMode": "none",
          "textMode": "value_and_name"
        }
      },
      {
        "id": 2,
        "title": "用户增长趋势",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 4},
        "targets": [
          {
            "expr": "xingyunpan_total_users",
            "legendFormat": "用户数",
            "refId": "A"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "short",
            "min": 0
          }
        }
      },
      {
        "id": 3,
        "title": "文件上传速率",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 4},
        "targets": [
          {
            "expr": "rate(xingyunpan_file_upload_success_total[5m]) * 60",
            "legendFormat": "上传速率 (文件/分钟)",
            "refId": "A"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "short",
            "min": 0
          }
        }
      },
      {
        "id": 4,
        "title": "存储使用趋势",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 12},
        "targets": [
          {
            "expr": "xingyunpan_total_storage_bytes",
            "legendFormat": "存储使用量",
            "refId": "A"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "bytes",
            "min": 0
          }
        }
      },
      {
        "id": 5,
        "title": "上传成功率",
        "type": "stat",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 12},
        "targets": [
          {
            "expr": "(rate(xingyunpan_file_upload_success_total[5m]) / (rate(xingyunpan_file_upload_success_total[5m]) + rate(xingyunpan_file_upload_failure_total[5m]) + 0.0001)) * 100",
            "legendFormat": "成功率",
            "refId": "A"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100,
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {"value": 0, "color": "red"},
                {"value": 90, "color": "yellow"},
                {"value": 95, "color": "green"}
              ]
            }
          }
        },
        "options": {
          "graphMode": "area",
          "textMode": "value_and_name"
        }
      },
      {
        "id": 6,
        "title": "API 请求分布 (Top 10)",
        "type": "piechart",
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 20},
        "targets": [
          {
            "expr": "topk(10, sum by (endpoint) (rate(xingyunpan_api_requests_by_endpoint[5m])))",
            "legendFormat": "{{endpoint}}",
            "refId": "A"
          }
        ],
        "options": {
          "legend": {
            "displayMode": "table",
            "placement": "right",
            "values": ["value", "percent"]
          }
        }
      }
    ]
  },
  "overwrite": true
}
EOF

echo "导入业务监控仪表盘..."
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -u "$GRAFANA_USER:$GRAFANA_PASS" \
  -d @/tmp/business-dashboard.json \
  "$GRAFANA_URL/api/dashboards/db")

if echo "$RESPONSE" | grep -q "success"; then
    echo "✅ 业务监控仪表盘导入成功"
    DASHBOARD_URL=$(echo "$RESPONSE" | grep -o '"url":"[^"]*"' | cut -d'"' -f4)
    echo ""
    echo "仪表盘访问地址："
    echo "  $GRAFANA_URL$DASHBOARD_URL"
else
    echo "❌ 仪表盘导入失败"
    echo "响应: $RESPONSE"
    echo ""
    echo "请手动导入仪表盘，参考文档："
    echo "  docs/business-dashboard-setup.md"
fi

# 清理临时文件
rm -f /tmp/business-dashboard.json

echo ""
echo "=========================================="
echo "设置完成"
echo "=========================================="
echo ""
echo "下一步："
echo "1. 确保应用已启动指标收集器"
echo "2. 访问 Grafana 查看业务监控仪表盘"
echo "3. 参考文档了解更多配置选项："
echo "   docs/business-dashboard-setup.md"
echo ""
