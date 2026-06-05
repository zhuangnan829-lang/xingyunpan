-- 查询性能测试脚本
-- 用于对比优化前后的查询性能

-- ============================================
-- 1. 测试文件列表查询性能
-- ============================================

-- 查询用户的文件列表（按 user_id 和 folder_id）
EXPLAIN SELECT * FROM user_files 
WHERE user_id = 1 AND folder_id = 0 AND deleted_at IS NULL;

-- ============================================
-- 2. 测试文件哈希查询性能（秒传功能）
-- ============================================

-- 查询相同哈希的文件
EXPLAIN SELECT * FROM physical_files 
WHERE file_hash = 'test_hash_value';

-- ============================================
-- 3. 测试分片上传任务查询性能
-- ============================================

-- 查询待清理的过期上传任务
EXPLAIN SELECT * FROM multipart_uploads 
WHERE status = 'pending' AND created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR);

-- ============================================
-- 4. 测试用户上传任务查询性能
-- ============================================

-- 查询用户的上传任务
EXPLAIN SELECT * FROM multipart_uploads 
WHERE user_id = 1 AND status = 'pending';

-- ============================================
-- 5. 测试文件删除任务查询性能
-- ============================================

-- 查询待处理的删除任务
EXPLAIN SELECT * FROM file_deletions 
WHERE status = 'pending' AND created_at < NOW();

-- ============================================
-- 6. 测试用户名查询性能（登录）
-- ============================================

-- 查询用户名
EXPLAIN SELECT * FROM users 
WHERE username = 'testuser';

-- ============================================
-- 7. 测试关联查询性能
-- ============================================

-- 查询文件及其物理文件信息
EXPLAIN SELECT uf.*, pf.* 
FROM user_files uf 
LEFT JOIN physical_files pf ON uf.real_file_id = pf.id 
WHERE uf.user_id = 1 AND uf.folder_id = 0 AND uf.deleted_at IS NULL;

-- ============================================
-- 8. 测试引用计数查询性能（垃圾回收）
-- ============================================

-- 查询引用计数为 0 的文件
EXPLAIN SELECT * FROM physical_files 
WHERE ref_count = 0;

