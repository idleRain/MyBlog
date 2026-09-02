// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// MediaHandlerInterface 媒体处理器接口
type MediaHandlerInterface interface {
	UploadFile(c *gin.Context)
	GetMedia(c *gin.Context)
	ListMedia(c *gin.Context)
	DeleteMedia(c *gin.Context)
}

// MediaHandler 媒体处理器实现
type MediaHandler struct {
	mediaService service.MediaServiceInterface
}

// NewMediaHandler 创建媒体处理器实例
func NewMediaHandler(mediaService service.MediaServiceInterface) MediaHandlerInterface {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// UploadFile 上传文件 POST /api/media/upload
func (h *MediaHandler) UploadFile(c *gin.Context) {
	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	// 解析上传文件，仅接受单文件上传。
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "上传文件参数错误: "+err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "打开上传文件失败: "+err.Error())
		return
	}
	defer file.Close()

	media, err := h.mediaService.UploadFile(
		fileHeader.Filename,
		file,
		fileHeader.Size,
		userID,
		c.ClientIP(),
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "文件上传成功", media)
}

// GetMedia 根据ID获取媒体文件 POST /api/media/get
func (h *MediaHandler) GetMedia(c *gin.Context) {
	type GetMediaRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req GetMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	media, err := h.mediaService.GetMedia(req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrMediaNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, media)
}

// ListMedia 分页查询媒体文件 POST /api/media/list
func (h *MediaHandler) ListMedia(c *gin.Context) {
	var req service.ListMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 从上下文读取操作者身份，未登录用户无法查看媒体列表。
	userID, userIDOK := getOperatorID(c)
	isAdmin, _ := c.Get("isAdmin")
	adminFlag, _ := isAdmin.(bool)

	// 未登录时默认按游客处理，仅能访问公开媒体，由服务层过滤。
	var uploaderID *uint
	if userIDOK {
		uploaderID = &userID
	}

	result, err := h.mediaService.ListMedia(&req, uploaderID, adminFlag)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteMedia 删除媒体文件 POST /api/media/delete
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	type DeleteMediaRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req DeleteMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	isAdmin, _ := c.Get("isAdmin")
	adminFlag, _ := isAdmin.(bool)

	if err := h.mediaService.DeleteMedia(req.ID, userID, adminFlag); err != nil {
		if errors.Is(err, repository.ErrMediaNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.Forbidden(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "文件删除成功", nil)
}
