@echo off
REM 设置超级管理员 - 便捷脚本

echo ========================================
echo 星云盘 - 设置超级管理员
echo ========================================
echo.
echo 此脚本将执行以下操作:
echo 1. 将用户 Sincerity (123456@xingyunnan.it.com) 设置为超级管理员
echo 2. 将其他所有用户设置为普通用户
echo.
echo 注意: 请确保该用户已经通过前端注册！
echo.
pause

echo.
echo 正在运行设置程序...
echo.

cd /d %~dp0..
go run scripts/create-super-admin.go

echo.
pause
