@echo off
chcp 65001 >nul
echo ========================================
echo 星云盘登录问题 - 一键诊断修复
echo ========================================
echo.

echo [步骤 1/5] 检查后端服务...
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 后端服务未运行
    echo.
    echo 正在启动后端服务...
    start "星云盘后端" cmd /k "title 星云盘后端服务 && cd /d %~dp0.. && go run cmd/server/main.go"
    echo ✅ 后端服务已启动,等待5秒...
    timeout /t 5 /nobreak >nul
) else (
    echo ✅ 后端服务正常运行
)
echo.

echo [步骤 2/5] 测试后端 API...
curl -s -X POST http://localhost:8080/api/v1/user/login ^
    -H "Content-Type: application/json" ^
    -d "{\"username\":\"admin2026\",\"password\":\"Admin123456\"}" >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 后端 API 调用失败
    echo.
    echo 可能原因:
    echo   1. 用户不存在
    echo   2. 数据库连接问题
    echo.
    echo 请运行以下命令创建测试用户:
    echo   start scripts\create-test-user.html
    pause
    exit /b 1
) else (
    echo ✅ 后端 API 正常
)
echo.

echo [步骤 3/5] 检查前端服务...
curl -s http://localhost:3000 >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 前端服务未运行
    echo.
    echo 正在启动前端服务...
    cd /d %~dp0..\frontend
    if not exist node_modules (
        echo 首次运行,正在安装依赖...
        call npm install
    )
    start "星云盘前端" cmd /k "title 星云盘前端服务 && npm run dev"
    echo ✅ 前端服务已启动,等待3秒...
    timeout /t 3 /nobreak >nul
    cd /d %~dp0
) else (
    echo ✅ 前端服务正常运行
)
echo.

echo [步骤 4/5] 检查前端配置...
if not exist "%~dp0..\frontend\.env.development" (
    echo ❌ 前端环境配置文件不存在
    echo.
    echo 正在创建配置文件...
    echo VITE_API_BASE_URL=http://localhost:8080 > "%~dp0..\frontend\.env.development"
    echo ✅ 配置文件已创建
    echo.
    echo ⚠️ 请重启前端服务使配置生效
    pause
) else (
    findstr /C:"VITE_API_BASE_URL" "%~dp0..\frontend\.env.development" >nul
    if %errorlevel% neq 0 (
        echo ❌ 前端配置文件缺少 API 地址配置
        echo.
        echo 正在修复配置...
        echo VITE_API_BASE_URL=http://localhost:8080 >> "%~dp0..\frontend\.env.development"
        echo ✅ 配置已修复
        echo.
        echo ⚠️ 请重启前端服务使配置生效
        pause
    ) else (
        echo ✅ 前端配置正确
    )
)
echo.

echo [步骤 5/5] 打开诊断工具...
echo.
echo ========================================
echo ✅ 基础检查完成!
echo ========================================
echo.
echo 📍 服务地址:
echo   • 后端: http://localhost:8080
echo   • 前端: http://localhost:3000
echo   • 登录: http://localhost:3000/#/login
echo.
echo 🔧 诊断工具:
echo   • 快速测试: scripts\quick-login-test.html
echo   • 完整诊断: scripts\diagnose-login-complete.html
echo.
echo 💡 下一步操作:
echo   1. 打开浏览器访问: http://localhost:3000/#/login
echo   2. 按 F12 打开开发者工具
echo   3. 切换到 Console 标签
echo   4. 输入用户名: admin2026
echo   5. 输入密码: Admin123456
echo   6. 点击登录按钮
echo   7. 查看控制台是否有错误信息
echo.
echo 如果仍然无法登录,请:
echo   1. 运行: start scripts\quick-login-test.html
echo   2. 点击"测试登录"按钮查看详细错误
echo.

choice /C YN /M "是否打开快速测试工具"
if errorlevel 2 goto :end

start scripts\quick-login-test.html

:end
echo.
pause
