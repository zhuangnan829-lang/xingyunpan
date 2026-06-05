@echo off
REM 路径: scripts/verify-indexes.bat
REM 验证数据库索引

echo ========================================
echo 验证数据库索引
echo ========================================
echo.

echo 运行索引验证脚本...
go run scripts/verify-indexes.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [错误] 索引验证失败
    exit /b 1
)

echo.
echo ========================================
echo ✅ 索引验证完成!
echo ========================================
