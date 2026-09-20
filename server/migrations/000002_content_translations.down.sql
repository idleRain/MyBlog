-- 分期迁移 000002 回滚：删除内容多语言翻译表
-- 翻译行随主实体级联删除，回滚仅移除翻译表本身，不影响主表数据。

DROP TABLE IF EXISTS `tag_translations`;
DROP TABLE IF EXISTS `category_translations`;
DROP TABLE IF EXISTS `article_translations`;
