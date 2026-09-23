// Package middleware HTTP横切中间件
package middleware

import (
	"fmt"
	"strings"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// IdentityProvider 当前用户身份解析抽象，由中间件消费方侧定义。
// 实现方封装 Bearer 解析、令牌校验与用户状态核查，屏蔽存储层细节。
type IdentityProvider interface {
	// Resolve 解析当前请求的用户身份，失败时已写入错误响应并中止请求链。
	Resolve(c *gin.Context) (*domain.User, error)
}

// tokenIdentityProvider 基于令牌表与用户仓储的身份解析实现。
type tokenIdentityProvider struct {
	tokenService service.TokenServiceInterface
	userRepo     repository.UserRepository
}

// NewIdentityProvider 创建身份解析器，依赖由组合根注入。
func NewIdentityProvider(tokenService service.TokenServiceInterface, userRepo repository.UserRepository) IdentityProvider {
	return &tokenIdentityProvider{
		tokenService: tokenService,
		userRepo:     userRepo,
	}
}

// Resolve 解析访问令牌并加载完整用户，逐项校验令牌、状态与角色有效性。
func (p *tokenIdentityProvider) Resolve(c *gin.Context) (*domain.User, error) {
	token := c.GetHeader("Authorization")

	if token == "" {
		response.Unauthorized(c, "未提供认证令牌")
		c.Abort()
		return nil, fmt.Errorf("no token")
	}

	// 移除 Bearer 前缀，无前缀时保持原值。
	token = strings.TrimPrefix(token, bearerTokenPrefix)

	// 校验访问令牌，服务端令牌表是身份的唯一权威。
	identity, err := p.tokenService.ValidateAccessToken(token)
	if err != nil {
		response.Unauthorized(c, "无效的认证令牌")
		c.Abort()
		return nil, err
	}

	// 从数据库查询用户信息。
	user, err := p.userRepo.GetByID(identity.UserID)
	if err != nil {
		response.Unauthorized(c, "用户不存在")
		c.Abort()
		return nil, err
	}

	// 验证用户状态。
	if user.Status != model.UserStatusActive {
		response.Forbidden(c, "用户已被禁用")
		c.Abort()
		return nil, fmt.Errorf("user disabled")
	}

	// 验证角色有效性，角色表为无状态纯函数。
	if !service.IsValidRole(user.Role) {
		response.Forbidden(c, "用户角色无效")
		c.Abort()
		return nil, fmt.Errorf("invalid role")
	}

	return user, nil
}
