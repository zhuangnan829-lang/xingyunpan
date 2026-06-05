-- 在宝塔 MySQL 终端中执行此脚本
-- 创建 xingyunpan 用户并授权

-- 方案 1：创建新用户 xingyunpan（推荐，更安全）
CREATE USER IF NOT EXISTS 'xingyunpan'@'%' IDENTIFIED BY 'RMGdedeMoMnMc';
GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'%';
FLUSH PRIVILEGES;

-- 方案 2：如果你想用 root 用户（不推荐）
-- GRANT ALL PRIVILEGES ON xingyunpan.* TO 'root'@'%' IDENTIFIED BY 'RMGdedeMoMnMc';
-- FLUSH PRIVILEGES;

-- 查看授权结果
SELECT user, host FROM mysql.user WHERE user IN ('root', 'xingyunpan');
SHOW GRANTS FOR 'xingyunpan'@'%';
