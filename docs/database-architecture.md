# MyBlog 数据库架构设计

## 概述

本文档定义了 MyBlog 项目的完整数据库架构设计，采用 MySQL 8.0 作为主数据库，使用 GORM 作为 ORM 框架。设计遵循博客系统的业务需求，支持用户管理、内容管理、评论系统、媒体管理、互动功能、系统监控等全面功能。

## 架构概览

**数据库规模**：21张完整功能表
- **核心业务表**：12张（用户、内容、评论、媒体、配置）
- **增强功能表**：9张（互动、监控、统计、通知）
- **迁移管理表**：1张（schema_migrations）

**最后更新**：2025年7月31日 - 完成数据库表结构全面完善

## 设计原则

1. **规范化设计** - 遵循第三范式，减少数据冗余
2. **性能优化** - 合理设计索引，支持高并发访问
3. **扩展性** - 支持水平和垂直扩展
4. **安全性** - 数据加密、权限控制、审计日志
5. **一致性** - 统一的命名规范和数据类型

## 数据库表结构设计

### 1. 用户管理模块

#### 1.1 users（用户表）

**功能**: 存储系统用户基本信息
```sql
CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名，全局唯一',
  `email` varchar(100) NOT NULL COMMENT '邮箱地址，全局唯一',
  `password` varchar(255) NOT NULL COMMENT '密码（bcrypt加密）',
  `nickname` varchar(50) DEFAULT NULL COMMENT '用户昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL',
  `bio` text COMMENT '个人简介',
  `birthday` date DEFAULT NULL COMMENT '生日',
  `role` varchar(20) NOT NULL DEFAULT 'user' COMMENT '用户角色：superadmin/admin/editor/user',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '用户状态：1-正常，0-禁用',
  `last_login_at` datetime(3) DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) DEFAULT NULL COMMENT '最后登录IP',
  `login_count` int unsigned DEFAULT '0' COMMENT '登录次数',
  `email_verified_at` datetime(3) DEFAULT NULL COMMENT '邮箱验证时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_last_login` (`last_login_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

#### 1.2 user_sessions（用户会话表）

**功能**: 管理用户登录会话和JWT令牌
```sql
CREATE TABLE `user_sessions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `refresh_token` varchar(255) NOT NULL COMMENT '刷新令牌',
  `access_token_hash` varchar(64) DEFAULT NULL COMMENT '访问令牌哈希值',
  `device_info` json DEFAULT NULL COMMENT '设备信息（浏览器、操作系统等）',
  `ip_address` varchar(45) DEFAULT NULL COMMENT '登录IP地址',
  `user_agent` text COMMENT '用户代理字符串',
  `expires_at` datetime(3) NOT NULL COMMENT '令牌过期时间',
  `last_used_at` datetime(3) DEFAULT NULL COMMENT '最后使用时间',
  `is_active` tinyint NOT NULL DEFAULT '1' COMMENT '会话状态：1-活跃，0-已注销',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refresh_token` (`refresh_token`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_ip_address` (`ip_address`),
  CONSTRAINT `fk_user_sessions_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';
```

### 2. 内容管理模块

#### 2.1 categories（分类表）

**功能**: 文章分类管理，支持层级结构
```sql
CREATE TABLE `categories` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `name` varchar(50) NOT NULL COMMENT '分类名称',
  `slug` varchar(50) NOT NULL COMMENT 'URL友好标识',
  `description` text COMMENT '分类描述',
  `cover_image` varchar(255) DEFAULT NULL COMMENT '分类封面图',
  `parent_id` int unsigned DEFAULT NULL COMMENT '父分类ID',
  `level` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '分类层级',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序权重',
  `article_count` int unsigned DEFAULT '0' COMMENT '文章数量',
  `is_featured` tinyint NOT NULL DEFAULT '0' COMMENT '是否为精选分类',
  `seo_title` varchar(100) DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` varchar(255) DEFAULT NULL COMMENT 'SEO描述',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_level` (`level`),
  KEY `idx_sort_order` (`sort_order`),
  KEY `idx_is_featured` (`is_featured`),
  KEY `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_categories_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类表';
```

#### 2.2 tags（标签表）

**功能**: 文章标签管理
```sql
CREATE TABLE `tags` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '标签ID',
  `name` varchar(30) NOT NULL COMMENT '标签名称',
  `slug` varchar(30) NOT NULL COMMENT 'URL友好标识',
  `color` varchar(7) DEFAULT '#808080' COMMENT '标签颜色（HEX格式）',
  `description` varchar(200) DEFAULT NULL COMMENT '标签描述',
  `usage_count` int unsigned DEFAULT '0' COMMENT '使用次数',
  `is_hot` tinyint NOT NULL DEFAULT '0' COMMENT '是否热门标签',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_usage_count` (`usage_count`),
  KEY `idx_is_hot` (`is_hot`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';
```

#### 2.3 articles（文章表）

**功能**: 存储博客文章内容和元信息
```sql
CREATE TABLE `articles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` varchar(200) NOT NULL COMMENT '文章标题',
  `slug` varchar(200) NOT NULL COMMENT 'URL友好标识',
  `summary` text COMMENT '文章摘要',
  `content` longtext NOT NULL COMMENT '文章内容（Markdown格式）',
  `content_html` longtext COMMENT '文章内容（HTML格式，缓存用）',
  `cover_image` varchar(500) DEFAULT NULL COMMENT '封面图片URL',
  `author_id` int unsigned NOT NULL COMMENT '作者ID',
  `category_id` int unsigned DEFAULT NULL COMMENT '主分类ID',
  `status` enum('draft','published','archived','private') NOT NULL DEFAULT 'draft' COMMENT '文章状态',
  `is_featured` tinyint NOT NULL DEFAULT '0' COMMENT '是否精选文章',
  `is_top` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶',
  `comment_enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否允许评论',
  `view_count` int unsigned DEFAULT '0' COMMENT '浏览量',
  `like_count` int unsigned DEFAULT '0' COMMENT '点赞数',
  `comment_count` int unsigned DEFAULT '0' COMMENT '评论数',
  `word_count` int unsigned DEFAULT '0' COMMENT '字数统计',
  `reading_time` int unsigned DEFAULT '0' COMMENT '预计阅读时间（分钟）',
  `seo_title` varchar(100) DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` varchar(255) DEFAULT NULL COMMENT 'SEO描述',
  `seo_keywords` varchar(200) DEFAULT NULL COMMENT 'SEO关键词',
  `published_at` datetime(3) DEFAULT NULL COMMENT '发布时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
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
```

#### 2.4 article_tags（文章标签关联表）

**功能**: 多对多关联文章和标签
```sql
CREATE TABLE `article_tags` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` int unsigned NOT NULL COMMENT '文章ID',
  `tag_id` int unsigned NOT NULL COMMENT '标签ID',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_tag` (`article_id`,`tag_id`),
  KEY `idx_tag_id` (`tag_id`),
  CONSTRAINT `fk_article_tags_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_tags_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签关联表';
```

### 3. 评论系统模块

#### 3.1 comments（评论表）

**功能**: 存储文章评论，支持多级回复
```sql
CREATE TABLE `comments` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `article_id` int unsigned NOT NULL COMMENT '文章ID',
  `user_id` int unsigned DEFAULT NULL COMMENT '用户ID（注册用户）',
  `parent_id` int unsigned DEFAULT NULL COMMENT '父评论ID（回复功能）',
  `root_id` int unsigned DEFAULT NULL COMMENT '根评论ID（便于查询评论树）',
  `level` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '评论层级',
  `author_name` varchar(50) DEFAULT NULL COMMENT '游客姓名',
  `author_email` varchar(100) DEFAULT NULL COMMENT '游客邮箱',
  `author_website` varchar(255) DEFAULT NULL COMMENT '游客网站',
  `author_ip` varchar(45) DEFAULT NULL COMMENT '评论者IP地址',
  `content` text NOT NULL COMMENT '评论内容',
  `content_html` text COMMENT '评论内容（HTML格式，缓存用）',
  `status` enum('pending','approved','rejected','spam','trash') NOT NULL DEFAULT 'pending' COMMENT '审核状态',
  `like_count` int unsigned DEFAULT '0' COMMENT '点赞数',
  `reply_count` int unsigned DEFAULT '0' COMMENT '回复数量',
  `user_agent` text COMMENT '用户代理',
  `is_author` tinyint NOT NULL DEFAULT '0' COMMENT '是否为文章作者回复',
  `is_pinned` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶评论',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
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
```

### 4. 媒体管理模块

#### 4.1 media_files（媒体文件表）

**功能**: 管理上传的图片、文档等媒体文件
```sql
CREATE TABLE `media_files` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `filename` varchar(255) NOT NULL COMMENT '原始文件名',
  `stored_name` varchar(255) NOT NULL COMMENT '存储文件名（UUID）',
  `file_path` varchar(500) NOT NULL COMMENT '文件存储路径',
  `file_url` varchar(500) NOT NULL COMMENT '文件访问URL',
  `thumbnail_url` varchar(500) DEFAULT NULL COMMENT '缩略图URL',
  `mime_type` varchar(100) NOT NULL COMMENT 'MIME类型',
  `file_size` bigint unsigned NOT NULL COMMENT '文件大小（字节）',
  `file_hash` varchar(64) DEFAULT NULL COMMENT '文件SHA256哈希值',
  `width` int unsigned DEFAULT NULL COMMENT '图片宽度',
  `height` int unsigned DEFAULT NULL COMMENT '图片高度',
  `alt_text` varchar(255) DEFAULT NULL COMMENT '替代文本（SEO用）',
  `uploader_id` int unsigned NOT NULL COMMENT '上传者ID',
  `upload_ip` varchar(45) DEFAULT NULL COMMENT '上传IP地址',
  `storage_type` enum('local','oss','s3','cos') NOT NULL DEFAULT 'local' COMMENT '存储类型',
  `folder` varchar(100) DEFAULT NULL COMMENT '文件夹分类',
  `usage_count` int unsigned DEFAULT '0' COMMENT '使用次数',
  `is_public` tinyint NOT NULL DEFAULT '1' COMMENT '是否公开访问',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
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
```

### 5. 系统配置模块

#### 5.1 settings（系统设置表）

**功能**: 存储系统全局配置信息
```sql
CREATE TABLE `settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '设置ID',
  `key_name` varchar(100) NOT NULL COMMENT '配置键名',
  `value` longtext COMMENT '配置值（支持JSON格式）',
  `default_value` longtext COMMENT '默认值',
  `description` varchar(255) DEFAULT NULL COMMENT '配置描述',
  `type` enum('string','number','boolean','json','array') NOT NULL DEFAULT 'string' COMMENT '值类型',
  `group_name` varchar(50) NOT NULL DEFAULT 'general' COMMENT '配置分组',
  `is_public` tinyint NOT NULL DEFAULT '0' COMMENT '是否公开（前端可访问）',
  `is_readonly` tinyint NOT NULL DEFAULT '0' COMMENT '是否只读',
  `validation_rule` varchar(200) DEFAULT NULL COMMENT '验证规则',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序权重',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key_name` (`key_name`),
  KEY `idx_group_name` (`group_name`),
  KEY `idx_is_public` (`is_public`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';
```

### 6. 统计和日志模块

#### 6.1 user_activities（用户活动日志表）

**功能**: 记录用户重要操作日志
```sql
CREATE TABLE `user_activities` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '活动ID',
  `user_id` int unsigned DEFAULT NULL COMMENT '用户ID',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) DEFAULT NULL COMMENT '资源类型（article、comment等）',
  `resource_id` int unsigned DEFAULT NULL COMMENT '资源ID',
  `description` varchar(255) DEFAULT NULL COMMENT '操作描述',
  `metadata` json DEFAULT NULL COMMENT '额外元数据',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COMMENT '用户代理',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_resource` (`resource_type`,`resource_id`),
  KEY `idx_ip_address` (`ip_address`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_user_activities_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户活动日志表';
```

#### 6.2 article_views（文章浏览统计表）

**功能**: 记录文章浏览详细统计
```sql
CREATE TABLE `article_views` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '浏览记录ID',
  `article_id` int unsigned NOT NULL COMMENT '文章ID',
  `user_id` int unsigned DEFAULT NULL COMMENT '用户ID（注册用户）',
  `visitor_id` varchar(64) DEFAULT NULL COMMENT '访客标识（匿名用户）',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COMMENT '用户代理',
  `referer` varchar(500) DEFAULT NULL COMMENT '来源页面',
  `view_date` date NOT NULL COMMENT '浏览日期',
  `view_count` int unsigned DEFAULT '1' COMMENT '当日浏览次数',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '首次浏览时间',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '最后浏览时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_visitor_date` (`article_id`,`visitor_id`,`view_date`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_view_date` (`view_date`),
  KEY `idx_ip_address` (`ip_address`),
  CONSTRAINT `fk_article_views_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_views_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章浏览统计表';
```

### 7. 增强功能模块

#### 7.1 article_categories（文章分类关联表）

**功能**: 支持文章多分类功能，文章可以属于多个分类
```sql
CREATE TABLE `article_categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `category_id` bigint unsigned NOT NULL COMMENT '分类ID',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_category` (`article_id`,`category_id`),
  KEY `idx_article_id` (`article_id`),
  KEY `idx_category_id` (`category_id`),
  CONSTRAINT `fk_article_categories_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_categories_category_id` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类关联表';
```

#### 7.2 operation_logs（操作日志表）

**功能**: 记录系统敏感操作，用于安全审计和问题追踪
```sql
CREATE TABLE `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '操作用户ID',
  `action` varchar(100) NOT NULL COMMENT '操作类型：login, logout, create_user, delete_article等',
  `resource_type` varchar(50) DEFAULT NULL COMMENT '资源类型：user, article, comment等',
  `resource_id` bigint unsigned DEFAULT NULL COMMENT '资源ID',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COMMENT '用户代理',
  `details` json DEFAULT NULL COMMENT '操作详情',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_action` (`user_id`,`action`),
  KEY `idx_resource` (`resource_type`,`resource_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_action` (`action`),
  CONSTRAINT `fk_operation_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
```

#### 7.3 article_likes（文章点赞表）

**功能**: 记录用户对文章的点赞行为
```sql
CREATE TABLE `article_likes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user_like` (`article_id`,`user_id`),
  KEY `idx_article_id` (`article_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_article_likes_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_likes_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章点赞表';
```

#### 7.4 comment_likes（评论点赞表）

**功能**: 记录用户对评论的点赞行为
```sql
CREATE TABLE `comment_likes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `comment_id` bigint unsigned NOT NULL COMMENT '评论ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_comment_user_like` (`comment_id`,`user_id`),
  KEY `idx_comment_id` (`comment_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_comment_likes_comment_id` FOREIGN KEY (`comment_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comment_likes_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论点赞表';
```

#### 7.5 article_bookmarks（文章收藏表）

**功能**: 用户文章收藏功能
```sql
CREATE TABLE `article_bookmarks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '收藏ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '收藏时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user_bookmark` (`article_id`,`user_id`),
  KEY `idx_article_id` (`article_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_article_bookmarks_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_bookmarks_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章收藏表';
```

#### 7.6 notifications（系统通知表）

**功能**: 系统通知管理，支持多种通知类型
```sql
CREATE TABLE `notifications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `user_id` bigint unsigned NOT NULL COMMENT '接收用户ID',
  `type` varchar(50) NOT NULL COMMENT '通知类型：comment_reply, article_like, system等',
  `title` varchar(255) NOT NULL COMMENT '通知标题',
  `content` text COMMENT '通知内容',
  `related_type` varchar(50) DEFAULT NULL COMMENT '关联资源类型',
  `related_id` bigint unsigned DEFAULT NULL COMMENT '关联资源ID',
  `is_read` tinyint(1) DEFAULT '0' COMMENT '是否已读',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_read` (`user_id`,`is_read`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_type` (`type`),
  KEY `idx_related` (`related_type`,`related_id`),
  CONSTRAINT `fk_notifications_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表';
```

#### 7.7 search_logs（搜索记录表）

**功能**: 记录用户搜索行为，支持搜索分析和优化
```sql
CREATE TABLE `search_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '搜索记录ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '搜索用户ID（游客为NULL）',
  `keyword` varchar(255) NOT NULL COMMENT '搜索关键词',
  `results_count` int DEFAULT '0' COMMENT '搜索结果数量',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COMMENT '用户代理',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '搜索时间',
  PRIMARY KEY (`id`),
  KEY `idx_keyword` (`keyword`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_ip_address` (`ip_address`),
  CONSTRAINT `fk_search_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索记录表';
```

#### 7.8 content_stats（内容统计表）

**功能**: 多维度内容统计，支持热门推荐和性能优化
```sql
CREATE TABLE `content_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '统计ID',
  `content_type` varchar(50) NOT NULL COMMENT '内容类型：article, tag, category',
  `content_id` bigint unsigned NOT NULL COMMENT '内容ID',
  `stat_type` varchar(50) NOT NULL COMMENT '统计类型：daily_views, weekly_views, monthly_views, likes_count',
  `stat_value` bigint unsigned DEFAULT '0' COMMENT '统计值',
  `stat_date` date NOT NULL COMMENT '统计日期',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_content_stat` (`content_type`,`content_id`,`stat_type`,`stat_date`),
  KEY `idx_stat_date` (`stat_date`),
  KEY `idx_content` (`content_type`,`content_id`),
  KEY `idx_stat_type` (`stat_type`),
  KEY `idx_stat_value` (`stat_value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容统计表';
```

#### 7.9 user_follows（用户关注表）

**功能**: 用户互关功能，支持社交化扩展
```sql
CREATE TABLE `user_follows` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关注关系ID',
  `follower_id` bigint unsigned NOT NULL COMMENT '关注者ID',
  `following_id` bigint unsigned NOT NULL COMMENT '被关注者ID',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '关注时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follow_relation` (`follower_id`,`following_id`),
  KEY `idx_follower_id` (`follower_id`),
  KEY `idx_following_id` (`following_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_user_follows_follower_id` FOREIGN KEY (`follower_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_follows_following_id` FOREIGN KEY (`following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `chk_user_follows_self` CHECK ((`follower_id` <> `following_id`))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注表';
```

## 数据关系图

### 核心业务关系
```
用户系统:
users (1) ←→ (N) user_sessions
users (1) ←→ (N) articles
users (1) ←→ (N) comments
users (1) ←→ (N) media_files
users (1) ←→ (N) user_activities

内容系统:
categories (1) ←→ (N) categories (自关联)
categories (1) ←→ (N) articles
articles (1) ←→ (N) comments
articles (N) ←→ (N) tags (通过 article_tags)
articles (N) ←→ (N) categories (通过 article_categories) ⭐新增
articles (1) ←→ (N) article_views

评论系统:
comments (1) ←→ (N) comments (自关联，支持多级回复)

统计系统:
articles (1) ←→ (N) article_views
users (1) ←→ (N) user_activities
```

### 增强功能关系
```
互动功能:
users (1) ←→ (N) article_likes ←→ (1) articles ⭐新增
users (1) ←→ (N) comment_likes ←→ (1) comments ⭐新增
users (1) ←→ (N) article_bookmarks ←→ (1) articles ⭐新增
users (N) ←→ (N) users (通过 user_follows, 关注关系) ⭐新增

通知系统:
users (1) ←→ (N) notifications (用户通知) ⭐新增

系统监控:
users (1) ←→ (N) operation_logs (操作审计) ⭐新增
users (1) ←→ (N) search_logs (搜索记录) ⭐新增
articles/tags/categories ←→ content_stats (内容统计) ⭐新增
```

### 完整关系总览
```
核心实体: users, articles, comments, categories, tags
关联表: article_tags, article_categories, user_follows
互动表: article_likes, comment_likes, article_bookmarks
监控表: operation_logs, search_logs, notifications
统计表: article_views, content_stats, user_activities
系统表: user_sessions, media_files, settings
```

## 索引设计策略

### 1. 主键索引
所有表都使用 `id` 作为主键，自增整型

### 2. 唯一索引
- `users.username`, `users.email`
- `categories.slug`, `articles.slug`, `tags.slug`
- `user_sessions.refresh_token`
- `media_files.stored_name`

### 3. 复合索引
**原有复合索引**：
- `article_tags(article_id, tag_id)` - 文章标签查询
- `comments(article_id, status, created_at)` - 文章评论分页
- `article_views(article_id, visitor_id, view_date)` - 访客统计去重

**新增复合索引**：
- `article_categories(article_id, category_id)` - 文章多分类查询 ⭐新增
- `operation_logs(user_id, action)` - 用户操作日志查询 ⭐新增
- `article_likes(article_id, user_id)` - 文章点赞去重 ⭐新增
- `comment_likes(comment_id, user_id)` - 评论点赞去重 ⭐新增
- `article_bookmarks(article_id, user_id)` - 文章收藏去重 ⭐新增
- `notifications(user_id, is_read)` - 用户未读通知查询 ⭐新增
- `user_follows(follower_id, following_id)` - 关注关系查询 ⭐新增
- `content_stats(content_type, content_id, stat_type, stat_date)` - 统计数据唯一性 ⭐新增

### 4. 覆盖索引
**原有覆盖索引**：
- `articles(status, published_at)` - 已发布文章列表
- `user_activities(user_id, created_at)` - 用户活动时间轴

**新增覆盖索引**：
- `operation_logs(resource_type, resource_id)` - 资源操作日志 ⭐新增
- `notifications(type, created_at)` - 通知类型时间排序 ⭐新增
- `content_stats(stat_date, stat_value)` - 统计数据排序 ⭐新增

## 数据库性能优化

### 1. 读写分离
```yaml
# 数据库连接配置示例
database:
  master:
    host: "master.mysql.local"
    port: 3306
  slaves:
    - host: "slave1.mysql.local"
      port: 3306
    - host: "slave2.mysql.local"
      port: 3306
```

### 2. 分表策略
```sql
-- 按月分区用户活动日志表（适用于高频写入场景）
ALTER TABLE user_activities 
PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
    PARTITION p202501 VALUES LESS THAN (202502),
    PARTITION p202502 VALUES LESS THAN (202503),
    PARTITION p202503 VALUES LESS THAN (202504)
    -- 可以继续添加更多分区
);
```

### 3. 缓存策略
- **MongoDB 缓存**：热点文章、用户会话、系统配置
- **本地缓存**：不经常变化的配置信息
- **CDN 缓存**：静态媒体文件

## 安全设计

### 1. 数据加密
- 用户密码：bcrypt 哈希加密（成本系数12）
- 敏感配置：环境变量 + Vault 管理
- 传输加密：全站 HTTPS

### 2. 权限控制
- 基于角色的访问控制（RBAC）
- API 频率限制
- IP 白名单/黑名单

### 3. 审计日志
- 用户操作日志记录
- 系统配置变更记录
- 登录失败记录

## 备份和恢复

### 1. 备份策略
```bash
# 全量备份（每日）
mysqldump --single-transaction --routines --triggers myblog > backup_$(date +%Y%m%d).sql

# 增量备份（实时）
mysqlbinlog --read-from-remote-server --host=localhost --raw --result-file=/backup/binlog/
```

### 2. 数据恢复
```bash
# 恢复全量备份
mysql myblog < backup_20250128.sql

# 恢复增量数据
mysqlbinlog binlog.000001 | mysql myblog
```

## 数据迁移计划

### 阶段一：核心功能迁移（✅ 已完成）
- ✅ 用户管理系统
- ✅ 权限控制系统  
- ✅ JWT 会话管理

### 阶段二：基础内容管理（✅ 已完成）
- ✅ 文章分类系统
- ✅ 文章内容管理
- ✅ 标签系统
- ✅ 评论系统
- ✅ 媒体文件管理

### 阶段三：数据库架构完善（✅ 已完成 - 2025年7月31日）
- ✅ 文章多分类支持（article_categories）
- ✅ 互动功能表（点赞、收藏、关注）
- ✅ 系统监控表（操作日志、搜索记录、通知）
- ✅ 统计分析表（内容统计、用户行为）
- ✅ GORM模型定义和关联关系
- ✅ 数据库迁移脚本（000002_enhance_database_tables）

### 阶段四：业务逻辑开发（⏳ 下一阶段）
- ⏳ 文章管理API接口
- ⏳ 互动功能API（点赞、收藏、关注）
- ⏳ 通知系统API
- ⏳ 搜索功能API
- ⏳ 统计分析API

### 阶段五：性能优化和扩展（⏳ 计划中）
- ⏳ MongoDB缓存集成
- ⏳ 搜索引擎集成（Elasticsearch）
- ⏳ 分表分库策略
- ⏳ CDN和静态资源优化

## 维护和监控

### 1. 数据库监控
- 连接数监控
- 慢查询分析
- 存储空间监控

### 2. 性能优化
- 定期分析慢查询日志
- 索引使用情况分析
- 表空间优化

### 3. 数据清理
- 定期清理过期会话
- 压缩历史日志数据
- 清理软删除数据

## 总结

本数据库架构设计文档涵盖了 MyBlog 项目的完整数据结构，从用户管理到内容管理，从权限控制到统计分析，从互动功能到系统监控，提供了一个现代化、可扩展、高性能、安全的数据库解决方案。

### 架构特色

**完整性**：
- **21张功能表**：覆盖博客系统的所有核心和扩展功能
- **模块化设计**：7大功能模块，清晰的职责划分
- **关系完整**：完善的外键约束和级联策略

**现代化**：
- **互动功能**：点赞、收藏、关注等社交化功能
- **智能统计**：多维度内容统计和用户行为分析
- **通知系统**：实时通知和消息推送支持
- **审计日志**：完整的操作记录和安全监控

**高性能**：
- **索引优化**：针对查询场景的复合索引和覆盖索引
- **数据类型**：统一使用高效的 BIGINT UNSIGNED 和 DATETIME(3)
- **缓存友好**：支持 MongoDB 和多级缓存策略
- **分片准备**：为高并发场景预留分表分库能力

**安全可靠**：
- **权限控制**：基于角色的访问控制（RBAC）
- **数据加密**：密码哈希、传输加密、敏感数据保护
- **审计追踪**：操作日志、用户活动、系统监控
- **备份恢复**：完整的数据备份和灾难恢复方案

### 实施状态

✅ **数据库架构**：21张表全部创建完成  
✅ **迁移管理**：golang-migrate 版本控制  
✅ **GORM模型**：完整的模型定义和关联关系  
✅ **索引优化**：高性能查询索引设计  
✅ **文档完善**：详细的架构设计和维护指南  

### 下一步规划

该架构为 MyBlog 项目提供了坚实的数据基础，支持：
- **快速开发**：完整的模型定义，加速业务逻辑开发
- **功能扩展**：预留的扩展点，支持新功能无缝集成
- **性能扩展**：分片、缓存、读写分离等优化方案
- **运维监控**：完善的监控指标和维护工具

随着业务发展，可以基于这个架构进行渐进式优化和功能扩展，确保系统的长期稳定运行。

---

**文档版本**：v2.0  
**最后更新**：2025年7月31日  
**更新内容**：完成数据库表结构全面完善，新增9张增强功能表