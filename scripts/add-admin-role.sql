-- 添加角色字段到用户表
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user' NOT NULL COMMENT '用户角色(admin/user)';

-- 将指定用户设置为超级管理员
UPDATE users 
SET role = 'admin' 
WHERE email = '123456@xingyunnan.it.com' AND username = 'Sincerity';

-- 确保其他所有用户都是普通用户
UPDATE users 
SET role = 'user' 
WHERE email != '123456@xingyunnan.it.com' OR username != 'Sincerity';
