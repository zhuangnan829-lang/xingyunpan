-- 数据库索引优化脚本
-- 用于提升查询性能，减少全表扫描

-- ============================================
-- 1. user_files 表索引
-- ============================================

-- 复合索引: user_id + folder_id + deleted_at
-- 用于文件列表查询，支持按用户和文件夹过滤
CREATE INDEX IF NOT EXISTS idx_user_files_user_folder 
ON user_files(user_id, folder_id, deleted_at);

-- 索引: real_file_id
-- 用于关联查询 physical_files 表
CREATE INDEX IF NOT EXISTS idx_user_files_real_file 
ON user_files(real_file_id);

-- 索引: parent_id
-- 用于查询子文件/文件夹
CREATE INDEX IF NOT EXISTS idx_user_files_parent 
ON user_files(parent_id);

-- 索引: is_folder
-- 用于区分文件和文件夹
CREATE INDEX IF NOT EXISTS idx_user_files_is_folder 
ON user_files(is_folder);

-- ============================================
-- 2. physical_files 表索引
-- ============================================

-- 唯一索引: file_hash
-- 用于秒传功能，快速查找相同哈希的文件
CREATE UNIQUE INDEX IF NOT EXISTS idx_physical_files_hash 
ON physical_files(file_hash);

-- 索引: ref_count
-- 用于查找引用计数为 0 的文件（垃圾回收）
CREATE INDEX IF NOT EXISTS idx_physical_files_ref_count 
ON physical_files(ref_count);

-- ============================================
-- 3. multipart_uploads 表索引
-- ============================================

-- 复合索引: status + created_at
-- 用于清理过期的分片上传任务
CREATE INDEX IF NOT EXISTS idx_multipart_status_created 
ON multipart_uploads(status, created_at);

-- 复合索引: user_id + status
-- 用于查询用户的上传任务
CREATE INDEX IF NOT EXISTS idx_multipart_user_status 
ON multipart_uploads(user_id, status);

-- 唯一索引: upload_id
-- 用于快速查找上传任务
CREATE UNIQUE INDEX IF NOT EXISTS idx_multipart_upload_id 
ON multipart_uploads(upload_id);

-- ============================================
-- 4. file_deletions 表索引
-- ============================================

-- 复合索引: status + created_at
-- 用于查找待处理的删除任务
CREATE INDEX IF NOT EXISTS idx_file_deletions_status_created 
ON file_deletions(status, created_at);

-- 索引: physical_file_id
-- 用于关联查询 physical_files 表
CREATE INDEX IF NOT EXISTS idx_file_deletions_physical_file 
ON file_deletions(physical_file_id);

-- ============================================
-- 5. users 表索引
-- ============================================

-- 唯一索引: username
-- 用于登录和注册时的用户名查找
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username 
ON users(username);

-- 唯一索引: email (如果有)
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email 
-- ON users(email);

-- ============================================
-- 6. 更新表统计信息
-- ============================================

-- 分析表，更新统计信息，帮助查询优化器选择最佳执行计划
ANALYZE TABLE users;
ANALYZE TABLE user_files;
ANALYZE TABLE physical_files;
ANALYZE TABLE multipart_uploads;
ANALYZE TABLE file_deletions;

-- ============================================
-- 7. 查看索引创建结果
-- ============================================

-- 查看 user_files 表的索引
SHOW INDEX FROM user_files;

-- 查看 physical_files 表的索引
SHOW INDEX FROM physical_files;

-- 查看 multipart_uploads 表的索引
SHOW INDEX FROM multipart_uploads;

-- 查看 file_deletions 表的索引
SHOW INDEX FROM file_deletions;

-- 查看 users 表的索引
SHOW INDEX FROM users;
