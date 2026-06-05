@echo off
echo ========================================
echo 星云盘 - 一键启动所有服务
echo ========================================
echo.

echo 此脚本将启动以下服务:
echo 1. 后端服务 (端口 8080)
echo 2. 前端服务 (端口 3000)
echo.
echo 注意: 请确保 MySQL 数据库已经启动
echo.
pause

echo.
echo ========================================
echo 步骤 1/2: 启动后端服务
echo ========================================
echo.

start "星云盘后端服务" cmd /k "cd /d %~dp0.. && go run cmd/server/main.go"

echo 等待后端服务启动...
timeout /t 5 /nobreak >nul

echo.
echo ========================================
echo 步骤 2/2: 启动前端服务
echo ========================================
echo.

start "星云盘前端服务" cmd /k "cd /d %~dp0..\frontend && npm run dev"

echo.
echo ========================================
echo ✅ 所有服务启动完成！
echo ========================================
echo.
echo 服务地址:
echo   • 后端 API: http://localhost:8080
echo   • 前端界面: http://localhost:3000
echo   • 登录页面: http://localhost:3000/#/login
echo.
echo 测试账号:
echo   • 用户名: admin2026
echo   • 密码: Admin123456
echo.
echo 诊断工具: 在浏览器中打开 scripts/diagnose-login-issue.html
echo.
echo 按任意键关闭此窗口...
pause >nul
