-- 分期迁移 000003：字典管理
-- 字典类型与字典项均为硬删除策略，删除类型时经外键级联清理字典项与翻译行。

CREATE TABLE `dict_types` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '字典类型ID',
  `code` varchar(50) NOT NULL COMMENT '字典码，业务侧唯一定位标识，如 tag_status',
  `name` varchar(50) NOT NULL COMMENT '字典类型名称，缺省语言',
  `description` varchar(200) DEFAULT NULL COMMENT '字典类型描述，缺省语言',
  `status` tinyint DEFAULT '1' COMMENT '生效状态：1-生效 0-停用',
  `sort_order` int DEFAULT '0' COMMENT '排序权重，数值小的靠前',
  `extra` json DEFAULT NULL COMMENT '预留扩展字段，存储颜色图标等展示元数据',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`),
  KEY `idx_dict_types_status` (`status`),
  KEY `idx_dict_types_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典类型表，状态类枚举的动态配置维度';

CREATE TABLE `dict_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '字典项ID',
  `type_id` bigint unsigned NOT NULL COMMENT '所属字典类型ID',
  `value` varchar(50) NOT NULL COMMENT '字典项值，同类型内唯一',
  `label` varchar(50) NOT NULL COMMENT '字典项显示名，缺省语言',
  `description` varchar(200) DEFAULT NULL COMMENT '字典项描述，缺省语言',
  `status` tinyint DEFAULT '1' COMMENT '生效状态：1-生效 0-停用',
  `sort_order` int DEFAULT '0' COMMENT '排序权重，数值小的靠前',
  `extra` json DEFAULT NULL COMMENT '预留扩展字段，存储颜色图标等展示元数据',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dict_item` (`type_id`,`value`),
  KEY `idx_dict_items_type_id` (`type_id`),
  KEY `idx_dict_items_status` (`status`),
  KEY `idx_dict_items_sort_order` (`sort_order`),
  CONSTRAINT `fk_dict_items_type` FOREIGN KEY (`type_id`) REFERENCES `dict_types` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典项表，字典类型下的可选值集合';

CREATE TABLE `dict_type_translations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '翻译ID',
  `type_id` bigint unsigned NOT NULL COMMENT '字典类型ID',
  `locale` varchar(10) NOT NULL COMMENT '语言标识，主子标签形式，如 zh、en',
  `name` varchar(50) DEFAULT NULL COMMENT '该语言字典类型名称',
  `description` varchar(200) DEFAULT NULL COMMENT '该语言字典类型描述',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dict_type_translation` (`type_id`,`locale`),
  KEY `idx_dict_type_translations_locale` (`locale`),
  CONSTRAINT `fk_dict_type_translations_type` FOREIGN KEY (`type_id`) REFERENCES `dict_types` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典类型翻译表，按语言存储名称与描述的翻译内容';

CREATE TABLE `dict_item_translations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '翻译ID',
  `item_id` bigint unsigned NOT NULL COMMENT '字典项ID',
  `locale` varchar(10) NOT NULL COMMENT '语言标识，主子标签形式，如 zh、en',
  `label` varchar(50) DEFAULT NULL COMMENT '该语言字典项显示名',
  `description` varchar(200) DEFAULT NULL COMMENT '该语言字典项描述',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dict_item_translation` (`item_id`,`locale`),
  KEY `idx_dict_item_translations_locale` (`locale`),
  CONSTRAINT `fk_dict_item_translations_item` FOREIGN KEY (`item_id`) REFERENCES `dict_items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典项翻译表，按语言存储显示名与描述的翻译内容';
