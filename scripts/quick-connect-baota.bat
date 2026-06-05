@echo off
chcp 65001 >nul
echo ========================================
echo 快速连接宝塔 MySQL
echo ========================================
echo.
echo 此脚本将帮助你：
echo 1. 测试宝塔 MySQL 连接
echo 2. 自动更新配置文件
echo 3. 启动应用程序
echo.
echo ========================================
echo.

REM 获取配置信息
set /p DB_HOST="宝塔服务器 IP 地址: "
set /p DB_PASSWORD="MySQL 密码: "

echo.
set /p DB_USERNAME="MySQL 用户名 [默认: root]: "
if "%DB_USERNAME%"=="" set DB_USERNAME=root

set /p DB_NAME="数据库名 [默认: xingyunpan]: "
if "%DB_NAME%"=="" set DB_NAME=xingyunpan

set DB_PORT=3306

echo.
echo ========================================
echo 步骤 1/3: 测试连接
echo ========================================
echo.

REM 导出环境变量供测试程序使用
set DB_HOST=%DB_HOST%
set DB_PORT=%DB_PORT%
set DB_USERNAME=%DB_USERNAME%
set DB_PASSWORD=%DB_PASSWORD%
set DB_NAME=%DB_NAME%

REM 运行测试
go run scripts/test-baota-connection.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ❌ 连接测试失败！
    echo.
    echo 请检查：
    echo 1. 宝塔面板是否已开放 MySQL 远程访问
    echo 2. 防火墙是否开放 3306 端口
    echo 3. IP 地址和密码是否正确
    echo.
    echo 详细说明请查看: docs\connect-baota-mysql.md
    echo.
    pause
    exit /b 1
)

echo.
echo ========================================
echo 步骤 2/3: 更新配置文件
echo ========================================
echo.

REM 备份配置文件
echo 正在备份配置文件...
copy configs\config.yaml configs\config.yaml.backup.%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2% >nul 2>&1

REM 使用 PowerShell 更新配置
powershell -ExecutionPolicy Bypass -File scripts/update-config-baota.ps1 -Host "%DB_HOST%" -Password "%DB_PASSWORD%" -Username "%DB_USERNAME%" -Database "%DB_NAME%"

echo.
echo ========================================
echo 步骤 3/3: 启动选项
echo ========================================
echo.
echo 配置已完成！请选择下一步操作：
echo.
echo [1] 运行数据库迁移（首次使用）
echo [2] 启动服务器
echo [3] 启动 Worker
echo [4] 退出
echo.
set /p CHOICE="请选择 [1-4]: "

if "%CHOICE%"=="1" (
    echo.
    echo 正在运行数据库迁移...
    go run scripts/migrate.go
    echo.
    echo 迁移完成！
    echo.
    set /p NEXT="是否启动服务器？[Y/N]: "
    if /i "%NEXT%"=="Y" (
        echo.
        echo 正在启动服务器...
        go run cmd/server/main.go
    )
) else if "%CHOICE%"=="2" (
    echo.
    echo 正在启动服务器...
    go run cmd/server/main.go
) else if "%CHOICE%"=="3" (
    echo.
    echo 正在启动 Worker...
    go run cmd/worker/main.go
) else (
    echo.
    echo 配置已保存，稍后可以手动启动应用。
)

echo.
pause
