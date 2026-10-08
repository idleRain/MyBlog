-- 分期迁移 000004：热点排序与昵称检索索引
-- articles.like_count 支撑按点赞数排序的文章列表查询，缺失时触发文件排序；
-- users.nickname 支撑用户列表按昵称的关键词检索路径，缺失时触发全表扫描。

CREATE INDEX `idx_articles_like_count` ON `articles` (`like_count`);
CREATE INDEX `idx_users_nickname` ON `users` (`nickname`);
