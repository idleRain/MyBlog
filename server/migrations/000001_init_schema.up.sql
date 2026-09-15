-- MyBlog 数据库迁移基线：以开发库 AutoMigrate 生成物为源固化（2026-09-15）。
-- 生产环境经 golang-migrate 执行本基线后，后续 schema 变更须以增量迁移演进，禁止回退 AutoMigrate。
-- 开发模式仍走 AutoMigrate，与本基线出现漂移时应以手工增量迁移对齐。

-- MySQL dump 10.13  Distrib 26.7.0, for Win64 (x86_64)
--
-- Host: localhost    Database: blog
-- ------------------------------------------------------
-- Server version	26.7.0

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `article_bookmarks`
--

DROP TABLE IF EXISTS `article_bookmarks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `article_bookmarks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '收藏ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `user_id` bigint unsigned NOT NULL COMMENT '收藏用户ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '收藏时间',
  `note` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收藏备注，由用户自行填写',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user_bookmark` (`article_id`,`user_id`),
  KEY `idx_article_bookmarks_article_id` (`article_id`),
  KEY `idx_article_bookmarks_user_id` (`user_id`),
  KEY `idx_article_bookmarks_created_at` (`created_at`),
  CONSTRAINT `fk_article_bookmarks_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_articles_bookmarks` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`),
  CONSTRAINT `fk_users_article_bookmarks` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章收藏表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_categories`
--

DROP TABLE IF EXISTS `article_categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `article_categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `category_id` bigint unsigned NOT NULL COMMENT '分类ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_category` (`article_id`,`category_id`),
  KEY `idx_article_categories_article_id` (`article_id`),
  KEY `idx_article_categories_category_id` (`category_id`),
  CONSTRAINT `fk_article_categories_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_categories_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类关联表，支持一文多分类';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_likes`
--

DROP TABLE IF EXISTS `article_likes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `article_likes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `user_id` bigint unsigned NOT NULL COMMENT '点赞用户ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user_like` (`article_id`,`user_id`),
  KEY `idx_article_likes_article_id` (`article_id`),
  KEY `idx_article_likes_user_id` (`user_id`),
  KEY `idx_article_likes_created_at` (`created_at`),
  CONSTRAINT `fk_article_likes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_articles_likes` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`),
  CONSTRAINT `fk_users_article_likes` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章点赞表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_revisions`
--

DROP TABLE IF EXISTS `article_revisions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `article_revisions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '修订ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `revision_no` bigint unsigned NOT NULL COMMENT '修订版本号，从 1 开始随文章 version 递增',
  `title` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '该版本文章标题快照',
  `summary` text COLLATE utf8mb4_unicode_ci COMMENT '该版本摘要快照',
  `content` longtext COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '该版本正文快照，Markdown 格式',
  `content_html` longtext COLLATE utf8mb4_unicode_ci COMMENT '该版本渲染后 HTML 快照',
  `word_count` bigint unsigned DEFAULT '0' COMMENT '该版本字数统计',
  `change_summary` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '变更说明，由作者填写',
  `editor_id` bigint unsigned DEFAULT NULL COMMENT '执行本次修订的用户ID',
  `is_autosave` tinyint(1) DEFAULT '0' COMMENT '是否编辑器自动保存产生的快照',
  `created_at` datetime(3) DEFAULT NULL COMMENT '修订时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_revision` (`article_id`,`revision_no`),
  KEY `idx_article_revisions_article_id` (`article_id`),
  KEY `idx_article_revisions_editor_id` (`editor_id`),
  KEY `idx_article_revisions_is_autosave` (`is_autosave`),
  KEY `idx_article_revisions_created_at` (`created_at`),
  CONSTRAINT `fk_article_revisions_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_revisions_editor` FOREIGN KEY (`editor_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章修订历史表，保存正文快照支持回滚';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_tags`
--

DROP TABLE IF EXISTS `article_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `article_tags` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `tag_id` bigint unsigned NOT NULL COMMENT '标签ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_tag` (`article_id`,`tag_id`),
  KEY `fk_article_tags_tag` (`tag_id`),
  CONSTRAINT `fk_article_tags_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_article_tags_tag` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签关联表，多对多挂载关系';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_views`
--

DROP TABLE IF EXISTS `article_views`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `article_views` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '浏览记录ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID，注册用户填写',
  `visitor_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访客标识，匿名用户填写',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COLLATE utf8mb4_unicode_ci COMMENT '用户代理',
  `referer` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '来源页面',
  `view_date` date DEFAULT NULL COMMENT '浏览日期',
  `view_count` bigint unsigned DEFAULT '1' COMMENT '当日浏览次数',
  `created_at` datetime(3) DEFAULT NULL COMMENT '首次浏览时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后浏览时间',
  `duration_seconds` bigint unsigned DEFAULT '0' COMMENT '页面停留时长，单位秒，由前端埋点上报',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_visitor_date` (`article_id`,`visitor_id`,`view_date`),
  KEY `idx_article_views_user_id` (`user_id`),
  KEY `idx_article_views_ip_address` (`ip_address`),
  KEY `idx_article_views_view_date` (`view_date`),
  CONSTRAINT `fk_article_views_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_articles_views` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章浏览统计表，按访客与日期去重计数';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `articles`
--

DROP TABLE IF EXISTS `articles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `articles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文章标题',
  `slug` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'URL友好标识',
  `summary` text COLLATE utf8mb4_unicode_ci COMMENT '文章摘要',
  `content` longtext COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文章内容，Markdown 格式',
  `content_html` longtext COLLATE utf8mb4_unicode_ci COMMENT '文章内容，渲染后的 HTML 缓存',
  `cover_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封面图片URL',
  `author_id` bigint unsigned NOT NULL COMMENT '作者ID',
  `category_id` bigint unsigned DEFAULT NULL COMMENT '主分类ID',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'draft' COMMENT '文章状态：draft/published/archived/private',
  `is_featured` tinyint(1) DEFAULT '0' COMMENT '是否精选文章',
  `is_top` tinyint(1) DEFAULT '0' COMMENT '是否置顶',
  `comment_enabled` tinyint(1) DEFAULT '1' COMMENT '是否允许评论',
  `view_count` bigint unsigned DEFAULT '0' COMMENT '浏览量',
  `like_count` bigint unsigned DEFAULT '0' COMMENT '点赞数',
  `comment_count` bigint unsigned DEFAULT '0' COMMENT '评论数',
  `word_count` bigint unsigned DEFAULT '0' COMMENT '字数统计',
  `reading_time` bigint unsigned DEFAULT '0' COMMENT '预计阅读时间，单位分钟',
  `seo_title` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'SEO描述',
  `seo_keywords` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'SEO关键词',
  `published_at` datetime(3) DEFAULT NULL COMMENT '发布时间',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `origin_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'original' COMMENT '来源类型：original-原创 translation-翻译 reprint-转载',
  `source_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '原文链接，原创文章为空',
  `source_author` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '原文作者，原创文章为空',
  `access_password` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访问密码哈希，仅私密文章生效，为空表示仅登录可见',
  `bookmark_count` bigint unsigned DEFAULT '0' COMMENT '收藏数',
  `version` bigint unsigned DEFAULT '1' COMMENT '内容版本号，每次保存正文递增，对应 article_revisions',
  `scheduled_at` datetime(3) DEFAULT NULL COMMENT '定时发布时间，到期后由调度任务发布',
  `edited_at` datetime(3) DEFAULT NULL COMMENT '正文最后编辑时间，用于展示已编辑标记',
  `archived_at` datetime(3) DEFAULT NULL COMMENT '归档时间，进入归档状态时写入',
  `last_comment_at` datetime(3) DEFAULT NULL COMMENT '最新评论时间，用于评论排序展示',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_articles_slug` (`slug`),
  KEY `idx_articles_author_id` (`author_id`),
  KEY `idx_articles_category_id` (`category_id`),
  KEY `idx_articles_status` (`status`),
  KEY `idx_articles_is_featured` (`is_featured`),
  KEY `idx_articles_is_top` (`is_top`),
  KEY `idx_articles_view_count` (`view_count`),
  KEY `idx_articles_published_at` (`published_at`),
  KEY `idx_articles_deleted_at` (`deleted_at`),
  KEY `idx_author_status` (`author_id`,`status`),
  KEY `idx_status_published` (`status`,`published_at`),
  KEY `idx_articles_origin_type` (`origin_type`),
  KEY `idx_articles_scheduled_at` (`scheduled_at`),
  FULLTEXT KEY `ft_articles_search` (`title`,`content`,`summary`) /*!50100 WITH PARSER `ngram` */ ,
  CONSTRAINT `fk_articles_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_categories_articles` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  CONSTRAINT `fk_users_articles` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表，博客核心内容实体';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `auth_tokens`
--

DROP TABLE IF EXISTS `auth_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `auth_tokens` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '令牌ID',
  `user_id` bigint unsigned NOT NULL COMMENT '所属用户ID',
  `token_hash` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '令牌哈希值，SHA256 十六进制',
  `token_type` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT 'password_reset' COMMENT '令牌用途：password_reset-密码重置 email_verify-邮箱验证',
  `expires_at` datetime(3) DEFAULT NULL COMMENT '令牌过期时间',
  `used_at` datetime(3) DEFAULT NULL COMMENT '令牌核销时间，为空表示未使用',
  `request_ip` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '申请令牌时的来源IP，用于安全审计',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_auth_tokens_token_hash` (`token_hash`),
  KEY `idx_auth_tokens_user_id` (`user_id`),
  KEY `idx_auth_tokens_token_type` (`token_type`),
  KEY `idx_auth_tokens_expires_at` (`expires_at`),
  CONSTRAINT `fk_auth_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证令牌表，支撑密码找回与邮箱验证流程';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `categories`
--

DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分类名称',
  `slug` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'URL友好标识',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '分类描述',
  `cover_image` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类封面图',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父分类ID，顶级分类为空',
  `level` tinyint unsigned DEFAULT '1' COMMENT '分类层级，顶级为 1',
  `sort_order` bigint DEFAULT '0' COMMENT '排序权重，数值小的靠前',
  `article_count` bigint unsigned DEFAULT '0' COMMENT '文章数量，发布时异步维护',
  `is_featured` tinyint(1) DEFAULT '0' COMMENT '是否为精选分类',
  `seo_title` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'SEO标题',
  `seo_description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'SEO描述',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `root_id` bigint unsigned DEFAULT NULL COMMENT '根分类ID，用于整棵子树的聚合查询',
  `path` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物化路径，形如 /1/5/12，用于一次查询取整棵子树',
  `status` tinyint DEFAULT '1' COMMENT '分类状态：1-显示 0-隐藏',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_categories_slug` (`slug`),
  KEY `idx_categories_parent_id` (`parent_id`),
  KEY `idx_categories_level` (`level`),
  KEY `idx_categories_sort_order` (`sort_order`),
  KEY `idx_categories_is_featured` (`is_featured`),
  KEY `idx_categories_deleted_at` (`deleted_at`),
  KEY `idx_categories_root_id` (`root_id`),
  KEY `idx_categories_status` (`status`),
  CONSTRAINT `fk_categories_children` FOREIGN KEY (`parent_id`) REFERENCES `categories` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类表，树形结构支撑栏目导航';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comment_likes`
--

DROP TABLE IF EXISTS `comment_likes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `comment_likes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `comment_id` bigint unsigned NOT NULL COMMENT '评论ID',
  `user_id` bigint unsigned NOT NULL COMMENT '点赞用户ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_comment_user_like` (`comment_id`,`user_id`),
  KEY `idx_comment_likes_comment_id` (`comment_id`),
  KEY `idx_comment_likes_user_id` (`user_id`),
  KEY `idx_comment_likes_created_at` (`created_at`),
  CONSTRAINT `fk_comment_likes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comments_likes` FOREIGN KEY (`comment_id`) REFERENCES `comments` (`id`),
  CONSTRAINT `fk_users_comment_likes` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论点赞表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comments`
--

DROP TABLE IF EXISTS `comments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `comments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID，注册用户填写',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父评论ID，根评论为空',
  `root_id` bigint unsigned DEFAULT NULL COMMENT '根评论ID，便于一次查询整棵评论树',
  `level` tinyint unsigned DEFAULT '1' COMMENT '评论层级，根评论为 1',
  `author_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '游客姓名',
  `author_email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '游客邮箱，仅管理端审计可见',
  `author_website` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '游客网站',
  `author_ip` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '评论者IP地址，用于反垃圾与封禁，仅管理端审计可见',
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '评论内容，Markdown 格式',
  `content_html` text COLLATE utf8mb4_unicode_ci COMMENT '评论内容，渲染后的 HTML 缓存',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '审核状态：pending/approved/rejected/spam/trash',
  `like_count` bigint unsigned DEFAULT '0' COMMENT '点赞数',
  `reply_count` bigint unsigned DEFAULT '0' COMMENT '回复数量',
  `user_agent` text COLLATE utf8mb4_unicode_ci COMMENT '用户代理，仅管理端审计可见',
  `is_author` tinyint(1) DEFAULT '0' COMMENT '是否为文章作者回复',
  `is_pinned` tinyint(1) DEFAULT '0' COMMENT '是否置顶评论',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `reported_count` bigint unsigned DEFAULT '0' COMMENT '被举报次数，达到阈值后进入待复核队列',
  `edited_at` datetime(3) DEFAULT NULL COMMENT '内容最后编辑时间，用于展示已编辑标记',
  PRIMARY KEY (`id`),
  KEY `idx_comments_article_id` (`article_id`),
  KEY `idx_comments_user_id` (`user_id`),
  KEY `idx_comments_parent_id` (`parent_id`),
  KEY `idx_comments_root_id` (`root_id`),
  KEY `idx_comments_author_ip` (`author_ip`),
  KEY `idx_comments_status` (`status`),
  KEY `idx_comments_created_at` (`created_at`),
  KEY `idx_comments_deleted_at` (`deleted_at`),
  KEY `idx_article_status_created` (`article_id`,`status`,`created_at`),
  KEY `idx_comments_is_pinned` (`is_pinned`),
  CONSTRAINT `fk_articles_comments` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`),
  CONSTRAINT `fk_comments_children` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`),
  CONSTRAINT `fk_comments_root` FOREIGN KEY (`root_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_users_comments` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表，树形结构支持多级回复与审核流';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `content_stats`
--

DROP TABLE IF EXISTS `content_stats`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `content_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '统计ID',
  `content_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '内容类型：article/tag/category',
  `content_id` bigint unsigned NOT NULL COMMENT '内容ID',
  `stat_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '统计类型：daily_views/weekly_views/likes_count 等',
  `stat_value` bigint unsigned DEFAULT '0' COMMENT '统计值',
  `stat_date` date DEFAULT NULL COMMENT '统计日期',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_content_stat` (`content_type`,`content_id`,`stat_type`,`stat_date`),
  KEY `idx_content_stats_content_type` (`content_type`),
  KEY `idx_content_stats_content_id` (`content_id`),
  KEY `idx_content_stats_stat_type` (`stat_type`),
  KEY `idx_content_stats_stat_value` (`stat_value`),
  KEY `idx_content_stats_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容统计表，多维度聚合指标';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `friendly_links`
--

DROP TABLE IF EXISTS `friendly_links`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `friendly_links` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '链接ID',
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '站点名称',
  `url` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '站点URL，全局唯一防止重复收录',
  `logo` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '站点图标或头像URL',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '站点简介',
  `contact_email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '站长联系邮箱',
  `sort_order` bigint DEFAULT '0' COMMENT '展示排序权重，数值小的靠前',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '链接状态：pending-待审核 active-展示中 hidden-已隐藏 rejected-已拒绝',
  `is_reciprocal` tinyint(1) DEFAULT '0' COMMENT '是否已确认对方回链',
  `note` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '管理员备注，如收录时间与沟通记录',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_friendly_links_url` (`url`),
  KEY `idx_friendly_links_sort_order` (`sort_order`),
  KEY `idx_friendly_links_status` (`status`),
  KEY `idx_friendly_links_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='友情链接表，互链申请与展示管理';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `media_files`
--

DROP TABLE IF EXISTS `media_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `media_files` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `filename` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '原始文件名',
  `stored_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '存储文件名，UUID 命名',
  `file_path` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件存储路径',
  `file_url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件访问URL',
  `thumbnail_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '缩略图URL',
  `mime_type` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'MIME类型',
  `file_size` bigint unsigned NOT NULL COMMENT '文件大小，单位字节',
  `file_hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件SHA256哈希值，用于秒传与去重',
  `width` bigint unsigned DEFAULT NULL COMMENT '图片宽度，单位像素',
  `height` bigint unsigned DEFAULT NULL COMMENT '图片高度，单位像素',
  `alt_text` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '替代文本，用于无障碍与SEO',
  `uploader_id` bigint unsigned NOT NULL COMMENT '上传者ID',
  `upload_ip` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '上传IP地址',
  `storage_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'local' COMMENT '存储类型：local/oss/s3/cos',
  `folder` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件夹分类',
  `usage_count` bigint unsigned DEFAULT '0' COMMENT '被正文引用次数，删除前需要校验',
  `is_public` tinyint(1) DEFAULT '1' COMMENT '是否公开访问',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `duration_seconds` bigint unsigned DEFAULT '0' COMMENT '音视频时长，单位秒，非媒体文件为 0',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '文件状态：active-可用 processing-处理中 failed-处理失败 lost-文件丢失',
  `processed_at` datetime(3) DEFAULT NULL COMMENT '缩略图等后处理完成时间，为空表示尚未处理',
  `download_count` bigint unsigned DEFAULT '0' COMMENT '累计下载次数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_media_files_stored_name` (`stored_name`),
  KEY `idx_media_files_mime_type` (`mime_type`),
  KEY `idx_media_files_file_hash` (`file_hash`),
  KEY `idx_media_files_uploader_id` (`uploader_id`),
  KEY `idx_media_files_storage_type` (`storage_type`),
  KEY `idx_media_files_folder` (`folder`),
  KEY `idx_media_files_deleted_at` (`deleted_at`),
  KEY `idx_media_files_status` (`status`),
  KEY `idx_media_files_is_public` (`is_public`),
  CONSTRAINT `fk_media_files_uploader` FOREIGN KEY (`uploader_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_users_media_files` FOREIGN KEY (`uploader_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='媒体文件表，管理上传资源元信息与生命周期';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `notifications`
--

DROP TABLE IF EXISTS `notifications`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `notifications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `user_id` bigint unsigned NOT NULL COMMENT '接收用户ID',
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知类型：comment_reply/article_like/system 等',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知标题',
  `content` text COLLATE utf8mb4_unicode_ci COMMENT '通知内容',
  `related_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联资源类型，如 article、comment',
  `related_id` bigint unsigned DEFAULT NULL COMMENT '关联资源ID',
  `is_read` tinyint(1) DEFAULT '0' COMMENT '是否已读',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `sender_id` bigint unsigned DEFAULT NULL COMMENT '触发用户ID，系统通知为空',
  `action_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '点击通知后的跳转地址',
  `read_at` datetime(3) DEFAULT NULL COMMENT '已读时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间，支持用户清理通知后后台留档',
  PRIMARY KEY (`id`),
  KEY `idx_notifications_user_id` (`user_id`),
  KEY `idx_notifications_type` (`type`),
  KEY `idx_notifications_related_type` (`related_type`),
  KEY `idx_notifications_is_read` (`is_read`),
  KEY `idx_notifications_created_at` (`created_at`),
  KEY `idx_user_read` (`user_id`,`is_read`),
  KEY `idx_notifications_sender_id` (`sender_id`),
  KEY `idx_notifications_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_notifications_sender` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_notifications_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_users_notifications` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表，站内消息中心';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `operation_logs`
--

DROP TABLE IF EXISTS `operation_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '操作用户ID，系统任务为空',
  `action` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型，如 login、create_article',
  `resource_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源类型，如 user、article、comment',
  `resource_id` bigint unsigned DEFAULT NULL COMMENT '资源ID',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COLLATE utf8mb4_unicode_ci COMMENT '用户代理',
  `details` json DEFAULT NULL COMMENT '操作详情，结构化快照',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'success' COMMENT '执行结果：success-成功 failed-失败',
  `error_message` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '失败原因，成功时为空',
  `duration_ms` bigint unsigned DEFAULT '0' COMMENT '操作耗时，单位毫秒',
  `trace_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '链路追踪ID，用于串联一次请求内的多条日志',
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_user_id` (`user_id`),
  KEY `idx_operation_logs_action` (`action`),
  KEY `idx_operation_logs_resource_type` (`resource_type`),
  KEY `idx_operation_logs_created_at` (`created_at`),
  KEY `idx_operation_logs_status` (`status`),
  KEY `idx_operation_logs_trace_id` (`trace_id`),
  CONSTRAINT `fk_operation_logs_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表，安全审计与问题追踪';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `search_logs`
--

DROP TABLE IF EXISTS `search_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `search_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '搜索记录ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '搜索用户ID，游客为空',
  `keyword` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '搜索关键词',
  `results_count` bigint DEFAULT '0' COMMENT '搜索结果数量',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COLLATE utf8mb4_unicode_ci COMMENT '用户代理',
  `created_at` datetime(3) DEFAULT NULL COMMENT '搜索时间',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'success' COMMENT '执行结果：success-成功 failed-失败',
  `duration_ms` bigint unsigned DEFAULT '0' COMMENT '搜索耗时，单位毫秒',
  PRIMARY KEY (`id`),
  KEY `idx_search_logs_user_id` (`user_id`),
  KEY `idx_search_logs_keyword` (`keyword`),
  KEY `idx_search_logs_ip_address` (`ip_address`),
  KEY `idx_search_logs_created_at` (`created_at`),
  CONSTRAINT `fk_search_logs_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_users_search_logs` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索记录表，搜索行为分析';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `settings`
--

DROP TABLE IF EXISTS `settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '设置ID',
  `key_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置键名，点分命名空间格式',
  `value` longtext COLLATE utf8mb4_unicode_ci COMMENT '配置值，支持JSON格式',
  `default_value` longtext COLLATE utf8mb4_unicode_ci COMMENT '默认值，用于还原出厂配置',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '配置描述',
  `type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'string' COMMENT '值类型：string/number/boolean/json/array',
  `group_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'general' COMMENT '配置分组',
  `is_public` tinyint(1) DEFAULT '0' COMMENT '是否公开，公开项允许前端读取',
  `is_readonly` tinyint(1) DEFAULT '0' COMMENT '是否只读，只读项由系统内部维护',
  `validation_rule` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '验证规则，如正则表达式或取值范围',
  `sort_order` bigint DEFAULT '0' COMMENT '排序权重',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `label` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '设置项显示名称，供管理界面渲染',
  `is_sensitive` tinyint(1) DEFAULT '0' COMMENT '是否敏感配置，输出时需要脱敏',
  `updated_by` bigint unsigned DEFAULT NULL COMMENT '最后更新该配置的用户ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_settings_key_name` (`key_name`),
  KEY `idx_settings_group_name` (`group_name`),
  KEY `idx_settings_is_public` (`is_public`),
  KEY `idx_settings_sort_order` (`sort_order`),
  KEY `idx_settings_is_sensitive` (`is_sensitive`),
  KEY `idx_settings_updated_by` (`updated_by`),
  CONSTRAINT `fk_settings_updated_by_user` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表，键值化全局配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tags`
--

DROP TABLE IF EXISTS `tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tags` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '标签ID',
  `name` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签名称',
  `slug` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'URL友好标识',
  `color` varchar(7) COLLATE utf8mb4_unicode_ci DEFAULT '#808080' COMMENT '标签颜色，HEX 格式',
  `description` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签描述',
  `usage_count` bigint unsigned DEFAULT '0' COMMENT '使用次数，文章挂载时异步维护',
  `is_hot` tinyint(1) DEFAULT '0' COMMENT '是否热门标签',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `status` tinyint DEFAULT '1' COMMENT '标签状态：1-启用 0-隐藏',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tags_name` (`name`),
  UNIQUE KEY `idx_tags_slug` (`slug`),
  KEY `idx_tags_usage_count` (`usage_count`),
  KEY `idx_tags_is_hot` (`is_hot`),
  KEY `idx_tags_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表，文章主题的轻量归类维度';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_activities`
--

DROP TABLE IF EXISTS `user_activities`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_activities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '活动ID',
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `action` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源类型：article/comment/user 等',
  `resource_id` bigint unsigned DEFAULT NULL COMMENT '资源ID',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作描述',
  `metadata` json DEFAULT NULL COMMENT '额外元数据',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COLLATE utf8mb4_unicode_ci COMMENT '用户代理',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'success' COMMENT '执行结果：success-成功 failed-失败',
  `error_message` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '失败原因，成功时为空',
  `duration_ms` bigint unsigned DEFAULT '0' COMMENT '操作耗时，单位毫秒',
  PRIMARY KEY (`id`),
  KEY `idx_user_activities_user_id` (`user_id`),
  KEY `idx_user_activities_action` (`action`),
  KEY `idx_user_activities_ip_address` (`ip_address`),
  KEY `idx_user_activities_created_at` (`created_at`),
  KEY `idx_user_created` (`user_id`,`created_at`),
  CONSTRAINT `fk_user_activities_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_users_activities` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户活动日志表，记录用户行为轨迹';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_follows`
--

DROP TABLE IF EXISTS `user_follows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_follows` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关注关系ID',
  `follower_id` bigint unsigned NOT NULL COMMENT '关注者ID',
  `following_id` bigint unsigned NOT NULL COMMENT '被关注者ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '关注时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follow_relation` (`follower_id`,`following_id`),
  KEY `idx_user_follows_follower_id` (`follower_id`),
  KEY `idx_user_follows_following_id` (`following_id`),
  KEY `idx_user_follows_created_at` (`created_at`),
  CONSTRAINT `fk_user_follows_follower` FOREIGN KEY (`follower_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_follows_following` FOREIGN KEY (`following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_users_followers` FOREIGN KEY (`following_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_users_following` FOREIGN KEY (`follower_id`) REFERENCES `users` (`id`),
  CONSTRAINT `chk_follow_self` CHECK ((`follower_id` <> `following_id`))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注表，社交关系维度';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_sessions`
--

DROP TABLE IF EXISTS `user_sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `refresh_token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '刷新令牌',
  `access_token_hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访问令牌哈希值',
  `device_info` json DEFAULT NULL COMMENT '设备信息，包含浏览器与操作系统等',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '登录IP地址',
  `user_agent` text COLLATE utf8mb4_unicode_ci COMMENT '用户代理字符串',
  `expires_at` datetime(3) DEFAULT NULL COMMENT '令牌过期时间',
  `last_used_at` datetime(3) DEFAULT NULL COMMENT '最后使用时间',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '会话状态：1-活跃，0-已注销',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `device_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'web' COMMENT '设备类型：web/mobile/tablet/desktop',
  `last_refresh_at` datetime(3) DEFAULT NULL COMMENT '刷新令牌最近轮换时间',
  `logout_at` datetime(3) DEFAULT NULL COMMENT '会话注销时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_sessions_refresh_token` (`refresh_token`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  KEY `idx_user_sessions_ip_address` (`ip_address`),
  KEY `idx_user_sessions_expires_at` (`expires_at`),
  KEY `idx_user_sessions_is_active` (`is_active`),
  KEY `idx_user_active` (`user_id`,`is_active`),
  KEY `idx_user_sessions_device_type` (`device_type`),
  CONSTRAINT `fk_user_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_users_sessions` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表，管理登录设备与令牌轮换';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名，全局唯一',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱地址，全局唯一',
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码，存储 bcrypt 哈希值',
  `nickname` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户昵称，为空时展示用户名',
  `avatar` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像URL',
  `bio` text COLLATE utf8mb4_unicode_ci COMMENT '个人简介',
  `birthday` date DEFAULT NULL COMMENT '生日',
  `role` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'user' COMMENT '用户角色：superadmin/admin/editor/user',
  `status` tinyint DEFAULT '1' COMMENT '用户状态：1-正常 0-禁用 2-锁定',
  `last_login_at` datetime(3) DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后登录IP，IPv6 最长 45 字符',
  `login_count` bigint unsigned DEFAULT '0' COMMENT '累计登录成功次数',
  `email_verified_at` datetime(3) DEFAULT NULL COMMENT '邮箱验证完成时间，为空表示未验证',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号，全局唯一，未绑定时为空',
  `cover_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '个人主页封面图URL',
  `website` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '个人网站URL',
  `location` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '常居地描述',
  `gender` tinyint DEFAULT NULL COMMENT '性别：0-未知 1-男 2-女',
  `timezone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'Asia/Shanghai' COMMENT '用户时区标识，IANA 命名格式',
  `locale` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'zh-CN' COMMENT '用户界面语言标识，BCP 47 格式',
  `failed_login_count` bigint unsigned DEFAULT '0' COMMENT '连续登录失败次数，登录成功后清零',
  `locked_until` datetime(3) DEFAULT NULL COMMENT '账户锁定截止时间，到期后可重新登录',
  `password_changed_at` datetime(3) DEFAULT NULL COMMENT '密码最后修改时间，用于安全审计',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '管理员备注，仅管理端可见',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_email` (`email`),
  UNIQUE KEY `idx_users_phone` (`phone`),
  KEY `idx_users_deleted_at` (`deleted_at`),
  KEY `idx_users_role` (`role`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_last_login_at` (`last_login_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表，存储账号身份、个人资料与安全状态';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping events for database 'blog'
--

--
-- Dumping routines for database 'blog'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-09-15 15:08:45
