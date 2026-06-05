@echo off
REM 生成 Swagger 文档脚本 (Windows)

echo ========================================
echo 星云盘 V2 - Swagger 文档生成
echo ========================================
echo.

REM 检查 swag 是否已安装
where swag >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] swag 命令未找到
    echo.
    echo 请先安装 swag CLI:
    echo   go install github.com/swaggo/swag/cmd/swag@latest
    echo.
    echo 然后确保 GOPATH\bin 在 PATH 中:
    echo   set PATH=%%PATH%%;%%GOPATH%%\bin
    echo.
    exit /b 1
)

echo [1/3] 检查 swag 版本...
swag --version
echo.

echo [2/3] 生成 Swagger 文档...
swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Swagger 文档生成失败
    exit /b 1
)
echo.

echo [3/3] 验证生成的文件...
if exist docs\swagger\docs.go (
    echo   ✓ docs/swagger/docs.go
) else (
    echo   ✗ docs/swagger/docs.go 未生成
)

if exist docs\swagger\swagger.json (
    echo   ✓ docs/swagger/swagger.json
) else (
    echo   ✗ docs/swagger/swagger.json 未生成
)

if exist docs\swagger\swagger.yaml (
    echo   ✓ docs/swagger/swagger.yaml
) else (
    echo   ✗ docs/swagger/swagger.yaml 未生成
)
echo.

echo ========================================
echo Swagger 文档生成完成！
echo ========================================
echo.
echo 启动服务器后访问:
echo   http://localhost:8080/swagger/index.html
echo.
