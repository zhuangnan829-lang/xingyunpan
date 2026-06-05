@echo off
REM 星云盘 V2 - Windows MySQL 启动脚本
REM Start MySQL Service on Windows

echo ========================================
echo 星云盘 V2 - MySQL 服务启动
echo XingYunPan V2 - Start MySQL Service
echo ========================================
echo.

echo [1/3] 检查 MySQL 服务状态...
echo Checking MySQL service status...
echo.

REM 尝试查找 MySQL 服务
for /f "tokens=1" %%i in ('sc query ^| findstr /i "mysql"') do (
    set MYSQL_SERVICE=%%i
    goto :found
)

echo ❌ 错误: 未找到 MySQL 服务
echo Error: MySQL service not found
echo.
echo 请确保已安装 MySQL 数据库
echo Please make sure MySQL is installed
echo.
echo 常见的 MySQL 服务名称:
echo Common MySQL service names:
echo - MySQL96
echo - MySQL80
echo - MySQL
echo - MySQL57
echo.
pause
exit /b 1

:found
echo ✓ 找到 MySQL 服务: %MYSQL_SERVICE%
echo Found MySQL service: %MYSQL_SERVICE%
echo.

echo [2/3] 启动 MySQL 服务...
echo Starting MySQL service...
net start %MYSQL_SERVICE%

if %errorlevel% neq 0 (
    echo.
    echo ❌ 启动失败，尝试以管理员身份运行此脚本
    echo Failed to start. Try running this script as Administrator
    echo.
    echo 右键点击此文件 -^> 以管理员身份运行
    echo Right-click this file -^> Run as Administrator
    echo.
    pause
    exit /b 1
)

echo.
echo [3/3] 验证 MySQL 连接...
echo Verifying MySQL connection...
echo.

REM 测试 MySQL 连接
mysql -u root -e "SELECT 1" 2>nul
if %errorlevel% equ 0 (
    echo ✓ MySQL 服务运行正常
    echo MySQL service is running successfully
) else (
    echo ⚠ MySQL 已启动，但无法连接
    echo MySQL started but connection test failed
    echo.
    echo 可能的原因:
    echo Possible reasons:
    echo 1. MySQL 客户端未安装或不在 PATH 中
    echo 2. root 用户需要密码
    echo 3. MySQL 正在初始化中，请稍等片刻
)

echo.
echo ========================================
echo MySQL 服务已启动
echo MySQL service started
echo ========================================
echo.
echo 下一步:
echo Next steps:
echo 1. 运行数据库迁移: go run scripts/migrate.go
echo 2. 启动服务器: go run cmd/server/main.go
echo 3. 运行测试: scripts\quick-start-test.bat
echo.

pause
