@echo off
REM 星云盘 V2 日志轮转脚本 (Windows)
REM Log Rotation Script for Xingyunpan V2 (Windows)

setlocal enabledelayedexpansion

REM 配置
set LOG_DIR=C:\data\xingyunpan\logs
set APP_LOG=%LOG_DIR%\app.log
set MAX_SIZE_MB=100
set MAX_AGE_DAYS=7

echo === 星云盘 V2 日志轮转 ===
echo 日志目录: %LOG_DIR%
echo 应用日志: %APP_LOG%
echo.

REM 检查日志目录是否存在
if not exist "%LOG_DIR%" (
    echo 错误: 日志目录不存在: %LOG_DIR%
    exit /b 1
)

REM 检查日志文件是否存在
if not exist "%APP_LOG%" (
    echo 警告: 日志文件不存在: %APP_LOG%
    echo 日志文件将在应用启动时自动创建
    exit /b 0
)

REM 获取当前日志文件大小（MB）
for %%A in ("%APP_LOG%") do set FILE_SIZE=%%~zA
set /a CURRENT_SIZE_MB=!FILE_SIZE! / 1048576
echo 当前日志文件大小: !CURRENT_SIZE_MB!MB

REM 如果日志文件超过最大大小，执行轮转
if !CURRENT_SIZE_MB! GEQ %MAX_SIZE_MB% (
    echo 日志文件超过 %MAX_SIZE_MB%MB，执行轮转...
    
    REM 生成时间戳
    for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
    set TIMESTAMP=!datetime:~0,8!_!datetime:~8,6!
    set BACKUP_FILE=%LOG_DIR%\app-!TIMESTAMP!.log
    
    REM 复制当前日志文件
    copy "%APP_LOG%" "!BACKUP_FILE!" >nul
    echo 已备份日志文件: !BACKUP_FILE!
    
    REM 清空当前日志文件
    type nul > "%APP_LOG%"
    echo 已清空当前日志文件
    
    REM 压缩备份文件（需要安装 7-Zip 或其他压缩工具）
    if exist "C:\Program Files\7-Zip\7z.exe" (
        "C:\Program Files\7-Zip\7z.exe" a -tgzip "!BACKUP_FILE!.gz" "!BACKUP_FILE!" >nul
        del "!BACKUP_FILE!"
        echo 已压缩备份文件: !BACKUP_FILE!.gz
    )
) else (
    echo 日志文件大小正常，无需轮转
)

echo.
echo === 清理旧日志文件 ===

REM 删除超过指定天数的日志文件
set DELETED_COUNT=0
for %%F in ("%LOG_DIR%\app-*.log" "%LOG_DIR%\app-*.log.gz") do (
    if exist "%%F" (
        REM 使用 forfiles 检查文件年龄
        forfiles /P "%LOG_DIR%" /M "%%~nxF" /D -%MAX_AGE_DAYS% /C "cmd /c del @path" 2>nul
        if !errorlevel! equ 0 (
            echo 已删除旧日志文件: %%F
            set /a DELETED_COUNT+=1
        )
    )
)

if !DELETED_COUNT! equ 0 (
    echo 没有需要删除的旧日志文件
) else (
    echo 共删除 !DELETED_COUNT! 个旧日志文件
)

echo.
echo === 日志统计 ===
dir /b "%LOG_DIR%\app*.log*" 2>nul | find /c /v "" > temp.txt
set /p FILE_COUNT=<temp.txt
del temp.txt
echo 当前日志文件数量: !FILE_COUNT!

echo.
echo === 日志轮转完成 ===

endlocal
