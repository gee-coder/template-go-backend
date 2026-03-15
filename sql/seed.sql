USE `nex_template`;

START TRANSACTION;

INSERT INTO `menus` (
  `id`, `created_at`, `updated_at`, `parent_id`, `name`, `title`, `path`, `component`, `icon`, `type`, `permission`, `sort`, `hidden`, `status`
) VALUES
  (1, NOW(3), NOW(3), 0, 'dashboard', '工作台', '/dashboard', 'views/dashboard/index.vue', 'House', 'menu', 'dashboard:view', 1, 0, 'enabled'),
  (2, NOW(3), NOW(3), 0, 'system', '系统管理', '/system', 'Layout', 'Setting', 'directory', '', 2, 0, 'enabled'),
  (3, NOW(3), NOW(3), 2, 'user', '用户管理', '/system/users', 'views/system/users/index.vue', 'User', 'menu', 'system:user:view', 1, 0, 'enabled'),
  (4, NOW(3), NOW(3), 2, 'user_write', '用户写入', '', '', '', 'button', 'system:user:write', 2, 0, 'enabled'),
  (5, NOW(3), NOW(3), 2, 'role', '角色管理', '/system/roles', 'views/system/roles/index.vue', 'Lock', 'menu', 'system:role:view', 3, 0, 'enabled'),
  (6, NOW(3), NOW(3), 2, 'role_write', '角色写入', '', '', '', 'button', 'system:role:write', 4, 0, 'enabled'),
  (7, NOW(3), NOW(3), 2, 'menu', '菜单管理', '/system/menus', 'views/system/menus/index.vue', 'Menu', 'menu', 'system:menu:view', 5, 0, 'enabled'),
  (8, NOW(3), NOW(3), 2, 'menu_write', '菜单写入', '', '', '', 'button', 'system:menu:write', 6, 0, 'enabled'),
  (9, NOW(3), NOW(3), 2, 'auth_setting', '认证设置', '/system/auth-settings', 'views/system/auth-settings/index.vue', 'Setting', 'menu', 'system:auth-setting:view', 7, 0, 'enabled'),
  (10, NOW(3), NOW(3), 2, 'auth_setting_write', '认证设置写入', '', '', '', 'button', 'system:auth-setting:write', 8, 0, 'enabled')
ON DUPLICATE KEY UPDATE
  `updated_at` = VALUES(`updated_at`),
  `parent_id` = VALUES(`parent_id`),
  `name` = VALUES(`name`),
  `title` = VALUES(`title`),
  `path` = VALUES(`path`),
  `component` = VALUES(`component`),
  `icon` = VALUES(`icon`),
  `type` = VALUES(`type`),
  `permission` = VALUES(`permission`),
  `sort` = VALUES(`sort`),
  `hidden` = VALUES(`hidden`),
  `status` = VALUES(`status`);

INSERT INTO `roles` (
  `id`, `created_at`, `updated_at`, `name`, `code`, `status`, `remark`
) VALUES
  (1, NOW(3), NOW(3), '超级管理员', 'super_admin', 'enabled', '模板初始化角色')
ON DUPLICATE KEY UPDATE
  `updated_at` = VALUES(`updated_at`),
  `name` = VALUES(`name`),
  `status` = VALUES(`status`),
  `remark` = VALUES(`remark`);

INSERT INTO `users` (
  `id`, `created_at`, `updated_at`, `username`, `nickname`, `email`, `phone`, `status`, `password`
) VALUES
  (1, NOW(3), NOW(3), 'admin', '系统管理员', 'admin@example.com', '18800000000', 'enabled', '$2a$10$MeTSW1eArpEklHBm0GgqkOrcse.itOPimmDJqp596oCE0v7gMcaX2')
ON DUPLICATE KEY UPDATE
  `updated_at` = VALUES(`updated_at`),
  `nickname` = VALUES(`nickname`),
  `email` = VALUES(`email`),
  `phone` = VALUES(`phone`),
  `status` = VALUES(`status`),
  `password` = VALUES(`password`);

INSERT IGNORE INTO `user_roles` (`user_id`, `role_id`) VALUES
  (1, 1);

INSERT IGNORE INTO `role_menus` (`role_id`, `menu_id`) VALUES
  (1, 1),
  (1, 2),
  (1, 3),
  (1, 4),
  (1, 5),
  (1, 6),
  (1, 7),
  (1, 8),
  (1, 9),
  (1, 10);

COMMIT;
