@echo off
REM 路径: scripts/run-query-analysis.bat
REM 查询性能分析脚本

echo ========================================
echo Phase 5 查询性能分析
echo ========================================
echo.

REM 检查 Go 环境
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未找到 Go 环境
    exit /b 1
)

REM 检查 MySQL 连接
echo 检查 MySQL 连接...
go run scripts/test-db-connection.go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo 错误: MySQL 连接失败
    echo 请确保 MySQL 正在运行并且配置正确
    exit /b 1
)

echo.
echo ========================================
echo 运行 EXPLAIN 分析
echo ========================================
echo.

REM 运行查询分析
go run scripts/analyze-queries.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo 错误: 查询分析失败
    exit /b 1
)

echo.
echo ========================================
echo 分析完成
echo ========================================
echo.

echo 优化建议:
echo   1. 确保所有索引都已创建（运行 scripts/migrations/phase5_schema.sql）
echo   2. 检查 EXPLAIN 输出中的 "key" 列，确认使用了正确的索引
echo   3. 如果 "type" 列显示 "ALL"，说明进行了全表扫描，需要优化
echo   4. 如果 "rows" 列数值很大，考虑添加更多过滤条件或优化索引
echo.

exit /b 0
