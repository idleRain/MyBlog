package model

import "gorm.io/gorm"

// Models 返回所有需要进行数据库迁移的模型
func Models() []interface{} {
	return []interface{}{
		// 用户模块
		&User{},
		&UserSession{},
		&UserActivity{},

		// 内容模块
		&Category{},
		&Tag{},
		&Article{},
		&ArticleTag{},
		&ArticleView{},

		// 评论模块
		&Comment{},

		// 媒体模块
		&MediaFile{},

		// 系统设置模块
		&Setting{},
	}
}

// AutoMigrate 执行自动迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
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
