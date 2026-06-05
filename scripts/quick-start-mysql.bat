@echo off
chcp 65001 >nul
title 星云盘 - MySQL 快速启动
color 0A

:menu
cls
echo.
echo ╔════════════════════════════════════════╗
echo ║     星云盘 V2 - MySQL 快速启动工具     ║
echo ╚════════════════════════════════════════╝
echo.
echo  请选择操作:
echo.
echo  [1] 检查 MySQL 状态
echo  [2] 启动 MySQL 服务
echo  [3] 测试数据库连接
echo  [4] 查看启动指南
echo  [5] 打开宝塔面板
echo  [0] 退出
echo.
echo ════════════════════════════════════════
echo.

set /p choice=请输入选项 (0-5): 

if "%choice%"=="1" goto check_status
if "%choice%"=="2" goto start_mysql
if "%choice%"=="3" goto test_connection
if "%choice%"=="4" goto show_guide
if "%choice%"=="5" goto open_panel
if "%choice%"=="0" goto end
goto menu

:check_status
cls
echo.
echo ════════════════════════════════════════
echo  检查 MySQL 状态
echo ════════════════════════════════════════
echo.

echo [1] 检查 MySQL 服务...
sc query | findstr /i mysql
if %errorlevel% neq 0 (
    echo ❌ 未找到 MySQL 服务
) else (
    echo.
)

echo.
echo [2] 检查 3306 端口...
netstat -ano | findstr :3306
if %errorlevel% neq 0 (
    echo ❌ 端口 3306 未被占用，MySQL 可能未运行
) else (
    echo ✓ 端口 3306 已被占用，MySQL 可能正在运行
)

echo.
echo [3] 检查宝塔 MySQL 目录...
if exist "C:\BtSoft\mysql\bin\mysqld.exe" (
    echo ✓ 找到: C:\BtSoft\mysql\
) else if exist "D:\BtSoft\mysql\bin\mysqld.exe" (
    echo ✓ 找到: D:\BtSoft\mysql\
) else (
    echo ❌ 未找到宝塔 MySQL 目录
)

echo.
pause
goto menu

:start_mysql
cls
echo.
echo ════════════════════════════════════════
echo  启动 MySQL 服务
echo ════════════════════════════════════════
echo.

echo 正在尝试启动 MySQL...
echo.

net start MySQL
if %errorlevel% equ 0 (
    echo.
    echo ✓ MySQL 服务启动成功！
    echo.
    echo 等待服务完全启动...
    timeout /t 3 /nobreak >nul
    goto test_connection
)

net start MySQL96
if %errorlevel% equ 0 (
    echo.
    echo ✓ MySQL96 服务启动成功！
    echo.
    echo 等待服务完全启动...
    timeout /t 3 /nobreak >nul
    goto test_connection
)

net start MySQL57
if %errorlevel% equ 0 (
    echo.
    echo ✓ MySQL57 服务启动成功！
    echo.
    echo 等待服务完全启动...
    timeout /t 3 /nobreak >nul
    goto test_connection
)

echo.
echo ❌ 无法通过命令行启动 MySQL
echo.
echo 建议:
echo 1. 以管理员身份运行此脚本
echo 2. 或在宝塔面板中手动启动 MySQL
echo 3. 或使用选项 [5] 打开宝塔面板
echo.
pause
goto menu

:test_connection
cls
echo.
echo ════════════════════════════════════════
echo  测试数据库连接
echo ════════════════════════════════════════
echo.

cd /d "%~dp0.."
go run scripts/test-db-connection.go

echo.
pause
goto menu

:show_guide
cls
echo.
echo ════════════════════════════════════════
echo  MySQL 启动指南
echo ════════════════════════════════════════
echo.
echo 方法 1: 使用宝塔面板（推荐）
echo ────────────────────────────────────────
echo 1. 打开浏览器访问: http://127.0.0.1:888
echo 2. 登录宝塔面板
echo 3. 进入"软件商店"或"已安装"
echo 4. 找到 MySQL，点击"启动"按钮
echo.
echo 方法 2: 使用服务管理器
echo ────────────────────────────────────────
echo 1. 按 Win+R，输入: services.msc
echo 2. 找到 MySQL 服务
echo 3. 右键点击"启动"
echo.
echo 方法 3: 使用命令行（需要管理员权限）
echo ────────────────────────────────────────
echo 1. 以管理员身份打开命令提示符
echo 2. 运行: net start MySQL
echo.
echo 详细文档: docs\start-mysql-guide.md
echo.
pause
goto menu

:open_panel
cls
echo.
echo ════════════════════════════════════════
echo  打开宝塔面板
echo ════════════════════════════════════════
echo.
echo 正在打开宝塔面板...
echo.
start http://127.0.0.1:888
echo.
echo 如果浏览器没有自动打开，请手动访问:
echo http://127.0.0.1:888
echo.
echo 或尝试:
echo http://localhost:888
echo.
pause
goto menu

:end
cls
echo.
echo 感谢使用星云盘 MySQL 快速启动工具！
echo.
timeout /t 2 /nobreak >nul
exit
