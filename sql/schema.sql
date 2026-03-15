CREATE DATABASE IF NOT EXISTS `nex_template`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;

USE `nex_template`;

CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `username` varchar(64) NOT NULL,
  `nickname` varchar(64) NOT NULL,
  `email` varchar(128) DEFAULT NULL,
  `phone` varchar(32) DEFAULT NULL,
  `avatar` varchar(255) DEFAULT 'default-01',
  `status` varchar(32) DEFAULT 'enabled',
  `password` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_email` (`email`),
  UNIQUE KEY `idx_users_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `code` varchar(64) NOT NULL,
  `status` varchar(32) DEFAULT 'enabled',
  `remark` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_roles_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `parent_id` bigint unsigned DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `title` varchar(64) NOT NULL,
  `path` varchar(255) DEFAULT NULL,
  `component` varchar(255) DEFAULT NULL,
  `icon` varchar(64) DEFAULT NULL,
  `type` varchar(32) DEFAULT 'menu',
  `permission` varchar(128) DEFAULT NULL,
  `sort` bigint DEFAULT '0',
  `hidden` tinyint(1) DEFAULT '0',
  `status` varchar(32) DEFAULT 'enabled',
  PRIMARY KEY (`id`),
  KEY `idx_menus_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `auth_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `enable_email_login` tinyint(1) DEFAULT '1',
  `enable_phone_login` tinyint(1) DEFAULT '1',
  `enable_email_registration` tinyint(1) DEFAULT '1',
  `enable_phone_registration` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `branding_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `app_title` varchar(128) DEFAULT NULL,
  `console_name` varchar(128) DEFAULT NULL,
  `product_tagline` varchar(255) DEFAULT NULL,
  `logo_mark_url` varchar(255) DEFAULT NULL,
  `favicon_url` varchar(255) DEFAULT NULL,
  `login_hero_url` varchar(255) DEFAULT NULL,
  `primary` varchar(16) DEFAULT NULL,
  `primary_dark` varchar(16) DEFAULT NULL,
  `shell_start` varchar(16) DEFAULT NULL,
  `shell_end` varchar(16) DEFAULT NULL,
  `hero_start` varchar(16) DEFAULT NULL,
  `hero_end` varchar(16) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `login_audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `username` varchar(64) DEFAULT NULL,
  `account` varchar(128) NOT NULL,
  `login_type` varchar(32) NOT NULL,
  `status` varchar(32) NOT NULL,
  `ip` varchar(64) DEFAULT NULL,
  `user_agent` varchar(255) DEFAULT NULL,
  `message` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_login_audit_logs_user_id` (`user_id`),
  KEY `idx_login_audit_logs_username` (`username`),
  KEY `idx_login_audit_logs_account` (`account`),
  KEY `idx_login_audit_logs_login_type` (`login_type`),
  KEY `idx_login_audit_logs_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `user_roles` (
  `user_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`user_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `role_menus` (
  `role_id` bigint unsigned NOT NULL,
  `menu_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`, `menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `contact_submissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `email` varchar(128) NOT NULL,
  `phone` varchar(32) DEFAULT NULL,
  `company` varchar(128) DEFAULT NULL,
  `message` text NOT NULL,
  `source` varchar(64) DEFAULT 'website',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
