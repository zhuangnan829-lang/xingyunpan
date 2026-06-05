-- 路径: scripts/analyze-query-performance.sql
-- Phase 5 查询性能分析脚本
-- 使用 EXPLAIN 分析查询计划，验证索引使用情况

-- ============================================================================
-- 1. 搜索查询分析
-- ============================================================================

-- 简单关键词搜索
EXPLAIN ANALYZE
SELECT * FROM user_files 
WHERE user_id = 1 
  AND deleted_at IS NULL
  AND file_name LIKE '%test%'
ORDER BY modified_at DESC
LIMIT 20;

-- 多条件过滤搜索
EXPLAIN ANALYZE
SELECT * FROM user_files 
WHERE user_id = 1 
  AND deleted_at IS NULL
  AND file_name LIKE '%document%'
  AND file_type = 'pdf'
  AND file_size >= 1024 AND file_size <= 10485760
  AND modified_at >= '2026-01-01'
ORDER BY modified_at DESC
LIMIT 20;

-- 文件夹范围搜索
EXPLAIN ANALYZE
SELECT * FROM user_files 
WHERE user_id = 1 
  AND folder_id = 'folder_123'
  AND deleted_at IS NULL
  AND file_name LIKE '%report%'
ORDER BY modified_at DESC
LIMIT 20;

-- ============================================================================
-- 2. 版本历史查询分析
-- ============================================================================

-- 获取文件版本历史
EXPLAIN ANALYZE
SELECT fv.*, u.username as uploader_name
FROM file_versions fv
LEFT JOIN users u ON fv.uploader_id = u.id
WHERE fv.file_id = 'file_123'
ORDER BY fv.version_number DESC;

-- 获取当前版本
EXPLAIN ANALYZE
SELECT * FROM file_versions
WHERE file_id = 'file_123' AND is_current = true
LIMIT 1;

-- 统计版本数量
EXPLAIN ANALYZE
SELECT COUNT(*) FROM file_versions
WHERE file_id = 'file_123';

-- ============================================================================
-- 3. 回收站查询分析
-- ============================================================================

-- 获取用户回收站列表
EXPLAIN ANALYZE
SELECT * FROM recycle_bin
WHERE user_id = 1
ORDER BY deleted_at DESC
LIMIT 20 OFFSET 0;

-- 获取过期项目（用于定时清理）
EXPLAIN ANALYZE
SELECT * FROM recycle_bin
WHERE expires_at < NOW()
LIMIT 100;

-- 深度分页查询
EXPLAIN ANALYZE
SELECT * FROM recycle_bin
WHERE user_id = 1
ORDER BY deleted_at DESC
LIMIT 20 OFFSET 500;

-- ============================================================================
-- 4. 分享查询分析
-- ============================================================================

-- 根据 share_token 查询分享
EXPLAIN ANALYZE
SELECT * FROM shares
WHERE share_token = 'abc123xyz789'
LIMIT 1;

-- 获取用户的所有分享
EXPLAIN ANALYZE
SELECT s.*, 
       (SELECT COUNT(*) FROM share_files WHERE share_id = s.id) as file_count
FROM shares s
WHERE s.user_id = 1
ORDER BY s.created_at DESC;

-- 获取分享的文件列表
EXPLAIN ANALYZE
SELECT uf.*
FROM share_files sf
JOIN user_files uf ON sf.file_id = uf.file_id
WHERE sf.share_id = 1;

-- ============================================================================
-- 5. 协作查询分析
-- ============================================================================

-- 获取文件的协作者列表
EXPLAIN ANALYZE
SELECT c.*, u.username
FROM collaborations c
JOIN users u ON c.collaborator_id = u.id
WHERE c.file_id = 'file_123';

-- 获取用户作为协作者的文件
EXPLAIN ANALYZE
SELECT uf.*, u.username as owner_name, c.permission
FROM collaborations c
JOIN user_files uf ON c.file_id = uf.file_id
JOIN users u ON uf.user_id = u.id
WHERE c.collaborator_id = 1;

-- 检查用户是否是协作者
EXPLAIN ANALYZE
SELECT permission FROM collaborations
WHERE file_id = 'file_123' AND collaborator_id = 1
LIMIT 1;

-- ============================================================================
-- 6. 索引使用情况检查
-- ============================================================================

-- 检查所有索引
SELECT 
    TABLE_NAME,
    INDEX_NAME,
    SEQ_IN_INDEX,
    COLUMN_NAME,
    CARDINALITY
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('user_files', 'shares', 'share_files', 'recycle_bin', 'file_versions', 'collaborations')
ORDER BY TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX;

-- 检查索引使用统计（需要启用 performance_schema）
SELECT 
    OBJECT_NAME as table_name,
    INDEX_NAME,
    COUNT_STAR as total_accesses,
    COUNT_READ as read_accesses,
    COUNT_FETCH as fetch_accesses
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE OBJECT_SCHEMA = DATABASE()
  AND OBJECT_NAME IN ('user_files', 'shares', 'recycle_bin', 'file_versions', 'collaborations')
ORDER BY COUNT_STAR DESC;

-- ============================================================================
-- 7. 慢查询分析
-- ============================================================================

-- 查看慢查询日志设置
SHOW VARIABLES LIKE 'slow_query%';
SHOW VARIABLES LIKE 'long_query_time';

-- 建议：启用慢查询日志
-- SET GLOBAL slow_query_log = 'ON';
-- SET GLOBAL long_query_time = 1;

-- ============================================================================
-- 8. 性能优化建议
-- ============================================================================

-- 如果搜索查询慢，检查以下索引是否存在：
-- SHOW INDEX FROM user_files WHERE Key_name IN ('idx_file_name', 'idx_modified_at', 'idx_file_size', 'idx_user_folder');

-- 如果版本查询慢，检查以下索引是否存在：
-- SHOW INDEX FROM file_versions WHERE Key_name IN ('idx_file_versions', 'idx_current_version');

-- 如果回收站查询慢，检查以下索引是否存在：
-- SHOW INDEX FROM recycle_bin WHERE Key_name = 'idx_user_expires';

-- 如果分享查询慢，检查以下索引是否存在：
-- SHOW INDEX FROM shares WHERE Key_name IN ('idx_share_token', 'idx_user_id', 'idx_expires_at');

-- 如果协作查询慢，检查以下索引是否存在：
-- SHOW INDEX FROM collaborations WHERE Key_name IN ('unique_collaboration', 'idx_collaborator', 'idx_file_owner');
