@echo off
REM Phase 5 数据库架构验证脚本
echo === Phase 5 数据库架构验证 ===
echo.

echo 正在运行数据库迁移...
go run scripts/migrate.go
if %ERRORLEVEL% NEQ 0 (
    echo 迁移失败！
    exit /b 1
)

echo.
echo === 验证完成 ===
echo 所有 Phase 5 表已成功创建
