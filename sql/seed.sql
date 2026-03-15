USE `nex_template`;

START TRANSACTION;

INSERT INTO `menus` (
  `id`, `created_at`, `updated_at`, `parent_id`, `name`, `title`, `path`, `component`, `icon`, `type`, `permission`, `sort`, `hidden`, `status`
) VALUES
  (1, NOW(3), NOW(3), 0, 'dashboard', 'Dashboard', '/dashboard', 'views/dashboard/index.vue', 'House', 'menu', 'dashboard:view', 1, 0, 'enabled'),
  (2, NOW(3), NOW(3), 0, 'system', 'System', '/system', 'Layout', 'Setting', 'directory', '', 2, 0, 'enabled'),
  (3, NOW(3), NOW(3), 2, 'user', 'Users', '/system/users', 'views/system/users/index.vue', 'User', 'menu', 'system:user:view', 1, 0, 'enabled'),
  (4, NOW(3), NOW(3), 2, 'user_write', 'User Write', '', '', '', 'button', 'system:user:write', 2, 0, 'enabled'),
  (5, NOW(3), NOW(3), 2, 'role', 'Roles', '/system/roles', 'views/system/roles/index.vue', 'Lock', 'menu', 'system:role:view', 3, 0, 'enabled'),
  (6, NOW(3), NOW(3), 2, 'role_write', 'Role Write', '', '', '', 'button', 'system:role:write', 4, 0, 'enabled'),
  (7, NOW(3), NOW(3), 2, 'menu', 'Menus', '/system/menus', 'views/system/menus/index.vue', 'Menu', 'menu', 'system:menu:view', 5, 0, 'enabled'),
  (8, NOW(3), NOW(3), 2, 'menu_write', 'Menu Write', '', '', '', 'button', 'system:menu:write', 6, 0, 'enabled'),
  (9, NOW(3), NOW(3), 2, 'auth_setting', 'Auth Settings', '/system/auth-settings', 'views/system/auth-settings/index.vue', 'Setting', 'menu', 'system:auth-setting:view', 7, 0, 'enabled'),
  (10, NOW(3), NOW(3), 2, 'auth_setting_write', 'Auth Settings Write', '', '', '', 'button', 'system:auth-setting:write', 8, 0, 'enabled'),
  (11, NOW(3), NOW(3), 2, 'branding_setting', 'Branding', '/system/branding-settings', 'views/system/branding-settings/index.vue', 'Grid', 'menu', 'system:branding-setting:view', 9, 0, 'enabled'),
  (12, NOW(3), NOW(3), 2, 'branding_setting_write', 'Branding Write', '', '', '', 'button', 'system:branding-setting:write', 10, 0, 'enabled'),
  (13, NOW(3), NOW(3), 2, 'login_audit', 'Login Audits', '/system/login-audits', 'views/system/login-audits/index.vue', 'Document', 'menu', 'system:login-audit:view', 11, 0, 'enabled')
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
  (1, NOW(3), NOW(3), 'Super Admin', 'super_admin', 'enabled', 'Template default role')
ON DUPLICATE KEY UPDATE
  `updated_at` = VALUES(`updated_at`),
  `name` = VALUES(`name`),
  `status` = VALUES(`status`),
  `remark` = VALUES(`remark`);

INSERT INTO `users` (
  `id`, `created_at`, `updated_at`, `username`, `nickname`, `email`, `phone`, `avatar`, `status`, `password`
) VALUES
  (1, NOW(3), NOW(3), 'admin', 'System Admin', 'admin@example.com', '18800000000', 'default-07', 'enabled', '$2a$10$MeTSW1eArpEklHBm0GgqkOrcse.itOPimmDJqp596oCE0v7gMcaX2')
ON DUPLICATE KEY UPDATE
  `updated_at` = VALUES(`updated_at`),
  `nickname` = VALUES(`nickname`),
  `email` = VALUES(`email`),
  `phone` = VALUES(`phone`),
  `avatar` = VALUES(`avatar`),
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
  (1, 10),
  (1, 11),
  (1, 12),
  (1, 13);

COMMIT;
