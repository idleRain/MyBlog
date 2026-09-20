-- 分期迁移 000002：内容多语言翻译表
-- 主表列恒为缺省语言内容，翻译行以 locale 区分其他语言；
-- 翻译行允许按字段部分填写，输出时缺失翻译的字段回退主列。

CREATE TABLE `article_translations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '翻译ID',
  `article_id` bigint unsigned NOT NULL COMMENT '文章ID',
  `locale` varchar(10) NOT NULL COMMENT '语言标识，主子标签形式，如 zh、en',
  `title` varchar(200) DEFAULT NULL COMMENT '该语言文章标题',
  `summary` text COMMENT '该语言文章摘要',
  `content` longtext COMMENT '该语言文章正文，Markdown 格式',
  `content_html` longtext COMMENT '该语言正文渲染后的 HTML 缓存',
  `word_count` bigint unsigned DEFAULT '0' COMMENT '该语言正文字数统计',
  `seo_title` varchar(100) DEFAULT NULL COMMENT '该语言SEO标题',
  `seo_description` varchar(255) DEFAULT NULL COMMENT '该语言SEO描述',
  `seo_keywords` varchar(200) DEFAULT NULL COMMENT '该语言SEO关键词',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_translation` (`article_id`,`locale`),
  KEY `idx_article_translations_locale` (`locale`),
  CONSTRAINT `fk_article_translations_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章翻译表，按语言存储标题摘要正文的翻译内容';

CREATE TABLE `category_translations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '翻译ID',
  `category_id` bigint unsigned NOT NULL COMMENT '分类ID',
  `locale` varchar(10) NOT NULL COMMENT '语言标识，主子标签形式，如 zh、en',
  `name` varchar(50) DEFAULT NULL COMMENT '该语言分类名称',
  `description` text COMMENT '该语言分类描述',
  `seo_title` varchar(100) DEFAULT NULL COMMENT '该语言SEO标题',
  `seo_description` varchar(255) DEFAULT NULL COMMENT '该语言SEO描述',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_translation` (`category_id`,`locale`),
  KEY `idx_category_translations_locale` (`locale`),
  CONSTRAINT `fk_category_translations_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分类翻译表，按语言存储分类名称与描述的翻译内容';

CREATE TABLE `tag_translations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '翻译ID',
  `tag_id` bigint unsigned NOT NULL COMMENT '标签ID',
  `locale` varchar(10) NOT NULL COMMENT '语言标识，主子标签形式，如 zh、en',
  `name` varchar(30) DEFAULT NULL COMMENT '该语言标签名称',
  `description` varchar(200) DEFAULT NULL COMMENT '该语言标签描述',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tag_translation` (`tag_id`,`locale`),
  KEY `idx_tag_translations_locale` (`locale`),
  CONSTRAINT `fk_tag_translations_tag` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签翻译表，按语言存储标签名称与描述的翻译内容';
