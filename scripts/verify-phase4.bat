@echo off
REM 路径: scripts/verify-phase4.bat
REM Phase 4 验证脚本

echo ========================================
echo Phase 4 验证脚本
echo ========================================
echo.

echo [1/4] 检查代码编译...
go build -o temp_server.exe cmd/server/main.go
if %errorlevel% neq 0 (
    echo ✗ 服务器编译失败
    exit /b 1
)
echo ✓ 服务器编译成功

go build -o temp_worker.exe cmd/worker/main.go
if %errorlevel% neq 0 (
    echo ✗ Worker 编译失败
    del temp_server.exe
    exit /b 1
)
echo ✓ Worker 编译成功
echo.

echo [2/4] 检查数据库迁移脚本...
go run scripts/migrate.go
if %errorlevel% neq 0 (
    echo ✗ 数据库迁移失败
    del temp_server.exe temp_worker.exe
    exit /b 1
)
echo ✓ 数据库迁移成功
echo.

echo [3/4] 检查关键文件存在...
set FILES_OK=1

if not exist "internal\model\multipart_upload.go" (
    echo ✗ multipart_upload.go 不存在
    set FILES_OK=0
)

if not exist "internal\model\file_deletion.go" (
    echo ✗ file_deletion.go 不存在
    set FILES_OK=0
)

if not exist "internal\repository\multipart_upload_repository.go" (
    echo ✗ multipart_upload_repository.go 不存在
    set FILES_OK=0
)

if not exist "internal\repository\file_deletion_repository.go" (
    echo ✗ file_deletion_repository.go 不存在
    set FILES_OK=0
)

if not exist "internal\service\multipart_service.go" (
    echo ✗ multipart_service.go 不存在
    set FILES_OK=0
)

if not exist "internal\controller\multipart_controller.go" (
    echo ✗ multipart_controller.go 不存在
    set FILES_OK=0
)

if not exist "pkg\redis\lock.go" (
    echo ✗ lock.go 不存在
    set FILES_OK=0
)

if not exist "pkg\redis\multipart.go" (
    echo ✗ multipart.go 不存在
    set FILES_OK=0
)

if not exist "internal\worker\file_deletion_worker.go" (
    echo ✗ file_deletion_worker.go 不存在
    set FILES_OK=0
)

if not exist "internal\worker\multipart_cleanup_worker.go" (
    echo ✗ multipart_cleanup_worker.go 不存在
    set FILES_OK=0
)

if %FILES_OK%==0 (
    echo ✗ 部分文件缺失
    del temp_server.exe temp_worker.exe
    exit /b 1
)
echo ✓ 所有关键文件存在
echo.

echo [4/4] 清理临时文件...
del temp_server.exe temp_worker.exe
echo ✓ 清理完成
echo.

echo ========================================
echo Phase 4 验证通过！
echo ========================================
echo.
echo 下一步:
echo 1. 启动服务器: go run cmd/server/main.go
echo 2. 启动 Worker: go run cmd/worker/main.go
echo 3. 测试分片上传 API
echo.

exit /b 0
