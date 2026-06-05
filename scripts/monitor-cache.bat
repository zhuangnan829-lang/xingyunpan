@echo off
REM 路径: scripts/monitor-cache.bat
REM Redis 缓存监控脚本

echo ========================================
echo Redis 缓存监控
echo ========================================
echo.

REM 检查 Redis 连接
echo 检查 Redis 连接...
redis-cli ping >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo 错误: Redis 连接失败
    echo 请确保 Redis 正在运行
    exit /b 1
)

echo Redis 连接正常
echo.

REM 运行监控脚本
go run scripts/monitor-cache.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo 错误: 监控脚本执行失败
    exit /b 1
)

echo.
echo ========================================
echo 监控完成
echo ========================================

exit /b 0
