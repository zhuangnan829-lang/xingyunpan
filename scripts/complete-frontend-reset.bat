@echo off
chcp 65001 >nul
echo ========================================
echo 完全重置前端环境
echo ========================================
echo.
echo ⚠️  警告: 这将删除 node_modules 并重新安装所有依赖
echo ⚠️  这个过程可能需要 5-10 分钟
echo.
pause
echo.

echo [步骤 1/6] 停止所有前端进程...
taskkill /F /FI "WINDOWTITLE eq 星云盘前端*" >nul 2>&1
taskkill /F /IM node.exe >nul 2>&1
timeout /t 3 /nobreak >nul
echo ✅ 前端进程已停止
echo.

echo [步骤 2/6] 进入前端目录...
cd /d %~dp0..\frontend
echo ✅ 当前目录: %CD%
echo.

echo [步骤 3/6] 删除缓存和构建文件...
if exist node_modules\.vite (
    rmdir /s /q node_modules\.vite
    echo ✅ Vite 缓存已删除
)
if exist dist (
    rmdir /s /q dist
    echo ✅ 构建目录已删除
)
if exist .vite (
    rmdir /s /q .vite
    echo ✅ .vite 目录已删除
)
echo.

echo [步骤 4/6] 删除 node_modules...
if exist node_modules (
    echo 正在删除 node_modules (这可能需要几分钟)...
    rmdir /s /q node_modules
    echo ✅ node_modules 已删除
) else (
    echo ℹ️  node_modules 不存在
)
echo.

echo [步骤 5/6] 重新安装依赖...
echo 正在运行 npm install (这可能需要 5-10 分钟)...
call npm install
if errorlevel 1 (
    echo ❌ npm install 失败
    echo.
    echo 请检查:
    echo   1. 是否安装了 Node.js
    echo   2. 网络连接是否正常
    echo   3. npm 配置是否正确
    echo.
    pause
    exit /b 1
)
echo ✅ 依赖安装完成
echo.

echo [步骤 6/6] 启动前端服务...
start "星云盘前端" cmd /k "title 星云盘前端服务 && npm run dev"
echo ✅ 前端服务已启动
echo.

echo ========================================
echo ✅ 重置完成!
echo ========================================
echo.
echo 📍 前端地址: http://localhost:3000
echo 📍 登录页面: http://localhost:3000/#/login
echo.
echo 💡 请等待 5-10 秒让前端服务完全启动
echo 💡 然后在浏览器中:
echo    1. 按 Ctrl+Shift+Delete 打开清除浏览器数据
echo    2. 选择"缓存的图片和文件"
echo    3. 点击"清除数据"
echo    4. 访问 http://localhost:3000/#/login
echo    5. 按 F12 打开开发者工具查看是否还有错误
echo.

timeout /t 10 /nobreak
pause
