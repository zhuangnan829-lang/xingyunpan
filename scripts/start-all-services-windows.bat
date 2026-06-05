@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
echo ========================================
echo 星云盘 - 启动所有服务
echo ========================================
echo.

REM 检查Go是否安装
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未检测到Go环境，请先安装Go
    echo 下载地址: https://golang.org/dl/
    pause
    exit /b 1
)

REM 检查Node.js是否安装
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未检测到Node.js环境，请先安装Node.js
    echo 下载地址: https://nodejs.org/
    pause
    exit /b 1
)

echo [1/4] 检查MySQL服务...
REM 自动检测MySQL服务名
set MYSQL_SERVICE=
set MYSQL_FOUND=0
for %%s in (MySQL96 MySQL80 MySQL MySQL57 MySQL56 MYSQL mysqld) do (
    sc query %%s >nul 2>&1
    if !errorlevel! equ 0 (
        set MYSQL_SERVICE=%%s
        set MYSQL_FOUND=1
        goto :mysql_found
    )
)
:mysql_found
if !MYSQL_FOUND! equ 0 (
    echo [错误] 未找到MySQL服务
    echo 请运行: scripts\fix-mysql-service.bat
    pause
    exit /b 1
)

echo 检测到MySQL服务: !MYSQL_SERVICE!
sc query !MYSQL_SERVICE! | find "RUNNING" >nul
if !errorlevel! neq 0 (
    echo [警告] MySQL服务未运行，尝试启动...
    net start !MYSQL_SERVICE!
    if !errorlevel! neq 0 (
        echo [错误] MySQL启动失败，请手动启动MySQL服务
        pause
        exit /b 1
    )
)
echo [✓] MySQL服务正常运行
echo.

echo [2/4] 检查Redis服务...
set REDIS_RUNNING=0

REM 方法1: 检查Windows服务形式的Redis
for %%s in (Redis redis) do (
    sc query %%s >nul 2>&1
    if !errorlevel! equ 0 (
        sc query %%s | find "RUNNING" >nul
        if !errorlevel! equ 0 (
            echo [✓] Redis Windows服务正常运行 (%%s)
            set REDIS_RUNNING=1
            goto :redis_check_done2
        )
    )
)

REM 方法2: 检查前台运行的redis-server.exe进程
tasklist /FI "IMAGENAME eq redis-server.exe" 2>NUL | find /I /N "redis-server.exe">NUL
if !errorlevel! equ 0 (
    echo [✓] Redis进程正常运行 (redis-server.exe)
    set REDIS_RUNNING=1
    goto :redis_check_done2
)

:redis_check_done2
if !REDIS_RUNNING! equ 0 (
    echo [警告] Redis未运行，请确保Redis已启动
    echo.
    echo 如果已安装 Redis Windows服务:
    echo   运行命令: net start Redis
    echo.
    echo 如果已安装 Redis 但未注册为服务:
    echo   运行 redis-server.exe
    echo.
    echo 如果未安装Redis，请访问: https://github.com/tporadowski/redis/releases
)
echo.

echo [3/4] 启动后端服务 (端口 8080)...
start "星云盘后端服务" cmd /k "cd /d %~dp0.. && echo 正在启动后端服务... && go run cmd/server/main.go"
echo [✓] 后端服务启动中...
timeout /t 3 /nobreak >nul
echo.

echo [4/4] 启动前端服务 (端口 3000)...
cd /d %~dp0..\frontend
if not exist node_modules (
    echo [提示] 首次运行，正在安装依赖...
    call npm install
)
start "星云盘前端服务" cmd /k "npm run dev"
echo [✓] 前端服务启动中...
echo.

echo ========================================
echo 服务启动完成！
echo ========================================
echo.
echo 后端服务: http://localhost:8080
echo 前端服务: http://localhost:3000
echo.
echo 等待5秒后自动打开诊断工具...
timeout /t 5 /nobreak >nul

REM 打开诊断工具
start "" "%~dp0diagnose-login-issue.html"

echo.
echo 按任意键关闭此窗口...
pause >nul
