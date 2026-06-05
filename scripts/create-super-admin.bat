@echo off
REM 创建超级管理员账户脚本
REM 用户名: Sincerity
REM 邮箱: 123456@xingyunnan.it.com

echo ========================================
echo 星云盘 - 创建超级管理员
echo ========================================
echo.

REM 从配置文件读取数据库连接信息
echo 正在读取数据库配置...
echo.

REM 设置数据库连接信息（从 configs/config.yaml 读取）
set DB_HOST=117.24.15.9
set DB_PORT=3306
set DB_USER=xingyunpan
set DB_PASS=xingyunpan123
set DB_NAME=xingyunpan_v2

echo 数据库信息:
echo   主机: %DB_HOST%:%DB_PORT%
echo   数据库: %DB_NAME%
echo   用户: %DB_USER%
echo.

echo 注意: 请确保用户 Sincerity (123456@xingyunnan.it.com) 已经注册
echo.
pause

echo.
echo 正在执行 SQL 脚本...
echo.

mysql -h %DB_HOST% -P %DB_PORT% -u %DB_USER% -p%DB_PASS% %DB_NAME% < scripts\create-super-admin.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo 超级管理员创建成功！
    echo ========================================
    echo.
    echo 超级管理员信息:
    echo   用户名: Sincerity
    echo   邮箱: 123456@xingyunnan.it.com
    echo   角色: admin
    echo.
    echo 其他用户已设置为普通用户 (role: user)
    echo.
) else (
    echo.
    echo ========================================
    echo 错误: 创建超级管理员失败
    echo ========================================
    echo.
    echo 可能的原因:
    echo 1. 用户还未注册，请先在前端注册
    echo 2. 数据库连接失败
    echo 3. MySQL 客户端未安装
    echo.
)

pause
