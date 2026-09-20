-- 分期迁移 000003 回滚：按依赖反序删除字典表

DROP TABLE IF EXISTS `dict_item_translations`;
DROP TABLE IF EXISTS `dict_type_translations`;
DROP TABLE IF EXISTS `dict_items`;
DROP TABLE IF EXISTS `dict_types`;
