@echo off
REM 星云盘 V2 性能压测脚本 (Windows)
REM 使用 Apache Bench (ab) 进行 API 性能测试

setlocal enabledelayedexpansion

REM 配置
set BASE_URL=http://localhost:8080
set API_BASE=%BASE_URL%/api/v1
set RESULTS_DIR=benchmark-results
set TIMESTAMP=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%
set REPORT_FILE=%RESULTS_DIR%\benchmark_%TIMESTAMP%.txt

REM 创建结果目录
if not exist %RESULTS_DIR% mkdir %RESULTS_DIR%

echo ========================================
echo 星云盘 V2 性能压测
echo ========================================
echo.

REM 检查 ab 是否安装
where ab >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] Apache Bench ^(ab^) 未安装
    echo.
    echo 请安装 Apache Bench:
    echo   1. 下载 Apache HTTP Server: https://www.apachelounge.com/download/
    echo   2. 解压后将 bin 目录添加到 PATH 环境变量
    echo   3. 或使用 chocolatey: choco install apache-httpd
    echo.
    pause
    exit /b 1
)

REM 检查服务器是否运行
echo 检查服务器状态...
curl -s %BASE_URL%/health >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 服务器未运行或无法访问 %BASE_URL%
    pause
    exit /b 1
)
echo [成功] 服务器运行正常
echo.

REM 初始化报告文件
echo ======================================== > %REPORT_FILE%
echo 星云盘 V2 性能压测报告 >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 测试时间: %date% %time% >> %REPORT_FILE%
echo 服务器地址: %BASE_URL% >> %REPORT_FILE%
echo 测试工具: Apache Bench ^(ab^) >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo 开始性能压测...
echo 结果将保存到: %REPORT_FILE%
echo.

REM ========================================
REM 测试 1: 健康检查 API
REM ========================================
echo [测试 1/4] 健康检查 API
echo. >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 测试: 健康检查 API >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo URL: %BASE_URL%/health >> %REPORT_FILE%
echo 方法: GET >> %REPORT_FILE%
echo 请求数: 1000 >> %REPORT_FILE%
echo 并发数: 10 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

ab -n 1000 -c 10 %BASE_URL%/health >> %REPORT_FILE% 2>&1
echo   [完成] 健康检查 API
echo.

REM ========================================
REM 测试 2: 用户注册 API
REM ========================================
echo [测试 2/4] 用户注册 API
echo. >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 测试: 用户注册 API >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo URL: %API_BASE%/user/register >> %REPORT_FILE%
echo 方法: POST >> %REPORT_FILE%
echo 请求数: 100 >> %REPORT_FILE%
echo 并发数: 5 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

ab -n 100 -c 5 -p scripts\test-data\register.json -T application/json %API_BASE%/user/register >> %REPORT_FILE% 2>&1
echo   [完成] 用户注册 API
echo.

REM ========================================
REM 测试 3: 用户登录 API
REM ========================================
echo [测试 3/4] 用户登录 API
echo. >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 测试: 用户登录 API >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo URL: %API_BASE%/user/login >> %REPORT_FILE%
echo 方法: POST >> %REPORT_FILE%
echo 请求数: 1000 >> %REPORT_FILE%
echo 并发数: 10 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

ab -n 1000 -c 10 -p scripts\test-data\login.json -T application/json %API_BASE%/user/login >> %REPORT_FILE% 2>&1
echo   [完成] 用户登录 API
echo.

REM ========================================
REM 测试 4: 分片上传初始化 API
REM ========================================
echo [测试 4/4] 分片上传初始化 API
echo. >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 测试: 分片上传初始化 API >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo URL: %API_BASE%/files/multipart/init >> %REPORT_FILE%
echo 方法: POST >> %REPORT_FILE%
echo 请求数: 100 >> %REPORT_FILE%
echo 并发数: 5 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

ab -n 100 -c 5 -p scripts\test-data\multipart-init.json -T application/json %API_BASE%/files/multipart/init >> %REPORT_FILE% 2>&1
echo   [完成] 分片上传初始化 API
echo.

REM ========================================
REM 生成摘要
REM ========================================
echo. >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 测试摘要 >> %REPORT_FILE%
echo ======================================== >> %REPORT_FILE%
echo 所有测试已完成 >> %REPORT_FILE%
echo 详细结果请查看上方各项测试数据 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ========================================
echo 性能压测完成！
echo ========================================
echo 报告文件: %REPORT_FILE%
echo.
echo 查看报告:
echo   type %REPORT_FILE%
echo.
pause
