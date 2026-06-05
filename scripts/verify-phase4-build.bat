@echo off
echo ========================================
echo Phase 4 构建验证脚本
echo ========================================
echo.

echo [1/4] 清理旧的构建文件...
if exist server.exe del server.exe
if exist worker.exe del worker.exe
echo 完成
echo.

echo [2/4] 下载依赖...
go mod download
if %errorlevel% neq 0 (
    echo 依赖下载失败
    exit /b 1
)
echo 完成
echo.

echo [3/4] 构建服务器...
go build -o server.exe cmd/server/main.go
if %errorlevel% neq 0 (
    echo 服务器构建失败
    exit /b 1
)
echo 完成
echo.

echo [4/4] 构建 Worker...
go build -o worker.exe cmd/worker/main.go
if %errorlevel% neq 0 (
    echo Worker 构建失败
    exit /b 1
)
echo 完成
echo.

echo ========================================
echo 构建成功！
echo ========================================
echo.
echo 生成的文件:
echo   - server.exe (服务器)
echo   - worker.exe (后台任务)
echo.
echo 运行服务器: server.exe
echo 运行 Worker: worker.exe
echo.
