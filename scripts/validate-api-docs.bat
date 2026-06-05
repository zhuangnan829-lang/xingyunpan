@echo off
REM API 文档验证脚本（Windows 版本）
REM 用于验证 Swagger/OpenAPI 文档的有效性

echo ==========================================
echo API 文档验证脚本
echo ==========================================
echo.

REM 检查文件是否存在
echo 1. 检查文档文件...
if exist "docs\api-swagger.yaml" (
    echo    ✓ docs\api-swagger.yaml 存在
) else (
    echo    × docs\api-swagger.yaml 不存在
    exit /b 1
)

if exist "docs\api-swagger.json" (
    echo    ✓ docs\api-swagger.json 存在
) else (
    echo    × docs\api-swagger.json 不存在
    exit /b 1
)

if exist "docs\API-DOCUMENTATION.md" (
    echo    ✓ docs\API-DOCUMENTATION.md 存在
) else (
    echo    × docs\API-DOCUMENTATION.md 不存在
    exit /b 1
)

echo.

REM 检查文件大小
echo 2. 检查文档大小...
for %%A in ("docs\api-swagger.yaml") do echo    - api-swagger.yaml: %%~zA bytes
for %%A in ("docs\api-swagger.json") do echo    - api-swagger.json: %%~zA bytes
for %%A in ("docs\API-DOCUMENTATION.md") do echo    - API-DOCUMENTATION.md: %%~zA bytes

for %%A in ("docs\api-swagger.yaml") do (
    if %%~zA GTR 10000 (
        echo    ✓ YAML 文档大小正常
    ) else (
        echo    ! YAML 文档可能不完整
    )
)

for %%A in ("docs\api-swagger.json") do (
    if %%~zA GTR 5000 (
        echo    ✓ JSON 文档大小正常
    ) else (
        echo    ! JSON 文档可能不完整
    )
)

echo.

REM 检查关键字段
echo 3. 检查关键字段...
findstr /C:"openapi: 3.0.3" docs\api-swagger.yaml >nul
if %errorlevel% equ 0 (
    echo    ✓ OpenAPI 版本正确
) else (
    echo    × OpenAPI 版本不正确
)

findstr /C:"title: 星云盘 V2 API" docs\api-swagger.yaml >nul
if %errorlevel% equ 0 (
    echo    ✓ API 标题正确
) else (
    echo    × API 标题不正确
)

findstr /C:"/api/v1/auth/register" docs\api-swagger.yaml >nul
if %errorlevel% equ 0 (
    echo    ✓ 包含注册接口
) else (
    echo    × 缺少注册接口
)

findstr /C:"/api/v1/file/upload" docs\api-swagger.yaml >nul
if %errorlevel% equ 0 (
    echo    ✓ 包含文件上传接口
) else (
    echo    × 缺少文件上传接口
)

echo.

REM 统计接口数量
echo 4. 统计接口数量...
for /f %%i in ('findstr /C:"operationId:" docs\api-swagger.yaml ^| find /c /v ""') do set api_count=%%i
echo    - 总接口数: %api_count%

if %api_count% GEQ 13 (
    echo    ✓ 接口数量正常（预期 13 个）
) else (
    echo    ! 接口数量可能不完整（预期 13 个）
)

echo.

REM 在线验证建议
echo 5. 在线验证建议...
echo    可以使用以下工具在线验证 API 文档:
echo    - Swagger Editor: https://editor.swagger.io/
echo    - Swagger Validator: https://validator.swagger.io/validator/debug
echo.

echo ==========================================
echo ✓ API 文档验证完成!
echo ==========================================
echo.
echo 查看文档:
echo   - YAML: docs\api-swagger.yaml
echo   - JSON: docs\api-swagger.json
echo   - 使用指南: docs\API-DOCUMENTATION.md
echo.
echo 在线查看:
echo   1. 访问 https://editor.swagger.io/
echo   2. 导入 docs\api-swagger.yaml
echo   3. 即可查看和测试 API
echo.

pause
