// Package service 业务逻辑层
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"MyBlog/internal/config"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"

	"github.com/google/uuid"
)

// MediaServiceInterface 媒体服务接口
type MediaServiceInterface interface {
	// 上传操作
	UploadFile(filename string, reader io.Reader, size int64, uploaderID uint, ipAddress string) (*model.MediaFile, error)

	// 查询操作
	GetMedia(id uint) (*model.MediaFile, error)
	ListMedia(req *ListMediaRequest, uploaderID *uint, isAdmin bool) (*MediaListResponse, error)

	// 删除操作
	DeleteMedia(id uint, operatorID uint, isAdmin bool) error
}

// ListMediaRequest 媒体列表请求
type ListMediaRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Folder   string `json:"folder"`
	MimeType string `json:"mimeType"`
}

// MediaListResponse 媒体列表响应
type MediaListResponse struct {
	Media    []*model.MediaFile `json:"media"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// MediaService 媒体服务实现
type MediaService struct {
	mediaRepo repository.MediaRepositoryInterface
	cfg       *config.Config
}

// NewMediaService 创建媒体服务实例
func NewMediaService(mediaRepo repository.MediaRepositoryInterface, cfg *config.Config) MediaServiceInterface {
	return &MediaService{
		mediaRepo: mediaRepo,
		cfg:       cfg,
	}
}

// UploadFile 上传文件到本地存储并落库，文件哈希相同则复用已有记录。
func (s *MediaService) UploadFile(filename string, reader io.Reader, size int64, uploaderID uint, ipAddress string) (*model.MediaFile, error) {
	// 校验文件大小是否超限。
	maxSize := int64(s.cfg.Media.MaxSizeMB) * 1024 * 1024
	if maxSize > 0 && size > maxSize {
		return nil, fmt.Errorf("文件大小超过限制 %dMB", s.cfg.Media.MaxSizeMB)
	}

	// 读取文件内容并计算哈希。
	content, err := io.ReadAll(io.LimitReader(reader, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败: %w", err)
	}
	fileHash := sha256.Sum256(content)
	hashHex := hex.EncodeToString(fileHash[:])

	// 文件哈希一致时直接复用已有记录，实现秒传去重。
	if existing, err := s.mediaRepo.GetByFileHash(hashHex); err == nil {
		return existing, nil
	}

	// 生成存储文件名与相对路径，按月分目录便于归档。
	storedName := uuid.NewString() + filepath.Ext(filename)
	subDir := timeNowStr()
	relativePath := filepath.Join(subDir, storedName)
	uploadDir := filepath.Join(s.cfg.Media.UploadDir, subDir)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 写入本地存储。
	fullPath := filepath.Join(s.cfg.Media.UploadDir, relativePath)
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 构建访问 URL。
	fileURL := strings.TrimRight(s.cfg.Media.BaseURL, "/") + "/" + filepath.ToSlash(relativePath)

	media := &model.MediaFile{
		Filename:    filename,
		StoredName:  storedName,
		FilePath:    fullPath,
		FileURL:     fileURL,
		MimeType:    detectContentType(content),
		FileSize:    uint64(size),
		FileHash:    hashHex,
		Status:      model.MediaStatusActive,
		UploaderID:  uploaderID,
		UploadIP:    ipAddress,
		StorageType: model.StorageTypeLocal,
		IsPublic:    true,
	}

	if err := s.mediaRepo.Create(media); err != nil {
		return nil, fmt.Errorf("保存媒体记录失败: %w", err)
	}

	return s.mediaRepo.GetByID(media.ID)
}

// GetMedia 根据ID获取媒体文件。
func (s *MediaService) GetMedia(id uint) (*model.MediaFile, error) {
	return s.mediaRepo.GetByID(id)
}

// ListMedia 分页查询媒体文件，非管理员仅能查看自己的文件。
func (s *MediaService) ListMedia(req *ListMediaRequest, uploaderID *uint, isAdmin bool) (*MediaListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.MediaListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Folder:   req.Folder,
		MimeType: req.MimeType,
	}

	// 非管理员仅返回自己上传的文件。
	if !isAdmin {
		params.UploaderID = uploaderID
	}

	mediaFiles, total, err := s.mediaRepo.List(params)
	if err != nil {
		return nil, err
	}

	return &MediaListResponse{
		Media:    mediaFiles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteMedia 删除媒体文件，上传者本人或管理员可操作。
func (s *MediaService) DeleteMedia(id uint, operatorID uint, isAdmin bool) error {
	media, err := s.mediaRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 权限校验：管理员可删除任意文件，否则仅上传者本人可删除。
	if !isAdmin && media.UploaderID != operatorID {
		return errors.New("没有删除此文件的权限")
	}

	// 删除物理文件，失败不影响数据库软删结果。
	_ = os.Remove(media.FilePath)

	return s.mediaRepo.Delete(id)
}

// timeNowStr 生成当前日期字符串，用于按天归档上传目录。
func timeNowStr() string {
	return time.Now().Format("2006/01")
}

// detectContentType 通过嗅探文件内容推断 MIME 类型。
func detectContentType(content []byte) string {
	return http.DetectContentType(content)
}
