package handler

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fixtureAssertCallPattern 匹配测试源码中金样本断言的调用，捕获金样本文件名。
// 消费关系以调用实参为准而非人工登记表，登记表一旦与断言脱节即失去约束力。
var fixtureAssertCallPattern = regexp.MustCompile(`assertResponseMatchesFixture\([^)]*"([^"]+\.json)"\)`)

// fixtureTSAnchorPath TS 侧契约锚定文件路径，用于校验金样本在双端均有锚点。
const fixtureTSAnchorPath = "../../../packages/api/src/contracts/__tests__/fixtures.test.ts"

// fixtureTSAnchorPattern 匹配 TS 锚定文件中对金样本的引用路径，捕获金样本文件名。
var fixtureTSAnchorPattern = regexp.MustCompile(`contracts/fixtures/([A-Za-z0-9._-]+\.json)`)

// TestFixtureCoverage 校验契约金样本与双端锚点一一对应。
// 只被单端消费的金样本仍会放过另一端的漂移，故 Go 断言与 TS 类型锚定两侧同时校验：
// 新增金样本却漏写任一端锚点、或锚点引用了不存在的金样本时一律变红。
func TestFixtureCoverage(t *testing.T) {
	declared := scanFixtureFiles(t)
	consumed := scanAssertedFixtures(t)
	anchored := scanTSAnchoredFixtures(t)

	for name := range declared {
		if !consumed[name] {
			t.Errorf("金样本 %s 未被 Go 侧断言消费", name)
		}
		if !anchored[name] {
			t.Errorf("金样本 %s 未被 TS 侧类型锚定", name)
		}
	}
	for name := range consumed {
		if !declared[name] {
			t.Errorf("断言引用的金样本 %s 不存在于 %s", name, fixtureContractPath)
		}
	}
	for name := range anchored {
		if !declared[name] {
			t.Errorf("TS 锚定引用的金样本 %s 不存在于 %s", name, fixtureContractPath)
		}
	}
}

// scanFixtureFiles 返回金样本目录下的全部 JSON 文件名集合。
func scanFixtureFiles(t *testing.T) map[string]bool {
	t.Helper()

	entries, err := os.ReadDir(fixtureContractPath)
	if err != nil {
		t.Fatalf("读取契约金样本目录失败: %v", err)
	}

	files := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files[entry.Name()] = true
		}
	}
	return files
}

// scanAssertedFixtures 扫描本包测试源码，返回被断言消费的金样本文件名集合。
func scanAssertedFixtures(t *testing.T) map[string]bool {
	t.Helper()

	sources, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatalf("枚举测试源码失败: %v", err)
	}

	consumed := map[string]bool{}
	for _, source := range sources {
		content, err := os.ReadFile(source)
		if err != nil {
			t.Fatalf("读取测试源码 %s 失败: %v", source, err)
		}
		for _, match := range fixtureAssertCallPattern.FindAllStringSubmatch(string(content), -1) {
			consumed[match[1]] = true
		}
	}
	return consumed
}

// scanTSAnchoredFixtures 读取 TS 侧锚定文件，返回其中引用的金样本文件名集合。
// 跨语言读取源码是刻意的：该文件的职责是锚定金样本，缺锚点即契约面覆盖不全。
func scanTSAnchoredFixtures(t *testing.T) map[string]bool {
	t.Helper()

	content, err := os.ReadFile(fixtureTSAnchorPath)
	if err != nil {
		t.Fatalf("读取 TS 契约锚定文件失败: %v", err)
	}

	anchored := map[string]bool{}
	for _, match := range fixtureTSAnchorPattern.FindAllStringSubmatch(string(content), -1) {
		anchored[match[1]] = true
	}
	return anchored
}
