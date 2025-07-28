package model

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MediaFile 媒体文件模型
type MediaFile struct {
	ID           uint           `json:"id" gorm:"primaryKey;comment:文件ID"`
	Filename     string         `json:"filename" gorm:"not null;size:255;comment:原始文件名"`
	StoredName   string         `json:"storedName" gorm:"uniqueIndex;not null;size:255;comment:存储文件名（UUID）"`
	FilePath     string         `json:"filePath" gorm:"not null;size:500;comment:文件存储路径"`
	FileURL      string         `json:"fileUrl" gorm:"not null;size:500;comment:文件访问URL"`
	ThumbnailURL string         `json:"thumbnailUrl" gorm:"size:500;comment:缩略图URL"`
	MimeType     string         `json:"mimeType" gorm:"not null;size:100;index;comment:MIME类型"`
	FileSize     uint64         `json:"fileSize" gorm:"not null;comment:文件大小（字节）"`
	FileHash     string         `json:"fileHash" gorm:"size:64;index;comment:文件SHA256哈希值"`
	Width        *uint          `json:"width" gorm:"comment:图片宽度"`
	Height       *uint          `json:"height" gorm:"comment:图片高度"`
	AltText      string         `json:"altText" gorm:"size:255;comment:替代文本（SEO用）"`
	UploaderID   uint           `json:"uploaderId" gorm:"not null;index;comment:上传者ID"`
	UploadIP     string         `json:"uploadIP" gorm:"size:45;comment:上传IP地址"`
	StorageType  StorageType    `json:"storageType" gorm:"default:local;index;comment:存储类型"`
	Folder       string         `json:"folder" gorm:"size:100;index;comment:文件夹分类"`
	UsageCount   uint           `json:"usageCount" gorm:"default:0;comment:使用次数"`
	IsPublic     bool           `json:"isPublic" gorm:"default:true;comment:是否公开访问"`
	CreatedAt    time.Time      `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Uploader User `json:"uploader" gorm:"foreignKey:UploaderID"`
}

// TableName 指定表名
func (MediaFile) TableName() string {
	return "media_files"
}

// 定义存储类型枚举
type StorageType string

const (
	StorageTypeLocal StorageType = "local" // 本地存储
	StorageTypeOSS   StorageType = "oss"   // 阿里云OSS
	StorageTypeS3    StorageType = "s3"    // AWS S3
	StorageTypeCOS   StorageType = "cos"   // 腾讯云COS
)

// 定义常用的MIME类型
const (
	MimeTypeJPEG = "image/jpeg"
	MimeTypePNG  = "image/png"
	MimeTypeGIF  = "image/gif"
	MimeTypeWEBP = "image/webp"
	MimeTypePDF  = "application/pdf"
	MimeTypeDOC  = "application/msword"
	MimeTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MimeTypeMP4  = "video/mp4"
	MimeTypeMP3  = "audio/mpeg"
)

// IsImage 检查是否为图片文件
func (m *MediaFile) IsImage() bool {
	return strings.HasPrefix(m.MimeType, "image/")
}

// IsVideo 检查是否为视频文件
func (m *MediaFile) IsVideo() bool {
	return strings.HasPrefix(m.MimeType, "video/")
}

// IsAudio 检查是否为音频文件
func (m *MediaFile) IsAudio() bool {
	return strings.HasPrefix(m.MimeType, "audio/")
}

// IsDocument 检查是否为文档文件
func (m *MediaFile) IsDocument() bool {
	documentTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
	}

	for _, docType := range documentTypes {
		if m.MimeType == docType {
			return true
		}
	}
	return false
}

// GetFileExtension 获取文件扩展名
func (m *MediaFile) GetFileExtension() string {
	return strings.ToLower(filepath.Ext(m.Filename))
}

// GetFileTypeCategory 获取文件类型分类
func (m *MediaFile) GetFileTypeCategory() string {
	if m.IsImage() {
		return "image"
	}
	if m.IsVideo() {
		return "video"
	}
	if m.IsAudio() {
		return "audio"
	}
	if m.IsDocument() {
		return "document"
	}
	return "other"
}

// GetFormattedFileSize 获取格式化的文件大小
func (m *MediaFile) GetFormattedFileSize() string {
	size := float64(m.FileSize)
	units := []string{"B", "KB", "MB", "GB", "TB"}

	for i, unit := range units {
		if size < 1024.0 || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f %s", size, unit)
			}
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024.0
	}

	return fmt.Sprintf("%.2f %s", size, units[len(units)-1])
}

// GetThumbnailURL 获取缩略图URL（如果没有则返回原图）
func (m *MediaFile) GetThumbnailURL() string {
	if m.ThumbnailURL != "" {
		return m.ThumbnailURL
	}
	if m.IsImage() {
		return m.FileURL
	}
	// 返回默认图标或空
	return ""
}

// HasDimensions 检查是否有尺寸信息
func (m *MediaFile) HasDimensions() bool {
	return m.Width != nil && m.Height != nil && *m.Width > 0 && *m.Height > 0
}

// GetAspectRatio 获取宽高比
func (m *MediaFile) GetAspectRatio() float64 {
	if !m.HasDimensions() {
		return 0
	}
	return float64(*m.Width) / float64(*m.Height)
}

// IncrementUsageCount 增加使用次数
func (m *MediaFile) IncrementUsageCount() {
	m.UsageCount++
}

// DecrementUsageCount 减少使用次数
func (m *MediaFile) DecrementUsageCount() {
	if m.UsageCount > 0 {
		m.UsageCount--
	}
}

// CanDelete 检查是否可以删除（仅上传者和管理员可删除）
func (m *MediaFile) CanDelete(currentUser *User) bool {
	if currentUser == nil {
		return false
	}

	// 管理员可以删除所有文件
	if currentUser.IsAdmin() {
		return true
	}

	// 上传者可以删除自己的文件
	return m.UploaderID == currentUser.ID
}

// CanEdit 检查是否可以编辑文件信息
func (m *MediaFile) CanEdit(currentUser *User) bool {
	if currentUser == nil {
		return false
	}

	// 管理员可以编辑所有文件
	if currentUser.IsAdmin() {
		return true
	}

	// 上传者可以编辑自己的文件
	return m.UploaderID == currentUser.ID
}

// SetDimensions 设置图片尺寸
func (m *MediaFile) SetDimensions(width, height uint) {
	m.Width = &width
	m.Height = &height
}

// SetThumbnail 设置缩略图URL
func (m *MediaFile) SetThumbnail(thumbnailURL string) {
	m.ThumbnailURL = thumbnailURL
}

// MarkAsPrivate 标记为私有文件
func (m *MediaFile) MarkAsPrivate() {
	m.IsPublic = false
}

// MarkAsPublic 标记为公开文件
func (m *MediaFile) MarkAsPublic() {
	m.IsPublic = true
}

// GetWebSafeURL 获取Web安全的URL（处理特殊字符）
func (m *MediaFile) GetWebSafeURL() string {
	// 可以在这里处理URL编码等安全措施
	return m.FileURL
}

// GenerateAltText 生成默认的替代文本
func (m *MediaFile) GenerateAltText() string {
	if m.AltText != "" {
		return m.AltText
	}

	// 基于文件名生成默认的alt文本
	filename := strings.TrimSuffix(m.Filename, filepath.Ext(m.Filename))
	// 移除特殊字符，用空格替换
	filename = strings.ReplaceAll(filename, "_", " ")
	filename = strings.ReplaceAll(filename, "-", " ")

	return strings.Title(filename)
}
