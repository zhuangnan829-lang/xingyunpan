-- 星云盘 V2 数据库初始化脚本
-- 此脚本会在 MySQL 容器首次启动时自动执行

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `xingyunpan` 
  DEFAULT CHARACTER SET utf8mb4 
  COLLATE utf8mb4_unicode_ci;

USE `xingyunpan`;

-- 注意：表结构将由 GORM 自动迁移创建
-- 这里只做一些初始化配置

SET FOREIGN_KEY_CHECKS = 1;

-- 输出初始化完成信息
SELECT '数据库初始化完成' AS message;
