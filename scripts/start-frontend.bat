@echo off
echo ========================================
echo 启动星云盘前端服务
echo ========================================
echo.

cd frontend

echo 检查 node_modules 是否存在...
if not exist "node_modules" (
    echo node_modules 不存在，正在安装依赖...
    call npm install
    if errorlevel 1 (
        echo.
        echo ❌ 依赖安装失败！
        echo 请检查 Node.js 和 npm 是否正确安装
        pause
        exit /b 1
    )
    echo ✅ 依赖安装完成
    echo.
)

echo 启动前端开发服务器...
echo 访问地址: http://localhost:3000
echo.
echo 按 Ctrl+C 停止服务器
echo.

call npm run dev

pause
