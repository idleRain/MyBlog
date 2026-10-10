// Package storage 提供本地文件存储的基础设施能力。
package storage

import (
	"fmt"
	"os"
)

// probeFilePrefix 可写性探测临时文件的前缀。
const probeFilePrefix = ".writable-probe-"

// dirPerm 上传目录的创建权限，与媒体上传路径一致。
const dirPerm = 0o755

// CheckWritable 校验目录可写，目录不存在时按上传语义创建。
// 判定依据是一次真实的文件创建与删除，不读取权限位。
func CheckWritable(dir string) error {
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	probeFile, err := os.CreateTemp(dir, probeFilePrefix)
	if err != nil {
		return fmt.Errorf("创建探测文件失败: %w", err)
	}

	probePath := probeFile.Name()
	if err := probeFile.Close(); err != nil {
		_ = os.Remove(probePath)
		return fmt.Errorf("关闭探测文件失败: %w", err)
	}

	if err := os.Remove(probePath); err != nil {
		return fmt.Errorf("清理探测文件失败: %w", err)
	}

	return nil
}
