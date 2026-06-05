@echo off
REM 准备性能压测环境

setlocal enabledelayedexpansion

set BASE_URL=http://localhost:8080
set API_BASE=%BASE_URL%/api/v1

echo ========================================
echo 准备性能压测环境
echo ========================================
echo.

REM 检查服务器是否运行
echo 检查服务器状态...
curl -s %BASE_URL%/health >nul 2>nul
if %errorlevel% neq 0 (
    echo [警告] 服务器未运行
    echo.
    echo 请先启动服务器:
    echo   cd cmd/server
    echo   go run main.go
    echo.
    echo 或者编译后运行:
    echo   go build -o bin/server.exe cmd/server/main.go
    echo   bin\server.exe
    echo.
    pause
    exit /b 1
)
echo [成功] 服务器运行正常
echo.

REM 创建测试用户
echo 创建测试用户...
curl -X POST %API_BASE%/user/register ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"testuser\",\"password\":\"Test123456\",\"email\":\"test@example.com\"}" ^
  -s -o nul -w "HTTP Status: %%{http_code}\n"

if %errorlevel% equ 0 (
    echo [成功] 测试用户创建完成 ^(如果用户已存在会返回错误，可以忽略^)
) else (
    echo [警告] 测试用户创建失败，可能已存在
)
echo.

REM 验证登录
echo 验证测试用户登录...
curl -X POST %API_BASE%/user/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"testuser\",\"password\":\"Test123456\"}" ^
  -s -o nul -w "HTTP Status: %%{http_code}\n"

if %errorlevel% equ 0 (
    echo [成功] 测试用户登录验证通过
) else (
    echo [错误] 测试用户登录失败
    pause
    exit /b 1
)
echo.

echo ========================================
echo 环境准备完成！
echo ========================================
echo.
echo 现在可以运行性能压测:
echo   scripts\benchmark.bat
echo.
pause
