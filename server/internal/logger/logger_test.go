package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain 把 gin 切到测试模式，避免其调试警告经已桥接的输出混入断言使用的日志缓冲区。
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// decodeRecord 解析一行 JSON 日志，解析失败即判定为失败。
func decodeRecord(t *testing.T, line string) map[string]any {
	t.Helper()

	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("日志行不是合法 JSON: %v，原文 %q", err, line)
	}
	return record
}

// records 按行解析缓冲区内的全部日志记录。
func records(t *testing.T, buffer *bytes.Buffer) []map[string]any {
	t.Helper()

	var decoded []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buffer.String()), "\n") {
		if line == "" {
			continue
		}
		decoded = append(decoded, decodeRecord(t, line))
	}
	return decoded
}

// TestInitWritesJSONRecords 验证初始化后输出的每行都是 JSON 且含时间、级别与消息。
func TestInitWritesJSONRecords(t *testing.T) {
	var buffer bytes.Buffer
	if err := Init(Options{Level: LevelInfo, Writer: &buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	L().Info("服务启动完成")

	decoded := records(t, &buffer)
	if len(decoded) != 1 {
		t.Fatalf("日志行数 = %d, 期望 1", len(decoded))
	}
	for _, key := range []string{"time", "level", "msg"} {
		if _, ok := decoded[0][key]; !ok {
			t.Errorf("日志记录缺少字段 %q: %v", key, decoded[0])
		}
	}
	if decoded[0]["msg"] != "服务启动完成" {
		t.Errorf("msg = %v, 期望 服务启动完成", decoded[0]["msg"])
	}
}

// TestInitFiltersBelowConfiguredLevel 验证低于配置级别的记录被丢弃。
func TestInitFiltersBelowConfiguredLevel(t *testing.T) {
	var buffer bytes.Buffer
	if err := Init(Options{Level: LevelWarn, Writer: &buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	L().Info("不该出现")
	L().Warn("应该出现")

	decoded := records(t, &buffer)
	if len(decoded) != 1 {
		t.Fatalf("日志行数 = %d, 期望 1，实际内容 %q", len(decoded), buffer.String())
	}
	if decoded[0]["msg"] != "应该出现" {
		t.Errorf("保留的记录 = %v, 期望 应该出现", decoded[0]["msg"])
	}
}

// TestInitRejectsUnknownLevelAndOutput 验证非法级别名与输出目标在启动期即报错。
func TestInitRejectsUnknownLevelAndOutput(t *testing.T) {
	cases := []struct {
		name string
		opts Options
	}{
		{name: "未知级别", opts: Options{Level: "verbose"}},
		{name: "未知输出目标", opts: Options{Level: LevelInfo, Output: "syslog"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := Init(testCase.opts); err == nil {
				t.Error("非法配置应返回错误")
			}
		})
	}
}

// TestInitWritesFileOutput 验证输出目标为文件时记录写入指定文件，且文件句柄可释放。
func TestInitWritesFileOutput(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "app.log")
	if err := Init(Options{Level: LevelInfo, Output: OutputFile, FilePath: logPath}); err != nil {
		t.Fatalf("初始化文件日志失败: %v", err)
	}

	L().Info("写入文件的记录")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	if !strings.Contains(string(content), "写入文件的记录") {
		t.Errorf("日志文件缺少记录: %q", string(content))
	}

	// 释放文件输出目标，否则临时目录在 Windows 上无法清理。
	if err := Close(); err != nil {
		t.Fatalf("关闭日志文件失败: %v", err)
	}
}

// TestCloseIsIdempotent 验证重复释放日志输出目标为空操作。
func TestCloseIsIdempotent(t *testing.T) {
	if err := Init(Options{Level: LevelInfo, Writer: &bytes.Buffer{}}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		if err := Close(); err != nil {
			t.Fatalf("第 %d 次释放失败: %v", attempt+1, err)
		}
	}
}

// TestFromGinCarriesRequestID 验证请求标识随 gin 上下文传递并写入记录。
func TestFromGinCarriesRequestID(t *testing.T) {
	var buffer bytes.Buffer
	if err := Init(Options{Level: LevelInfo, Writer: &buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	AttachRequestID(ctx, "req-abc")
	FromGin(ctx).Info("带请求标识的记录")

	decoded := records(t, &buffer)
	if len(decoded) != 1 {
		t.Fatalf("日志行数 = %d, 期望 1", len(decoded))
	}
	if decoded[0][FieldRequestID] != "req-abc" {
		t.Errorf("%s = %v, 期望 req-abc", FieldRequestID, decoded[0][FieldRequestID])
	}
}

// TestFromGinWithoutRequestIDOmitsField 验证上下文无标识时不写入空字段。
func TestFromGinWithoutRequestIDOmitsField(t *testing.T) {
	var buffer bytes.Buffer
	if err := Init(Options{Level: LevelInfo, Writer: &buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	FromGin(ctx).Info("无请求标识的记录")

	decoded := records(t, &buffer)
	if _, ok := decoded[0][FieldRequestID]; ok {
		t.Errorf("无标识时不应写入 %s 字段: %v", FieldRequestID, decoded[0])
	}
}

// TestStdlibLogBridgeKeepsLevelAndJSON 验证标准库 log 输出转为 JSON 且级别标记被识别。
func TestStdlibLogBridgeKeepsLevelAndJSON(t *testing.T) {
	var buffer bytes.Buffer
	if err := Init(Options{Level: LevelDebug, Writer: &buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	log.Printf("[ERROR] 未分类的服务错误：%v", "示例错误")
	log.Printf("不带级别标记的记录")

	decoded := records(t, &buffer)
	if len(decoded) != 2 {
		t.Fatalf("日志行数 = %d, 期望 2，实际内容 %q", len(decoded), buffer.String())
	}
	if decoded[0]["level"] != slog.LevelError.String() {
		t.Errorf("首行 level = %v, 期望 ERROR", decoded[0]["level"])
	}
	if strings.Contains(decoded[0]["msg"].(string), "[ERROR]") {
		t.Errorf("级别标记不应残留在消息中: %v", decoded[0]["msg"])
	}
	if decoded[1]["level"] != slog.LevelInfo.String() {
		t.Errorf("次行 level = %v, 期望 INFO", decoded[1]["level"])
	}
}

// TestInitBridgesGinOutput 验证 gin 自身的调试输出同样以 JSON 记录写出，不留明文行。
func TestInitBridgesGinOutput(t *testing.T) {
	var buffer bytes.Buffer
	if err := Init(Options{Level: LevelInfo, Writer: &buffer}); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	if _, err := gin.DefaultWriter.Write([]byte("[GIN-debug] GET /api/health\n")); err != nil {
		t.Fatalf("写入 gin 调试输出失败: %v", err)
	}

	decoded := records(t, &buffer)
	if len(decoded) != 1 {
		t.Fatalf("日志行数 = %d, 期望 1，实际内容 %q", len(decoded), buffer.String())
	}
	if decoded[0]["msg"] != "[GIN-debug] GET /api/health" {
		t.Errorf("msg = %v, 期望原样保留调试文案", decoded[0]["msg"])
	}
}
