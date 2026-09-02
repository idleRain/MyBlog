package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// UserFollowRoutes 用户关注路由模块
type UserFollowRoutes struct {
	followHandler handler.UserFollowHandlerInterface
	jwtService    service.JWTService
}

// NewUserFollowRoutes 创建用户关注路由模块
func NewUserFollowRoutes(
	followHandler handler.UserFollowHandlerInterface,
	jwtService service.JWTService,
) *UserFollowRoutes {
	return &UserFollowRoutes{
		followHandler: followHandler,
		jwtService:    jwtService,
	}
}

// RegisterRoutes 注册用户关注相关路由
func (uf *UserFollowRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 关注与取消关注接口，需要登录。
	authUsers := api.Group("/users")
	authUsers.Use(middleware.Auth(uf.jwtService))
	{
		authUsers.POST("/follow", uf.followHandler.Follow)         // 关注用户
		authUsers.POST("/unfollow", uf.followHandler.Unfollow)     // 取消关注
	}

	// 关注关系查询接口，公开可访问。
	publicUsers := api.Group("/users")
	{
		publicUsers.POST("/followers", uf.followHandler.ListFollowers) // 粉丝列表
		publicUsers.POST("/following", uf.followHandler.ListFollowing) // 关注列表
	}
}
