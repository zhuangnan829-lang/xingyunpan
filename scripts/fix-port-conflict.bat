@echo off
chcp 65001 >nul
echo ========================================
echo 端口冲突修复工具
echo ========================================
echo.

echo [1] 查找占用8080端口的进程
echo [2] 关闭占用8080端口的进程
echo [3] 修改配置使用其他端口(8081)
echo [4] 退出
echo.

set /p choice="请选择操作 (1-4): "

if "%choice%"=="1" goto find_process
if "%choice%"=="2" goto kill_process
if "%choice%"=="3" goto change_port
if "%choice%"=="4" goto end

:find_process
echo.
echo 正在查找占用8080端口的进程...
netstat -ano | findstr :8080
echo.
echo 上面显示的最后一列是进程ID(PID)
pause
goto end

:kill_process
echo.
echo 正在查找占用8080端口的进程...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do (
    set pid=%%a
    goto found_pid
)
echo 未找到占用8080端口的进程
pause
goto end

:found_pid
echo 找到进程 PID: %pid%
tasklist /FI "PID eq %pid%"
echo.
set /p confirm="确认要关闭这个进程吗? (Y/N): "
if /i "%confirm%"=="Y" (
    taskkill /F /PID %pid%
    echo 进程已关闭
) else (
    echo 操作已取消
)
pause
goto end

:change_port
echo.
echo 正在修改配置文件,将端口改为8081...
echo.

REM 备份配置文件
copy configs\config.yaml configs\config.yaml.backup >nul 2>&1
copy frontend\.env.development frontend\.env.development.backup >nul 2>&1

REM 修改后端配置
powershell -Command "(Get-Content configs\config.yaml) -replace 'port: 8080', 'port: 8081' -replace 'http://localhost:8080', 'http://localhost:8081' | Set-Content configs\config.yaml"

REM 修改前端配置
powershell -Command "(Get-Content frontend\.env.development) -replace 'http://localhost:8080', 'http://localhost:8081' | Set-Content frontend\.env.development"

echo 配置文件已修改:
echo   - 后端端口: 8080 → 8081
echo   - 前端API地址: http://localhost:8081
echo.
echo 原配置文件已备份为 .backup
echo.
echo 请重新启动服务
pause
goto end

:end
