@echo off
REM 星云盘 V2 数据库备份脚本 (Windows)
REM Database Backup Script for Xingyunpan V2 (Windows)

setlocal enabledelayedexpansion

REM 配置
if "%BACKUP_DIR%"=="" set BACKUP_DIR=C:\backup\xingyunpan
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=3306
if "%DB_USER%"=="" set DB_USER=xingyunpan
if "%DB_NAME%"=="" set DB_NAME=xingyunpan
if "%RETENTION_DAYS%"=="" set RETENTION_DAYS=30

REM 生成时间戳
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set DATE=!datetime:~0,8!_!datetime:~8,6!
set BACKUP_FILE=%BACKUP_DIR%\xingyunpan_!DATE!.sql

echo === 星云盘 V2 数据库备份开始 ===
echo 备份目录: %BACKUP_DIR%
echo 数据库: %DB_NAME%@%DB_HOST%:%DB_PORT%
echo 备份文件: %BACKUP_FILE%
echo.

REM 检查必需的环境变量
if "%DB_PASSWORD%"=="" (
    echo 错误: DB_PASSWORD 环境变量未设置
    exit /b 1
)

REM 创建备份目录
if not exist "%BACKUP_DIR%" (
    echo 创建备份目录: %BACKUP_DIR%
    mkdir "%BACKUP_DIR%"
)

REM 检查 mysqldump 是否存在
where mysqldump >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: mysqldump 命令不存在，请安装 MySQL 客户端
    exit /b 1
)

REM 执行数据库备份
echo 开始备份数据库...
mysqldump ^
    --host=%DB_HOST% ^
    --port=%DB_PORT% ^
    --user=%DB_USER% ^
    --password=%DB_PASSWORD% ^
    --single-transaction ^
    --quick ^
    --lock-tables=false ^
    --routines ^
    --triggers ^
    --events ^
    %DB_NAME% > "%BACKUP_FILE%" 2>&1

if %errorlevel% neq 0 (
    echo 错误: 数据库备份失败
    del "%BACKUP_FILE%" 2>nul
    exit /b 1
)

REM 验证备份文件不为空
for %%A in ("%BACKUP_FILE%") do set FILE_SIZE=%%~zA
if !FILE_SIZE! equ 0 (
    echo 错误: 备份文件为空
    del "%BACKUP_FILE%"
    exit /b 1
)

set /a BACKUP_SIZE_MB=!FILE_SIZE! / 1048576
echo 备份完成，文件大小: !BACKUP_SIZE_MB!MB

REM 压缩备份文件（需要安装 7-Zip）
if exist "C:\Program Files\7-Zip\7z.exe" (
    echo 压缩备份文件...
    "C:\Program Files\7-Zip\7z.exe" a -tgzip "%BACKUP_FILE%.gz" "%BACKUP_FILE%" >nul
    if !errorlevel! equ 0 (
        del "%BACKUP_FILE%"
        set COMPRESSED_FILE=%BACKUP_FILE%.gz
        for %%A in ("!COMPRESSED_FILE!") do set COMPRESSED_SIZE=%%~zA
        set /a COMPRESSED_SIZE_MB=!COMPRESSED_SIZE! / 1048576
        echo 压缩完成，压缩后大小: !COMPRESSED_SIZE_MB!MB
    ) else (
        echo 警告: 压缩失败，保留未压缩文件
    )
) else (
    echo 警告: 未找到 7-Zip，跳过压缩
)

REM 清理旧备份
echo 清理超过 %RETENTION_DAYS% 天的旧备份...
set DELETED_COUNT=0
for %%F in ("%BACKUP_DIR%\xingyunpan_*.sql" "%BACKUP_DIR%\xingyunpan_*.sql.gz") do (
    if exist "%%F" (
        forfiles /P "%BACKUP_DIR%" /M "%%~nxF" /D -%RETENTION_DAYS% /C "cmd /c del @path" 2>nul
        if !errorlevel! equ 0 (
            echo 删除旧备份: %%~nxF
            set /a DELETED_COUNT+=1
        )
    )
)

if !DELETED_COUNT! equ 0 (
    echo 没有需要删除的旧备份
) else (
    echo 共删除 !DELETED_COUNT! 个旧备份文件
)

REM 备份统计
echo.
echo === 备份统计 ===
dir /b "%BACKUP_DIR%\xingyunpan_*.sql*" 2>nul | find /c /v "" > temp.txt
set /p BACKUP_COUNT=<temp.txt
del temp.txt
echo 当前备份文件数量: !BACKUP_COUNT!

echo.
echo === 数据库备份完成 ===

endlocal
exit /b 0
