@echo off
echo ========================================
echo 星云盘 V2 - 设置超级管理员
echo ========================================
echo.

echo 正在设置超级管理员账户...
echo 用户名: Sincerity
echo 邮箱: 123456@xingyunnan.it.com
echo.

go run scripts/setup-admin.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo 设置完成!
    echo ========================================
) else (
    echo.
    echo ========================================
    echo 设置失败，请检查错误信息
    echo ========================================
)

pause
