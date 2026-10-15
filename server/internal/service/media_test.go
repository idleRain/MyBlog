package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeMediaRepo 媒体仓储的测试替身。
type fakeMediaRepo struct {
	repository.MediaRepositoryInterface
	media     []*model.MediaFile
	getByHash func(hash string, uploaderID uint) (*model.MediaFile, error)
	getByID   func(id uint) (*model.MediaFile, error)
}

func (f *fakeMediaRepo) Create(media *model.MediaFile) error {
	media.ID = uint(len(f.media) + 1)
	f.media = append(f.media, media)
	return nil
}

// GetByFileHashAndUploader 默认按上传者过滤，与生产仓储的限定语义一致。
func (f *fakeMediaRepo) GetByFileHashAndUploader(hash string, uploaderID uint) (*model.MediaFile, error) {
	if f.getByHash != nil {
		return f.getByHash(hash, uploaderID)
	}
	for _, media := range f.media {
		if media.FileHash == hash && media.UploaderID == uploaderID {
			return media, nil
		}
	}
	return nil, domain.ErrMediaNotFound
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
	return nil, domain.ErrMediaNotFound
}

func (f *fakeMediaRepo) Delete(id uint) error {
	for i, media := range f.media {
		if media.ID == id {
			f.media = append(f.media[:i], f.media[i+1:]...)
			return nil
		}
	}
	return domain.ErrMediaNotFound
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
	// 超限属业务校验失败，handler 据此分档为 400 而非内部错误。
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("超限上传应包装 ErrInvalidRequest，实际为 %v", err)
	}
}

// testBytesPerMegabyte 单兆字节的字节数，供测试构造大小上限的边界内容。
const testBytesPerMegabyte = 1024 * 1024

// TestUploadFileRejectsNonPositiveMaxSizeConfig 验证大小上限非正时上传被拒绝。
// 服务实例可被直接构造，配置校验不覆盖该路径，上限合法性的判定必须在服务层独立成立。
func TestUploadFileRejectsNonPositiveMaxSizeConfig(t *testing.T) {
	repo := &fakeMediaRepo{}
	svc := newTestMediaService(t, repo)
	svc.cfg.Media.MaxSizeMB = 0

	_, err := svc.UploadFile("pic.png", bytes.NewReader(pngTestContent), int64(len(pngTestContent)), 1, "127.0.0.1")
	if err == nil {
		t.Fatal("大小上限为零时上传应返回错误")
	}
	// 上限非法属服务端配置错误，handler 据此分档为 500 而非客户端 400。
	if errors.Is(err, ErrInvalidRequest) {
		t.Errorf("大小上限非法不应分类为请求参数错误，实际为 %v", err)
	}
	if len(repo.media) != 0 {
		t.Error("被拒绝的文件不应落库")
	}
}

// TestUploadFileRecordsActualStoredSize 验证落库大小取自实际读取字节数而非入参声明值。
// 声明值远大于实际内容时，记录大小必须与存储内容一致，否则读回的大小与实际不符。
func TestUploadFileRecordsActualStoredSize(t *testing.T) {
	const maxSizeMB = 1
	declaredSizeBytes := int64(testBytesPerMegabyte) / 2

	repo := &fakeMediaRepo{}
	svc := newTestMediaService(t, repo)
	svc.cfg.Media.MaxSizeMB = maxSizeMB

	media, err := svc.UploadFile("pic.png", bytes.NewReader(pngTestContent), declaredSizeBytes, 1, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	if media.FileSize != uint64(len(pngTestContent)) {
		t.Errorf("FileSize = %d, 期望等于实际字节数 %d", media.FileSize, len(pngTestContent))
	}
}

// TestUploadFileRejectsContentExceedingMaxSize 验证声明大小在上限内但实际内容超限时被拒绝。
// 声明值由调用方提供，不能作为放行依据，超限判定以实际读取字节数为准。
func TestUploadFileRejectsContentExceedingMaxSize(t *testing.T) {
	const maxSizeMB = 1
	declaredSizeBytes := int64(testBytesPerMegabyte)
	content := bytes.Repeat([]byte("a"), testBytesPerMegabyte+1)

	repo := &fakeMediaRepo{}
	svc := newTestMediaService(t, repo)
	svc.cfg.Media.MaxSizeMB = maxSizeMB

	_, err := svc.UploadFile("note.txt", bytes.NewReader(content), declaredSizeBytes, 1, "127.0.0.1")
	if err == nil {
		t.Fatal("实际内容超过上限的上传应返回错误")
	}
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("超限上传应包装 ErrInvalidRequest，实际为 %v", err)
	}
	if len(repo.media) != 0 {
		t.Error("被拒绝的文件不应落库")
	}
}

// TestUploadFileDeduplicates 验证同一上传者的相同内容文件通过哈希去重复用已有记录。
func TestUploadFileDeduplicates(t *testing.T) {
	// 计算与上传内容一致的哈希，模拟同一上传者已上传过该内容的场景。
	content := []byte("same content")
	hashBytes := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hashBytes[:])

	const uploaderID uint = 1
	existing := &model.MediaFile{ID: 1, FileHash: contentHash, FileURL: "/uploads/2026/01/a.png", Filename: "a.png", UploaderID: uploaderID}
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{existing},
		getByHash: func(hash string, ownerID uint) (*model.MediaFile, error) {
			if hash == contentHash && ownerID == uploaderID {
				return existing, nil
			}
			return nil, domain.ErrMediaNotFound
		},
	}
	svc := newTestMediaService(t, repo)

	// 上传相同内容，应命中已有记录。
	media, err := svc.UploadFile("b.png", bytes.NewReader(content), int64(len(content)), uploaderID, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}

	if media.ID != 1 {
		t.Errorf("去重后应返回已有记录，实际 ID = %d", media.ID)
	}
	if got := len(repo.media); got != 1 {
		t.Errorf("去重命中不应新建记录，实际记录数 = %d", got)
	}
}

// TestUploadFileDoesNotReuseOtherUploaderRecord 验证仓储忽略上传者过滤时服务层仍拒绝跨用户复用。
// 仓储实现不受业务层信任，归属判断由业务层独立承担。
// 该场景下调用方必须获得属于自己的记录，不得回出他人的存储路径、归属与原始文件名。
func TestUploadFileDoesNotReuseOtherUploaderRecord(t *testing.T) {
	content := []byte("shared bytes across uploaders")
	hashBytes := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hashBytes[:])

	const otherUploaderID uint = 1
	const currentUploaderID uint = 2
	otherRecord := &model.MediaFile{
		ID:         1,
		Filename:   "other-owner.png",
		FilePath:   `C:\srv\uploads\2026\01\other-owner.png`,
		FileHash:   contentHash,
		UploaderID: otherUploaderID,
	}
	// 钩子无视上传者参数并强行返回他人记录，模拟忽略过滤条件的仓储实现。
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{otherRecord},
		getByHash: func(hash string, _ uint) (*model.MediaFile, error) {
			if hash == contentHash {
				return otherRecord, nil
			}
			return nil, domain.ErrMediaNotFound
		},
	}
	svc := newTestMediaService(t, repo)

	media, err := svc.UploadFile("mine.png", bytes.NewReader(content), int64(len(content)), currentUploaderID, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}

	if media.UploaderID != currentUploaderID {
		t.Errorf("上传结果应归属于当前上传者，实际 uploaderId = %d", media.UploaderID)
	}
	if media.FilePath == otherRecord.FilePath {
		t.Errorf("上传结果不应回出他人存储路径，实际 filePath = %s", media.FilePath)
	}
	if media.Filename != "mine.png" {
		t.Errorf("上传结果应保留本次提交的原始文件名，实际 filename = %s", media.Filename)
	}
	if got := len(repo.media); got != 2 {
		t.Errorf("跨上传者命中应新增属于本人的记录，实际记录数 = %d", got)
	}
}

// TestUploadFileDeduplicatesOwnRecordAmongOthers 验证同哈希存在多位上传者记录时，去重落在调用方自己的记录上。
// 限定上传者之前，查询只会返回主键最小的一行，非首位上传者每次重复上传都会新建副本。
func TestUploadFileDeduplicatesOwnRecordAmongOthers(t *testing.T) {
	content := []byte("content held by two uploaders")
	hashBytes := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hashBytes[:])

	const otherUploaderID uint = 1
	const currentUploaderID uint = 2
	otherRecord := &model.MediaFile{ID: 1, Filename: "other.png", FileHash: contentHash, UploaderID: otherUploaderID}
	ownRecord := &model.MediaFile{ID: 2, Filename: "mine.png", FileURL: "/uploads/2026/01/mine.png", FileHash: contentHash, UploaderID: currentUploaderID}
	// 钩子按上传者分流，模拟按上传者限定查询的生产仓储。
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{otherRecord, ownRecord},
		getByHash: func(hash string, ownerID uint) (*model.MediaFile, error) {
			if hash != contentHash {
				return nil, domain.ErrMediaNotFound
			}
			if ownerID == otherUploaderID {
				return otherRecord, nil
			}
			if ownerID == currentUploaderID {
				return ownRecord, nil
			}
			return nil, domain.ErrMediaNotFound
		},
	}
	svc := newTestMediaService(t, repo)

	media, err := svc.UploadFile("mine.png", bytes.NewReader(content), int64(len(content)), currentUploaderID, "127.0.0.1")
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}

	if media.ID != ownRecord.ID {
		t.Errorf("去重应命中调用方自己的记录，实际 ID = %d", media.ID)
	}
	if got := len(repo.media); got != 2 {
		t.Errorf("去重命中不应新建记录，实际记录数 = %d", got)
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

// TestGetMediaOwnership 验证媒体详情的水平越权防护，与列表的可见范围保持一致。
func TestGetMediaOwnership(t *testing.T) {
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{
			{ID: 1, UploaderID: 1, FilePath: "/tmp/a.png"},
		},
	}
	svc := newTestMediaService(t, repo)

	// 非上传者、非管理员访问他人文件时按不存在同响应，防止经 ID 枚举探测。
	if _, err := svc.GetMedia(1, 2, false); !errors.Is(err, domain.ErrMediaNotFound) {
		t.Errorf("非上传者访问他人文件应按不存在处理，实际: %v", err)
	}

	// 上传者本人可读取自己的文件，作为归属性校验的防误伤断言。
	if _, err := svc.GetMedia(1, 1, false); err != nil {
		t.Errorf("上传者本人读取失败: %v", err)
	}

	// 管理员可读取任意文件。
	if _, err := svc.GetMedia(1, 2, true); err != nil {
		t.Errorf("管理员读取失败: %v", err)
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
	// 类型不在白名单属业务校验失败，handler 据此分档为 400 而非内部错误。
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("非白名单类型应包装 ErrInvalidRequest，实际为 %v", err)
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

// TestDeleteMediaLogsOrphanFileFailure 验证物理删除失败时事件写入日志且软删结果不受影响。
func TestDeleteMediaLogsOrphanFileFailure(t *testing.T) {
	// 物理路径指向一个非空目录，os.Remove 对非空目录返回错误且非不存在，
	// 等价于删除受阻的孤儿场景。
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("构造非空目录失败: %v", err)
	}
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{
			{ID: 1, UploaderID: 1, FilePath: dir},
		},
	}
	svc := newTestMediaService(t, repo)

	// 捕获标准 logger 输出，验证失败事件确实入日志。
	logOutput := &bytes.Buffer{}
	originalOutput := log.Writer()
	log.SetOutput(logOutput)
	defer log.SetOutput(originalOutput)

	if err := svc.DeleteMedia(1, 1, false); err != nil {
		t.Fatalf("物理删除失败不应阻塞软删结果: %v", err)
	}

	logged := logOutput.String()
	if !strings.Contains(logged, "媒体物理文件删除失败") || !strings.Contains(logged, dir) {
		t.Errorf("删除失败事件应携带路径入日志，实际为 %q", logged)
	}

	// 软删仍应执行：数据库侧记录被移除。
	if got := len(repo.media); got != 0 {
		t.Errorf("软删结果不应受物理删除失败影响，剩余记录 %d", got)
	}
}

// TestDeleteMediaIgnoresMissingFile 验证物理文件已不存在时按完成处理，不产生孤儿告警。
func TestDeleteMediaIgnoresMissingFile(t *testing.T) {
	repo := &fakeMediaRepo{
		media: []*model.MediaFile{
			{ID: 1, UploaderID: 1, FilePath: t.TempDir() + "/missing.png"},
		},
	}
	svc := newTestMediaService(t, repo)

	logOutput := &bytes.Buffer{}
	originalOutput := log.Writer()
	log.SetOutput(logOutput)
	defer log.SetOutput(originalOutput)

	if err := svc.DeleteMedia(1, 1, false); err != nil {
		t.Fatalf("文件已不存在时的删除不应失败: %v", err)
	}
	if strings.Contains(logOutput.String(), "媒体物理文件删除失败") {
		t.Error("文件已不存在不构成孤儿事件，不应告警")
	}
}
