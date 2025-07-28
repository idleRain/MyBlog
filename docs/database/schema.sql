-- MyBlog 数据库架构设计 SQL 脚本
-- 用途: 数据库设计文档和参考实现
-- 注意: 此文件仅作为文档参考，实际数据库结构由 GORM 模型自动迁移管理

-- 设置字符集和校对规则
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ===================================
-- 1. 用户管理模块
-- ===================================

-- 1.1 用户表 (已存在，更新结构)
ALTER TABLE `users` 
ADD COLUMN `bio` TEXT COMMENT '个人简介' AFTER `avatar`,
ADD COLUMN `last_login_at` DATETIME(3) DEFAULT NULL COMMENT '最后登录时间' AFTER `status`,
ADD COLUMN `last_login_ip` VARCHAR(45) DEFAULT NULL COMMENT '最后登录IP' AFTER `last_login_at`,
ADD COLUMN `login_count` INT UNSIGNED DEFAULT 0 COMMENT '登录次数' AFTER `last_login_ip`,
ADD COLUMN `email_verified_at` DATETIME(3) DEFAULT NULL COMMENT '邮箱验证时间' AFTER `login_count`,
ADD INDEX `idx_role` (`role`),
ADD INDEX `idx_last_login` (`last_login_at`);

-- 1.2 用户会话表
CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
  `refresh_token` VARCHAR(255) NOT NULL COMMENT '刷新令牌',
  `access_token_hash` VARCHAR(64) DEFAULT NULL COMMENT '访问令牌哈希值',
  `device_info` JSON DEFAULT NULL COMMENT '设备信息（浏览器、操作系统等）',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT '登录IP地址',
  `user_agent` TEXT COMMENT '用户代理字符串',
  `expires_at` DATETIME(3) NOT NULL COMMENT '令牌过期时间',
  `last_used_at` DATETIME(3) DEFAULT NULL COMMENT '最后使用时间',
  `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT '会话状态：1-活跃，0-已注销',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refresh_token` (`refresh_token`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_ip_address` (`ip_address`),
  CONSTRAINT `fk_user_sessions_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- ===================================
-- 2. 内容管理模块
-- ===================================

-- 2.1 文章分类表
CREATE TABLE IF NOT EXISTS `categories` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `name` VARCHAR(50) NOT NULL COMMENT '分类名称',
  `slug` VARCHAR(50) NOT NULL COMMENT 'URL友好标识',
  `description` TEXT COMMENT '分类描述',
  `cover_image` VARCHAR(255) DEFAULT NULL COMMENT '分类封面图',
  `parent_id` INT UNSIGNED DEFAULT NULL COMMENT '父分类ID',
  `level` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分类层级',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序权重',
  `article_count` INT UNSIGNED DEFAULT 0 COMMENT '文章数量',
  `is_featured` TINYINT NOT NULL DEFAULT 0 COMMENT '是否为精选分类',
  `seo_title` VARCHAR(100) DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` VARCHAR(255) DEFAULT NULL COMMENT 'SEO描述',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_level` (`level`),
  KEY `idx_sort_order` (`sort_order`),
  KEY `idx_is_featured` (`is_featured`),
  KEY `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_categories_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类表';

-- 2.2 标签表
CREATE TABLE IF NOT EXISTS `tags` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '标签ID',
  `name` VARCHAR(30) NOT NULL COMMENT '标签名称',
  `slug` VARCHAR(30) NOT NULL COMMENT 'URL友好标识',
  `color` VARCHAR(7) DEFAULT '#808080' COMMENT '标签颜色（HEX格式）',
  `description` VARCHAR(200) DEFAULT NULL COMMENT '标签描述',
  `usage_count` INT UNSIGNED DEFAULT 0 COMMENT '使用次数',
  `is_hot` TINYINT NOT NULL DEFAULT 0 COMMENT '是否热门标签',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_usage_count` (`usage_count`),
  KEY `idx_is_hot` (`is_hot`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

-- 2.3 文章表
CREATE TABLE IF NOT EXISTS `articles` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` VARCHAR(200) NOT NULL COMMENT '文章标题',
  `slug` VARCHAR(200) NOT NULL COMMENT 'URL友好标识',
  `summary` TEXT COMMENT '文章摘要',
  `content` LONGTEXT NOT NULL COMMENT '文章内容（Markdown格式）',
  `content_html` LONGTEXT COMMENT '文章内容（HTML格式，缓存用）',
  `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '封面图片URL',
  `author_id` INT UNSIGNED NOT NULL COMMENT '作者ID',
  `category_id` INT UNSIGNED DEFAULT NULL COMMENT '主分类ID',
  `status` ENUM('draft','published','archived','private') NOT NULL DEFAULT 'draft' COMMENT '文章状态',
  `is_featured` TINYINT NOT NULL DEFAULT 0 COMMENT '是否精选文章',
  `is_top` TINYINT NOT NULL DEFAULT 0 COMMENT '是否置顶',
  `comment_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否允许评论',
  `view_count` INT UNSIGNED DEFAULT 0 COMMENT '浏览量',
  `like_count` INT UNSIGNED DEFAULT 0 COMMENT '点赞数',
  `comment_count` INT UNSIGNED DEFAULT 0 COMMENT '评论数',
  `word_count` INT UNSIGNED DEFAULT 0 COMMENT '字数统计',
  `reading_time` INT UNSIGNED DEFAULT 0 COMMENT '预计阅读时间（分钟）',
  `seo_title` VARCHAR(100) DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` VARCHAR(255) DEFAULT NULL COMMENT 'SEO描述',
  `seo_keywords` VARCHAR(200) DEFAULT NULL COMMENT 'SEO关键词',
  `published_at` DATETIME(3) DEFAULT NULL COMMENT '发布时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_author_id` (`author_id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_featured` (`is_featured`),
  KEY `idx_is_top` (`is_top`),
  KEY `idx_published_at` (`published_at`),
  KEY `idx_view_count` (`view_count`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_status_published` (`status`,`published_at`),
  CONSTRAINT `fk_articles_author_id` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_articles_category_id` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表';

-- 2.4 文章标签关联表
CREATE TABLE IF NOT EXISTS `article_tags` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` INT UNSIGNED NOT NULL COMMENT '文章ID',
  `tag_id` INT UNSIGNED NOT NULL COMMENT '标签ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_tag` (`article_id`,`tag_id`),
  KEY `idx_tag_id` (`tag_id`),
  CONSTRAINT `fk_article_tags_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_tags_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签关联表';

-- ===================================
-- 3. 评论系统模块
-- ===================================

-- 3.1 评论表
CREATE TABLE IF NOT EXISTS `comments` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `article_id` INT UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` INT UNSIGNED DEFAULT NULL COMMENT '用户ID（注册用户）',
  `parent_id` INT UNSIGNED DEFAULT NULL COMMENT '父评论ID（回复功能）',
  `root_id` INT UNSIGNED DEFAULT NULL COMMENT '根评论ID（便于查询评论树）',
  `level` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '评论层级',
  `author_name` VARCHAR(50) DEFAULT NULL COMMENT '游客姓名',
  `author_email` VARCHAR(100) DEFAULT NULL COMMENT '游客邮箱',
  `author_website` VARCHAR(255) DEFAULT NULL COMMENT '游客网站',
  `author_ip` VARCHAR(45) DEFAULT NULL COMMENT '评论者IP地址',
  `content` TEXT NOT NULL COMMENT '评论内容',
  `content_html` TEXT COMMENT '评论内容（HTML格式，缓存用）',
  `status` ENUM('pending','approved','rejected','spam','trash') NOT NULL DEFAULT 'pending' COMMENT '审核状态',
  `like_count` INT UNSIGNED DEFAULT 0 COMMENT '点赞数',
  `reply_count` INT UNSIGNED DEFAULT 0 COMMENT '回复数量',
  `user_agent` TEXT COMMENT '用户代理',
  `is_author` TINYINT NOT NULL DEFAULT 0 COMMENT '是否为文章作者回复',
  `is_pinned` TINYINT NOT NULL DEFAULT 0 COMMENT '是否置顶评论',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_article_id` (`article_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_root_id` (`root_id`),
  KEY `idx_status` (`status`),
  KEY `idx_author_ip` (`author_ip`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_article_status_created` (`article_id`,`status`,`created_at`),
  CONSTRAINT `fk_comments_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comments_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_comments_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comments_root_id` FOREIGN KEY (`root_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- ===================================
-- 4. 媒体管理模块
-- ===================================

-- 4.1 媒体文件表
CREATE TABLE IF NOT EXISTS `media_files` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `filename` VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `stored_name` VARCHAR(255) NOT NULL COMMENT '存储文件名（UUID）',
  `file_path` VARCHAR(500) NOT NULL COMMENT '文件存储路径',
  `file_url` VARCHAR(500) NOT NULL COMMENT '文件访问URL',
  `thumbnail_url` VARCHAR(500) DEFAULT NULL COMMENT '缩略图URL',
  `mime_type` VARCHAR(100) NOT NULL COMMENT 'MIME类型',
  `file_size` BIGINT UNSIGNED NOT NULL COMMENT '文件大小（字节）',
  `file_hash` VARCHAR(64) DEFAULT NULL COMMENT '文件SHA256哈希值',
  `width` INT UNSIGNED DEFAULT NULL COMMENT '图片宽度',
  `height` INT UNSIGNED DEFAULT NULL COMMENT '图片高度',
  `alt_text` VARCHAR(255) DEFAULT NULL COMMENT '替代文本（SEO用）',
  `uploader_id` INT UNSIGNED NOT NULL COMMENT '上传者ID',
  `upload_ip` VARCHAR(45) DEFAULT NULL COMMENT '上传IP地址',
  `storage_type` ENUM('local','oss','s3','cos') NOT NULL DEFAULT 'local' COMMENT '存储类型',
  `folder` VARCHAR(100) DEFAULT NULL COMMENT '文件夹分类',
  `usage_count` INT UNSIGNED DEFAULT 0 COMMENT '使用次数',
  `is_public` TINYINT NOT NULL DEFAULT 1 COMMENT '是否公开访问',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stored_name` (`stored_name`),
  KEY `idx_uploader_id` (`uploader_id`),
  KEY `idx_mime_type` (`mime_type`),
  KEY `idx_file_hash` (`file_hash`),
  KEY `idx_storage_type` (`storage_type`),
  KEY `idx_folder` (`folder`),
  KEY `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_media_files_uploader_id` FOREIGN KEY (`uploader_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='媒体文件表';

-- ===================================
-- 5. 系统配置模块
-- ===================================

-- 5.1 系统设置表
CREATE TABLE IF NOT EXISTS `settings` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '设置ID',
  `key_name` VARCHAR(100) NOT NULL COMMENT '配置键名',
  `value` LONGTEXT COMMENT '配置值（支持JSON格式）',
  `default_value` LONGTEXT COMMENT '默认值',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '配置描述',
  `type` ENUM('string','number','boolean','json','array') NOT NULL DEFAULT 'string' COMMENT '值类型',
  `group_name` VARCHAR(50) NOT NULL DEFAULT 'general' COMMENT '配置分组',
  `is_public` TINYINT NOT NULL DEFAULT 0 COMMENT '是否公开（前端可访问）',
  `is_readonly` TINYINT NOT NULL DEFAULT 0 COMMENT '是否只读',
  `validation_rule` VARCHAR(200) DEFAULT NULL COMMENT '验证规则',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序权重',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key_name` (`key_name`),
  KEY `idx_group_name` (`group_name`),
  KEY `idx_is_public` (`is_public`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- ===================================
-- 6. 统计和日志模块
-- ===================================

-- 6.1 用户活动日志表
CREATE TABLE IF NOT EXISTS `user_activities` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动ID',
  `user_id` INT UNSIGNED DEFAULT NULL COMMENT '用户ID',
  `action` VARCHAR(50) NOT NULL COMMENT '操作类型',
  `resource_type` VARCHAR(50) DEFAULT NULL COMMENT '资源类型（article、comment等）',
  `resource_id` INT UNSIGNED DEFAULT NULL COMMENT '资源ID',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '操作描述',
  `metadata` JSON DEFAULT NULL COMMENT '额外元数据',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` TEXT COMMENT '用户代理',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_resource` (`resource_type`,`resource_id`),
  KEY `idx_ip_address` (`ip_address`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_user_activities_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户活动日志表';

-- 6.2 文章浏览统计表
CREATE TABLE IF NOT EXISTS `article_views` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '浏览记录ID',
  `article_id` INT UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` INT UNSIGNED DEFAULT NULL COMMENT '用户ID（注册用户）',
  `visitor_id` VARCHAR(64) DEFAULT NULL COMMENT '访客标识（匿名用户）',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` TEXT COMMENT '用户代理',
  `referer` VARCHAR(500) DEFAULT NULL COMMENT '来源页面',
  `view_date` DATE NOT NULL COMMENT '浏览日期',
  `view_count` INT UNSIGNED DEFAULT 1 COMMENT '当日浏览次数',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '首次浏览时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '最后浏览时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_visitor_date` (`article_id`,`visitor_id`,`view_date`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_view_date` (`view_date`),
  KEY `idx_ip_address` (`ip_address`),
  CONSTRAINT `fk_article_views_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_views_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章浏览统计表';

-- ===================================
-- 7. 默认数据插入
-- ===================================

-- 插入默认系统配置
INSERT INTO `settings` (`key_name`, `value`, `description`, `type`, `group_name`, `is_public`) VALUES
('site_name', 'MyBlog', '网站名称', 'string', 'general', 1),
('site_description', '一个基于 Go + SvelteKit 的现代化博客系统', '网站描述', 'string', 'general', 1),
('site_keywords', 'blog,go,svelte,typescript', '网站关键词', 'string', 'seo', 1),
('site_author', 'MyBlog Team', '网站作者', 'string', 'general', 1),
('articles_per_page', '10', '每页文章数量', 'number', 'content', 1),
('comment_enabled', '1', '是否启用评论系统', 'boolean', 'comment', 1),
('comment_auto_approve', '0', '评论是否自动审核通过', 'boolean', 'comment', 0),
('upload_max_size', '10485760', '文件上传最大大小（字节）', 'number', 'media', 0),
('allowed_file_types', '["jpg","jpeg","png","gif","pdf","doc","docx"]', '允许上传的文件类型', 'json', 'media', 0),
('cache_enabled', '1', '是否启用缓存系统', 'boolean', 'cache', 0),
('cache_expire', '3600', '缓存过期时间（秒）', 'number', 'cache', 0),
('mongo_host', 'localhost', 'MongoDB服务器地址', 'string', 'cache', 0),
('mongo_port', '27017', 'MongoDB端口', 'string', 'cache', 0),
('mongo_database', 'myblog_cache', 'MongoDB数据库名', 'string', 'cache', 0),
('mongo_username', '', 'MongoDB用户名', 'string', 'cache', 0),
('mongo_password', '', 'MongoDB密码', 'string', 'cache', 0),
('mongo_auth_source', 'admin', 'MongoDB认证数据库', 'string', 'cache', 0);

-- 插入默认分类
INSERT INTO `categories` (`name`, `slug`, `description`, `sort_order`) VALUES
('技术', 'tech', '技术相关文章', 1),
('生活', 'life', '生活感悟和日常', 2),
('随笔', 'notes', '随笔和思考', 3);

-- 插入默认标签
INSERT INTO `tags` (`name`, `slug`, `color`) VALUES
('Go', 'go', '#00ADD8'),
('JavaScript', 'javascript', '#F7DF1E'),
('TypeScript', 'typescript', '#3178C6'),
('Svelte', 'svelte', '#FF3E00'),
('MySQL', 'mysql', '#4479A1'),
('MongoDB', 'mongodb', '#47A248');

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;