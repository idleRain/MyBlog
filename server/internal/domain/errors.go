// Package domain 领域层：实体、DTO 与跨层共享的领域语义。
package domain

import "errors"

// 资源不存在类哨兵错误，handler 统一映射为 404。
// 哨兵定义在领域层作为全系统唯一的错误语义真相源，
// service 与 handler 直接引用本包，禁止为引用错误语义而依赖持久化包。
var (
	ErrArticleNotFound      = errors.New("文章不存在")
	ErrCategoryNotFound     = errors.New("分类不存在")
	ErrTagNotFound          = errors.New("标签不存在")
	ErrCommentNotFound      = errors.New("评论不存在")
	ErrMediaNotFound        = errors.New("媒体文件不存在")
	ErrNotificationNotFound = errors.New("通知不存在")
	ErrFriendlyLinkNotFound = errors.New("友情链接不存在")
	ErrSettingNotFound      = errors.New("设置项不存在")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrFollowNotFound       = errors.New("关注关系不存在")
	ErrDictTypeNotFound     = errors.New("字典类型不存在")
	ErrDictItemNotFound     = errors.New("字典项不存在")
)
