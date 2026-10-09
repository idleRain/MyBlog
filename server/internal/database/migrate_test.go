package database

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"MyBlog/internal/model"
)

// 迁移文件的结构不变量以文件名与 SQL 文本为唯一依据，因此本文件的所有用例都是纯静态检查，
// 不连接数据库。可静态验证的部分包括版本序列的完整性、每个版本 up/down 成对，
// 以及 up 中创建的每张表在所属 down 中被回收。
// 需要真实数据库才能验证的部分不在本文件覆盖范围内，任何结论都不得超出静态证据。
const (
	migrationsDirectoryName = "migrations"
	upFileSuffix            = ".up.sql"
	downFileSuffix          = ".down.sql"
)

// 版本号前缀固定为六位十进制数字，后接下划线分隔的描述。
var migrationFileNamePattern = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.(up|down)\.sql$`)

var (
	createTablePattern = regexp.MustCompile("(?i)CREATE\\s+TABLE\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?`?([a-zA-Z0-9_]+)`?")
	dropTablePattern   = regexp.MustCompile("(?i)DROP\\s+TABLE\\s+(?:IF\\s+EXISTS\\s+)?`?([a-zA-Z0-9_]+)`?")
)

// 不实现 TableName 方法的实体，GORM 按命名策略复数化推导表名。
// 该映射是双轨校验的已知例外，新增此类实体时必须在此登记，否则校验会失败。
var pluralizedTableModels = map[string]string{
	"users": "domain.User",
}

// migrationPair 描述一个版本号下的 up 与 down 文件。
type migrationPair struct {
	version  uint
	name     string
	upPath   string
	downPath string
}

// migrationsDir 返回仓库中迁移目录的绝对路径。
// 以本文件的编译期位置为基准向上回退到 server 根目录，使用例不受执行目录影响。
func migrationsDir(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法获取当前测试文件路径")
	}

	// 本文件位于 server/internal/database，迁移目录位于 server/migrations。
	serverRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	directory := filepath.Join(serverRoot, migrationsDirectoryName)
	if info, err := os.Stat(directory); err != nil || !info.IsDir() {
		t.Fatalf("迁移目录不存在或不是目录: %s", directory)
	}

	return directory
}

// collectMigrationPairs 扫描迁移目录并按版本号归集成对结果。
func collectMigrationPairs(t *testing.T) []migrationPair {
	t.Helper()

	directory := migrationsDir(t)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("读取迁移目录失败: %v", err)
	}

	byVersion := make(map[uint]*migrationPair)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := migrationFileNamePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			t.Errorf("迁移文件名不符合 六位版本号_描述.{up,down}.sql 规范: %s", entry.Name())
			continue
		}

		version, err := strconv.ParseUint(matches[1], 10, 32)
		if err != nil {
			t.Errorf("迁移文件名版本号无法解析: %s", entry.Name())
			continue
		}

		pair, exists := byVersion[uint(version)]
		if !exists {
			pair = &migrationPair{version: uint(version), name: matches[2]}
			byVersion[uint(version)] = pair
		}

		fullPath := filepath.Join(directory, entry.Name())
		switch matches[3] {
		case "up":
			if pair.upPath != "" {
				t.Errorf("版本 %d 存在重复的 up 文件", version)
			}
			pair.upPath = fullPath
		case "down":
			if pair.downPath != "" {
				t.Errorf("版本 %d 存在重复的 down 文件", version)
			}
			pair.downPath = fullPath
		}
	}

	pairs := make([]migrationPair, 0, len(byVersion))
	for _, pair := range byVersion {
		pairs = append(pairs, *pair)
	}
	sort.Slice(pairs, func(left, right int) bool { return pairs[left].version < pairs[right].version })

	return pairs
}

// readSQL 读取迁移文件内容。
func readSQL(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取迁移文件失败 %s: %v", path, err)
	}

	return string(content)
}

// extractTableNames 按给定正则抽取 SQL 中的表名集合。
func extractTableNames(pattern *regexp.Regexp, content string) map[string]struct{} {
	names := make(map[string]struct{})
	for _, match := range pattern.FindAllStringSubmatch(content, -1) {
		names[strings.ToLower(match[1])] = struct{}{}
	}

	return names
}

// 每个版本必须同时具备 up 与 down，且版本号自 1 起连续无缺号。
// 缺 down 会让回滚在生产上无路可走，缺号则意味着历史被改写。
func TestMigrationVersionsArePairedAndContinuous(t *testing.T) {
	pairs := collectMigrationPairs(t)

	if len(pairs) == 0 {
		t.Fatal("未发现任何迁移版本")
	}

	for index, pair := range pairs {
		expectedVersion := uint(index + 1)
		if pair.version != expectedVersion {
			t.Errorf("版本序列不连续: 第 %d 个版本为 %d, 期望 %d", index+1, pair.version, expectedVersion)
		}
		if pair.upPath == "" {
			t.Errorf("版本 %d 缺少 up 文件", pair.version)
		}
		if pair.downPath == "" {
			t.Errorf("版本 %d 缺少 down 文件", pair.version)
		}
	}
}

// up 中创建的每张表都必须在其所属版本的 down 中被回收，否则回滚会留下残余表。
func TestEachMigrationDropsEveryTableItCreates(t *testing.T) {
	for _, pair := range collectMigrationPairs(t) {
		if pair.upPath == "" || pair.downPath == "" {
			continue
		}

		t.Run(pair.name, func(t *testing.T) {
			created := extractTableNames(createTablePattern, readSQL(t, pair.upPath))
			dropped := extractTableNames(dropTablePattern, readSQL(t, pair.downPath))

			for table := range created {
				if _, ok := dropped[table]; !ok {
					t.Errorf("up 创建了表 %s，但同版本 down 未回收该表", table)
				}
			}
		})
	}
}

// 静态证据只能证明 up 与 down 各自的文本内容，无法证明回滚后数据库回到初始状态。
// 此用例仅锚定「up 创建的表在 down 中有对应回收语句」这一必要条件。
func TestMigrationFilesAreNotEmpty(t *testing.T) {
	for _, pair := range collectMigrationPairs(t) {
		for label, path := range map[string]string{"up": pair.upPath, "down": pair.downPath} {
			if path == "" {
				continue
			}
			if len(strings.TrimSpace(readSQL(t, path))) == 0 {
				t.Errorf("版本 %d 的 %s 文件为空", pair.version, label)
			}
		}
	}
}

// 双轨一致性：golang-migrate 的建表清单与 GORM AutoMigrate 的模型清单必须互相覆盖。
// 迁移多出的表说明模型层遗漏了实体，模型多出的表说明生产建表会缺表；
// 两类偏差都会在「开发用 AutoMigrate、生产用 migrate」的切换中爆发。
func TestMigrationTablesAndModelsCoverEachOther(t *testing.T) {
	migrationTables := make(map[string]struct{})
	for _, pair := range collectMigrationPairs(t) {
		if pair.upPath == "" {
			continue
		}
		for table := range extractTableNames(createTablePattern, readSQL(t, pair.upPath)) {
			migrationTables[table] = struct{}{}
		}
	}

	modelTables := make(map[string]struct{})
	for _, entity := range model.Models() {
		named, ok := entity.(interface{ TableName() string })
		if !ok {
			continue
		}
		modelTables[named.TableName()] = struct{}{}
	}
	for table := range pluralizedTableModels {
		modelTables[table] = struct{}{}
	}

	for table := range migrationTables {
		if _, ok := modelTables[table]; !ok {
			t.Errorf("迁移创建了表 %s，但 model.Models() 中没有任何实体映射到该表", table)
		}
	}
	for table := range modelTables {
		if _, ok := migrationTables[table]; !ok {
			t.Errorf("模型映射了表 %s，但迁移文件中没有任何 CREATE TABLE 创建该表", table)
		}
	}
}
