@echo off
echo ========================================
echo 星云盘 V2 监控系统验证脚本
echo ========================================
echo.

echo [1/6] 检查 Docker 是否运行...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] Docker 未安装或未运行
    exit /b 1
)
echo [成功] Docker 已安装

echo.
echo [2/6] 检查监控容器状态...
docker ps --filter "name=prometheus" --format "{{.Names}}: {{.Status}}"
docker ps --filter "name=grafana" --format "{{.Names}}: {{.Status}}"
docker ps --filter "name=alertmanager" --format "{{.Names}}: {{.Status}}"
docker ps --filter "name=node-exporter" --format "{{.Names}}: {{.Status}}"

echo.
echo [3/6] 检查 Prometheus 端点...
curl -s http://localhost:9090/-/healthy >nul 2>&1
if %errorlevel% neq 0 (
    echo [警告] Prometheus 端点不可访问
) else (
    echo [成功] Prometheus 运行正常
)

echo.
echo [4/6] 检查 Grafana 端点...
curl -s http://localhost:3000/api/health >nul 2>&1
if %errorlevel% neq 0 (
    echo [警告] Grafana 端点不可访问
) else (
    echo [成功] Grafana 运行正常
)

echo.
echo [5/6] 检查应用 Metrics 端点...
curl -s http://localhost:8080/metrics >nul 2>&1
if %errorlevel% neq 0 (
    echo [警告] 应用 Metrics 端点不可访问
    echo [提示] 请确保星云盘应用正在运行
) else (
    echo [成功] 应用 Metrics 端点可访问
)

echo.
echo [6/6] 检查 Prometheus Targets...
echo [提示] 请访问 http://localhost:9090/targets 查看采集目标状态

echo.
echo ========================================
echo 验证完成！
echo ========================================
echo.
echo 访问以下 URL 查看监控系统：
echo - Prometheus: http://localhost:9090
echo - Grafana: http://localhost:3000 (admin/admin)
echo - Alertmanager: http://localhost:9093
echo - 应用 Metrics: http://localhost:8080/metrics
echo.
echo 详细配置步骤请参考：
echo - docs/monitoring-setup.md
echo - docs/grafana-setup.md
echo.

pause
