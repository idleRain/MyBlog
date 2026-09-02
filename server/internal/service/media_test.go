package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeMediaRepo 媒体仓储的测试替身。
type fakeMediaRepo struct {
	repository.MediaRepositoryInterface
	media      []*model.MediaFile
	getByHash  func(hash string) (*model.MediaFile, error)
	getByID    func(id uint) (*model.MediaFile, error)
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

	// 非上传者、非管理员删除应被拒绝。
	err := svc.DeleteMedia(1, 2, false)
	if err == nil {
		t.Fatal("非上传者删除应返回错误")
	}
	if !strings.Contains(err.Error(), "没有删除") {
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
