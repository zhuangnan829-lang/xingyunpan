@echo off
REM 路径: scripts/verify.bat
REM Phase 1 快速验证脚本 (Windows)

echo === 星云盘 V2 Phase 1 验证脚本 ===
echo.

REM 1. 检查 Docker 服务
echo 1. 检查 Docker 服务...
docker-compose ps >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Docker Compose 可用
) else (
    echo [FAIL] Docker Compose 不可用
)

REM 2. 检查 MySQL
echo.
echo 2. 检查 MySQL...
docker exec xingyunpan-mysql mysql -uroot -ppassword -e "SELECT 1" >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] MySQL 连接正常
) else (
    echo [FAIL] MySQL 连接失败
)

REM 3. 检查 Redis
echo.
echo 3. 检查 Redis...
docker exec xingyunpan-redis redis-cli PING >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Redis 连接正常
) else (
    echo [FAIL] Redis 连接失败
)

REM 4. 运行单元测试
echo.
echo 4. 运行单元测试...
echo    - 测试日志工具...
go test ./pkg/logger/ -run TestInit >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 日志工具测试
) else (
    echo [FAIL] 日志工具测试
)

echo    - 测试响应格式...
go test ./pkg/response/ -run TestSuccess >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 响应格式测试
) else (
    echo [FAIL] 响应格式测试
)

echo    - 测试配置管理...
go test ./internal/config/ -run TestLoadConfig >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 配置管理测试
) else (
    echo [FAIL] 配置管理测试
)

echo    - 测试数据模型...
go test ./internal/model/ -run TestBaseModel >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 数据模型测试
) else (
    echo [FAIL] 数据模型测试
)

REM 5. 检查数据库表
echo.
echo 5. 检查数据库表...
docker exec xingyunpan-mysql mysql -uroot -ppassword xingyunpan -e "SHOW TABLES;" 2>nul | findstr /C:"users" >nul
if %errorlevel% equ 0 (
    echo [OK] 数据库表已创建
) else (
    echo [WARN] 数据库表未创建
    echo    运行迁移: go run scripts/migrate.go
)

REM 6. 检查服务器
echo.
echo 6. 检查服务器...
curl -s http://localhost:8080/ping >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 服务器正在运行
    
    REM 测试健康检查
    curl -s http://localhost:8080/health | findstr /C:"200" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [OK] 健康检查通过
    ) else (
        echo [FAIL] 健康检查失败
    )
) else (
    echo [WARN] 服务器未运行
    echo    启动服务器: go run cmd/server/main.go
)

echo.
echo === 验证完成 ===
echo.
echo 详细测试文档: docs/phase1-integration-test.md
pause
