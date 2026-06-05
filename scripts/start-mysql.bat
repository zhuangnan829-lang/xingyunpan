@echo off
chcp 65001 >nul
echo 正在启动 MySQL...
echo.

REM 检查是否有管理员权限
net session >nul 2>&1
if %errorlevel% == 0 (
    echo [管理员模式] 正在运行...
    echo.
    powershell -ExecutionPolicy Bypass -File "%~dp0start-mysql.ps1"
) else (
    echo [普通模式] 尝试启动...
    echo 注意: 某些操作可能需要管理员权限
    echo.
    powershell -ExecutionPolicy Bypass -File "%~dp0start-mysql.ps1"
)

pause
