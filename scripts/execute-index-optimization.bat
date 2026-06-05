@echo off
REM 执行数据库索引优化脚本（Windows 版本）
REM 使用方法: scripts\execute-index-optimization.bat

echo === 数据库索引优化 ===
echo.

REM 设置数据库连接信息（从环境变量读取，或使用默认值）
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

REM 步骤 1: 测量优化前的查询性能
echo [步骤 1/4] 测量优化前的查询性能...
echo.
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% < scripts\measure-query-performance.sql > before-optimization.txt 2>&1
if errorlevel 1 (
    echo 错误: 无法连接到数据库或执行查询
    echo 请检查数据库连接信息和 MySQL 客户端是否已安装
    exit /b 1
)
echo 优化前的查询计划已保存到 before-optimization.txt
echo.

REM 步骤 2: 执行索引创建
echo [步骤 2/4] 执行索引创建...
echo.
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% < scripts\add-indexes.sql
if errorlevel 1 (
    echo 错误: 索引创建失败
    exit /b 1
)
echo 索引创建成功
echo.

REM 步骤 3: 验证索引创建
echo [步骤 3/4] 验证索引创建...
echo.
echo 检查 user_files 表索引:
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% -e "SHOW INDEX FROM user_files;"
echo.
echo 检查 physical_files 表索引:
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% -e "SHOW INDEX FROM physical_files;"
echo.
echo 检查 multipart_uploads 表索引:
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% -e "SHOW INDEX FROM multipart_uploads;"
echo.

REM 步骤 4: 测量优化后的查询性能
echo [步骤 4/4] 测量优化后的查询性能...
echo.
mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASSWORD% %DB_NAME% < scripts\measure-query-performance.sql > after-optimization.txt 2>&1
echo 优化后的查询计划已保存到 after-optimization.txt
echo.

echo === 索引优化完成 ===
echo.
echo 对比文件:
echo   - before-optimization.txt  (优化前)
echo   - after-optimization.txt   (优化后)
echo.
echo 请使用以下命令对比性能提升:
echo   fc before-optimization.txt after-optimization.txt
echo.

