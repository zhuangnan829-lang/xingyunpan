@echo off
REM 星云盘 V2 业务监控仪表盘快速设置脚本 (Windows)

echo ==========================================
echo 星云盘 V2 业务监控仪表盘设置
echo ==========================================
echo.

REM 检查 Grafana 是否运行
echo 检查 Grafana 状态...
curl -s http://localhost:3000/api/health >nul 2>&1
if errorlevel 1 (
    echo X Grafana 未运行，请先启动监控系统：
    echo    docker-compose -f docker-compose.monitoring.yml up -d
    exit /b 1
)
echo √ Grafana 正在运行
echo.

echo 创建业务监控仪表盘配置...

REM 创建临时 JSON 文件
echo { > %TEMP%\business-dashboard.json
echo   "dashboard": { >> %TEMP%\business-dashboard.json
echo     "title": "星云盘业务监控", >> %TEMP%\business-dashboard.json
echo     "tags": ["xingyunpan", "business"], >> %TEMP%\business-dashboard.json
echo     "timezone": "browser", >> %TEMP%\business-dashboard.json
echo     "schemaVersion": 36, >> %TEMP%\business-dashboard.json
echo     "version": 1, >> %TEMP%\business-dashboard.json
echo     "refresh": "30s", >> %TEMP%\business-dashboard.json
echo     "panels": [ >> %TEMP%\business-dashboard.json
echo       { >> %TEMP%\business-dashboard.json
echo         "id": 1, >> %TEMP%\business-dashboard.json
echo         "title": "关键业务指标", >> %TEMP%\business-dashboard.json
echo         "type": "stat", >> %TEMP%\business-dashboard.json
echo         "gridPos": {"h": 4, "w": 24, "x": 0, "y": 0}, >> %TEMP%\business-dashboard.json
echo         "targets": [ >> %TEMP%\business-dashboard.json
echo           {"expr": "xingyunpan_total_users", "legendFormat": "总用户数", "refId": "A"}, >> %TEMP%\business-dashboard.json
echo           {"expr": "xingyunpan_total_files", "legendFormat": "总文件数", "refId": "B"}, >> %TEMP%\business-dashboard.json
echo           {"expr": "xingyunpan_total_storage_bytes / 1024 / 1024 / 1024", "legendFormat": "总存储 (GB)", "refId": "C"}, >> %TEMP%\business-dashboard.json
echo           {"expr": "increase(xingyunpan_file_upload_success_total[24h])", "legendFormat": "今日上传数", "refId": "D"} >> %TEMP%\business-dashboard.json
echo         ] >> %TEMP%\business-dashboard.json
echo       } >> %TEMP%\business-dashboard.json
echo     ] >> %TEMP%\business-dashboard.json
echo   }, >> %TEMP%\business-dashboard.json
echo   "overwrite": true >> %TEMP%\business-dashboard.json
echo } >> %TEMP%\business-dashboard.json

echo 导入业务监控仪表盘...
curl -s -X POST ^
  -H "Content-Type: application/json" ^
  -u "admin:admin" ^
  -d @%TEMP%\business-dashboard.json ^
  http://localhost:3000/api/dashboards/db > %TEMP%\grafana-response.txt

findstr /C:"success" %TEMP%\grafana-response.txt >nul
if errorlevel 1 (
    echo X 仪表盘导入失败
    echo 响应内容：
    type %TEMP%\grafana-response.txt
    echo.
    echo 请手动导入仪表盘，参考文档：
    echo   docs/business-dashboard-setup.md
) else (
    echo √ 业务监控仪表盘导入成功
    echo.
    echo 仪表盘访问地址：
    echo   http://localhost:3000/d/business-monitoring
)

REM 清理临时文件
del %TEMP%\business-dashboard.json >nul 2>&1
del %TEMP%\grafana-response.txt >nul 2>&1

echo.
echo ==========================================
echo 设置完成
echo ==========================================
echo.
echo 下一步：
echo 1. 确保应用已启动指标收集器
echo 2. 访问 Grafana 查看业务监控仪表盘
echo 3. 参考文档了解更多配置选项：
echo    docs/business-dashboard-setup.md
echo.

pause
