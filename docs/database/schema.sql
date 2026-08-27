-- MyBlog 数据库架构设计 SQL 脚本
-- 用途: 数据库设计文档和参考实现，共 23 张业务表。
-- 注意: 此文件仅作为文档参考，实际数据库结构由 GORM 模型自动迁移管理。
-- 约定: 时间字段统一 datetime(3) 精度；状态类字段使用命名常量枚举；
--       树形结构使用 parent_id、root_id、level 三件套并辅以 path 物化路径；
--       多对多关系使用独立关联表并声明复合唯一索引。

-- 设置字符集和校对规则
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ===================================
-- 1. 用户管理模块
-- ===================================

-- 1.1 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名，全局唯一',
  `email` VARCHAR(100) NOT NULL COMMENT '邮箱地址，全局唯一',
  `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号，全局唯一，未绑定时为空',
  `password` VARCHAR(255) NOT NULL COMMENT '密码，存储 bcrypt 哈希值',
  `nickname` VARCHAR(50) DEFAULT NULL COMMENT '用户昵称，为空时展示用户名',
  `avatar` VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
  `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '个人主页封面图URL',
  `bio` TEXT DEFAULT NULL COMMENT '个人简介',
  `website` VARCHAR(255) DEFAULT NULL COMMENT '个人网站URL',
  `location` VARCHAR(100) DEFAULT NULL COMMENT '常居地描述',
  `gender` TINYINT DEFAULT NULL COMMENT '性别：0-未知 1-男 2-女',
  `birthday` DATE DEFAULT NULL COMMENT '生日',
  `timezone` VARCHAR(50) DEFAULT 'Asia/Shanghai' COMMENT '用户时区标识，IANA 命名格式',
  `locale` VARCHAR(10) DEFAULT 'zh-CN' COMMENT '用户界面语言标识，BCP 47 格式',
  `role` VARCHAR(20) DEFAULT 'user' COMMENT '用户角色：superadmin/admin/editor/user',
  `status` TINYINT DEFAULT 1 COMMENT '用户状态：1-正常 0-禁用 2-锁定',
  `failed_login_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '连续登录失败次数，登录成功后清零',
  `locked_until` DATETIME(3) DEFAULT NULL COMMENT '账户锁定截止时间，到期后可重新登录',
  `password_changed_at` DATETIME(3) DEFAULT NULL COMMENT '密码最后修改时间，用于安全审计',
  `last_login_at` DATETIME(3) DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(45) DEFAULT NULL COMMENT '最后登录IP，IPv6 最长 45 字符',
  `login_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '累计登录成功次数',
  `email_verified_at` DATETIME(3) DEFAULT NULL COMMENT '邮箱验证完成时间，为空表示未验证',
  `remark` VARCHAR(500) DEFAULT NULL COMMENT '管理员备注，仅管理端可见',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_phone` (`phone`),
  KEY `idx_users_role` (`role`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_last_login_at` (`last_login_at`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表，存储账号身份、个人资料与安全状态';

-- 1.2 用户会话表
CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `refresh_token` VARCHAR(255) NOT NULL COMMENT '刷新令牌',
  `access_token_hash` VARCHAR(64) DEFAULT NULL COMMENT '访问令牌哈希值',
  `device_info` JSON DEFAULT NULL COMMENT '设备信息，包含浏览器与操作系统等',
  `device_type` VARCHAR(20) DEFAULT 'web' COMMENT '设备类型：web/mobile/tablet/desktop',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT '登录IP地址',
  `user_agent` TEXT DEFAULT NULL COMMENT '用户代理字符串',
  `expires_at` DATETIME(3) NOT NULL COMMENT '令牌过期时间',
  `last_used_at` DATETIME(3) DEFAULT NULL COMMENT '最后使用时间',
  `last_refresh_at` DATETIME(3) DEFAULT NULL COMMENT '刷新令牌最近轮换时间',
  `logout_at` DATETIME(3) DEFAULT NULL COMMENT '会话注销时间',
  `is_active` TINYINT(1) DEFAULT 1 COMMENT '会话状态：1-活跃，0-已注销',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_sessions_refresh_token` (`refresh_token`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  KEY `idx_user_sessions_device_type` (`device_type`),
  KEY `idx_user_sessions_ip_address` (`ip_address`),
  KEY `idx_user_sessions_expires_at` (`expires_at`),
  KEY `idx_user_sessions_is_active` (`is_active`),
  KEY `idx_user_active` (`user_id`,`is_active`),
  CONSTRAINT `fk_user_sessions_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表，管理登录设备与令牌轮换';

-- 1.3 用户活动日志表
CREATE TABLE IF NOT EXISTS `user_activities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '用户ID',
  `action` VARCHAR(50) NOT NULL COMMENT '操作类型',
  `resource_type` VARCHAR(50) DEFAULT NULL COMMENT '资源类型：article/comment/user 等',
  `resource_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '资源ID',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '操作描述',
  `status` VARCHAR(20) DEFAULT 'success' COMMENT '执行结果：success-成功 failed-失败',
  `error_message` VARCHAR(500) DEFAULT NULL COMMENT '失败原因，成功时为空',
  `duration_ms` BIGINT UNSIGNED DEFAULT 0 COMMENT '操作耗时，单位毫秒',
  `metadata` JSON DEFAULT NULL COMMENT '额外元数据',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` TEXT DEFAULT NULL COMMENT '用户代理',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_activities_user_id` (`user_id`),
  KEY `idx_user_activities_action` (`action`),
  KEY `idx_user_activities_ip_address` (`ip_address`),
  KEY `idx_user_activities_created_at` (`created_at`),
  KEY `idx_user_created` (`user_id`,`created_at`),
  CONSTRAINT `fk_user_activities_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户活动日志表，记录用户行为轨迹';

-- 1.4 认证令牌表
CREATE TABLE IF NOT EXISTS `auth_tokens` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '令牌ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '所属用户ID',
  `token_hash` VARCHAR(64) NOT NULL COMMENT '令牌哈希值，SHA256 十六进制',
  `token_type` VARCHAR(30) DEFAULT 'password_reset' COMMENT '令牌用途：password_reset-密码重置 email_verify-邮箱验证',
  `expires_at` DATETIME(3) NOT NULL COMMENT '令牌过期时间',
  `used_at` DATETIME(3) DEFAULT NULL COMMENT '令牌核销时间，为空表示未使用',
  `request_ip` VARCHAR(45) DEFAULT NULL COMMENT '申请令牌时的来源IP，用于安全审计',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_auth_tokens_token_hash` (`token_hash`),
  KEY `idx_auth_tokens_user_id` (`user_id`),
  KEY `idx_auth_tokens_token_type` (`token_type`),
  KEY `idx_auth_tokens_expires_at` (`expires_at`),
  CONSTRAINT `fk_auth_tokens_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证令牌表，支撑密码找回与邮箱验证流程';

-- ===================================
-- 2. 内容管理模块
-- ===================================

-- 2.1 文章分类表
CREATE TABLE IF NOT EXISTS `categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `name` VARCHAR(50) NOT NULL COMMENT '分类名称',
  `slug` VARCHAR(50) NOT NULL COMMENT 'URL友好标识',
  `description` TEXT DEFAULT NULL COMMENT '分类描述',
  `cover_image` VARCHAR(255) DEFAULT NULL COMMENT '分类封面图',
  `parent_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父分类ID，顶级分类为空',
  `root_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '根分类ID，用于整棵子树的聚合查询',
  `level` TINYINT UNSIGNED DEFAULT 1 COMMENT '分类层级，顶级为 1',
  `path` VARCHAR(100) DEFAULT NULL COMMENT '物化路径，形如 /1/5/12，用于一次查询取整棵子树',
  `sort_order` INT DEFAULT 0 COMMENT '排序权重，数值小的靠前',
  `status` TINYINT DEFAULT 1 COMMENT '分类状态：1-显示 0-隐藏',
  `article_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '文章数量，发布时异步维护',
  `is_featured` TINYINT(1) DEFAULT 0 COMMENT '是否为精选分类',
  `seo_title` VARCHAR(100) DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` VARCHAR(255) DEFAULT NULL COMMENT 'SEO描述',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_categories_slug` (`slug`),
  KEY `idx_categories_parent_id` (`parent_id`),
  KEY `idx_categories_root_id` (`root_id`),
  KEY `idx_categories_level` (`level`),
  KEY `idx_categories_sort_order` (`sort_order`),
  KEY `idx_categories_status` (`status`),
  KEY `idx_categories_is_featured` (`is_featured`),
  KEY `idx_categories_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_categories_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类表，树形结构支撑栏目导航';

-- 2.2 标签表
CREATE TABLE IF NOT EXISTS `tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '标签ID',
  `name` VARCHAR(30) NOT NULL COMMENT '标签名称',
  `slug` VARCHAR(30) NOT NULL COMMENT 'URL友好标识',
  `color` VARCHAR(7) DEFAULT '#808080' COMMENT '标签颜色，HEX 格式',
  `description` VARCHAR(200) DEFAULT NULL COMMENT '标签描述',
  `status` TINYINT DEFAULT 1 COMMENT '标签状态：1-启用 0-隐藏',
  `usage_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '使用次数，文章挂载时异步维护',
  `is_hot` TINYINT(1) DEFAULT 0 COMMENT '是否热门标签',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tags_name` (`name`),
  UNIQUE KEY `uk_tags_slug` (`slug`),
  KEY `idx_tags_status` (`status`),
  KEY `idx_tags_usage_count` (`usage_count`),
  KEY `idx_tags_is_hot` (`is_hot`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表，文章主题的轻量归类维度';

-- 2.3 文章表
CREATE TABLE IF NOT EXISTS `articles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` VARCHAR(200) NOT NULL COMMENT '文章标题',
  `slug` VARCHAR(200) NOT NULL COMMENT 'URL友好标识',
  `summary` TEXT DEFAULT NULL COMMENT '文章摘要',
  `content` LONGTEXT NOT NULL COMMENT '文章内容，Markdown 格式',
  `content_html` LONGTEXT DEFAULT NULL COMMENT '文章内容，渲染后的 HTML 缓存',
  `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '封面图片URL',
  `author_id` BIGINT UNSIGNED NOT NULL COMMENT '作者ID',
  `category_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '主分类ID',
  `status` VARCHAR(20) DEFAULT 'draft' COMMENT '文章状态：draft/published/archived/private',
  `origin_type` VARCHAR(20) DEFAULT 'original' COMMENT '来源类型：original-原创 translation-翻译 reprint-转载',
  `source_url` VARCHAR(500) DEFAULT NULL COMMENT '原文链接，原创文章为空',
  `source_author` VARCHAR(50) DEFAULT NULL COMMENT '原文作者，原创文章为空',
  `access_password` VARCHAR(255) DEFAULT NULL COMMENT '访问密码哈希，仅私密文章生效，为空表示仅登录可见',
  `is_featured` TINYINT(1) DEFAULT 0 COMMENT '是否精选文章',
  `is_top` TINYINT(1) DEFAULT 0 COMMENT '是否置顶',
  `comment_enabled` TINYINT(1) DEFAULT 1 COMMENT '是否允许评论',
  `view_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '浏览量',
  `like_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '点赞数',
  `bookmark_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '收藏数',
  `comment_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '评论数',
  `word_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '字数统计',
  `reading_time` BIGINT UNSIGNED DEFAULT 0 COMMENT '预计阅读时间，单位分钟',
  `version` BIGINT UNSIGNED DEFAULT 1 COMMENT '内容版本号，每次保存正文递增，对应 article_revisions',
  `seo_title` VARCHAR(100) DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` VARCHAR(255) DEFAULT NULL COMMENT 'SEO描述',
  `seo_keywords` VARCHAR(200) DEFAULT NULL COMMENT 'SEO关键词',
  `scheduled_at` DATETIME(3) DEFAULT NULL COMMENT '定时发布时间，到期后由调度任务发布',
  `published_at` DATETIME(3) DEFAULT NULL COMMENT '发布时间',
  `edited_at` DATETIME(3) DEFAULT NULL COMMENT '正文最后编辑时间，用于展示已编辑标记',
  `archived_at` DATETIME(3) DEFAULT NULL COMMENT '归档时间，进入归档状态时写入',
  `last_comment_at` DATETIME(3) DEFAULT NULL COMMENT '最新评论时间，用于评论排序展示',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_articles_slug` (`slug`),
  KEY `idx_articles_author_id` (`author_id`),
  KEY `idx_articles_category_id` (`category_id`),
  KEY `idx_articles_status` (`status`),
  KEY `idx_articles_origin_type` (`origin_type`),
  KEY `idx_articles_is_featured` (`is_featured`),
  KEY `idx_articles_is_top` (`is_top`),
  KEY `idx_articles_view_count` (`view_count`),
  KEY `idx_articles_scheduled_at` (`scheduled_at`),
  KEY `idx_articles_published_at` (`published_at`),
  KEY `idx_articles_deleted_at` (`deleted_at`),
  KEY `idx_status_published` (`status`,`published_at`),
  KEY `idx_author_status` (`author_id`,`status`),
  CONSTRAINT `fk_articles_author_id` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_articles_category_id` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表，博客核心内容实体';

-- 2.4 文章标签关联表
CREATE TABLE IF NOT EXISTS `article_tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_tag` (`article_id`,`tag_id`),
  CONSTRAINT `fk_article_tags_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_tags_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签关联表，多对多挂载关系';

-- 2.5 文章分类关联表
CREATE TABLE IF NOT EXISTS `article_categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `category_id` BIGINT UNSIGNED NOT NULL COMMENT '分类ID',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_category` (`article_id`,`category_id`),
  KEY `idx_article_categories_article_id` (`article_id`),
  KEY `idx_article_categories_category_id` (`category_id`),
  CONSTRAINT `fk_article_categories_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_categories_category_id` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类关联表，支持一文多分类';

-- 2.6 文章浏览统计表
CREATE TABLE IF NOT EXISTS `article_views` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '浏览记录ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '用户ID，注册用户填写',
  `visitor_id` VARCHAR(64) DEFAULT NULL COMMENT '访客标识，匿名用户填写',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` TEXT DEFAULT NULL COMMENT '用户代理',
  `referer` VARCHAR(500) DEFAULT NULL COMMENT '来源页面',
  `view_date` DATE DEFAULT NULL COMMENT '浏览日期',
  `view_count` BIGINT UNSIGNED DEFAULT 1 COMMENT '当日浏览次数',
  `duration_seconds` BIGINT UNSIGNED DEFAULT 0 COMMENT '页面停留时长，单位秒，由前端埋点上报',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '首次浏览时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '最后浏览时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_visitor_date` (`article_id`,`visitor_id`,`view_date`),
  KEY `idx_article_views_user_id` (`user_id`),
  KEY `idx_article_views_ip_address` (`ip_address`),
  CONSTRAINT `fk_article_views_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_views_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章浏览统计表，按访客与日期去重计数';

-- 2.7 文章修订历史表
CREATE TABLE IF NOT EXISTS `article_revisions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '修订ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `revision_no` BIGINT UNSIGNED NOT NULL COMMENT '修订版本号，从 1 开始随文章 version 递增',
  `title` VARCHAR(200) DEFAULT NULL COMMENT '该版本文章标题快照',
  `summary` TEXT DEFAULT NULL COMMENT '该版本摘要快照',
  `content` LONGTEXT NOT NULL COMMENT '该版本正文快照，Markdown 格式',
  `content_html` LONGTEXT DEFAULT NULL COMMENT '该版本渲染后 HTML 快照',
  `word_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '该版本字数统计',
  `change_summary` VARCHAR(255) DEFAULT NULL COMMENT '变更说明，由作者填写',
  `editor_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '执行本次修订的用户ID',
  `is_autosave` TINYINT(1) DEFAULT 0 COMMENT '是否编辑器自动保存产生的快照',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '修订时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_revision` (`article_id`,`revision_no`),
  KEY `idx_article_revisions_article_id` (`article_id`),
  KEY `idx_article_revisions_editor_id` (`editor_id`),
  KEY `idx_article_revisions_is_autosave` (`is_autosave`),
  KEY `idx_article_revisions_created_at` (`created_at`),
  CONSTRAINT `fk_article_revisions_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_revisions_editor_id` FOREIGN KEY (`editor_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章修订历史表，保存正文快照支持回滚';

-- ===================================
-- 3. 评论系统模块
-- ===================================

-- 3.1 评论表
CREATE TABLE IF NOT EXISTS `comments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '用户ID，注册用户填写',
  `parent_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '父评论ID，根评论为空',
  `root_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '根评论ID，便于一次查询整棵评论树',
  `level` TINYINT UNSIGNED DEFAULT 1 COMMENT '评论层级，根评论为 1',
  `author_name` VARCHAR(50) DEFAULT NULL COMMENT '游客姓名',
  `author_email` VARCHAR(100) DEFAULT NULL COMMENT '游客邮箱',
  `author_website` VARCHAR(255) DEFAULT NULL COMMENT '游客网站',
  `author_ip` VARCHAR(45) DEFAULT NULL COMMENT '评论者IP地址，用于反垃圾与封禁',
  `content` TEXT NOT NULL COMMENT '评论内容，Markdown 格式',
  `content_html` TEXT DEFAULT NULL COMMENT '评论内容，渲染后的 HTML 缓存',
  `status` VARCHAR(20) DEFAULT 'pending' COMMENT '审核状态：pending/approved/rejected/spam/trash',
  `like_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '点赞数',
  `reply_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '回复数量',
  `reported_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '被举报次数，达到阈值后进入待复核队列',
  `user_agent` TEXT DEFAULT NULL COMMENT '用户代理',
  `is_author` TINYINT(1) DEFAULT 0 COMMENT '是否为文章作者回复',
  `is_pinned` TINYINT(1) DEFAULT 0 COMMENT '是否置顶评论',
  `edited_at` DATETIME(3) DEFAULT NULL COMMENT '内容最后编辑时间，用于展示已编辑标记',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_comments_article_id` (`article_id`),
  KEY `idx_comments_user_id` (`user_id`),
  KEY `idx_comments_parent_id` (`parent_id`),
  KEY `idx_comments_root_id` (`root_id`),
  KEY `idx_comments_status` (`status`),
  KEY `idx_comments_author_ip` (`author_ip`),
  KEY `idx_comments_is_pinned` (`is_pinned`),
  KEY `idx_comments_created_at` (`created_at`),
  KEY `idx_comments_deleted_at` (`deleted_at`),
  KEY `idx_article_status_created` (`article_id`,`status`,`created_at`),
  CONSTRAINT `fk_comments_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comments_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_comments_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comments_root_id` FOREIGN KEY (`root_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表，树形结构支持多级回复与审核流';

-- 3.2 评论点赞表
CREATE TABLE IF NOT EXISTS `comment_likes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `comment_id` BIGINT UNSIGNED NOT NULL COMMENT '评论ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '点赞用户ID',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_comment_user_like` (`comment_id`,`user_id`),
  KEY `idx_comment_likes_comment_id` (`comment_id`),
  KEY `idx_comment_likes_user_id` (`user_id`),
  KEY `idx_comment_likes_created_at` (`created_at`),
  CONSTRAINT `fk_comment_likes_comment_id` FOREIGN KEY (`comment_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comment_likes_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论点赞表';

-- ===================================
-- 4. 互动模块
-- ===================================

-- 4.1 文章点赞表
CREATE TABLE IF NOT EXISTS `article_likes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '点赞用户ID',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user_like` (`article_id`,`user_id`),
  KEY `idx_article_likes_article_id` (`article_id`),
  KEY `idx_article_likes_user_id` (`user_id`),
  KEY `idx_article_likes_created_at` (`created_at`),
  CONSTRAINT `fk_article_likes_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_likes_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章点赞表';

-- 4.2 文章收藏表
CREATE TABLE IF NOT EXISTS `article_bookmarks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '收藏ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '收藏用户ID',
  `note` VARCHAR(500) DEFAULT NULL COMMENT '收藏备注，由用户自行填写',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '收藏时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user_bookmark` (`article_id`,`user_id`),
  KEY `idx_article_bookmarks_article_id` (`article_id`),
  KEY `idx_article_bookmarks_user_id` (`user_id`),
  KEY `idx_article_bookmarks_created_at` (`created_at`),
  CONSTRAINT `fk_article_bookmarks_article_id` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_bookmarks_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章收藏表';

-- 4.3 用户关注表
CREATE TABLE IF NOT EXISTS `user_follows` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关注关系ID',
  `follower_id` BIGINT UNSIGNED NOT NULL COMMENT '关注者ID',
  `following_id` BIGINT UNSIGNED NOT NULL COMMENT '被关注者ID',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '关注时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follow_relation` (`follower_id`,`following_id`),
  KEY `idx_user_follows_follower_id` (`follower_id`),
  KEY `idx_user_follows_following_id` (`following_id`),
  KEY `idx_user_follows_created_at` (`created_at`),
  CONSTRAINT `chk_follow_self` CHECK (`follower_id` <> `following_id`),
  CONSTRAINT `fk_user_follows_follower_id` FOREIGN KEY (`follower_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_follows_following_id` FOREIGN KEY (`following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注表，社交关系维度';

-- 4.4 系统通知表
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收用户ID',
  `sender_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '触发用户ID，系统通知为空',
  `type` VARCHAR(50) NOT NULL COMMENT '通知类型：comment_reply/article_like/system 等',
  `title` VARCHAR(255) NOT NULL COMMENT '通知标题',
  `content` TEXT DEFAULT NULL COMMENT '通知内容',
  `action_url` VARCHAR(500) DEFAULT NULL COMMENT '点击通知后的跳转地址',
  `related_type` VARCHAR(50) DEFAULT NULL COMMENT '关联资源类型，如 article、comment',
  `related_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联资源ID',
  `is_read` TINYINT(1) DEFAULT 0 COMMENT '是否已读',
  `read_at` DATETIME(3) DEFAULT NULL COMMENT '已读时间',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间，支持用户清理通知后后台留档',
  PRIMARY KEY (`id`),
  KEY `idx_notifications_user_id` (`user_id`),
  KEY `idx_notifications_sender_id` (`sender_id`),
  KEY `idx_notifications_type` (`type`),
  KEY `idx_notifications_related_type` (`related_type`),
  KEY `idx_notifications_is_read` (`is_read`),
  KEY `idx_notifications_created_at` (`created_at`),
  KEY `idx_notifications_deleted_at` (`deleted_at`),
  KEY `idx_user_read` (`user_id`,`is_read`),
  CONSTRAINT `fk_notifications_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_notifications_sender_id` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表，站内消息中心';

-- ===================================
-- 5. 媒体管理模块
-- ===================================

-- 5.1 媒体文件表
CREATE TABLE IF NOT EXISTS `media_files` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `filename` VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `stored_name` VARCHAR(255) NOT NULL COMMENT '存储文件名，UUID 命名',
  `file_path` VARCHAR(500) NOT NULL COMMENT '文件存储路径',
  `file_url` VARCHAR(500) NOT NULL COMMENT '文件访问URL',
  `thumbnail_url` VARCHAR(500) DEFAULT NULL COMMENT '缩略图URL',
  `mime_type` VARCHAR(100) NOT NULL COMMENT 'MIME类型',
  `file_size` BIGINT UNSIGNED NOT NULL COMMENT '文件大小，单位字节',
  `file_hash` VARCHAR(64) DEFAULT NULL COMMENT '文件SHA256哈希值，用于秒传与去重',
  `width` BIGINT UNSIGNED DEFAULT NULL COMMENT '图片宽度，单位像素',
  `height` BIGINT UNSIGNED DEFAULT NULL COMMENT '图片高度，单位像素',
  `duration_seconds` BIGINT UNSIGNED DEFAULT 0 COMMENT '音视频时长，单位秒，非媒体文件为 0',
  `alt_text` VARCHAR(255) DEFAULT NULL COMMENT '替代文本，用于无障碍与SEO',
  `status` VARCHAR(20) DEFAULT 'active' COMMENT '文件状态：active-可用 processing-处理中 failed-处理失败 lost-文件丢失',
  `processed_at` DATETIME(3) DEFAULT NULL COMMENT '缩略图等后处理完成时间，为空表示尚未处理',
  `uploader_id` BIGINT UNSIGNED NOT NULL COMMENT '上传者ID',
  `upload_ip` VARCHAR(45) DEFAULT NULL COMMENT '上传IP地址',
  `storage_type` VARCHAR(20) DEFAULT 'local' COMMENT '存储类型：local/oss/s3/cos',
  `folder` VARCHAR(100) DEFAULT NULL COMMENT '文件夹分类',
  `usage_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '被正文引用次数，删除前需要校验',
  `download_count` BIGINT UNSIGNED DEFAULT 0 COMMENT '累计下载次数',
  `is_public` TINYINT(1) DEFAULT 1 COMMENT '是否公开访问',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_media_files_stored_name` (`stored_name`),
  KEY `idx_media_files_mime_type` (`mime_type`),
  KEY `idx_media_files_file_hash` (`file_hash`),
  KEY `idx_media_files_status` (`status`),
  KEY `idx_media_files_uploader_id` (`uploader_id`),
  KEY `idx_media_files_storage_type` (`storage_type`),
  KEY `idx_media_files_folder` (`folder`),
  KEY `idx_media_files_is_public` (`is_public`),
  KEY `idx_media_files_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_media_files_uploader_id` FOREIGN KEY (`uploader_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='媒体文件表，管理上传资源元信息与生命周期';

-- ===================================
-- 6. 站点运营模块
-- ===================================

-- 6.1 系统设置表
CREATE TABLE IF NOT EXISTS `settings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '设置ID',
  `key_name` VARCHAR(100) NOT NULL COMMENT '配置键名，点分命名空间格式',
  `label` VARCHAR(100) DEFAULT NULL COMMENT '设置项显示名称，供管理界面渲染',
  `value` LONGTEXT DEFAULT NULL COMMENT '配置值，支持JSON格式',
  `default_value` LONGTEXT DEFAULT NULL COMMENT '默认值，用于还原出厂配置',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '配置描述',
  `type` VARCHAR(20) DEFAULT 'string' COMMENT '值类型：string/number/boolean/json/array',
  `group_name` VARCHAR(50) DEFAULT 'general' COMMENT '配置分组',
  `is_public` TINYINT(1) DEFAULT 0 COMMENT '是否公开，公开项允许前端读取',
  `is_readonly` TINYINT(1) DEFAULT 0 COMMENT '是否只读，只读项由系统内部维护',
  `is_sensitive` TINYINT(1) DEFAULT 0 COMMENT '是否敏感配置，输出时需要脱敏',
  `validation_rule` VARCHAR(200) DEFAULT NULL COMMENT '验证规则，如正则表达式或取值范围',
  `sort_order` INT DEFAULT 0 COMMENT '排序权重',
  `updated_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '最后更新该配置的用户ID',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settings_key_name` (`key_name`),
  KEY `idx_settings_group_name` (`group_name`),
  KEY `idx_settings_is_public` (`is_public`),
  KEY `idx_settings_is_sensitive` (`is_sensitive`),
  KEY `idx_settings_sort_order` (`sort_order`),
  KEY `idx_settings_updated_by` (`updated_by`),
  CONSTRAINT `fk_settings_updated_by` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表，键值化全局配置';

-- 6.2 友情链接表
CREATE TABLE IF NOT EXISTS `friendly_links` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '链接ID',
  `name` VARCHAR(50) NOT NULL COMMENT '站点名称',
  `url` VARCHAR(255) NOT NULL COMMENT '站点URL，全局唯一防止重复收录',
  `logo` VARCHAR(500) DEFAULT NULL COMMENT '站点图标或头像URL',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '站点简介',
  `contact_email` VARCHAR(100) DEFAULT NULL COMMENT '站长联系邮箱',
  `sort_order` INT DEFAULT 0 COMMENT '展示排序权重，数值小的靠前',
  `status` VARCHAR(20) DEFAULT 'pending' COMMENT '链接状态：pending-待审核 active-展示中 hidden-已隐藏 rejected-已拒绝',
  `is_reciprocal` TINYINT(1) DEFAULT 0 COMMENT '是否已确认对方回链',
  `note` VARCHAR(255) DEFAULT NULL COMMENT '管理员备注，如收录时间与沟通记录',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_friendly_links_url` (`url`),
  KEY `idx_friendly_links_sort_order` (`sort_order`),
  KEY `idx_friendly_links_status` (`status`),
  KEY `idx_friendly_links_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='友情链接表，互链申请与展示管理';

-- ===================================
-- 7. 统计和日志模块
-- ===================================

-- 7.1 操作日志表
CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '操作用户ID，系统任务为空',
  `action` VARCHAR(100) NOT NULL COMMENT '操作类型，如 login、create_article',
  `resource_type` VARCHAR(50) DEFAULT NULL COMMENT '资源类型，如 user、article、comment',
  `resource_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '资源ID',
  `status` VARCHAR(20) DEFAULT 'success' COMMENT '执行结果：success-成功 failed-失败',
  `error_message` VARCHAR(500) DEFAULT NULL COMMENT '失败原因，成功时为空',
  `duration_ms` BIGINT UNSIGNED DEFAULT 0 COMMENT '操作耗时，单位毫秒',
  `trace_id` VARCHAR(64) DEFAULT NULL COMMENT '链路追踪ID，用于串联一次请求内的多条日志',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` TEXT DEFAULT NULL COMMENT '用户代理',
  `details` JSON DEFAULT NULL COMMENT '操作详情，结构化快照',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_user_id` (`user_id`),
  KEY `idx_operation_logs_action` (`action`),
  KEY `idx_operation_logs_resource_type` (`resource_type`),
  KEY `idx_operation_logs_status` (`status`),
  KEY `idx_operation_logs_trace_id` (`trace_id`),
  KEY `idx_operation_logs_created_at` (`created_at`),
  CONSTRAINT `fk_operation_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表，安全审计与问题追踪';

-- 7.2 搜索记录表
CREATE TABLE IF NOT EXISTS `search_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '搜索记录ID',
  `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '搜索用户ID，游客为空',
  `keyword` VARCHAR(255) NOT NULL COMMENT '搜索关键词',
  `results_count` INT DEFAULT 0 COMMENT '搜索结果数量',
  `status` VARCHAR(20) DEFAULT 'success' COMMENT '执行结果：success-成功 failed-失败',
  `duration_ms` BIGINT UNSIGNED DEFAULT 0 COMMENT '搜索耗时，单位毫秒',
  `ip_address` VARCHAR(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` TEXT DEFAULT NULL COMMENT '用户代理',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '搜索时间',
  PRIMARY KEY (`id`),
  KEY `idx_search_logs_user_id` (`user_id`),
  KEY `idx_search_logs_keyword` (`keyword`),
  KEY `idx_search_logs_ip_address` (`ip_address`),
  KEY `idx_search_logs_created_at` (`created_at`),
  CONSTRAINT `fk_search_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索记录表，搜索行为分析';

-- 7.3 内容统计表
CREATE TABLE IF NOT EXISTS `content_stats` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '统计ID',
  `content_type` VARCHAR(50) NOT NULL COMMENT '内容类型：article/tag/category',
  `content_id` BIGINT UNSIGNED NOT NULL COMMENT '内容ID',
  `stat_type` VARCHAR(50) NOT NULL COMMENT '统计类型：daily_views/weekly_views/likes_count 等',
  `stat_value` BIGINT UNSIGNED DEFAULT 0 COMMENT '统计值',
  `stat_date` DATE DEFAULT NULL COMMENT '统计日期',
  `created_at` DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_content_stat` (`content_type`,`content_id`,`stat_type`,`stat_date`),
  KEY `idx_content_stats_content_type` (`content_type`),
  KEY `idx_content_stats_content_id` (`content_id`),
  KEY `idx_content_stats_stat_type` (`stat_type`),
  KEY `idx_content_stats_stat_value` (`stat_value`),
  KEY `idx_content_stats_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容统计表，多维度聚合指标';

-- ===================================
-- 8. 默认数据插入
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
('upload_max_size', '10485760', '文件上传最大大小，单位字节', 'number', 'media', 0),
('allowed_file_types', '["jpg","jpeg","png","gif","pdf","doc","docx"]', '允许上传的文件类型', 'json', 'media', 0),
('cache_enabled', '1', '是否启用缓存系统', 'boolean', 'cache', 0),
('cache_expire', '3600', '缓存过期时间，单位秒', 'number', 'cache', 0),
('mail_password', '', 'SMTP 邮箱密码，敏感配置', 'string', 'mail', 0);

-- 插入默认分类
INSERT INTO `categories` (`name`, `slug`, `description`, `sort_order`, `status`) VALUES
('技术', 'tech', '技术相关文章', 1, 1),
('生活', 'life', '生活感悟和日常', 2, 1),
('随笔', 'notes', '随笔和思考', 3, 1);

-- 插入默认标签
INSERT INTO `tags` (`name`, `slug`, `color`, `status`) VALUES
('Go', 'go', '#00ADD8', 1),
('JavaScript', 'javascript', '#F7DF1E', 1),
('TypeScript', 'typescript', '#3178C6', 1),
('Svelte', 'svelte', '#FF3E00', 1),
('MySQL', 'mysql', '#4479A1', 1);

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;
