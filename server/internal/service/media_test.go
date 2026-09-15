package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeMediaRepo 媒体仓储的测试替身。
type fakeMediaRepo struct {
	repository.MediaRepositoryInterface
	media     []*model.MediaFile
	getByHash func(hash string) (*model.MediaFile, error)
	getByID   func(id uint) (*model.MediaFile, error)
}

func (f *fakeMediaRepo) Create(media *model.MediaFile) error {
	media.ID = uint(len(f.media) + 1)
	f.media = append(f.media, media)
	return nil
}

func (f *fakeMediaRepo) GetByFileHash(hash string) (*model.MediaFile, error) {
	if f.getByHash != nil {
		return f.getByHash(hash)
	}
	for _, media := range f.media {
		if media.FileHash == hash {
			return media, nil
		}
	}
	return nil, repository.ErrMediaNotFound
}

func (f *fakeMediaRepo) GetByID(id uint) (*model.MediaFile, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	for _, media := range f.media {
		if media.ID == id {
			return media, nil
		}
	}
	return nil, repository.ErrMediaNotFound
}

func (f *fakeMediaRepo) Delete(id uint) error {
	for i, media := range f.media {
		if media.ID == id {
			f.media = append(f.media[:i], f.media[i+1:]...)
			return nil
		}
	}
	return repository.ErrMediaNotFound
}

// newTestMediaService 创建注入测试替身的媒体服务实例。
func newTestMediaService(t *testing.T, repo repository.MediaRepositoryInterface) *MediaService {
	cfg := &config.Config{}
	cfg.Media.UploadDir = t.TempDir() // 使用临时目录，避免污染工作区
	cfg.Media.BaseURL = "/uploads"
	cfg.Media.MaxSizeMB = 10
	return NewMediaService(repo, cfg).(*MediaService)
}

// TestUploadFileOversize 验证超过大小限制的文件被拒绝。
func TestUploadFileOversize(t *testing.T) {
	svc := newTestMediaService(t, &fakeMediaRepo{})
	// 构造 11MB 大小的内容。
	content := bytes.Repeat([]byte("a"), 11*1024*1024)

	_, err := svc.UploadFile("big.png", bytes.NewReader(content), int64(len(content)), 1, "127.0.0.1")
	if err == nil {
		t.Fatal("超过大小限制的上传应返回错误")
	}
	if !strings.Contains(err.Error(), "大小超过限制") {
		t.Errorf("错误信息不符合预期: %v", err)
	}
}

// TestUploadFileDeduplicates 验证相同内容文件通过哈希去重复用已有记录。
func TestUploadFileDeduplicates(t *testing.T) {
	// 计算与上传内容一致的哈希，模拟文件已存在的场景。
	content := []byte("same content")
	hashBytes := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hashBytes[:])

	existing := &model.MediaFile{ID: 1, FileHash: contentHash, FileURL: "/uploads/2026/01/a.png", Filename: "a.png"}
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{existing},
		getByHash: func(hash string) (*model.MediaFile, error) {
			if hash == contentHash {
				return existing, nil
			}
			return nil, repository.ErrMediaNotFound
		},
	}
	svc := newTestMediaService(t, repo)

	// 上传相同内容，应命中已有记录。
	media, err := svc.UploadFile("b.png", bytes.NewReader(content), int64(len(content)), 1, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}

	if media.ID != 1 {
		t.Errorf("去重后应返回已有记录，实际 ID = %d", media.ID)
	}
}

// TestDeleteMediaPermission 验证非上传者删除他人文件被拒绝。
func TestDeleteMediaPermission(t *testing.T) {
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{
			{ID: 1, UploaderID: 1, FilePath: "/tmp/a.png"},
		},
	}
	svc := newTestMediaService(t, repo)

	// 非上传者、非管理员删除应被拒绝，且错误可经哨兵识别映射为 403。
	err := svc.DeleteMedia(1, 2, false)
	if err == nil {
		t.Fatal("非上传者删除应返回错误")
	}
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("错误信息不符合预期: %v", err)
	}

	// 管理员可删除任意文件。
	if err := svc.DeleteMedia(1, 2, true); err != nil {
		t.Errorf("管理员删除失败: %v", err)
	}
}

// TestDeleteMediaOwnFile 验证上传者本人可删除自己的文件。
func TestDeleteMediaOwnFile(t *testing.T) {
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{
			{ID: 1, UploaderID: 1, FilePath: "/tmp/a.png"},
		},
	}
	svc := newTestMediaService(t, repo)

	if err := svc.DeleteMedia(1, 1, false); err != nil {
		t.Errorf("上传者本人删除失败: %v", err)
	}
}

// pngTestContent 构造可被内容嗅探识别为 image/png 的最小内容。
var pngTestContent = append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 8)...)

// TestUploadFileRejectsDisallowedType 验证白名单开启后非白名单类型被拒绝且不落库。
func TestUploadFileRejectsDisallowedType(t *testing.T) {
	repo := &fakeMediaRepo{}
	svc := newTestMediaService(t, repo)
	svc.cfg.Media.AllowedTypes = []string{"image/png"}

	content := []byte("plain text content")
	_, err := svc.UploadFile("note.txt", bytes.NewReader(content), int64(len(content)), 1, "127.0.0.1")
	if err == nil {
		t.Fatal("非白名单类型的上传应返回错误")
	}
	if !strings.Contains(err.Error(), "不支持的文件类型") {
		t.Errorf("错误信息不符合预期: %v", err)
	}
	if len(repo.media) != 0 {
		t.Error("被拒绝的文件不应落库")
	}

	if _, err := svc.UploadFile("pic.png", bytes.NewReader(pngTestContent), int64(len(pngTestContent)), 1, "127.0.0.1"); err != nil {
		t.Fatalf("白名单内类型上传失败: %v", err)
	}
}

// TestUploadFileDerivesStoredExtension 验证存储扩展名由内容嗅探结果推导，杜绝伪造扩展名。
func TestUploadFileDerivesStoredExtension(t *testing.T) {
	repo := &fakeMediaRepo{}
	svc := newTestMediaService(t, repo)
	svc.cfg.Media.AllowedTypes = []string{"image/png"}

	media, err := svc.UploadFile("polyglot.html", bytes.NewReader(pngTestContent), int64(len(pngTestContent)), 1, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	if !strings.HasSuffix(media.StoredName, ".png") {
		t.Errorf("存储扩展名应由嗅探类型推导为 .png，实际为 %s", media.StoredName)
	}
}

// TestUploadFileKeepsExtensionWhenUnrestricted 验证白名单关闭时保持原始扩展名。
func TestUploadFileKeepsExtensionWhenUnrestricted(t *testing.T) {
	repo := &fakeMediaRepo{}
	svc := newTestMediaService(t, repo)

	content := []byte("plain text content")
	media, err := svc.UploadFile("note.txt", bytes.NewReader(content), int64(len(content)), 1, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	if !strings.HasSuffix(media.StoredName, ".txt") {
		t.Errorf("白名单关闭时应保留原始扩展名，实际为 %s", media.StoredName)
	}
}
