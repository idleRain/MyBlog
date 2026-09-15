-- 回滚迁移基线：删除全部业务表。
-- schema_migrations 表由 golang-migrate 自行管理，不在本脚本删除范围内。

SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `user_sessions`;
DROP TABLE IF EXISTS `user_follows`;
DROP TABLE IF EXISTS `user_activities`;
DROP TABLE IF EXISTS `tags`;
DROP TABLE IF EXISTS `settings`;
DROP TABLE IF EXISTS `search_logs`;
DROP TABLE IF EXISTS `operation_logs`;
DROP TABLE IF EXISTS `notifications`;
DROP TABLE IF EXISTS `media_files`;
DROP TABLE IF EXISTS `friendly_links`;
DROP TABLE IF EXISTS `content_stats`;
DROP TABLE IF EXISTS `comments`;
DROP TABLE IF EXISTS `comment_likes`;
DROP TABLE IF EXISTS `categories`;
DROP TABLE IF EXISTS `auth_tokens`;
DROP TABLE IF EXISTS `articles`;
DROP TABLE IF EXISTS `article_views`;
DROP TABLE IF EXISTS `article_tags`;
DROP TABLE IF EXISTS `article_revisions`;
DROP TABLE IF EXISTS `article_likes`;
DROP TABLE IF EXISTS `article_categories`;
DROP TABLE IF EXISTS `article_bookmarks`;
SET FOREIGN_KEY_CHECKS = 1;
