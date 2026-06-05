@echo off
chcp 65001 >nul
echo ========================================
echo 修复并重启前端服务
echo ========================================
echo.

echo [步骤 1/4] 停止现有前端进程...
taskkill /F /FI "WINDOWTITLE eq 星云盘前端*" >nul 2>&1
taskkill /F /IM node.exe /FI "WINDOWTITLE eq *npm*" >nul 2>&1
timeout /t 2 /nobreak >nul
echo ✅ 前端进程已停止
echo.

echo [步骤 2/4] 清理前端缓存...
cd /d %~dp0..\frontend
if exist node_modules\.vite (
    rmdir /s /q node_modules\.vite
    echo ✅ Vite 缓存已清理
)
if exist dist (
    rmdir /s /q dist
    echo ✅ 构建目录已清理
)
echo.

echo [步骤 3/4] 检查环境配置...
if not exist .env.development (
    echo 创建 .env.development 文件...
    echo VITE_API_BASE_URL=http://localhost:8080 > .env.development
    echo ✅ 配置文件已创建
) else (
    echo ✅ 配置文件存在
)
echo.

echo [步骤 4/4] 启动前端服务...
start "星云盘前端" cmd /k "title 星云盘前端服务 && npm run dev"
echo ✅ 前端服务已启动
echo.

echo ========================================
echo ✅ 修复完成!
echo ========================================
echo.
echo 📍 前端地址: http://localhost:3000
echo 📍 登录页面: http://localhost:3000/#/login
echo.
echo 💡 请等待 3-5 秒让前端服务完全启动
echo 💡 然后刷新浏览器页面
echo.
echo 如果问题仍然存在:
echo   1. 按 F12 打开浏览器开发者工具
echo   2. 切换到 Console 标签
echo   3. 清除所有错误 (点击 🚫 图标)
echo   4. 刷新页面 (Ctrl+F5 强制刷新)
echo   5. 点击登录按钮
echo   6. 查看是否还有错误
echo.

timeout /t 5 /nobreak
pause
