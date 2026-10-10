package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckWritableCreatesMissingDir 验证目录缺失时按上传语义创建并判定可写。
func TestCheckWritableCreatesMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "uploads")

	if err := CheckWritable(dir); err != nil {
		t.Fatalf("目录缺失时应创建并判定可写，实际返回错误: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("探测后目录应存在，实际 stat 失败: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("探测后路径应为目录，实际为 %v", info.Mode())
	}
}

// TestCheckWritableLeavesNoResidue 验证探测成功后在目录中不留残留文件。
func TestCheckWritableLeavesNoResidue(t *testing.T) {
	dir := t.TempDir()

	if err := CheckWritable(dir); err != nil {
		t.Fatalf("可写目录应判定通过，实际返回错误: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), probeFilePrefix) {
			t.Errorf("探测文件未清理: %s", entry.Name())
		}
	}
	if len(entries) != 0 {
		t.Errorf("可写目录在探测后应保持为空，实际含 %d 个条目", len(entries))
	}
}

// TestCheckWritableRejectsDirUnderFile 验证父路径为普通文件时判定不可写。
func TestCheckWritableRejectsDirUnderFile(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("准备阻隔文件失败: %v", err)
	}

	err := CheckWritable(filepath.Join(blocker, "uploads"))
	if err == nil {
		t.Fatal("父路径为文件时应返回错误，实际返回 nil")
	}
	if !strings.Contains(err.Error(), "创建目录失败") {
		t.Errorf("错误信息应指明创建目录失败，实际为: %v", err)
	}
}

// TestCheckWritableRejectsPathOccupiedByFile 验证目标路径已被普通文件占用时判定不可写。
func TestCheckWritableRejectsPathOccupiedByFile(t *testing.T) {
	occupied := filepath.Join(t.TempDir(), "uploads")
	if err := os.WriteFile(occupied, []byte("occupied"), 0o644); err != nil {
		t.Fatalf("准备占位文件失败: %v", err)
	}

	if err := CheckWritable(occupied); err == nil {
		t.Fatal("目标路径为文件时应返回错误，实际返回 nil")
	}
}
