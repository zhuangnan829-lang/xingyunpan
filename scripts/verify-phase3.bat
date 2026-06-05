@echo off
REM 路径: scripts/verify-phase3.bat
REM Phase 3 验证脚本 - 检查所有修改是否正确

echo ========================================
echo Phase 3 验证脚本
echo ========================================
echo.

echo [检查 1/5] 验证代码编译...
go build ./...
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 代码编译失败
    exit /b 1
)
echo ✅ 代码编译成功
echo.

echo [检查 2/5] 验证 Service 层测试...
go test ./internal/service -v -count=1
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Service 层测试失败
    exit /b 1
)
echo ✅ Service 层测试通过
echo.

echo [检查 3/5] 验证 Repository 层测试...
go test ./internal/repository -v -count=1
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Repository 层测试失败
    exit /b 1
)
echo ✅ Repository 层测试通过
echo.

echo [检查 4/5] 验证 Controller 层测试...
go test ./internal/controller -v -count=1
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Controller 层测试失败
    exit /b 1
)
echo ✅ Controller 层测试通过
echo.

echo [检查 5/5] 验证其他组件测试...
go test ./pkg/... -v -count=1
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 其他组件测试失败
    exit /b 1
)
echo ✅ 其他组件测试通过
echo.

echo ========================================
echo ✅ Phase 3 所有验证通过!
echo ========================================
echo.
echo 下一步:
echo 1. 检查数据库迁移(复合索引)
echo 2. 更新 cmd/server/main.go 集成所有修改
echo 3. 运行集成测试
echo.
