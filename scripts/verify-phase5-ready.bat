@echo off
REM Phase 5 前后端联调准备检查脚本

echo ========================================
echo Phase 5 前后端联调准备检查
echo ========================================
echo.

set ERRORS=0

REM 1. 检查后端服务器
echo [1/6] 检查后端服务器...
curl -s http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo ❌ 后端服务器未运行
    echo    请运行: go run cmd/server/main.go
    set /a ERRORS+=1
) else (
    echo ✅ 后端服务器运行中
)
echo.

REM 2. 检查 Redis
echo [2/6] 检查 Redis 服务...
redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo ❌ Redis 未运行
    echo    请启动 Redis 服务
    set /a ERRORS+=1
) else (
    echo ✅ Redis 运行中
)
echo.

REM 3. 检查数据库连接
echo [3/6] 检查数据库连接...
go run scripts/test-db-connection.go >nul 2>&1
if errorlevel 1 (
    echo ❌ 数据库连接失败
    echo    请检查数据库配置
    set /a ERRORS+=1
) else (
    echo ✅ 数据库连接正常
)
echo.

REM 4. 检查 Phase 5 路由
echo [4/6] 检查 Phase 5 路由注册...
curl -s http://localhost:8080/api/v1/ >nul 2>&1
if errorlevel 1 (
    echo ❌ API 路由不可访问
    set /a ERRORS+=1
) else (
    echo ✅ API 路由可访问
)
echo.

REM 5. 检查 Swagger 文档
echo [5/6] 检查 Swagger 文档...
curl -s http://localhost:8080/swagger/index.html >nul 2>&1
if errorlevel 1 (
    echo ⚠️  Swagger 文档不可访问（非阻塞）
) else (
    echo ✅ Swagger 文档可访问
)
echo.

REM 6. 检查前端服务器
echo [6/6] 检查前端服务器...
curl -s http://localhost:5173 >nul 2>&1
if errorlevel 1 (
    echo ⚠️  前端服务器未运行（可选）
    echo    如需测试前端，请运行: cd frontend ^&^& npm run dev
) else (
    echo ✅ 前端服务器运行中
)
echo.

REM 总结
echo ========================================
if %ERRORS% EQU 0 (
    echo ✅ 所有检查通过，可以开始联调
    echo.
    echo 下一步:
    echo 1. 运行集成测试: scripts\test-phase5-integration.bat
    echo 2. 打开前端页面: http://localhost:5173
    echo 3. 查看 API 文档: http://localhost:8080/swagger/index.html
) else (
    echo ❌ 发现 %ERRORS% 个问题，请先解决
    exit /b 1
)
echo ========================================
