// Go 文件监听器 - 仅监听 server 目录
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
	lastRun      time.Time
	building     bool
	buildMutex   sync.Mutex
	process      *exec.Cmd
	processMutex sync.Mutex
)

// 需要监听的目录（相对于 server 目录）
var watchDirs = []string{
	"cmd",
	"internal",
	"pkg",
	"configs",
}

// 需要跳过的目录
var skipDirs = []string{
	"tmp",
	"bin",
	"vendor",
	".git",
	"logs",
	"node_modules",
	".svelte-kit",
	"dist",
	"build",
}

func main() {
	fmt.Println("🚀 启动 Go 热更新监听器")
	fmt.Println("💡 按 Ctrl+C 停止监听")
	fmt.Println()

	// 初始编译
	if err := buildAndRun(); err != nil {
		log.Fatal("初始编译失败:", err)
	}

	// 设置初始检查时间
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

	// 停止旧进程
	stopProcess()

	// 创建输出目录
	if err := os.MkdirAll("tmp", 0755); err != nil {
		fmt.Printf("❌ 创建tmp目录失败: %v\n", err)
		return err
	}

	// 编译
	var outputPath string
	if runtime.GOOS == "windows" {
		outputPath = "tmp/myblog.exe"
	} else {
		outputPath = "tmp/myblog"
	}

	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/myblog")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ 编译失败: %v\n", err)
		return err
	}

	fmt.Println("✅ 编译成功")

	// 启动新进程
	startProcess(outputPath)

	return nil
}

// startProcess 启动应用进程
func startProcess(outputPath string) {
	processMutex.Lock()
	defer processMutex.Unlock()

	fmt.Println("🚀 启动应用...")

	process = exec.Command(outputPath)
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	// 使用 Start() 代替 Run()，避免阻塞
	if err := process.Start(); err != nil {
		fmt.Printf("❌ 启动应用失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 应用启动成功 (PID: %d)\n", process.Process.Pid)

	// 在后台等待进程结束
	go func() {
		if err := process.Wait(); err != nil {
			fmt.Printf("⚠️ 应用退出: %v\n", err)
		}
	}()
}

// stopProcess 停止应用进程
func stopProcess() {
	processMutex.Lock()
	defer processMutex.Unlock()

	if process != nil && process.Process != nil {
		fmt.Println("🛑 停止旧进程...")

		// 优雅停止
		if err := process.Process.Kill(); err != nil {
			fmt.Printf("⚠️ 停止进程失败: %v\n", err)
		}

		// 等待进程结束，进程已被终止时错误无意义，显式忽略返回值
		_ = process.Wait()
		process = nil
	}
}

// watchFiles 监听文件变化
func watchFiles() {
	serverDir, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前目录:", err)
	}

	fmt.Printf("🔍 监听目录: %s\n", serverDir)
	fmt.Printf("📁 监听子目录: %s\n", strings.Join(watchDirs, ", "))
	fmt.Println()

	for {
		time.Sleep(1 * time.Second)

		changed := false
		latestModTime := time.Time{}

		// 只遍历指定的目录
		for _, dir := range watchDirs {
			dirPath := filepath.Join(serverDir, dir)
			if !dirExists(dirPath) {
				continue
			}

			err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}

				// 跳过目录
				if info.IsDir() {
					if shouldSkipDir(info.Name()) {
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
					fmt.Printf("📄 文件变化: %s (修改时间: %v)\n", path, info.ModTime())
					changed = true
					if info.ModTime().After(latestModTime) {
						latestModTime = info.ModTime()
					}
				}

				return nil
			})

			if err != nil {
				fmt.Printf("⚠️ 监听目录 %s 出错: %v\n", dir, err)
			}
		}

		if changed {
			lastRun = latestModTime
			fmt.Println("🔄 检测到文件变化，重新编译...")

			// 重新编译和运行
			if err := buildAndRun(); err != nil {
				fmt.Printf("❌ 重新编译失败: %v\n", err)
			} else {
				fmt.Println("🎉 热更新完成")
			}
			fmt.Println()
		}
	}
}

// dirExists 检查目录是否存在
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// shouldSkipDir 判断是否应该跳过目录
func shouldSkipDir(dirName string) bool {
	for _, skipDir := range skipDirs {
		if dirName == skipDir {
			return true
		}
	}
	return false
}
