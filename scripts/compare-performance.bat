@echo off
REM 性能对比脚本 (Windows)
REM 使用方法: scripts\compare-performance.bat <before_file> <after_file>

if "%1"=="" (
    echo 使用方法: scripts\compare-performance.bat ^<before_file^> ^<after_file^>
    echo 示例: scripts\compare-performance.bat benchmark-before.txt benchmark-after.txt
    exit /b 1
)

if "%2"=="" (
    echo 使用方法: scripts\compare-performance.bat ^<before_file^> ^<after_file^>
    echo 示例: scripts\compare-performance.bat benchmark-before.txt benchmark-after.txt
    exit /b 1
)

set BEFORE_FILE=%1
set AFTER_FILE=%2

echo === 性能对比分析 ===
echo.

if not exist %BEFORE_FILE% (
    echo 错误: 文件不存在: %BEFORE_FILE%
    exit /b 1
)

if not exist %AFTER_FILE% (
    echo 错误: 文件不存在: %AFTER_FILE%
    exit /b 1
)

echo 优化前: %BEFORE_FILE%
echo 优化后: %AFTER_FILE%
echo.

echo 正在分析性能数据...
echo.

REM 这里可以添加更复杂的分析逻辑
echo === 对比结果 ===
echo.
echo 请手动对比以下指标:
echo   1. 平均响应时间 (Time per request)
echo   2. P95 响应时间
echo   3. P99 响应时间
echo   4. 吞吐量 (Requests per second)
echo   5. 错误率
echo.
echo 预期改善:
echo   - 响应时间减少 30%%+
echo   - 吞吐量提升 40%%+
echo   - 数据库查询时间减少 50%%+
echo.
