@echo off
chcp 65001 >nul
echo ========================================
echo 诊断登录问题
echo ========================================
echo.

cd /d %~dp0..\frontend

echo [检查 1/5] 检查 Node.js 和 npm...
node --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Node.js 未安装或不在 PATH 中
    echo 请安装 Node.js: https://nodejs.org/
    pause
    exit /b 1
)
echo ✅ Node.js 版本:
node --version
npm --version
echo.

echo [检查 2/5] 检查前端目录结构...
if not exist "src\stores\user.ts" (
    echo ❌ 找不到 src\stores\user.ts
    pause
    exit /b 1
)
if not exist "src\utils\auth.ts" (
    echo ❌ 找不到 src\utils\auth.ts
    pause
    exit /b 1
)
echo ✅ 文件结构正常
echo.

echo [检查 3/5] 检查环境配置...
if not exist ".env.development" (
    echo ⚠️  .env.development 不存在，正在创建...
    echo VITE_API_BASE_URL=http://localhost:8080 > .env.development
    echo ✅ 已创建 .env.development
) else (
    echo ✅ .env.development 存在
    type .env.development
)
echo.

echo [检查 4/5] 检查 node_modules...
if not exist "node_modules" (
    echo ⚠️  node_modules 不存在
    echo 需要运行 npm install
    echo.
    set /p install="是否现在安装依赖? (y/n): "
    if /i "%install%"=="y" (
        echo 正在安装依赖...
        call npm install
    ) else (
        echo 请手动运行: npm install
        pause
        exit /b 1
    )
) else (
    echo ✅ node_modules 存在
)
echo.

echo [检查 5/5] 检查前端进程...
tasklist /FI "IMAGENAME eq node.exe" 2>NUL | find /I /N "node.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo ✅ 前端进程正在运行
    echo.
    echo 建议操作:
    echo   1. 停止当前前端进程
    echo   2. 清理缓存
    echo   3. 重新启动
    echo.
    set /p restart="是否现在重启前端? (y/n): "
    if /i "%restart%"=="y" (
        goto :restart_frontend
    )
) else (
    echo ⚠️  前端进程未运行
    echo.
    set /p start="是否现在启动前端? (y/n): "
    if /i "%start%"=="y" (
        goto :start_frontend
    )
)
echo.
echo ========================================
echo 诊断完成
echo ========================================
pause
exit /b 0

:restart_frontend
echo.
echo 正在重启前端...
taskkill /F /IM node.exe >nul 2>&1
timeout /t 2 /nobreak >nul

if exist node_modules\.vite (
    rmdir /s /q node_modules\.vite
    echo ✅ 已清理 Vite 缓存
)
if exist dist (
    rmdir /s /q dist
    echo ✅ 已清理构建目录
)

:start_frontend
echo.
echo 正在启动前端服务...
start "星云盘前端" cmd /k "title 星云盘前端服务 && npm run dev"
echo ✅ 前端服务已启动
echo.
echo 📍 前端地址: http://localhost:3000
echo 📍 登录页面: http://localhost:3000/#/login
echo.
echo 💡 重要提示:
echo    1. 等待 5-10 秒让服务启动
echo    2. 在浏览器中按 Ctrl+F5 强制刷新
echo    3. 如果还有问题，清除浏览器缓存:
echo       - 按 Ctrl+Shift+Delete
echo       - 选择"缓存的图片和文件"
echo       - 点击"清除数据"
echo    4. 按 F12 打开开发者工具查看错误
echo.
timeout /t 5 /nobreak
pause
exit /b 0
