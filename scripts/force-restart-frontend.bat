@echo off
chcp 65001 >nul
echo ========================================
echo 强制清理并重启前端服务
echo ========================================
echo.

echo [步骤 1/6] 停止所有 Node.js 进程...
taskkill /F /IM node.exe >nul 2>&1
timeout /t 2 /nobreak >nul
echo ✅ Node.js 进程已停止
echo.

echo [步骤 2/6] 清理前端缓存和构建文件...
cd /d %~dp0..\frontend

if exist node_modules\.vite (
    rmdir /s /q node_modules\.vite
    echo ✅ Vite 缓存已清理
)

if exist node_modules\.cache (
    rmdir /s /q node_modules\.cache
    echo ✅ Node 缓存已清理
)

if exist dist (
    rmdir /s /q dist
    echo ✅ 构建目录已清理
)

if exist .vite (
    rmdir /s /q .vite
    echo ✅ .vite 目录已清理
)
echo.

echo [步骤 3/6] 清理浏览器缓存提示...
echo ⚠️  请手动清理浏览器缓存:
echo    方法1: 按 Ctrl+Shift+Delete 打开清理窗口
echo    方法2: 按 Ctrl+F5 强制刷新页面
echo    方法3: 在开发者工具中右键刷新按钮,选择"清空缓存并硬性重新加载"
echo.
timeout /t 3 /nobreak >nul

echo [步骤 4/6] 检查环境配置...
if not exist .env.development (
    echo 创建 .env.development 文件...
    echo VITE_API_BASE_URL=http://localhost:8080 > .env.development
    echo ✅ 配置文件已创建
) else (
    echo ✅ 配置文件存在
)
echo.

echo [步骤 5/6] 验证代码修复...
findstr /C:"import { getToken, removeToken } from '@/utils/auth'" src\api\request.ts >nul
if %errorlevel% equ 0 (
    echo ✅ request.ts 导入语句正确
) else (
    echo ❌ request.ts 导入语句有问题
    echo 请检查 frontend/src/api/request.ts 文件
    pause
    exit /b 1
)
echo.

echo [步骤 6/6] 启动前端服务...
start "星云盘前端" cmd /k "title 星云盘前端服务 && npm run dev"
echo ✅ 前端服务已启动
echo.

echo ========================================
echo ✅ 强制重启完成!
echo ========================================
echo.
echo 📍 前端地址: http://localhost:3000
echo 📍 登录页面: http://localhost:3000/#/login
echo.
echo 💡 重要提示:
echo    1. 等待 5-10 秒让前端服务完全启动
echo    2. 在浏览器中按 Ctrl+Shift+Delete 清除缓存
echo    3. 或者按 Ctrl+F5 强制刷新页面
echo    4. 然后尝试登录
echo.
echo 测试账号:
echo    用户名: admin2026
echo    密码: Admin123456
echo.

timeout /t 5 /nobreak
pause
