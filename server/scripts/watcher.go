// 简单的Go文件监听器 - 适用于Windows环境
// 使用方法: go run scripts/watcher.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	lastRun    time.Time
	building   bool
	buildMutex sync.Mutex
)

func main() {
	fmt.Println("🚀 启动简单Go热更新监听器")
	fmt.Println("💡 监听 .go 文件变化，自动重新编译和运行")
	fmt.Println("💡 按 Ctrl+C 停止监听")
	fmt.Println()

	// 初始编译
	if err := buildAndRun(); err != nil {
		log.Fatal("初始编译失败:", err)
	}

	// 设置初始检查时间，避免冷启动重复触发
	lastRun = time.Now()

	// 监听文件变化
	watchFiles()
}

// buildAndRun 编译并运行程序
func buildAndRun() error {
	buildMutex.Lock()
	defer buildMutex.Unlock()

	if building {
		return nil
	}
	building = true
	defer func() { building = false }()

	fmt.Println("🔨 编译中...")

	// 创建tmp目录
	if err := os.MkdirAll("tmp", 0755); err != nil {
		fmt.Printf("❌ 创建tmp目录失败: %v\n", err)
		return err
	}

	// 编译 (跨平台)
	var outputPath string
	if runtime.GOOS == "windows" {
		outputPath = "tmp/myblog.exe"
	} else {
		outputPath = "tmp/myblog"
	}

	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/myblog")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ 编译失败: %v\n", err)
		return err
	}

	fmt.Println("✅ 编译成功")

	// 运行
	go func() {
		fmt.Println("🚀 启动应用...")
		cmd := exec.Command(outputPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️ 应用退出: %v\n", err)
		}
	}()

	return nil
}

// watchFiles 监听文件变化
func watchFiles() {
	// 获取当前脚本所在目录
	scriptDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("无法获取脚本目录:", err)
	}
	// server目录是脚本目录的上一级
	serverDir := filepath.Dir(scriptDir)

	for {
		time.Sleep(1 * time.Second)

		changed := false
		err = filepath.Walk(serverDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// 跳过不需要监听的目录
			if info.IsDir() {
				name := info.Name()
				if name == "tmp" || name == "vendor" || name == ".git" || name == "logs" {
					return filepath.SkipDir
				}
				return nil
			}

			// 只监听 .go 文件
			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			// 检查文件修改时间
			if info.ModTime().After(lastRun) {
				changed = true
				return filepath.SkipDir
			}

			return nil
		})
		if err != nil {
			fmt.Printf("⚠️ 文件监听出错: %v\n", err)
			continue
		}

		if changed {
			lastRun = time.Now()
			fmt.Println("📝 检测到文件变化，重新编译...")

			// 杀死旧进程 (跨平台)
			if runtime.GOOS == "windows" {
				if err := exec.Command("taskkill", "/F", "/IM", "myblog.exe").Run(); err != nil {
					fmt.Printf("⚠️ 终止Windows进程失败: %v\n", err)
				}
			} else {
				if err := exec.Command("pkill", "-f", "myblog").Run(); err != nil {
					fmt.Printf("⚠️ 终止进程失败: %v\n", err)
				}
			}

			// 重新编译和运行
			time.Sleep(500 * time.Millisecond) // 等待进程完全退出
			if err := buildAndRun(); err != nil {
				fmt.Printf("❌ 重新编译失败: %v\n", err)
			}
		}
	}
}
