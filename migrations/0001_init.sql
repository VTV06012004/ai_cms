-- 0001_init: schema ban đầu, khớp với schema do AutoMigrate tạo trước khi có migration có phiên bản.

CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(191) NOT NULL,
  `password_hash` longtext NOT NULL,
  `name` longtext,
  `role` enum('admin','editor','author') DEFAULT 'author',
  `root_admin_key` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_email` (`email`),
  UNIQUE KEY `idx_users_root_admin_key` (`root_admin_key`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` longtext NOT NULL,
  `slug` varchar(255) NOT NULL,
  `description` longtext,
  `parent_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_categories_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `tags` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `slug` varchar(255) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_tags_name` (`name`),
  UNIQUE KEY `uni_tags_slug` (`slug`),
  KEY `idx_tags_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `location` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `menu_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `menu_id` bigint unsigned DEFAULT NULL,
  `title` longtext,
  `type` longtext,
  `target_id` bigint unsigned DEFAULT NULL,
  `url` longtext,
  `sort_order` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_menu_items_deleted_at` (`deleted_at`),
  KEY `fk_menus_menu_items` (`menu_id`),
  CONSTRAINT `fk_menus_menu_items` FOREIGN KEY (`menu_id`) REFERENCES `menus` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `posts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` longtext NOT NULL,
  `slug` varchar(255) NOT NULL,
  `content` longtext,
  `excerpt` text,
  `status` enum('draft','review','approved','rejected','published','scheduled') NOT NULL DEFAULT 'draft',
  `rejection_reason` text,
  `author_id` bigint unsigned DEFAULT NULL,
  `moderator_id` bigint unsigned DEFAULT NULL,
  `category_id` bigint unsigned DEFAULT NULL,
  `seo_title` longtext,
  `seo_description` longtext,
  `scheduled_at` datetime(3) DEFAULT NULL,
  `published_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `active_slug` varchar(255) GENERATED ALWAYS AS ((case when (`deleted_at` is null) then `slug` else NULL end)) STORED,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_posts_active_slug` (`active_slug`),
  KEY `idx_posts_author_id` (`author_id`),
  KEY `idx_posts_category_id` (`category_id`),
  KEY `idx_posts_scheduled_at` (`scheduled_at`),
  KEY `idx_posts_deleted_at` (`deleted_at`),
  KEY `fk_posts_moderator` (`moderator_id`),
  CONSTRAINT `fk_posts_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_posts_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  CONSTRAINT `fk_posts_moderator` FOREIGN KEY (`moderator_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `post_tags` (
  `post_model_id` bigint unsigned NOT NULL,
  `tag_model_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`post_model_id`,`tag_model_id`),
  KEY `fk_post_tags_tag_model` (`tag_model_id`),
  CONSTRAINT `fk_post_tags_post_model` FOREIGN KEY (`post_model_id`) REFERENCES `posts` (`id`),
  CONSTRAINT `fk_post_tags_tag_model` FOREIGN KEY (`tag_model_id`) REFERENCES `tags` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `revisions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `post_id` bigint unsigned DEFAULT NULL,
  `title` longtext NOT NULL,
  `slug` longtext NOT NULL,
  `content` text,
  `status` varchar(191) DEFAULT 'draft',
  `category_id` bigint unsigned DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_revisions_category_id` (`category_id`),
  KEY `idx_revisions_user_id` (`user_id`),
  KEY `idx_revisions_deleted_at` (`deleted_at`),
  KEY `idx_revisions_post_id` (`post_id`),
  CONSTRAINT `fk_revisions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `action` longtext NOT NULL,
  `target` longtext,
  `details` text,
  `ip_address` longtext,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_audit_logs_user_id` (`user_id`),
  CONSTRAINT `fk_audit_logs_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;