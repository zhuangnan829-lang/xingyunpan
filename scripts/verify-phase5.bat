@echo off
REM Phase 5 部署检查脚本 (Windows)
REM 验证 Phase 5 所有组件是否正确部署

setlocal enabledelayedexpansion

echo ==========================================
echo Phase 5 部署检查开始
echo ==========================================
echo.

set PASSED=0
set FAILED=0

REM 1. 检查配置文件
echo 1. 检查配置文件...
echo -------------------------------------------

if exist "configs\config.yaml" (
    echo [PASS] 配置文件存在
    set /a PASSED+=1
) else if exist "configs\config.prod.yaml" (
    echo [PASS] 生产配置文件存在
    set /a PASSED+=1
) else (
    echo [FAIL] 配置文件不存在
    set /a FAILED+=1
)

REM 2. 检查必需的环境变量
echo.
echo 2. 检查环境变量...
echo -------------------------------------------

if defined DB_PASSWORD (
    echo [PASS] DB_PASSWORD 已设置
    set /a PASSED+=1
) else (
    echo [FAIL] DB_PASSWORD 未设置
    set /a FAILED+=1
)

if defined JWT_SECRET (
    echo [PASS] JWT_SECRET 已设置
    set /a PASSED+=1
) else (
    echo [FAIL] JWT_SECRET 未设置
    set /a FAILED+=1
)

REM 3. 检查服务是否运行
echo.
echo 3. 检查服务状态...
echo -------------------------------------------

set BASE_URL=http://localhost:8080
if defined BASE_URL_ENV (
    set BASE_URL=%BASE_URL_ENV%
)

REM 使用 curl 检查健康端点
curl -s -o nul -w "%%{http_code}" "%BASE_URL%/health" > temp_health.txt 2>nul
set /p HEALTH_CODE=<temp_health.txt
del temp_health.txt 2>nul

if "%HEALTH_CODE%"=="200" (
    echo [PASS] 服务健康检查通过
    set /a PASSED+=1
) else (
    echo [FAIL] 服务健康检查失败 ^(HTTP %HEALTH_CODE%^)
    set /a FAILED+=1
    echo [WARN] 请确保服务正在运行
)

REM 4. 检查 Swagger 文档
echo.
echo 4. 检查 API 文档...
echo -------------------------------------------

curl -s -o nul -w "%%{http_code}" "%BASE_URL%/swagger/index.html" > temp_swagger.txt 2>nul
set /p SWAGGER_CODE=<temp_swagger.txt
del temp_swagger.txt 2>nul

if "%SWAGGER_CODE%"=="200" (
    echo [PASS] Swagger 文档可访问
    set /a PASSED+=1
) else (
    echo [WARN] Swagger 文档不可访问 ^(HTTP %SWAGGER_CODE%^)
)

REM 5. 检查 Prometheus metrics
echo.
echo 5. 检查监控指标...
echo -------------------------------------------

curl -s -o nul -w "%%{http_code}" "%BASE_URL%/metrics" > temp_metrics.txt 2>nul
set /p METRICS_CODE=<temp_metrics.txt
del temp_metrics.txt 2>nul

if "%METRICS_CODE%"=="200" (
    echo [PASS] Prometheus metrics 端点可访问
    set /a PASSED+=1
) else (
    echo [WARN] Prometheus metrics 端点不可访问 ^(HTTP %METRICS_CODE%^)
)

REM 6. 检查迁移脚本
echo.
echo 6. 检查迁移脚本...
echo -------------------------------------------

if exist "scripts\migrations\phase5_schema.sql" (
    echo [PASS] 迁移脚本存在
    set /a PASSED+=1
) else (
    echo [FAIL] 迁移脚本不存在
    set /a FAILED+=1
)

if exist "scripts\migrations\phase5_rollback.sql" (
    echo [PASS] 回滚脚本存在
    set /a PASSED+=1
) else (
    echo [FAIL] 回滚脚本不存在
    set /a FAILED+=1
)

REM 7. 总结
echo.
echo ==========================================
echo 检查完成
echo ==========================================
echo 通过: %PASSED%
echo 失败: %FAILED%
echo.

if %FAILED% EQU 0 (
    echo [SUCCESS] Phase 5 部署检查全部通过！
    exit /b 0
) else (
    echo [ERROR] Phase 5 部署检查发现 %FAILED% 个问题，请修复后重试
    exit /b 1
)
