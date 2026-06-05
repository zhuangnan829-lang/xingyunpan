@echo off
REM 验证 Phase 5 API 文档完整性脚本 (Windows)

echo ========================================
echo Phase 5 API 文档验证
echo ========================================
echo.

set ERROR_COUNT=0

echo [1/5] 检查 OpenAPI 规范文件...
if exist docs\api-swagger.yaml (
    echo   ✓ docs/api-swagger.yaml 存在
) else (
    echo   ✗ docs/api-swagger.yaml 不存在
    set /a ERROR_COUNT+=1
)

if exist docs\api-swagger.json (
    echo   ✓ docs/api-swagger.json 存在
) else (
    echo   ✗ docs/api-swagger.json 不存在
    set /a ERROR_COUNT+=1
)
echo.

echo [2/5] 检查 Postman 集合...
if exist docs\postman\phase5-api.json (
    echo   ✓ docs/postman/phase5-api.json 存在
) else (
    echo   ✗ docs/postman/phase5-api.json 不存在
    set /a ERROR_COUNT+=1
)
echo.

echo [3/5] 检查文档指南...
if exist docs\phase5-api-documentation.md (
    echo   ✓ docs/phase5-api-documentation.md 存在
) else (
    echo   ✗ docs/phase5-api-documentation.md 不存在
    set /a ERROR_COUNT+=1
)

if exist docs\phase5-api-setup-guide.md (
    echo   ✓ docs/phase5-api-setup-guide.md 存在
) else (
    echo   ✗ docs/phase5-api-setup-guide.md 不存在
    set /a ERROR_COUNT+=1
)

if exist docs\generate-swagger.md (
    echo   ✓ docs/generate-swagger.md 存在
) else (
    echo   ✗ docs/generate-swagger.md 不存在
    set /a ERROR_COUNT+=1
)
echo.

echo [4/5] 检查 Swagger 注释...
findstr /C:"@Summary" internal\controller\share_controller.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ share_controller.go 包含 Swagger 注释
) else (
    echo   ✗ share_controller.go 缺少 Swagger 注释
    set /a ERROR_COUNT+=1
)

findstr /C:"@Summary" internal\controller\search_controller.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ search_controller.go 包含 Swagger 注释
) else (
    echo   ✗ search_controller.go 缺少 Swagger 注释
    set /a ERROR_COUNT+=1
)

findstr /C:"@Summary" internal\controller\recycle_controller.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ recycle_controller.go 包含 Swagger 注释
) else (
    echo   ✗ recycle_controller.go 缺少 Swagger 注释
    set /a ERROR_COUNT+=1
)

findstr /C:"@Summary" internal\controller\version_controller.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ version_controller.go 包含 Swagger 注释
) else (
    echo   ✗ version_controller.go 缺少 Swagger 注释
    set /a ERROR_COUNT+=1
)

findstr /C:"@Summary" internal\controller\collaboration_controller.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ collaboration_controller.go 包含 Swagger 注释
) else (
    echo   ✗ collaboration_controller.go 缺少 Swagger 注释
    set /a ERROR_COUNT+=1
)
echo.

echo [5/5] 检查 main.go Swagger 配置...
findstr /C:"@title" cmd\server\main.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ main.go 包含 Swagger API 信息
) else (
    echo   ✗ main.go 缺少 Swagger API 信息
    set /a ERROR_COUNT+=1
)

findstr /C:"ginSwagger" cmd\server\main.go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo   ✓ main.go 已注册 Swagger UI 路由
) else (
    echo   ✗ main.go 未注册 Swagger UI 路由
    set /a ERROR_COUNT+=1
)
echo.

echo ========================================
if %ERROR_COUNT% EQU 0 (
    echo ✓ 所有文档检查通过！
    echo.
    echo 下一步:
    echo   1. 运行 scripts\generate-swagger.bat 生成 Swagger UI
    echo   2. 启动服务器: go run cmd/server/main.go
    echo   3. 访问 http://localhost:8080/swagger/index.html
    echo.
    exit /b 0
) else (
    echo ✗ 发现 %ERROR_COUNT% 个问题
    echo.
    echo 请检查上述错误并修复
    echo.
    exit /b 1
)
