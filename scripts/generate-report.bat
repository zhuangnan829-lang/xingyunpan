@echo off
REM 生成性能压测报告

setlocal enabledelayedexpansion

set RESULTS_DIR=benchmark-results
set DOCS_DIR=docs
set TIMESTAMP=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%
set REPORT_FILE=%DOCS_DIR%\performance-report-%TIMESTAMP%.md

echo ========================================
echo 生成性能压测报告
echo ========================================
echo.

REM 检查结果目录
if not exist %RESULTS_DIR% (
    echo [错误] 结果目录不存在: %RESULTS_DIR%
    echo 请先运行性能压测: scripts\benchmark.bat
    pause
    exit /b 1
)

REM 查找最新的结果文件
for /f "delims=" %%i in ('dir /b /o-d %RESULTS_DIR%\benchmark_*.txt 2^>nul') do (
    set LATEST_FILE=%RESULTS_DIR%\%%i
    goto :found
)

:found
if not defined LATEST_FILE (
    echo [错误] 未找到压测结果文件
    pause
    exit /b 1
)

echo 源文件: %LATEST_FILE%
echo 报告文件: %REPORT_FILE%
echo.

REM 创建报告文件
echo # 星云盘 V2 性能压测报告 > %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ## 报告信息 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo - **测试日期**: %date% %time% >> %REPORT_FILE%
echo - **测试环境**: 本地开发环境 >> %REPORT_FILE%
echo - **服务器地址**: http://localhost:8080 >> %REPORT_FILE%
echo - **测试工具**: Apache Bench ^(ab^) >> %REPORT_FILE%
echo - **原始数据**: %LATEST_FILE% >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ## 执行摘要 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 本报告记录了星云盘 V2 系统的性能压测结果，包括核心 API 的吞吐量、响应时间和错误率等关键指标。 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ## 测试场景 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ^| 测试场景 ^| API 端点 ^| 请求数 ^| 并发数 ^| 方法 ^| >> %REPORT_FILE%
echo ^|---------|---------|--------|--------|------^| >> %REPORT_FILE%
echo ^| 健康检查 ^| GET /health ^| 1000 ^| 10 ^| GET ^| >> %REPORT_FILE%
echo ^| 用户注册 ^| POST /api/v1/user/register ^| 100 ^| 5 ^| POST ^| >> %REPORT_FILE%
echo ^| 用户登录 ^| POST /api/v1/user/login ^| 1000 ^| 10 ^| POST ^| >> %REPORT_FILE%
echo ^| 分片上传初始化 ^| POST /api/v1/files/multipart/init ^| 100 ^| 5 ^| POST ^| >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ## 性能指标 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ### 关键指标说明 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo - **吞吐量 ^(req/s^)**: 每秒处理的请求数，越高越好 >> %REPORT_FILE%
echo - **平均响应时间 ^(ms^)**: 单个请求的平均响应时间，越低越好 >> %REPORT_FILE%
echo - **P95 响应时间 ^(ms^)**: 95%% 的请求在此时间内完成 >> %REPORT_FILE%
echo - **P99 响应时间 ^(ms^)**: 99%% 的请求在此时间内完成 >> %REPORT_FILE%
echo - **错误率 ^(%%^)**: 失败请求占总请求的比例，越低越好 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ### 测试结果 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 详细的测试结果请查看原始数据文件: %LATEST_FILE% >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo #### 1. 健康检查 API >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ```text >> %REPORT_FILE%
findstr /C:"测试: 健康检查 API" /C:"URL:" /C:"方法:" /C:"请求数:" /C:"并发数:" /C:"Requests per second:" /C:"Time per request:" /C:"Failed requests:" /C:"Percentage of the requests served within a certain time" %LATEST_FILE% | findstr /V "========" >> %REPORT_FILE% 2>nul
echo ``` >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo #### 2. 用户注册 API >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ```text >> %REPORT_FILE%
findstr /C:"测试: 用户注册 API" /C:"URL:" /C:"方法:" /C:"请求数:" /C:"并发数:" /C:"Requests per second:" /C:"Time per request:" /C:"Failed requests:" /C:"Percentage of the requests served within a certain time" %LATEST_FILE% | findstr /V "========" >> %REPORT_FILE% 2>nul
echo ``` >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo #### 3. 用户登录 API >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ```text >> %REPORT_FILE%
findstr /C:"测试: 用户登录 API" /C:"URL:" /C:"方法:" /C:"请求数:" /C:"并发数:" /C:"Requests per second:" /C:"Time per request:" /C:"Failed requests:" /C:"Percentage of the requests served within a certain time" %LATEST_FILE% | findstr /V "========" >> %REPORT_FILE% 2>nul
echo ``` >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo #### 4. 分片上传初始化 API >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ```text >> %REPORT_FILE%
findstr /C:"测试: 分片上传初始化 API" /C:"URL:" /C:"方法:" /C:"请求数:" /C:"并发数:" /C:"Requests per second:" /C:"Time per request:" /C:"Failed requests:" /C:"Percentage of the requests served within a certain time" %LATEST_FILE% | findstr /V "========" >> %REPORT_FILE% 2>nul
echo ``` >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ## 性能分析 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ### 性能瓶颈 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 根据测试结果，识别出以下性能瓶颈: >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 1. **待分析**: 请根据实际测试结果填写 >> %REPORT_FILE%
echo 2. **待分析**: 请根据实际测试结果填写 >> %REPORT_FILE%
echo 3. **待分析**: 请根据实际测试结果填写 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ### 优化建议 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo #### 短期优化 ^(1-2 天^) >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 1. **添加数据库索引** >> %REPORT_FILE%
echo    - user_files 表: ^(user_id, folder_id, deleted_at^) >> %REPORT_FILE%
echo    - physical_files 表: ^(file_hash^) >> %REPORT_FILE%
echo    - multipart_uploads 表: ^(status, created_at^) >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 2. **启用 Redis 缓存** >> %REPORT_FILE%
echo    - 缓存用户信息 ^(5 分钟 TTL^) >> %REPORT_FILE%
echo    - 缓存文件列表 ^(1 分钟 TTL^) >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 3. **调整连接池** >> %REPORT_FILE%
echo    - 数据库: max_open_conns = 100 >> %REPORT_FILE%
echo    - Redis: pool_size = 10 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo #### 中期优化 ^(1 周^) >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 1. **查询优化** >> %REPORT_FILE%
echo    - 减少 N+1 查询 >> %REPORT_FILE%
echo    - 使用 JOIN 替代多次查询 >> %REPORT_FILE%
echo    - 使用 GORM Preload >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 2. **缓存策略** >> %REPORT_FILE%
echo    - 实现多级缓存 >> %REPORT_FILE%
echo    - 缓存预热 >> %REPORT_FILE%
echo    - 缓存失效策略 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ## 结论 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ### 性能评估 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo - **整体性能**: 待评估 >> %REPORT_FILE%
echo - **瓶颈识别**: 待分析 >> %REPORT_FILE%
echo - **优化效果**: 待验证 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ### 下一步行动 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 1. **立即执行** >> %REPORT_FILE%
echo    - [ ] 分析测试结果 >> %REPORT_FILE%
echo    - [ ] 识别性能瓶颈 >> %REPORT_FILE%
echo    - [ ] 制定优化计划 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 2. **短期计划** >> %REPORT_FILE%
echo    - [ ] 实施数据库索引优化 >> %REPORT_FILE%
echo    - [ ] 启用 Redis 缓存 >> %REPORT_FILE%
echo    - [ ] 调整连接池配置 >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ## 附录 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ### 完整测试数据 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo 完整的测试数据请查看: %LATEST_FILE% >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo ### 参考资料 >> %REPORT_FILE%
echo. >> %REPORT_FILE%
echo - [Apache Bench 文档](https://httpd.apache.org/docs/2.4/programs/ab.html) >> %REPORT_FILE%
echo - [Go 性能优化](https://github.com/dgryski/go-perfbook) >> %REPORT_FILE%
echo - [MySQL 性能优化](https://dev.mysql.com/doc/refman/8.0/en/optimization.html) >> %REPORT_FILE%
echo - [Redis 性能优化](https://redis.io/docs/management/optimization/) >> %REPORT_FILE%
echo. >> %REPORT_FILE%

echo ---
echo.
echo [成功] 报告生成完成: %REPORT_FILE%
echo.
echo 查看报告:
echo   type %REPORT_FILE%
echo.
pause
