@echo off
REM Phase 5 数据库迁移脚本 - Windows 版本
chcp 65001 >nul

REM 数据库配置
set DB_HOST=117.24.15.9
set DB_PORT=3306
set DB_USER=xingyunpan
set DB_PASS=RMGdedeMoMnMc
set DB_NAME=xingyunpan

echo 开始执行 Phase 5 数据库迁移...
echo 数据库: %DB_NAME% @ %DB_HOST%:%DB_PORT%
echo.

REM 检查 mysql 命令是否可用
where mysql >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 错误: 找不到 mysql 命令
    echo 请确保 MySQL 客户端已安装并添加到 PATH 环境变量
    pause
    exit /b 1
)

REM 执行迁移
mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER% -p%DB_PASS% %DB_NAME% < scripts\migrations\phase5_schema.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Phase 5 数据库迁移成功完成！
    echo.
    echo 已创建的表：
    echo   - shares ^(分享表^)
    echo   - share_files ^(分享文件关联表^)
    echo   - recycle_bin ^(回收站表^)
    echo   - file_versions ^(文件版本表^)
    echo   - collaborations ^(协作表^)
    echo.
    echo 已添加的索引：
    echo   - idx_file_name ^(文件名索引^)
    echo   - idx_modified_at ^(修改时间索引^)
    echo   - idx_file_size ^(文件大小索引^)
) else (
    echo.
    echo ❌ 数据库迁移失败，请检查错误信息
    pause
    exit /b 1
)

pause
