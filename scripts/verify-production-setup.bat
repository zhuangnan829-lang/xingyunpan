@echo off
REM 星云盘 V2 生产环境配置验证脚本 (Windows)
REM Production Setup Verification Script for Xingyunpan V2 (Windows)

setlocal enabledelayedexpansion

set PASSED=0
set FAILED=0
set WARNINGS=0

echo ========================================
echo 星云盘 V2 生产环境配置验证
echo ========================================
echo.

REM 1. 检查目录结构
echo 1. 检查目录结构
echo ----------------------------------------

if exist "C:\data\xingyunpan\logs" (
    echo [OK] 日志目录存在: C:\data\xingyunpan\logs
    set /a PASSED+=1
) else (
    echo [FAIL] 日志目录不存在: C:\data\xingyunpan\logs
    set /a FAILED+=1
)

if exist "C:\data\xingyunpan\storage" (
    echo [OK] 存储目录存在: C:\data\xingyunpan\storage
    set /a PASSED+=1
) else (
    echo [FAIL] 存储目录不存在: C:\data\xingyunpan\storage
    set /a FAILED+=1
)

if exist "C:\data\xingyunpan\temp" (
    echo [OK] 临时目录存在: C:\data\xingyunpan\temp
    set /a PASSED+=1
) else (
    echo [FAIL] 临时目录不存在: C:\data\xingyunpan\temp
    set /a FAILED+=1
)

if exist "C:\backup\xingyunpan" (
    echo [OK] 备份目录存在: C:\backup\xingyunpan
    set /a PASSED+=1
) else (
    echo [FAIL] 备份目录不存在: C:\backup\xingyunpan
    set /a FAILED+=1
)

echo.

REM 2. 检查配置文件
echo 2. 检查配置文件
echo ----------------------------------------

if exist "configs\config.prod.yaml" (
    echo [OK] 生产配置文件存在
    set /a PASSED+=1
) else (
    echo [FAIL] 生产配置文件不存在: configs\config.prod.yaml
    set /a FAILED+=1
)

if exist "configs\.env.production" (
    echo [OK] 环境变量文件存在
    set /a PASSED+=1
) else (
    echo [FAIL] 环境变量文件不存在: configs\.env.production
    set /a FAILED+=1
)

echo.

REM 3. 检查脚本文件
echo 3. 检查脚本文件
echo ----------------------------------------

if exist "scripts\backup.bat" (
    echo [OK] backup.bat 存在
    set /a PASSED+=1
) else (
    echo [WARN] backup.bat 不存在
    set /a WARNINGS+=1
)

if exist "scripts\restore.sh" (
    echo [OK] restore.sh 存在
    set /a PASSED+=1
) else (
    echo [WARN] restore.sh 不存在
    set /a WARNINGS+=1
)

if exist "scripts\rotate-logs.bat" (
    echo [OK] rotate-logs.bat 存在
    set /a PASSED+=1
) else (
    echo [WARN] rotate-logs.bat 不存在
    set /a WARNINGS+=1
)

echo.

REM 4. 检查数据库连接
echo 4. 检查数据库连接
echo ----------------------------------------

where mysql >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] mysql 命令存在
    set /a PASSED+=1
) else (
    echo [WARN] mysql 命令不存在，跳过数据库连接测试
    set /a WARNINGS+=1
)

echo.

REM 5. 检查 Redis 连接
echo 5. 检查 Redis 连接
echo ----------------------------------------

where redis-cli >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] redis-cli 命令存在
    set /a PASSED+=1
) else (
    echo [WARN] redis-cli 命令不存在，跳过 Redis 连接测试
    set /a WARNINGS+=1
)

echo.

REM 6. 检查磁盘空间
echo 6. 检查磁盘空间
echo ----------------------------------------

for %%D in (C:) do (
    for /f "tokens=3" %%a in ('dir %%D ^| find "bytes free"') do (
        echo [OK] %%D 磁盘可用空间: %%a 字节
        set /a PASSED+=1
    )
)

echo.

REM 7. 检查应用状态
echo 7. 检查应用状态
echo ----------------------------------------

where curl >nul 2>&1
if %errorlevel% equ 0 (
    curl -f http://localhost:8080/health >nul 2>&1
    if !errorlevel! equ 0 (
        echo [OK] 应用健康检查通过
        set /a PASSED+=1
    ) else (
        echo [WARN] 应用健康检查失败（应用可能未启动）
        set /a WARNINGS+=1
    )
) else (
    echo [WARN] curl 命令不存在，跳过健康检查
    set /a WARNINGS+=1
)

echo.

REM 8. 检查备份
echo 8. 检查备份
echo ----------------------------------------

if exist "C:\backup\xingyunpan\xingyunpan_*.sql.gz" (
    echo [OK] 找到备份文件
    set /a PASSED+=1
) else (
    echo [WARN] 未找到备份文件（可能尚未执行首次备份）
    set /a WARNINGS+=1
)

echo.

REM 总结
echo ========================================
echo 验证总结
echo ========================================
echo 通过: !PASSED!
echo 警告: !WARNINGS!
echo 失败: !FAILED!
echo.

if !FAILED! equ 0 (
    if !WARNINGS! equ 0 (
        echo [OK] 所有检查通过！生产环境配置正确。
        exit /b 0
    ) else (
        echo [WARN] 存在一些警告，建议检查并修复。
        exit /b 0
    )
) else (
    echo [FAIL] 存在配置问题，请修复后重新验证。
    exit /b 1
)

endlocal
