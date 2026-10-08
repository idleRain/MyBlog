-- 回滚热点排序与昵称检索索引。

DROP INDEX `idx_users_nickname` ON `users`;
DROP INDEX `idx_articles_like_count` ON `articles`;
