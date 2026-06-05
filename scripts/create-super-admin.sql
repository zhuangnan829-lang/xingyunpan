-- 创建超级管理员账户
-- 用户名: Sincerity
-- 邮箱: 123456@xingyunnan.it.com
-- 密码: #A456718293a (需要先注册后再更新角色)

-- 注意：这个脚本假设用户已经通过注册接口创建
-- 如果用户还不存在，请先通过前端注册页面注册

-- 将指定用户设置为超级管理员
UPDATE users 
SET role = 'admin' 
WHERE email = '123456@xingyunnan.it.com' AND username = 'Sincerity';

-- 将其他所有用户设置为普通用户
UPDATE users 
SET role = 'user' 
WHERE email != '123456@xingyunnan.it.com' OR username != 'Sincerity';

-- 验证结果
SELECT id, username, email, role, created_at 
FROM users 
ORDER BY role DESC, created_at ASC;
