@echo off
REM 分析性能压测结果

setlocal enabledelayedexpansion

set RESULTS_DIR=benchmark-results

echo ========================================
echo 性能压测结果分析
echo ========================================
echo.

REM 检查结果目录是否存在
if not exist %RESULTS_DIR% (
    echo [错误] 结果目录不存在: %RESULTS_DIR%
    echo 请先运行性能压测: scripts\benchmark.bat
    pause
    exit /b 1
)

REM 查找最新的结果文件
for /f "delims=" %%i in ('dir /b /o-d %RESULTS_DIR%\benchmark_*.txt 2^>nul') do (
    set LATEST_FILE=%RESULTS_DIR%\%%i
    goto :found
)

:found
if not defined LATEST_FILE (
    echo [错误] 未找到压测结果文件
    echo 请先运行性能压测: scripts\benchmark.bat
    pause
    exit /b 1
)

echo 分析文件: %LATEST_FILE%
echo.

REM 提取关键指标
echo ========================================
echo 性能指标摘要
echo ========================================
echo.

REM 使用 findstr 提取关键行
echo [健康检查 API]
findstr /C:"测试: 健康检查 API" %LATEST_FILE% >nul 2>nul
if %errorlevel% equ 0 (
    findstr /C:"Requests per second:" %LATEST_FILE% | findstr /A:0 /C:"Requests per second:" | more +0
    findstr /C:"Time per request:" %LATEST_FILE% | findstr /A:0 /C:"Time per request:" | more +0
    findstr /C:"Failed requests:" %LATEST_FILE% | findstr /A:0 /C:"Failed requests:" | more +0
)
echo.

echo [用户注册 API]
findstr /C:"测试: 用户注册 API" %LATEST_FILE% >nul 2>nul
if %errorlevel% equ 0 (
    findstr /C:"Requests per second:" %LATEST_FILE% | findstr /A:1 /C:"Requests per second:" | more +1
    findstr /C:"Time per request:" %LATEST_FILE% | findstr /A:1 /C:"Time per request:" | more +1
    findstr /C:"Failed requests:" %LATEST_FILE% | findstr /A:1 /C:"Failed requests:" | more +1
)
echo.

echo [用户登录 API]
findstr /C:"测试: 用户登录 API" %LATEST_FILE% >nul 2>nul
if %errorlevel% equ 0 (
    findstr /C:"Requests per second:" %LATEST_FILE% | findstr /A:2 /C:"Requests per second:" | more +2
    findstr /C:"Time per request:" %LATEST_FILE% | findstr /A:2 /C:"Time per request:" | more +2
    findstr /C:"Failed requests:" %LATEST_FILE% | findstr /A:2 /C:"Failed requests:" | more +2
)
echo.

echo [分片上传初始化 API]
findstr /C:"测试: 分片上传初始化 API" %LATEST_FILE% >nul 2>nul
if %errorlevel% equ 0 (
    findstr /C:"Requests per second:" %LATEST_FILE% | findstr /A:3 /C:"Requests per second:" | more +3
    findstr /C:"Time per request:" %LATEST_FILE% | findstr /A:3 /C:"Time per request:" | more +3
    findstr /C:"Failed requests:" %LATEST_FILE% | findstr /A:3 /C:"Failed requests:" | more +3
)
echo.

echo ========================================
echo 完整报告: %LATEST_FILE%
echo ========================================
echo.
pause
