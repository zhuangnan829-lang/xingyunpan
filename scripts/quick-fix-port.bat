@echo off
chcp 65001 >nul
echo ========================================
echo 快速修复8080端口冲突
echo ========================================
echo.

echo 正在查找占用8080端口的进程...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do (
    set pid=%%a
    goto found
)

echo 未找到占用8080端口的进程,端口可用
pause
exit /b 0

:found
echo.
echo 找到占用8080端口的进程:
tasklist /FI "PID eq %pid%" /FO TABLE
echo.
echo 进程ID: %pid%
echo.
echo 正在尝试关闭该进程...
taskkill /F /PID %pid%

if %errorlevel% equ 0 (
    echo.
    echo ✓ 进程已成功关闭
    echo ✓ 8080端口现在可用
    echo.
    echo 现在可以启动后端服务了
) else (
    echo.
    echo ✗ 无法关闭进程,可能需要管理员权限
    echo.
    echo 请右键点击此脚本,选择"以管理员身份运行"
    echo 或者运行 fix-port-conflict.bat 选择修改端口
)

echo.
pause
