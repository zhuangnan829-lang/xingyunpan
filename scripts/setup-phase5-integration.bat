@echo off
REM Phase 5 前后端联调环境准备脚本

echo ========================================
echo Phase 5 前后端联调环境准备
echo ========================================
echo.

REM 1. 安装 Swagger 依赖
echo [1/5] 安装 Swagger 依赖...
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
if errorlevel 1 (
    echo ❌ Swagger 依赖安装失败
    exit /b 1
)
echo ✅ Swagger 依赖安装完成
echo.

REM 2. 生成 Swagger 文档
echo [2/5] 生成 Swagger 文档...
swag init -g cmd/server/main.go -o docs/swagger
if errorlevel 1 (
    echo ⚠️  Swagger 文档生成失败（可能需要手动生成）
) else (
    echo ✅ Swagger 文档生成完成
)
echo.

REM 3. 执行数据库迁移
echo [3/5] 执行数据库迁移...
go run scripts/migrate.go
if errorlevel 1 (
    echo ❌ 数据库迁移失败
    exit /b 1
)
echo ✅ 数据库迁移完成
echo.

REM 4. 检查 Redis
echo [4/5] 检查 Redis 服务...
redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo ⚠️  Redis 未运行，请手动启动
) else (
    echo ✅ Redis 运行中
)
echo.

REM 5. 编译服务器
echo [5/5] 编译服务器...
go build -o bin/server.exe cmd/server/main.go
if errorlevel 1 (
    echo ❌ 服务器编译失败
    exit /b 1
)
echo ✅ 服务器编译完成
echo.

echo ========================================
echo ✅ 环境准备完成
echo.
echo 下一步:
echo 1. 启动服务器: bin\server.exe
echo 2. 启动前端: cd frontend ^&^& npm run dev
echo 3. 运行集成测试: scripts\test-phase5-integration.bat
echo 4. 查看 API 文档: http://localhost:8080/swagger/index.html
echo ========================================
