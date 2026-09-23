// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// HandleServiceError 将 service 层返回的错误按语义映射为对应的 HTTP 响应。
// "资源不存在"哨兵错误统一映射为 404，权限不足哨兵错误映射为 403，
// 业务规则校验失败映射为 400，其余作为内部错误处理，避免一律返回 500。
// 后续接入错误码契约时，可将本表迁移为基于错误码的判定，保持单一映射点。
func HandleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPermissionDenied):
		response.Forbidden(c, err.Error())
	case errors.Is(err, domain.ErrArticleNotFound),
		errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrTagNotFound),
		errors.Is(err, domain.ErrCommentNotFound),
		errors.Is(err, domain.ErrMediaNotFound),
		errors.Is(err, domain.ErrNotificationNotFound),
		errors.Is(err, domain.ErrFriendlyLinkNotFound),
		errors.Is(err, domain.ErrSettingNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrFollowNotFound),
		errors.Is(err, domain.ErrDictTypeNotFound),
		errors.Is(err, domain.ErrDictItemNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrInvalidRequest),
		errors.Is(err, service.ErrUsernameTaken),
		errors.Is(err, service.ErrEmailTaken):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
