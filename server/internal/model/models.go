package model

import (
	"time"

	"gorm.io/gorm"
)

// Models 返回所有需要进行数据库迁移的模型
func Models() []interface{} {
	return []interface{}{
		// 用户模块
		&User{},
		&UserSession{},
		&UserActivity{},
		&AuthToken{},

		// 内容模块
		&Category{},
		&Tag{},
		&Article{},
		&ArticleTag{},
		&ArticleCategory{},
		&ArticleView{},
		&ArticleRevision{},

		// 评论模块
		&Comment{},
		&CommentLike{},

		// 互动模块
		&ArticleLike{},
		&ArticleBookmark{},
		&UserFollow{},
		&Notification{},

		// 媒体模块
		&MediaFile{},

		// 站点运营模块
		&Setting{},
		&FriendlyLink{},

		// 统计与日志模块
		&OperationLog{},
		&SearchLog{},
		&ContentStats{},
	}
}

// mysqlTableOptions 建表时统一附加的物理选项，确保存储引擎与字符集在全部表上保持一致。
const mysqlTableOptions = "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"

// AutoMigrate 执行自动迁移，建表附加统一表选项，迁移完成后同步表注释。
func AutoMigrate(db *gorm.DB) error {
	if err := db.Set("gorm:table_options", mysqlTableOptions).AutoMigrate(Models()...); err != nil {
		return err
	}
	return syncTableComments(db)
}

// 定义常用的查询作用域

// Published 查询已发布的文章
func Published(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", ArticleStatusPublished)
}

// Approved 查询已审核通过的评论
func Approved(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", CommentStatusApproved)
}

// Active 查询活跃用户
func Active(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", UserStatusActive)
}

// Public 查询公开的媒体文件
func Public(db *gorm.DB) *gorm.DB {
	return db.Where("is_public = ?", true)
}

// Featured 查询精选内容
func Featured(db *gorm.DB) *gorm.DB {
	return db.Where("is_featured = ?", true)
}

// OrderByCreatedAt 按创建时间排序
func OrderByCreatedAt(desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("created_at DESC")
		}
		return db.Order("created_at ASC")
	}
}

// OrderByUpdatedAt 按更新时间排序
func OrderByUpdatedAt(desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("updated_at DESC")
		}
		return db.Order("updated_at ASC")
	}
}

// OrderByPublishedAt 按发布时间排序
func OrderByPublishedAt(desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("published_at DESC")
		}
		return db.Order("published_at ASC")
	}
}

// OrderByViewCount 按浏览量排序
func OrderByViewCount(desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("view_count DESC")
		}
		return db.Order("view_count ASC")
	}
}

// WithAuthor 预加载作者信息
func WithAuthor(db *gorm.DB) *gorm.DB {
	return db.Preload("Author")
}

// WithCategory 预加载分类信息
func WithCategory(db *gorm.DB) *gorm.DB {
	return db.Preload("Category")
}

// WithTags 预加载标签信息
func WithTags(db *gorm.DB) *gorm.DB {
	return db.Preload("Tags")
}

// WithUser 预加载用户信息
func WithUser(db *gorm.DB) *gorm.DB {
	return db.Preload("User")
}

// WithUploader 预加载上传者信息
func WithUploader(db *gorm.DB) *gorm.DB {
	return db.Preload("Uploader")
}

// ========== 增强功能查询作用域 ==========

// Unread 查询未读通知
func Unread(db *gorm.DB) *gorm.DB {
	return db.Where("is_read = ?", false)
}

// ByAction 根据操作类型查询日志
func ByAction(action string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("action = ?", action)
	}
}

// ByContentType 根据内容类型查询统计
func ByContentType(contentType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("content_type = ?", contentType)
	}
}

// ByStatType 根据统计类型查询
func ByStatType(statType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("stat_type = ?", statType)
	}
}

// ThisWeek 查询本周数据
func ThisWeek(db *gorm.DB) *gorm.DB {
	return db.Where("created_at >= ?", time.Now().AddDate(0, 0, -7))
}

// ThisMonth 查询本月数据
func ThisMonth(db *gorm.DB) *gorm.DB {
	return db.Where("created_at >= ?", time.Now().AddDate(0, -1, 0))
}

// WithFollower 预加载关注者信息
func WithFollower(db *gorm.DB) *gorm.DB {
	return db.Preload("Follower")
}

// WithFollowing 预加载被关注者信息
func WithFollowing(db *gorm.DB) *gorm.DB {
	return db.Preload("Following")
}

// WithRelated 预加载通知关联信息
func WithRelated(db *gorm.DB) *gorm.DB {
	return db.Preload("User")
}

// OrderByStatValue 按统计值排序
func OrderByStatValue(desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("stat_value DESC")
		}
		return db.Order("stat_value ASC")
	}
}
