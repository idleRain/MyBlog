# MyBlog 数据库架构设计

## 概述

本文档定义 MyBlog 项目的完整数据库架构设计。采用 MySQL 8.0 作为主数据库，使用 GORM 作为 ORM 框架，覆盖用户管理、内容管理（含多语言翻译）、评论系统、媒体管理、互动功能、字典管理、系统监控等模块。完整 DDL 见 `docs/database/schema.sql`，实际表结构以 GORM 模型定义为准；生产环境经 golang-migrate 执行 `server/migrations/` 增量迁移，开发模式启动时通过 AutoMigrate 同步。

**数据库规模**：30 张业务表，按职责划分为 8 个模块。
- 用户模块 4 张：users、user_sessions、user_activities、auth_tokens
- 内容模块 10 张：categories、tags、articles、article_tags、article_categories、article_views、article_revisions、article_translations、category_translations、tag_translations
- 评论模块 2 张：comments、comment_likes
- 互动模块 4 张：article_likes、article_bookmarks、user_follows、notifications
- 媒体模块 1 张：media_files
- 站点运营模块 2 张：settings、friendly_links
- 字典模块 4 张：dict_types、dict_items、dict_type_translations、dict_item_translations
- 统计日志模块 3 张：operation_logs、search_logs、content_stats

**最后更新**：表清单与模型层对齐（v3.1）。

## 设计原则

表结构的生命周期与 comment 约定、唯一索引与普通索引、OnDelete 策略、时间字段精度、多段写库事务化已被根 `AGENTS.md` 第 5 节「后端约定」的数据库条文覆盖，本文档不再复述。本节仅保留该处未覆盖的三条本架构约定：

1. **规范化设计**：遵循第三范式，多对多关系使用独立关联表，避免冗余存储。
2. **树形结构**：分类与评论统一使用 parent_id、root_id、level 三件套，分类额外使用 path 物化路径支持整棵子树的一次查询。
3. **状态取值与删除形态**：状态类字段使用命名常量枚举并预留合法取值，如用户状态预留 2-锁定、媒体文件预留 processing/failed；软删除列统一使用 deleted_at，字典等读多写少的配置表采用硬删除以避免与唯一索引的占位冲突。

## 模块与表结构

### 1. 用户管理模块

#### users（用户表）
承载账号身份、个人资料与安全状态三类信息。
- 身份字段：username、email、phone 均全局唯一，phone 为可空指针类型，空值不参与唯一约束冲突。
- 资料字段：nickname、avatar、cover_image、bio、website、location、gender、birthday、timezone、locale。
- 安全字段：failed_login_count、locked_until、password_changed_at、last_login_at、last_login_ip、login_count、email_verified_at、remark。
- 状态：status 使用 tinyint 命名常量，1-正常 0-禁用 2-锁定；remark 与密码相关字段不对外输出。
- 索引：username/email/phone 唯一索引，role/status/nickname/last_login_at/deleted_at 普通索引。

#### user_sessions（用户会话表）
管理登录设备与令牌轮换。
- refresh_token 唯一索引，device_type 记录设备类型，logout_at 与 last_refresh_at 记录会话生命周期。
- 复合索引 idx_user_active(user_id, is_active) 支撑活跃会话列表查询。

#### user_activities（用户活动日志表）
记录用户行为轨迹。
- 新增 status/error_message/duration_ms 描述操作结果，复合索引 idx_user_created(user_id, created_at) 支撑时间轴。

#### auth_tokens（认证令牌表）
支撑密码找回与邮箱验证等带时效的一次性流程。
- 令牌原文不下库，仅存储 SHA256 哈希值，token_hash 唯一索引。
- token_type 区分用途，expires_at 与 used_at 控制时效与一次性核销。

### 2. 内容管理模块

#### categories（分类表）
树形结构支撑栏目导航，parent_id/root_id/level/path 四字段描述层级。
- path 为物化路径，形如 /1/5/12，可在单次查询中取得整棵子树。
- status 控制展示与隐藏，article_count 由发布逻辑异步维护。

#### tags（标签表）
文章主题的轻量归类维度，name/slug 唯一索引，status 控制启用与隐藏。

#### articles（文章表）
博客核心内容实体。
- 生命周期字段：scheduled_at 定时发布、published_at、edited_at、archived_at、last_comment_at。
- 来源字段：origin_type 区分原创/翻译/转载，source_url 与 source_author 记录出处。
- 权限字段：access_password 支持私密文章密码访问，comment_enabled 控制评论开关。
- 计数与版本：view_count/like_count/bookmark_count/comment_count 冗余计数，version 对应 article_revisions 修订版本。
- 索引：slug 唯一索引；复合索引 idx_status_published 支撑已发布列表，idx_author_status 支撑作者文章列表。

#### article_tags / article_categories（关联表）
多对多挂载关系，复合唯一索引确保一文一标签、一文一分类不重复。

#### article_views（浏览统计表）
按 article_id、visitor_id、view_date 三元组复合唯一索引去重计数，duration_seconds 记录停留时长。

#### article_revisions（修订历史表）
保存每次正文保存后的快照，支持版本回滚与差异对比。
- (article_id, revision_no) 复合唯一索引保证版本号连续，is_autosave 区分手动保存与自动保存。

#### article_translations / category_translations / tag_translations（内容翻译表）
主表列恒为缺省语言内容，翻译行以 locale 区分其他语言，允许按字段部分填写，输出时缺失翻译的字段回退主列。
- 均在 (主表 ID, locale) 上建立复合唯一索引，locale 单列索引支撑按语言批量查询。
- 文章翻译行额外建 ngram 全文索引 ft_article_translations_search，与主表索引共同支撑多语言检索。

### 3. 评论系统模块

#### comments（评论表）
树形结构支持多级回复与审核流，parent_id/root_id/level 描述评论树。
- status 审核流：pending/approved/rejected/spam/trash，reported_count 支撑举报复核。
- edited_at 记录编辑痕迹，复合索引 idx_article_status_created(article_id, status, created_at) 支撑文章评论分页。

#### comment_likes（评论点赞表）
复合唯一索引确保同一用户对同一评论仅一次有效点赞。

### 4. 互动模块

#### article_likes / article_bookmarks（点赞与收藏表）
复合唯一索引杜绝重复点赞与重复收藏，article_bookmarks 增加 note 字段。

#### user_follows（用户关注表）
复合唯一索引 uk_follow_relation 杜绝重复关注，CHECK 约束 chk_follow_self 阻止自我关注。

#### notifications（系统通知表）
站内消息中心。
- sender_id 记录触发用户，action_url 记录跳转地址，read_at 与 is_read 记录已读状态。
- 软删除支持用户清理通知后后台留档，复合索引 idx_user_read(user_id, is_read) 支撑未读统计。

### 5. 媒体管理模块

#### media_files（媒体文件表）
管理上传资源元信息与生命周期。
- 生命周期字段：status 区分可用/处理中/处理失败/文件丢失，processed_at 记录后处理完成时间。
- 元信息字段：width/height/duration_seconds/alt_text、file_hash 支撑秒传去重、storage_type 区分存储后端。
- usage_count 记录被正文引用次数，删除前需要校验。

### 6. 站点运营模块

#### settings（系统设置表）
键值化全局配置。
- label 展示名称，is_sensitive 显式标记敏感项，输出时统一脱敏。
- updated_by 记录最后更新人，外键 SET NULL。

#### friendly_links（友情链接表）
管理互链申请与展示的完整生命周期。
- url 唯一索引防止重复收录，status 审核流：pending/active/hidden/rejected，is_reciprocal 记录回链状态。

### 7. 字典管理模块

#### dict_types / dict_items（字典类型与字典项表）
字典为读多写少的配置数据，承载状态类枚举的动态配置维度。
- dict_types.code 唯一索引作为业务侧定位标识；status 命名常量枚举 1-生效 0-停用，sort_order 控制排序，extra(json) 预留颜色图标等展示元数据。
- dict_items 在 (type_id, value) 上建立复合唯一索引确保同类型内取值唯一，type_id 单列索引支撑按类型取项。
- 两表均采用硬删除策略，不设 deleted_at，避免软删除与唯一索引的占位冲突；删除类型时经外键级联清理字典项与翻译行。

#### dict_type_translations / dict_item_translations（字典翻译表）
与内容翻译表同构，在 (type_id, locale)、(item_id, locale) 上建立复合唯一索引，locale 单列索引支撑按语言批量查询。

### 8. 统计和日志模块

#### operation_logs（操作日志表）
安全审计与问题追踪。
- status/error_message/duration_ms 描述操作结果，trace_id 串联一次请求内的多条日志。

#### search_logs（搜索记录表）
搜索行为分析，keyword、results_count、duration_ms、status 记录搜索质量。

#### content_stats（内容统计表）
多维度聚合指标，(content_type, content_id, stat_type, stat_date) 复合唯一索引防止重复入账。

## 索引设计策略

### 唯一索引（防重与幂等）
- users：username、email、phone
- user_sessions：refresh_token
- auth_tokens：token_hash
- categories：slug；tags：name、slug；articles：slug
- media_files：stored_name
- settings：key_name；friendly_links：url
- dict_types：code
- 复合唯一：article_tags(article_id, tag_id)、article_categories(article_id, category_id)、article_views(article_id, visitor_id, view_date)、article_revisions(article_id, revision_no)、article_translations(article_id, locale)、category_translations(category_id, locale)、tag_translations(tag_id, locale)、article_likes(article_id, user_id)、comment_likes(comment_id, user_id)、article_bookmarks(article_id, user_id)、user_follows(follower_id, following_id)、dict_items(type_id, value)、dict_type_translations(type_id, locale)、dict_item_translations(item_id, locale)、content_stats(content_type, content_id, stat_type, stat_date)

### 复合索引（高频查询）
- user_sessions(user_id, is_active)：活跃会话列表
- user_activities(user_id, created_at)：用户活动时间轴
- articles(status, published_at)：已发布文章列表
- articles(author_id, status)：作者文章列表
- comments(article_id, status, created_at)：文章评论分页
- notifications(user_id, is_read)：未读通知统计

### 全文索引（多语言检索）
- articles(title, content, summary)：ft_articles_search，ngram 分词，支撑中文按双字切分匹配
- article_translations(title, content, summary)：ft_article_translations_search，覆盖非默认语言的检索

## 约束与删除策略

- 级联删除：文章删除级联清理标签关联、分类关联、浏览记录、修订历史、翻译行、评论、点赞与收藏；用户删除级联清理会话、令牌、其文章、文章点赞、评论点赞、收藏、关注与通知；字典类型删除级联清理字典项与翻译行。
- 置空删除：用户删除后其活动日志、操作日志、搜索日志与评论保留记录但外键置空；文章修订的 editor_id、通知的 sender_id、设置的 updated_by 同样声明 SET NULL。
- 无删除策略声明：media_files.uploader_id 未声明 OnDelete，按数据库默认 RESTRICT 处理，删除用户前须先处理其媒体记录。
- 检查约束：user_follows 禁止自我关注。

## 数据一致性约定

状态类字段的命名常量枚举与软删除列以 `internal/model` 的常量定义与实体 tag 为准，时间字段精度与多段写库事务口径以根 `AGENTS.md` 第 5 节「后端约定」为准，本文档不再复述。本架构的实现结果：冗余计数字段为 view_count、like_count、bookmark_count、comment_count、reply_count、article_count、usage_count，在业务写入时同步维护。

## 演进与运维约定

### 迁移双轨

- **生产轨（golang-migrate 严格轨）**：非 debug 模式启动时经 `server/internal/database` 的迁移链路执行 `server/migrations/` 增量迁移。基线为 `000001_init_schema`，由开发库 AutoMigrate 生成物导出固化，生产部署以迁移版本为准。
- **开发轨（AutoMigrate）**：debug 模式保留 GORM AutoMigrate 快速同步，模型改动直接生效；与迁移基线出现漂移时以手写增量迁移对齐。
- **演进纪律**：后续任何 schema 变更必须手写增量迁移并随版本提交，禁止回退生产轨到 AutoMigrate，禁止仅依赖 AutoMigrate 演进生产表结构。

1. **新增表**：在 `server/internal/model/` 创建模型，加入 `Models()` 注册，并在 `table_comments.go` 补充表注释；生产轨同步手写增量迁移建表。
2. **修改表结构**：修改 GORM 模型后开发环境由 AutoMigrate 自动同步；生产环境以手写增量迁移执行，新增字段与索引对既有数据无影响。
3. **新字段约束**：新增字段必须携带 comment、贴合真实数据长度，并考虑是否补充索引。
4. **数据清理**：定期清理过期会话、过期认证令牌、软删除数据与历史日志。
5. **备份策略**：每日全量 `mysqldump` 并打包上传目录；当前链路不归档 binlog，故 RPO 为 24 小时。基于 binlog 的时间点恢复属可选增强，方案见 `docs/operations/backup-restore.md` 第 7 节。

## 文档版本

**文档版本**：v3.1
**最后更新**：表清单与模型层对齐
**更新内容**：表数量由 23 张校正为 30 张，补入内容翻译表 3 张与字典模块 4 张；补齐全文索引、字典与翻译表的索引清单；按模型层修正删除策略（媒体上传者外键无 OnDelete 声明）；设计原则与数据一致性约定中与根 `AGENTS.md` 重复的条文压缩为指针，仅保留规范化、树形结构与状态/删除形态约定。
