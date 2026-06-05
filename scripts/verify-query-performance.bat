@echo off
REM 查询性能验证脚本 - Windows 版本
REM 用于验证索引优化后的查询性能提升

echo ========================================
echo 查询性能验证
echo ========================================
echo.

REM 检查数据库连接信息
if "%DB_PASSWORD%"=="" (
    echo 错误: 请设置 DB_PASSWORD 环境变量
    echo.
    echo 使用方法:
    echo   set DB_HOST=localhost
    echo   set DB_PORT=3306
    echo   set DB_USER=xingyunpan
    echo   set DB_PASSWORD=your_password
    echo   set DB_NAME=xingyunpan
    echo   scripts\verify-query-performance.bat
    echo.
    exit /b 1
)

REM 设置默认值
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=3306
if "%DB_USER%"=="" set DB_USER=xingyunpan
if "%DB_NAME%"=="" set DB_NAME=xingyunpan

echo 数据库连接信息:
echo   Host: %DB_HOST%
echo   Port: %DB_PORT%
echo   User: %DB_USER%
echo   Database: %DB_NAME%
echo.

REM 运行 Go 程序
echo 正在执行查询性能测试...
echo.

go run scripts\verify-query-performance.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo 错误: 查询性能测试失败
    exit /b 1
)

echo.
echo ========================================
echo 验证完成
echo ========================================
echo.
echo 报告已生成，请查看上方输出
echo.
echo 下一步:
echo   1. 如果所有查询都使用了索引，说明优化成功
echo   2. 如果查询时间都在 10ms 以内，性能优秀
echo   3. 如果有查询未使用索引，请检查索引是否创建成功
echo   4. 运行 ANALYZE TABLE 更新统计信息
echo.
