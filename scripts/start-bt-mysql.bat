@echo off
chcp 65001 >nul
echo ========================================
echo 宝塔 MySQL 启动脚本
echo ========================================
echo.

echo [方法 1] 尝试通过 Windows 服务启动 MySQL
echo.

REM 尝试常见的 MySQL 服务名
echo 正在查找 MySQL 服务...
sc query | findstr /i mysql

echo.
echo 尝试启动 MySQL 服务...
net start MySQL
if %errorlevel% equ 0 (
    echo ✓ MySQL 服务启动成功！
    goto :test_connection
)

net start MySQL96
if %errorlevel% equ 0 (
    echo ✓ MySQL96 服务启动成功！
    goto :test_connection
)

net start MySQL57
if %errorlevel% equ 0 (
    echo ✓ MySQL57 服务启动成功！
    goto :test_connection
)

echo.
echo [方法 2] 尝试通过宝塔目录启动 MySQL
echo.

REM 检查常见的宝塔安装路径
set BT_MYSQL_PATHS=C:\BtSoft\mysql D:\BtSoft\mysql E:\BtSoft\mysql C:\phpstudy_pro\Extensions\MySQL5.7.26 D:\phpstudy_pro\Extensions\MySQL5.7.26

for %%p in (%BT_MYSQL_PATHS%) do (
    if exist "%%p\bin\mysqld.exe" (
        echo 找到 MySQL: %%p
        echo.
        echo 尝试启动 MySQL 服务...
        
        REM 检查是否有 my.ini
        if exist "%%p\my.ini" (
            echo 使用配置文件: %%p\my.ini
            "%%p\bin\mysqld.exe" --defaults-file="%%p\my.ini" --console
        ) else (
            echo 使用默认配置启动
            "%%p\bin\mysqld.exe" --console
        )
        
        goto :end
    )
)

echo.
echo [方法 3] 手动指导
echo.
echo ❌ 未能自动启动 MySQL，请手动操作：
echo.
echo 1. 打开宝塔面板（通常在浏览器访问 http://127.0.0.1:888）
echo 2. 登录宝塔面板
echo 3. 进入"软件商店"或"已安装"
echo 4. 找到 MySQL，点击"启动"按钮
echo.
echo 或者：
echo 1. 打开"服务"管理器（Win+R 输入 services.msc）
echo 2. 找到 MySQL 相关服务
echo 3. 右键点击"启动"
echo.
goto :end

:test_connection
echo.
echo ========================================
echo 测试数据库连接
echo ========================================
echo.
timeout /t 3 /nobreak >nul
go run scripts/test-db-connection.go
goto :end

:end
echo.
echo ========================================
echo 脚本执行完成
echo ========================================
pause
