// 依赖装配文件：与 main.go 同属 cmd/myblog 组合根，按业务域集中构造全部依赖。
// A4 约束下依赖只允许在组合根内构造，本文件是组合根的延伸而非新的实例化点。
package main

import (
	"MyBlog/internal/config"
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/router"
	"MyBlog/internal/service"

	"gorm.io/gorm"
)

// newDependencies 按业务域装配全部仓储、服务与处理器，返回路由层依赖。
func newDependencies(cfg *config.Config, db *gorm.DB) *router.Dependencies {
	repos := newRepositories(db)
	services := newServices(cfg, repos)
	handlers := newHandlers(services)

	return &router.Dependencies{
		UserHandler:         handlers.user,
		ArticleHandler:      handlers.article,
		CategoryHandler:     handlers.category,
		TagHandler:          handlers.tag,
		CommentHandler:      handlers.comment,
		MediaHandler:        handlers.media,
		SettingHandler:      handlers.setting,
		FriendlyLinkHandler: handlers.friendlyLink,
		StatsHandler:        handlers.stats,
		NotificationHandler: handlers.notification,
		UserFollowHandler:   handlers.userFollow,
		JWTService:          services.jwt,
		IdentityProvider:    services.identity,
		RBACService:         services.rbac,
	}
}

// appRepositories 集中持有全部仓储实例。
type appRepositories struct {
	user         repository.UserRepository
	article      repository.ArticleRepositoryInterface
	category     repository.CategoryRepositoryInterface
	tag          repository.TagRepositoryInterface
	comment      repository.CommentRepositoryInterface
	media        repository.MediaRepositoryInterface
	setting      repository.SettingRepositoryInterface
	friendlyLink repository.FriendlyLinkRepositoryInterface
	stats        repository.StatsRepositoryInterface
	notification repository.NotificationRepositoryInterface
	follow       repository.UserFollowRepositoryInterface
}

// newRepositories 构造全部仓储实例。
func newRepositories(db *gorm.DB) *appRepositories {
	return &appRepositories{
		user:         repository.NewUserRepository(db),
		article:      repository.NewArticleRepository(db),
		category:     repository.NewCategoryRepository(db),
		tag:          repository.NewTagRepository(db),
		comment:      repository.NewCommentRepository(db),
		media:        repository.NewMediaRepository(db),
		setting:      repository.NewSettingRepository(db),
		friendlyLink: repository.NewFriendlyLinkRepository(db),
		stats:        repository.NewStatsRepository(db),
		notification: repository.NewNotificationRepository(db),
		follow:       repository.NewUserFollowRepository(db),
	}
}

// appServices 集中持有全部服务实例。
type appServices struct {
	jwt          service.JWTService
	identity     middleware.IdentityProvider
	rbac         service.RBACService
	user         service.UserService
	article      service.ArticleServiceInterface
	category     service.CategoryServiceInterface
	tag          service.TagServiceInterface
	comment      service.CommentServiceInterface
	media        service.MediaServiceInterface
	setting      service.SettingServiceInterface
	friendlyLink service.FriendlyLinkServiceInterface
	stats        service.StatsServiceInterface
	notification service.NotificationServiceInterface
	follow       service.UserFollowServiceInterface
}

// newServices 构造全部服务实例，跨域依赖经仓储参数显式传递。
func newServices(cfg *config.Config, repos *appRepositories) *appServices {
	// 从配置加载 RBAC 权限表，作为运行期唯一权威。
	service.LoadRBACConfig(cfg.RBAC.RoleHierarchy, cfg.RBAC.RolePermissions)

	jwtService := service.NewJWTService(cfg)
	rbacService := service.NewRBACService()

	return &appServices{
		jwt:          jwtService,
		identity:     middleware.NewIdentityProvider(jwtService, repos.user),
		rbac:         rbacService,
		user:         service.NewUserService(repos.user, jwtService, rbacService),
		article:      service.NewArticleService(repos.article, repos.user, rbacService, repos.stats, repos.notification),
		category:     service.NewCategoryService(repos.category),
		tag:          service.NewTagService(repos.tag),
		comment:      service.NewCommentService(repos.comment, repos.article, repos.setting, repos.notification),
		media:        service.NewMediaService(repos.media, cfg),
		setting:      service.NewSettingService(repos.setting),
		friendlyLink: service.NewFriendlyLinkService(repos.friendlyLink),
		stats:        service.NewStatsService(repos.stats),
		notification: service.NewNotificationService(repos.notification),
		follow:       service.NewUserFollowService(repos.follow, repos.user, repos.notification, repos.article),
	}
}

// appHandlers 集中持有全部处理器实例。
type appHandlers struct {
	user         handler.UserHandlerInterface
	article      handler.ArticleHandlerInterface
	category     handler.CategoryHandlerInterface
	tag          handler.TagHandlerInterface
	comment      handler.CommentHandlerInterface
	media        handler.MediaHandlerInterface
	setting      handler.SettingHandlerInterface
	friendlyLink handler.FriendlyLinkHandlerInterface
	stats        handler.StatsHandlerInterface
	notification handler.NotificationHandlerInterface
	userFollow   handler.UserFollowHandlerInterface
}

// newHandlers 构造全部处理器实例。
func newHandlers(services *appServices) *appHandlers {
	return &appHandlers{
		user:         handler.NewUserHandler(services.user),
		article:      handler.NewArticleHandler(services.article),
		category:     handler.NewCategoryHandler(services.category),
		tag:          handler.NewTagHandler(services.tag),
		comment:      handler.NewCommentHandler(services.comment),
		media:        handler.NewMediaHandler(services.media),
		setting:      handler.NewSettingHandler(services.setting),
		friendlyLink: handler.NewFriendlyLinkHandler(services.friendlyLink),
		stats:        handler.NewStatsHandler(services.stats),
		notification: handler.NewNotificationHandler(services.notification),
		userFollow:   handler.NewUserFollowHandler(services.follow),
	}
}
