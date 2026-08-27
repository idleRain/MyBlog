package model

import (
	"fmt"

	"gorm.io/gorm"
)

// tableComments 维护各表的业务注释。GORM 的 AutoMigrate 不支持表级注释标签，因此在此集中声明并在迁移后同步。
var tableComments = map[string]string{
	// 用户模块
	"users":           "用户表，存储账号身份、个人资料与安全状态",
	"user_sessions":   "用户会话表，管理登录设备与令牌轮换",
	"user_activities": "用户活动日志表，记录用户行为轨迹",
	"auth_tokens":     "认证令牌表，支撑密码找回与邮箱验证流程",

	// 内容模块
	"categories":         "文章分类表，树形结构支撑栏目导航",
	"tags":               "标签表，文章主题的轻量归类维度",
	"articles":           "文章表，博客核心内容实体",
	"article_tags":       "文章标签关联表，多对多挂载关系",
	"article_categories": "文章分类关联表，支持一文多分类",
	"article_views":      "文章浏览统计表，按访客与日期去重计数",
	"article_revisions":  "文章修订历史表，保存正文快照支持回滚",

	// 评论模块
	"comments":      "评论表，树形结构支持多级回复与审核流",
	"comment_likes": "评论点赞表",

	// 互动模块
	"article_likes":     "文章点赞表",
	"article_bookmarks": "文章收藏表",
	"user_follows":      "用户关注表，社交关系维度",
	"notifications":     "系统通知表，站内消息中心",

	// 媒体模块
	"media_files": "媒体文件表，管理上传资源元信息与生命周期",

	// 站点运营模块
	"settings":       "系统设置表，键值化全局配置",
	"friendly_links": "友情链接表，互链申请与展示管理",

	// 统计与日志模块
	"operation_logs": "操作日志表，安全审计与问题追踪",
	"search_logs":    "搜索记录表，搜索行为分析",
	"content_stats":  "内容统计表，多维度聚合指标",
}

// syncTableComments 在迁移完成后将表注释同步到数据库，保证库内自描述能力。
// 表名与注释均来自上方常量表，不存在 SQL 注入面。
func syncTableComments(db *gorm.DB) error {
	for table, comment := range tableComments {
		alterSQL := fmt.Sprintf("ALTER TABLE `%s` COMMENT '%s'", table, comment)
		if err := db.Exec(alterSQL).Error; err != nil {
			return fmt.Errorf("同步表 %s 注释失败: %w", table, err)
		}
	}
	return nil
}
