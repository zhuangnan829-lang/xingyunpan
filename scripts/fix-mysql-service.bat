@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
echo ========================================
echo 星云盘系统 - 完整启动脚本
echo ========================================
echo.

echo [检查] 正在检查系统环境...
echo.

REM 检查 MySQL 服务
echo [1/4] 检查 MySQL 服务...
set MYSQL_SERVICE=
set MYSQL_FOUND=0
for %%s in (MySQL96 MySQL80 MySQL MySQL57 MySQL56 MYSQL mysqld) do (
    sc query %%s >nul 2>&1
    if !errorlevel! equ 0 (
        set MYSQL_SERVICE=%%s
        set MYSQL_FOUND=1
        goto :mysql_check
    )
)
:mysql_check
if !MYSQL_FOUND! equ 0 (
    echo [警告] 未找到 MySQL 服务
    echo [提示] 请运行: scripts\fix-mysql-service.bat
    echo.
    pause
    exit /b 1
)

echo 检测到MySQL服务: !MYSQL_SERVICE!
sc query !MYSQL_SERVICE! | find "RUNNING" >nul
if !errorlevel! neq 0 (
    echo [警告] MySQL 服务未运行,尝试启动...
    net start !MYSQL_SERVICE!
    if !errorlevel! neq 0 (
        echo [错误] MySQL 启动失败
        echo [提示] 请手动启动或运行: scripts\fix-mysql-service.bat
        pause
        exit /b 1
    )
)
echo [✓] MySQL 服务正常运行
echo.

REM 检查 Redis 服务
echo [2/4] 检查 Redis 服务...
set REDIS_RUNNING=0

REM 检查Windows服务形式的Redis
for %%s in (Redis redis) do (
    sc query %%s >nul 2>&1
    if !errorlevel! equ 0 (
        sc query %%s | find "RUNNING" >nul
        if !errorlevel! equ 0 (
            echo [✓] Redis Windows服务正常运行 (%%s)
            set REDIS_RUNNING=1
            goto :redis_check_done
        )
    )
)

REM 检查前台运行的redis-server.exe进程
tasklist /FI "IMAGENAME eq redis-server.exe" 2>NUL | find /I /N "redis-server.exe">NUL
if !errorlevel! equ 0 (
    echo [✓] Redis进程正常运行 (redis-server.exe)
    set REDIS_RUNNING=1
    goto :redis_check_done
)

:redis_check_done
if !REDIS_RUNNING! equ 0 (
    echo [警告] Redis 服务未运行
    echo [提示] 系统将在无Redis模式下启动(性能会降低)
    echo.
    choice /C YN /M "是否继续启动(不使用Redis缓存)"
    if errorlevel 2 exit /b 1
    echo [!] 将在无 Redis 模式下启动
) else (
    echo [✓] Redis 服务正常运行，跳过降级提示
)
echo.

REM 检查配置文件
echo [3/4] 检查配置文件...
if not exist "configs\config.yaml" (
    echo [警告] 配置文件不存在
    if exist "configs\config.yaml.example" (
        echo [提示] 正在从示例文件创建配置...
        copy "configs\config.yaml.example" "configs\config.yaml" >nul
        echo [✓] 配置文件已创建
    ) else (
        echo [错误] 找不到配置文件模板
        pause
        exit /b 1
    )
) else (
    echo [✓] 配置文件存在
)
echo.

REM 检查端口占用
echo [4/4] 检查端口占用...
netstat -ano | findstr :8080 | findstr LISTENING >nul
if %errorlevel% equ 0 (
    echo [警告] 8080端口已被占用!
    echo.
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do (
        echo [进程] PID: %%a
        tasklist /FI "PID eq %%a" /FO TABLE | findstr /V "="
        echo.
        echo [解决方案]
        echo   1. 自动关闭占用进程 (运行 quick-fix-port.bat)
        echo   2. 修改后端端口配置 (运行 fix-port-conflict.bat)
        echo   3. 手动处理后重试
        echo.
        choice /C 123 /M "请选择操作"
        if errorlevel 3 (
            echo [提示] 请手动关闭占用8080端口的程序后重新运行此脚本
            pause
            exit /b 1
        )
        if errorlevel 2 (
            call scripts\fix-port-conflict.bat
            exit /b 0
        )
        if errorlevel 1 (
            call scripts\quick-fix-port.bat
            echo.
            echo [继续] 端口已释放,继续启动...
        )
        goto port_check_done
    )
) else (
    echo [✓] 8080端口可用
)
:port_check_done
echo.

echo ========================================
echo 正在启动服务...
echo ========================================
echo.

REM 启动后端服务
echo [启动] 后端服务 (端口 8080)...
start "星云盘后端" cmd /k "title 星云盘后端服务 && cd /d %~dp0.. && go run cmd/server/main.go"
echo [✓] 后端服务启动中...
echo.

REM 等待后端启动
echo [等待] 等待后端服务就绪 (5秒)...
timeout /t 5 /nobreak >nul
echo.

REM 启动前端服务
echo [启动] 前端服务 (端口 3000)...
cd /d %~dp0..\frontend
if not exist node_modules (
    echo [提示] 首次运行,正在安装依赖...
    call npm install
)
start "星云盘前端" cmd /k "title 星云盘前端服务 && npm run dev"
echo [✓] 前端服务启动中...
echo.

REM 等待前端启动
echo [等待] 等待前端服务就绪 (3秒)...
timeout /t 3 /nobreak >nul
echo.

echo ========================================
echo ✅ 系统启动完成!
echo ========================================
echo.
echo 📍 服务地址:
echo   • 后端 API:  http://localhost:8080
echo   • 前端界面:  http://localhost:3000
echo   • 登录页面:  http://localhost:3000/#/login
echo.
echo 👤 测试账号:
echo   • 用户名: admin2026
echo   • 密码:   Admin123456
echo.
echo 🔧 诊断工具:
echo   • 在浏览器打开: scripts\diagnose-login-issue.html
echo.
echo 📊 健康检查:
echo   • 运行命令: curl http://localhost:8080/health
echo.
echo 💡 提示:
echo   • 后端和前端服务在独立窗口运行
echo   • 关闭对应窗口即可停止服务
echo   • 查看日志请查看对应服务窗口
echo.

REM 询问是否打开浏览器
choice /C YN /M "是否自动打开浏览器"
if errorlevel 2 goto :end
echo.
echo [打开] 正在打开浏览器...
start http://localhost:3000/#/login

:end
echo.
echo 按任意键关闭此窗口...
pause >nul