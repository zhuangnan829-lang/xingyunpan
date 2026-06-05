@echo off
chcp 65001 >nul
echo ========================================
echo 星云盘数据库诊断工具
echo ========================================
echo.

echo [1] 检查 MySQL 服务状态
echo.
sc query MySQL
if %errorlevel% neq 0 (
    echo 尝试查找其他 MySQL 服务...
    sc query | findstr /i mysql
)
echo.

echo [2] 检查 3306 端口占用情况
echo.
netstat -ano | findstr :3306
if %errorlevel% neq 0 (
    echo 端口 3306 未被占用，MySQL 可能未启动
)
echo.

echo [3] 检查宝塔 MySQL 服务
echo.
if exist "C:\BtSoft\mysql\bin\mysql.exe" (
    echo ✓ 找到宝塔 MySQL: C:\BtSoft\mysql\bin\mysql.exe
) else (
    echo ✗ 未找到宝塔 MySQL
)
echo.

echo [4] 尝试使用 mysql 命令连接
echo.
where mysql
if %errorlevel% equ 0 (
    echo 找到 mysql 命令，尝试连接...
    mysql -h 127.0.0.1 -P 3306 -u xingyunpan -puKcgJLKyWHHJzm -e "SELECT 'Connection OK' as status;"
) else (
    echo mysql 命令不在 PATH 中
)
echo.

echo ========================================
echo 诊断完成
echo ========================================
echo.
echo 建议:
echo 1. 如果 MySQL 服务未启动，请在宝塔面板中启动 MySQL
echo 2. 确认数据库用户名和密码是否正确
echo 3. 检查防火墙设置是否允许 3306 端口
echo.
pause
