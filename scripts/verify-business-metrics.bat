@echo off
REM 验证业务监控指标脚本 (Windows)

echo ==========================================
echo 验证星云盘 V2 业务监控指标
echo ==========================================
echo.

set METRICS_URL=http://localhost:8080/metrics
set PROMETHEUS_URL=http://localhost:9090

REM 检查应用是否运行
echo 1. 检查应用状态...
curl -s %METRICS_URL% >nul 2>&1
if errorlevel 1 (
    echo X 应用未运行或 /metrics 端点不可访问
    echo    请确保应用已启动
    exit /b 1
)
echo √ 应用正在运行
echo.

REM 检查业务指标是否暴露
echo 2. 检查业务指标...

curl -s %METRICS_URL% > %TEMP%\metrics.txt

findstr /C:"xingyunpan_total_users" %TEMP%\metrics.txt >nul
if errorlevel 1 (
    echo   X 总用户数指标未找到
) else (
    echo   √ 总用户数 ^(xingyunpan_total_users^)
    findstr /R "^xingyunpan_total_users [0-9]" %TEMP%\metrics.txt
)

findstr /C:"xingyunpan_total_files" %TEMP%\metrics.txt >nul
if errorlevel 1 (
    echo   X 总文件数指标未找到
) else (
    echo   √ 总文件数 ^(xingyunpan_total_files^)
    findstr /R "^xingyunpan_total_files [0-9]" %TEMP%\metrics.txt
)

findstr /C:"xingyunpan_total_storage_bytes" %TEMP%\metrics.txt >nul
if errorlevel 1 (
    echo   X 总存储使用量指标未找到
) else (
    echo   √ 总存储使用量 ^(xingyunpan_total_storage_bytes^)
    findstr /R "^xingyunpan_total_storage_bytes [0-9]" %TEMP%\metrics.txt
)

findstr /C:"xingyunpan_file_upload_success_total" %TEMP%\metrics.txt >nul
if errorlevel 1 (
    echo   X 上传成功总数指标未找到
) else (
    echo   √ 上传成功总数 ^(xingyunpan_file_upload_success_total^)
    findstr /R "^xingyunpan_file_upload_success_total [0-9]" %TEMP%\metrics.txt
)

findstr /C:"xingyunpan_file_upload_failure_total" %TEMP%\metrics.txt >nul
if errorlevel 1 (
    echo   X 上传失败总数指标未找到
) else (
    echo   √ 上传失败总数 ^(xingyunpan_file_upload_failure_total^)
    findstr /R "^xingyunpan_file_upload_failure_total [0-9]" %TEMP%\metrics.txt
)

findstr /C:"xingyunpan_api_requests_by_endpoint" %TEMP%\metrics.txt >nul
if errorlevel 1 (
    echo   X API 请求分布指标未找到
) else (
    echo   √ API 请求分布 ^(xingyunpan_api_requests_by_endpoint^)
)

echo.

REM 检查 Prometheus 是否采集到数据
echo 3. 检查 Prometheus 数据采集...
curl -s "%PROMETHEUS_URL%/api/v1/query?query=xingyunpan_total_users" > %TEMP%\prom-result.txt 2>nul
if errorlevel 1 (
    echo ⚠ Prometheus 未运行或不可访问
    echo    跳过 Prometheus 检查
) else (
    findstr /C:"success" %TEMP%\prom-result.txt >nul
    if errorlevel 1 (
        echo X Prometheus 未采集到数据
        type %TEMP%\prom-result.txt
    ) else (
        echo √ Prometheus 已采集到业务指标
    )
)

echo.

REM 生成测试报告
echo ==========================================
echo 验证报告
echo ==========================================
echo.
echo 指标端点: %METRICS_URL%
echo Prometheus: %PROMETHEUS_URL%
echo.
echo 业务指标状态：
findstr /R "^xingyunpan_total_users [0-9]" %TEMP%\metrics.txt
findstr /R "^xingyunpan_total_files [0-9]" %TEMP%\metrics.txt
findstr /R "^xingyunpan_total_storage_bytes [0-9]" %TEMP%\metrics.txt
echo.
echo 下一步：
echo 1. 访问 Grafana 查看业务监控仪表盘
echo    http://localhost:3000
echo 2. 如果仪表盘未创建，运行：
echo    scripts\setup-business-dashboard.bat
echo.

REM 清理临时文件
del %TEMP%\metrics.txt >nul 2>&1
del %TEMP%\prom-result.txt >nul 2>&1

pause
