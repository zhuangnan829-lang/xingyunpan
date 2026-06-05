-- Phase 5 数据库迁移脚本
-- 创建 5 个新表和搜索优化索引

-- 1. 创建 shares 表
CREATE TABLE IF NOT EXISTS `shares` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    `share_token` VARCHAR(64) NOT NULL COMMENT '分享令牌',
    `access_code_hash` VARCHAR(255) DEFAULT NULL COMMENT '访问密码哈希',
    `expires_at` DATETIME DEFAULT NULL COMMENT '过期时间',
    `download_count` INT NOT NULL DEFAULT 0 COMMENT '下载次数',
    `access_count` INT NOT NULL DEFAULT 0 COMMENT '访问次数',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_share_token` (`share_token`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_expires_at` (`expires_at`),
    KEY `idx_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_shares_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享表';

-- 2. 创建 share_files 表
CREATE TABLE IF NOT EXISTS `share_files` (
    `share_id` BIGINT UNSIGNED NOT NULL COMMENT '分享ID',
    `file_id` BIGINT UNSIGNED NOT NULL COMMENT '文件ID',
    PRIMARY KEY (`share_id`, `file_id`),
    KEY `idx_file_id` (`file_id`),
    CONSTRAINT `fk_share_files_share` FOREIGN KEY (`share_id`) REFERENCES `shares` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_share_files_file` FOREIGN KEY (`file_id`) REFERENCES `user_files` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享文件关联表';

-- 3. 创建 recycle_bin 表
CREATE TABLE IF NOT EXISTS `recycle_bin` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `file_id` BIGINT UNSIGNED NOT NULL COMMENT '文件ID',
    `file_name` VARCHAR(255) NOT NULL COMMENT '文件名',
    `file_size` BIGINT NOT NULL COMMENT '文件大小',
    `file_type` VARCHAR(100) DEFAULT NULL COMMENT '文件类型',
    `original_path` VARCHAR(1000) NOT NULL COMMENT '原始路径',
    `deleted_at` DATETIME NOT NULL COMMENT '删除时间',
    `expires_at` DATETIME NOT NULL COMMENT '过期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_expires` (`user_id`, `expires_at`),
    CONSTRAINT `fk_recycle_bin_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回收站表';

-- 4. 创建 file_versions 表
CREATE TABLE IF NOT EXISTS `file_versions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `file_id` BIGINT UNSIGNED NOT NULL COMMENT '文件ID',
    `version_number` INT NOT NULL COMMENT '版本号',
    `physical_file_id` BIGINT UNSIGNED NOT NULL COMMENT '物理文件ID',
    `file_size` BIGINT NOT NULL COMMENT '文件大小',
    `uploader_id` BIGINT UNSIGNED NOT NULL COMMENT '上传者ID',
    `is_current` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否当前版本',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_file_versions` (`file_id`, `version_number`),
    KEY `idx_current_version` (`file_id`, `is_current`),
    KEY `idx_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_file_versions_file` FOREIGN KEY (`file_id`) REFERENCES `user_files` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_file_versions_physical` FOREIGN KEY (`physical_file_id`) REFERENCES `physical_files` (`id`),
    CONSTRAINT `fk_file_versions_uploader` FOREIGN KEY (`uploader_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件版本表';

-- 5. 创建 collaborations 表
CREATE TABLE IF NOT EXISTS `collaborations` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `file_id` BIGINT UNSIGNED NOT NULL COMMENT '文件ID',
    `owner_id` BIGINT UNSIGNED NOT NULL COMMENT '所有者ID',
    `collaborator_id` BIGINT UNSIGNED NOT NULL COMMENT '协作者ID',
    `permission` VARCHAR(20) NOT NULL COMMENT '权限(view/download/edit)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_collaboration` (`file_id`, `collaborator_id`),
    KEY `idx_file_owner` (`file_id`, `owner_id`),
    KEY `idx_collaborator` (`collaborator_id`),
    KEY `idx_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_collaborations_file` FOREIGN KEY (`file_id`) REFERENCES `user_files` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_collaborations_owner` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`),
    CONSTRAINT `fk_collaborations_collaborator` FOREIGN KEY (`collaborator_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='协作表';

-- 6. 为 user_files 表添加搜索优化索引
-- 检查索引是否存在，如果不存在则创建
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE table_schema = DATABASE() 
               AND table_name = 'user_files' 
               AND index_name = 'idx_file_name');
SET @sqlstmt := IF(@exist = 0, 
    'CREATE INDEX idx_file_name ON user_files(file_name)', 
    'SELECT ''Index idx_file_name already exists'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE table_schema = DATABASE() 
               AND table_name = 'user_files' 
               AND index_name = 'idx_modified_at');
SET @sqlstmt := IF(@exist = 0, 
    'CREATE INDEX idx_modified_at ON user_files(updated_at)', 
    'SELECT ''Index idx_modified_at already exists'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE table_schema = DATABASE() 
               AND table_name = 'user_files' 
               AND index_name = 'idx_file_size');
SET @sqlstmt := IF(@exist = 0, 
    'CREATE INDEX idx_file_size ON user_files(file_size)', 
    'SELECT ''Index idx_file_size already exists'' AS message');
PREPARE stmt FROM @sqlstmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- idx_user_folder 索引已经在 user_file.go 中定义，这里不需要重复创建

-- 完成
SELECT 'Phase 5 schema migration completed successfully' AS status;
