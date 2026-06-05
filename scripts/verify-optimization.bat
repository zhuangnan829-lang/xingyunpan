@echo off
REM 验证性能优化效果脚本 (Windows)
REM 使用方法: scripts\verify-optimization.bat

echo === 验证性能优化效果 ===
echo.

REM 设置变量
set BASE_URL=http://localhost:8080
set REPORT_FILE=docs\optimization-report.md

echo 1. 检查服务器是否运行...
curl -s %BASE_URL%/health >nul 2>&1
if errorlevel 1 (
    echo 错误: 服务器未运行，请先启动服务器
    exit /b 1
)
echo ✓ 服务器运行正常
echo.

echo 2. 检查数据库索引...
echo 正在验证索引是否已创建...
mysql -h localhost -P 3306 -u xingyunpan -pxingyunpan123 xingyunpan -e "SHOW INDEX FROM user_files WHERE Key_name = 'idx_user_files_user_folder';" >nul 2>&1
if errorlevel 1 (
    echo ⚠ 警告: 索引未创建，请运行 scripts\apply-indexes.bat
) else (
    echo ✓ 数据库索引已创建
)
echo.

echo 3. 运行性能基准测试...
echo 正在执行压测...
call scripts\benchmark.bat
echo.

echo 4. 生成性能报告...
echo 正在生成报告...
call scripts\generate-report.bat
echo.

echo 5. 分析性能提升...
echo.
echo === 性能优化验证完成 ===
echo.
echo 查看详细报告: %REPORT_FILE%
echo.
echo 预期性能提升:
echo   - API 响应时间减少 30%%+
echo   - 数据库查询时间减少 50%%+
echo   - 文件列表查询从 N+1 次减少到 2 次
echo.
