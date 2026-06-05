@echo off
REM 应用数据库索引优化脚本 (Windows)
REM 使用方法: scripts\apply-indexes.bat

echo === 应用数据库索引优化 ===
echo.

REM 从环境变量读取数据库连接信息
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=3306
if "%DB_USER%"=="" set DB_USER=xingyunpan
if "%DB_PASSWORD%"=="" set DB_PASSWORD=xingyunpan123
if "%DB_NAME%"=="" set DB_NAME=xingyunpan

echo 数据库连接信息:
echo   Host: %DB_HOST%
echo   Port: %DB_PORT%
echo   User: %DB_USER%
echo   Database: %DB_NAME%
echo.

REM 执行索引创建脚本
echo 执行索引创建脚本...
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% < scripts\add-indexes.sql

if errorlevel 1 (
    echo.
    echo === 索引优化失败 ===
    exit /b 1
)

echo.
echo === 索引优化完成 ===
echo.
echo 可以使用以下命令验证索引:
echo   mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% -e "SHOW INDEX FROM user_files;"
