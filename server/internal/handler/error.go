// Package handler HTTP请求处理层
package handler

import (
	"context"
	"errors"
	"log/slog"

	"MyBlog/internal/domain"
	"MyBlog/internal/logger"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// unclassifiedErrorText 未分类错误的响应文案。
// 原始错误可能携带表名、列名或 SQL 片段，仅入日志不进入响应体。
const unclassifiedErrorText = "服务器内部错误"

// serviceErrorMapping 服务层错误到传输层分档的注册项。
// Name 为哨兵错误在 contracts/errors.yaml 的 mapping 节中的引用名，
// 契约锚定测试按此名双向比对，使注册表与契约同增同减。
type serviceErrorMapping struct {
	Name string
	Err  error
	Code int
}

// serviceErrorMappings 错误分档的唯一权威表，切片顺序即匹配优先级。
// 新增哨兵错误必须同时登记本表与 contracts/errors.yaml 的 mapping 节，
// 任一侧缺失由 TestErrorMappingMatchesContract 拦截，禁止在模块内另起私有映射分支。
var serviceErrorMappings = []serviceErrorMapping{
	{Name: "service.ErrUnauthenticated", Err: service.ErrUnauthenticated, Code: response.CodeAuth},
	{Name: "service.ErrPermissionDenied", Err: service.ErrPermissionDenied, Code: response.CodeForbid},

	{Name: "domain.ErrArticleNotFound", Err: domain.ErrArticleNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrCategoryNotFound", Err: domain.ErrCategoryNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrTagNotFound", Err: domain.ErrTagNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrCommentNotFound", Err: domain.ErrCommentNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrMediaNotFound", Err: domain.ErrMediaNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrNotificationNotFound", Err: domain.ErrNotificationNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrFriendlyLinkNotFound", Err: domain.ErrFriendlyLinkNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrSettingNotFound", Err: domain.ErrSettingNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrUserNotFound", Err: domain.ErrUserNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrFollowNotFound", Err: domain.ErrFollowNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrDictTypeNotFound", Err: domain.ErrDictTypeNotFound, Code: response.CodeNotFound},
	{Name: "domain.ErrDictItemNotFound", Err: domain.ErrDictItemNotFound, Code: response.CodeNotFound},

	{Name: "service.ErrInvalidRequest", Err: service.ErrInvalidRequest, Code: response.CodeInvalid},
	{Name: "service.ErrUsernameTaken", Err: service.ErrUsernameTaken, Code: response.CodeInvalid},
	{Name: "service.ErrEmailTaken", Err: service.ErrEmailTaken, Code: response.CodeInvalid},
}

// HandleServiceError 将 service 层返回的错误按注册表映射为对应的 HTTP 响应。
// 命中哨兵错误时分档返回 401 / 403 / 404 / 400；未命中时归入内部错误，
// 响应固定为脱敏文案，原始错误只写日志，避免内部细节经响应体泄漏。
// 注册表由 contracts/errors.yaml 的 mapping 节锚定，契约变更须先改契约再改表。
func HandleServiceError(c *gin.Context, err error) {
	for _, mapping := range serviceErrorMappings {
		if errors.Is(err, mapping.Err) {
			response.Error(c, mapping.Code, err.Error())
			return
		}
	}

	// 非哨兵错误统一脱敏，原始信息落日志供排障，防止内部细节经响应体泄漏。
	// 日志带请求标识与用户标识，运维可据此定位到具体请求。
	attributes := []slog.Attr{slog.Any(logger.FieldError, err)}
	if operatorID, ok := getOperatorID(c); ok {
		attributes = append(attributes, slog.Any(logger.FieldUserID, operatorID))
	}
	logger.FromGin(c).LogAttrs(
		context.Background(), slog.LevelError, "未分类的服务错误", attributes...,
	)
	response.InternalError(c, unclassifiedErrorText)
}
