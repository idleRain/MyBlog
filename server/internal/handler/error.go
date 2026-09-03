// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// HandleServiceError 将 service 层返回的错误按语义映射为对应的 HTTP 响应。
// 已知的"资源不存在"哨兵错误统一映射为 404，其余作为内部错误处理，避免一律返回 500。
// 后续接入错误码契约时，可将本表迁移为基于错误码的判定，保持单一映射点。
func HandleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrArticleNotFound),
		errors.Is(err, repository.ErrCategoryNotFound),
		errors.Is(err, repository.ErrTagNotFound),
		errors.Is(err, repository.ErrCommentNotFound),
		errors.Is(err, repository.ErrMediaNotFound),
		errors.Is(err, repository.ErrNotificationNotFound),
		errors.Is(err, repository.ErrFriendlyLinkNotFound),
		errors.Is(err, repository.ErrSettingNotFound),
		errors.Is(err, repository.ErrUserNotFound),
		errors.Is(err, repository.ErrFollowNotFound):
		response.NotFound(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
