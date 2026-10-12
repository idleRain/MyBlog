package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"testing"

	"gopkg.in/yaml.v3"
)

// errorsContractPath 错误码契约文件相对路径，基于本测试包所在目录计算。
const errorsContractPath = "../../../contracts/errors.yaml"

// responseSourcePath 传输层响应码常量所在源文件相对路径。
// 常量集合以源码为唯一来源枚举，避免测试内另存一份手写清单而与新增常量脱节。
const responseSourcePath = "../../pkg/response/response.go"

// errorsContract 错误码契约中参与锚定的两节。
type errorsContract struct {
	// Transport 传输层响应码集合，键为响应体 code 字段的取值。
	Transport map[int]string `yaml:"transport"`
	// Mapping 哨兵错误到传输层码的分档，键为码的字符串形式，值为哨兵引用名。
	Mapping map[string][]string `yaml:"mapping"`
}

// loadErrorsContract 读取并解析错误码契约，任一节缺失即判定契约不可用。
func loadErrorsContract(t *testing.T) errorsContract {
	t.Helper()

	raw, err := os.ReadFile(errorsContractPath)
	if err != nil {
		t.Fatalf("读取错误码契约失败: %v", err)
	}

	var contract errorsContract
	if err := yaml.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("解析错误码契约失败: %v", err)
	}
	if len(contract.Transport) == 0 {
		t.Fatalf("错误码契约 %s 缺少 transport 节", errorsContractPath)
	}
	if contract.Mapping == nil {
		t.Fatalf("错误码契约 %s 缺少 mapping 节", errorsContractPath)
	}

	return contract
}

// responseCodesFromSource 解析 pkg/response 源码，枚举导出的整型响应码常量。
// 常量取值必须写成整型字面量，改用表达式时在此报错而不是静默漏检。
func responseCodesFromSource(t *testing.T) map[int]string {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), responseSourcePath, nil, 0)
	if err != nil {
		t.Fatalf("解析 %s 失败: %v", responseSourcePath, err)
	}

	codes := map[int]string{}
	for _, declaration := range file.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.CONST {
			continue
		}
		for _, spec := range generic.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, name := range values.Names {
				if !name.IsExported() {
					continue
				}
				if index >= len(values.Values) {
					t.Errorf("%s.%s 未显式赋值，锚定测试无法读取其取值", responseSourcePath, name.Name)
					continue
				}
				literal, ok := values.Values[index].(*ast.BasicLit)
				if !ok || literal.Kind != token.INT {
					t.Errorf("%s.%s 的取值不是整型字面量，锚定测试无法读取", responseSourcePath, name.Name)
					continue
				}
				code, err := strconv.Atoi(literal.Value)
				if err != nil {
					t.Errorf("%s.%s 的取值 %q 无法解析为响应码: %v", responseSourcePath, name.Name, literal.Value, err)
					continue
				}
				codes[code] = name.Name
			}
		}
	}
	if len(codes) == 0 {
		t.Fatalf("未从 %s 解析到任何导出的响应码常量", responseSourcePath)
	}

	return codes
}

// registryNamesByCode 按传输层码归拢注册表中的哨兵引用名。
func registryNamesByCode() map[int][]string {
	namesByCode := map[int][]string{}
	for _, mapping := range serviceErrorMappings {
		namesByCode[mapping.Code] = append(namesByCode[mapping.Code], mapping.Name)
	}
	return namesByCode
}

// contractNamesByCode 把契约 mapping 节的字符串码键解析为响应码后归拢。
func contractNamesByCode(t *testing.T, mapping map[string][]string) map[int][]string {
	t.Helper()

	namesByCode := map[int][]string{}
	for rawCode, names := range mapping {
		code, err := strconv.Atoi(rawCode)
		if err != nil {
			t.Errorf("契约 mapping 的分类键 %q 不是合法响应码: %v", rawCode, err)
			continue
		}
		namesByCode[code] = names
	}
	return namesByCode
}

// TestErrorMappingMatchesContract 校验错误分档注册表与错误码契约双向一致。
// 契约是错误语义的唯一编辑点，pkg/response 源码是传输层码的唯一来源，
// 注册表是运行期唯一映射权威，三者在服务端两两比对；前端常量由
// packages/api 的 errors.test.ts 与本文件同源锚定。任一侧单独改动都会变红：
// 契约漏登记源码中的响应码、注册表多出或缺少哨兵、分档码不是传输层码三类情况均在此暴露。
func TestErrorMappingMatchesContract(t *testing.T) {
	contract := loadErrorsContract(t)
	transportCodes := responseCodesFromSource(t)

	assertSameTransportCodes(t, contract.Transport, transportCodes)
	assertSameMappings(t, contractNamesByCode(t, contract.Mapping), transportCodes)
}

// assertSameTransportCodes 断言契约 transport 节与源码中的响应码常量集合完全一致。
func assertSameTransportCodes(t *testing.T, declared map[int]string, implemented map[int]string) {
	t.Helper()

	for code, meaning := range declared {
		if _, ok := implemented[code]; !ok {
			t.Errorf("契约 transport 声明了响应码 %d（%s），%s 未定义该常量", code, meaning, responseSourcePath)
		}
	}
	for code, name := range implemented {
		if _, ok := declared[code]; !ok {
			t.Errorf("%s 定义的 %s = %d 未登记到契约 transport 节", responseSourcePath, name, code)
		}
	}
}

// assertSameMappings 断言契约 mapping 节与注册表的码集合及哨兵名集合双向一致。
// 留空分类（例如 500 表示未命中哨兵时的兜底归类）在两侧同为空时视为一致。
func assertSameMappings(t *testing.T, contractNames map[int][]string, transportCodes map[int]string) {
	t.Helper()

	registryNames := registryNamesByCode()
	for _, code := range unionCodes(contractNames, registryNames) {
		if _, ok := transportCodes[code]; !ok {
			t.Errorf("分档使用了响应码 %d，该码不在传输层响应码集合内", code)
		}

		declared, inContract := contractNames[code]
		registered, inRegistry := registryNames[code]
		switch {
		case !inContract && len(registered) > 0:
			t.Errorf("注册表存在映射到码 %d 的哨兵，契约 mapping 缺少 '%d' 分类，实际登记: %v", code, code, registered)
		case !inRegistry && len(declared) > 0:
			t.Errorf("契约 mapping 的 '%d' 分类在注册表中没有对应哨兵，契约声明: %v", code, declared)
		default:
			assertSameNameSet(t, code, declared, registered)
		}
	}
}

// unionCodes 汇总两侧出现的全部响应码，保证单侧独有的码也能进入比对。
func unionCodes(left, right map[int][]string) []int {
	seen := map[int]bool{}
	codes := make([]int, 0, len(left)+len(right))
	for _, source := range []map[int][]string{left, right} {
		for code := range source {
			if !seen[code] {
				seen[code] = true
				codes = append(codes, code)
			}
		}
	}
	sort.Ints(codes)
	return codes
}

// assertSameNameSet 断言契约与注册表的哨兵引用名集合互为子集，双向缺失分别报出。
func assertSameNameSet(t *testing.T, code int, declared, registered []string) {
	t.Helper()

	declaredSet := map[string]bool{}
	for _, name := range declared {
		declaredSet[name] = true
	}
	registeredSet := map[string]bool{}
	for _, name := range registered {
		registeredSet[name] = true
	}

	for _, name := range sortedNames(declaredSet) {
		if !registeredSet[name] {
			t.Errorf("码 %d：契约声明的 %s 未登记到注册表", code, name)
		}
	}
	for _, name := range sortedNames(registeredSet) {
		if !declaredSet[name] {
			t.Errorf("码 %d：注册表登记的 %s 未写入契约 mapping 节", code, name)
		}
	}
}

// sortedNames 返回集合中的名字并排序，使失败信息稳定可比对。
func sortedNames(names map[string]bool) []string {
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	return sorted
}
