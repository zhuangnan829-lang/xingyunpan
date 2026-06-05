@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
echo ========================================
echo 星云盘 - 一键修复所有问题
echo ========================================
echo.

echo [步骤 1/5] 检查环境依赖...
echo.

REM 检查Go
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [❌] Go未安装
    echo     下载: https://golang.org/dl/
    set HAS_ERROR=1
) else (
    for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
    echo [✓] Go已安装: %GO_VERSION%
)

REM 检查Node.js
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [❌] Node.js未安装
    echo     下载: https://nodejs.org/
    set HAS_ERROR=1
) else (
    for /f "tokens=1" %%i in ('node -v') do set NODE_VERSION=%%i
    echo [✓] Node.js已安装: %NODE_VERSION%
)

REM 检查MySQL - 自动检测服务名
set MYSQL_CHECK_FOUND=0
for %%s in (MySQL96 MySQL80 MySQL MySQL57 MySQL56 MYSQL mysqld) do (
    sc query %%s >nul 2>&1
    if !errorlevel! equ 0 (
        echo [✓] MySQL服务已安装: %%s
        set MYSQL_CHECK_FOUND=1
        goto :mysql_check_done
    )
)
:mysql_check_done
if !MYSQL_CHECK_FOUND! equ 0 (
    echo [❌] MySQL服务未找到
    set HAS_ERROR=1
)

if defined HAS_ERROR (
    echo.
    echo [错误] 缺少必要的环境依赖，请先安装
    pause
    exit /b 1
)

echo.
echo [步骤 2/5] 启动MySQL服务...

REM 自动检测MySQL服务名称
set MYSQL_SERVICE=
set MYSQL_FOUND=0

REM 尝试常见的MySQL服务名称
for %%s in (MySQL96 MySQL80 MySQL MySQL57 MySQL56 MYSQL mysqld) do (
    sc query %%s >nul 2>&1
    if !errorlevel! equ 0 (
        set MYSQL_SERVICE=%%s
        set MYSQL_FOUND=1
        goto :MYSQL_DETECTED
    )
)

:MYSQL_DETECTED
if !MYSQL_FOUND! equ 0 (
    echo [❌] MySQL服务未找到
    echo.
    echo 请运行: scripts\fix-mysql-service.bat
    echo 或手动安装MySQL: https://dev.mysql.com/downloads/
    pause
    exit /b 1
)

echo 检测到MySQL服务: !MYSQL_SERVICE!

REM 检查服务是否运行
sc query !MYSQL_SERVICE! | find "RUNNING" >nul
if !errorlevel! neq 0 (
    echo 正在启动MySQL...
    net start !MYSQL_SERVICE!
    if !errorlevel! neq 0 (
        echo [❌] MySQL启动失败
        echo 请运行: scripts\fix-mysql-service.bat
        pause
        exit /b 1
    )
)
echo [✓] MySQL运行中
echo.

echo [步骤 3/5] 检查数据库配置...
echo 正在测试数据库连接...
cd /d %~dp0..
go run scripts/test-db-simple.go
if %errorlevel% neq 0 (
    echo [❌] 数据库连接失败
    echo.
    echo 请检查配置文件: configs/config.yaml
    echo 确保数据库用户名、密码、数据库名正确
    pause
    exit /b 1
)
echo [✓] 数据库连接正常
echo.

echo [步骤 4/5] 启动后端服务...
REM 检查端口8080是否被占用
netstat -ano | findstr ":8080" | findstr "LISTENING" >nul
if %errorlevel% equ 0 (
    echo [提示] 端口8080已被占用，可能后端已在运行
) else (
    start "星云盘后端" cmd /k "cd /d %~dp0.. && go run cmd/server/main.go"
    echo [✓] 后端服务启动中...
    echo 等待后端服务初始化...
    timeout /t 5 /nobreak >nul
)
echo.

echo [步骤 5/5] 启动前端服务...
cd /d %~dp0..\frontend

REM 检查依赖
if not exist node_modules (
    echo [提示] 首次运行，安装前端依赖（可能需要几分钟）...
    call npm install
    if %errorlevel% neq 0 (
        echo [❌] 依赖安装失败
        pause
        exit /b 1
    )
)

REM 检查端口3000
netstat -ano | findstr ":3000" | findstr "LISTENING" >nul
if %errorlevel% equ 0 (
    echo [提示] 端口3000已被占用，可能前端已在运行
) else (
    start "星云盘前端" cmd /k "npm run dev"
    echo [✓] 前端服务启动中...
)

echo.
echo ========================================
echo ✅ 所有服务已启动！
echo ========================================
echo.
echo 📍 服务地址:
echo    后端: http://localhost:8080
echo    前端: http://localhost:3000
echo.
echo 📍 测试账号:
echo    管理员: admin2026 / Admin123456
echo    普通用户: Sincerity / #A456718293a
echo.
echo 等待10秒后打开诊断工具验证...
timeout /t 10 /nobreak >nul

REM 打开诊断工具
start "" "%~dp0diagnose-login-issue.html"

echo.
echo 💡 提示:
echo    - 如果诊断工具显示服务未运行，请等待服务完全启动
echo    - 后端服务启动需要5-10秒
echo    - 前端服务启动需要10-20秒
echo.
echo 按任意键关闭此窗口...
pause >nul
