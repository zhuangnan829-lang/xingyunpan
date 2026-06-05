-- Phase 5 数据库回滚脚本
-- 删除 Phase 5 创建的所有表和索引

-- 警告：此脚本将删除所有 Phase 5 数据，请谨慎执行！
-- 建议在执行前备份数据库

-- 1. 删除 collaborations 表
DROP TABLE IF EXISTS `collaborations`;

-- 2. 删除 file_versions 表
DROP TABLE IF EXISTS `file_versions`;

-- 3. 删除 recycle_bin 表
DROP TABLE IF EXISTS `recycle_bin`;

-- 4. 删除 share_files 表
DROP TABLE IF EXISTS `share_files`;

-- 5. 删除 shares 表
DROP TABLE IF EXISTS `shares`;

-- 6. 删除 user_files 表的搜索优化索引
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE table_schema = DATABASE() 
               AND table_name = 'user_files' 
               AND index_name = 'idx_file_name');
SET @sqlstmt := IF(@exist > 0, 
    'DROP INDEX idx_file_name ON user_files', 
    'SELECT ''Index idx_file_name does not exist'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE table_schema = DATABASE() 
               AND table_name = 'user_files' 
               AND index_name = 'idx_modified_at');
SET @sqlstmt := IF(@exist > 0, 
    'DROP INDEX idx_modified_at ON user_files', 
    'SELECT ''Index idx_modified_at does not exist'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE table_schema = DATABASE() 
               AND table_name = 'user_files' 
               AND index_name = 'idx_file_size');
SET @sqlstmt := IF(@exist > 0, 
    'DROP INDEX idx_file_size ON user_files', 
    'SELECT ''Index idx_file_size does not exist'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 完成
SELECT 'Phase 5 rollback completed successfully' AS status;
SELECT 'All Phase 5 tables and indexes have been removed' AS message;
