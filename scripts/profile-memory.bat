@echo off
REM 内存性能分析脚本 (Windows)
REM 使用方法: scripts\profile-memory.bat

echo === 内存性能分析 ===
echo.

REM 检查 pprof 端点是否可用
curl -s http://localhost:6060/debug/pprof/ >nul 2>&1
if errorlevel 1 (
    echo 错误: pprof 端点不可用，请确保服务器运行在开发模式
    exit /b 1
)

echo 1. 采集 Heap profile...
curl -o heap.prof http://localhost:6060/debug/pprof/heap

echo.
echo 2. 采集 Allocs profile...
curl -o allocs.prof http://localhost:6060/debug/pprof/allocs

echo.
echo 3. 分析 Heap profile...
go tool pprof -http=:8082 heap.prof

echo.
echo === 内存性能分析完成 ===
echo Profile 文件: heap.prof, allocs.prof
echo 可以使用以下命令查看:
echo   go tool pprof heap.prof
echo   go tool pprof -http=:8082 heap.prof
echo   go tool pprof allocs.prof
