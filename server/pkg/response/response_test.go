package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// 统一响应信封固定以 HTTP 200 承载业务码，客户端据响应体的 code 字段判定语义。
// 本文件锚定信封的两个易漂移面：五档响应码的取值，以及 data 键在成功与失败路径下的出现规则。
const (
	defaultSuccessMessage = "操作成功"
	testMessage           = "自定义消息"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// responseEnvelope 是测试侧对响应报文的独立解码结构。
// 刻意不复用生产 Response 类型，避免实现改动同时改变断言依据而失去守护效果。
type responseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// invokeHandler 在测试上下文内执行一次响应写入并返回解码后的报文与 HTTP 状态码。
func invokeHandler(t *testing.T, handler gin.HandlerFunc) (responseEnvelope, int) {
	t.Helper()

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/test", nil)

	handler(context)

	var envelope responseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v, 原文: %s", err, recorder.Body.String())
	}

	return envelope, recorder.Code
}

func TestSuccessEnvelope(t *testing.T) {
	envelope, status := invokeHandler(t, func(c *gin.Context) {
		Success(c, gin.H{"id": 7})
	})

	if status != http.StatusOK {
		t.Errorf("HTTP 状态码 = %d, 期望 %d", status, http.StatusOK)
	}
	if envelope.Code != CodeSuccess {
		t.Errorf("业务码 = %d, 期望 %d", envelope.Code, CodeSuccess)
	}
	if envelope.Message != defaultSuccessMessage {
		t.Errorf("消息 = %q, 期望 %q", envelope.Message, defaultSuccessMessage)
	}
	if string(envelope.Data) != `{"id":7}` {
		t.Errorf("data = %s, 期望 {\"id\":7}", envelope.Data)
	}
}

func TestSuccessWithMessageOverridesDefaultMessage(t *testing.T) {
	envelope, _ := invokeHandler(t, func(c *gin.Context) {
		SuccessWithMessage(c, testMessage, nil)
	})

	if envelope.Code != CodeSuccess {
		t.Errorf("业务码 = %d, 期望 %d", envelope.Code, CodeSuccess)
	}
	if envelope.Message != testMessage {
		t.Errorf("消息 = %q, 期望 %q", envelope.Message, testMessage)
	}
}

// data 为 nil 时 omitempty 会省略整个键，而不是输出 null。
// 契约锁的 fixture 依赖该语义，实现若改为显式输出 null 必须同步修订金样本。
func TestSuccessOmitsDataKeyWhenPayloadIsNil(t *testing.T) {
	envelope, _ := invokeHandler(t, func(c *gin.Context) {
		Success(c, nil)
	})

	if envelope.Data != nil {
		t.Errorf("data 键 = %s, 期望被省略", envelope.Data)
	}
	if envelope.Data != nil && string(envelope.Data) == "null" {
		t.Errorf("data 键输出为 null, 期望整个键被省略")
	}
	if envelope.Code != CodeSuccess {
		t.Errorf("业务码 = %d, 期望 %d", envelope.Code, CodeSuccess)
	}
}

// omitempty 的判定依据是接口内的具体值，非 nil 的空切片不会被省略。
// 实测 encoding/json 的语义为：nil 省略键，[]T{} 与 map[] 分别输出 [] 与 {}，
// 理由是接口内已装箱的具体类型不是 nil。此用例把该边界固定下来，
// 使消费方明确「data 键可能为空集合」是契约的一部分，而非实现疏漏。
func TestSuccessKeepsDataKeyWhenPayloadIsEmptyCollection(t *testing.T) {
	envelope, _ := invokeHandler(t, func(c *gin.Context) {
		Success(c, []string{})
	})

	if envelope.Data == nil {
		t.Fatalf("data 键被省略, 期望保留为空数组")
	}
	if string(envelope.Data) != `[]` {
		t.Errorf("data = %s, 期望 []", envelope.Data)
	}
}

// 错误响应一律不携带 data 键，这是错误分档契约的组成部分。
func TestErrorEnvelopeOmitsDataKey(t *testing.T) {
	envelope, status := invokeHandler(t, func(c *gin.Context) {
		Error(c, CodeError, "内部错误")
	})

	if status != http.StatusOK {
		t.Errorf("HTTP 状态码 = %d, 期望 %d", status, http.StatusOK)
	}
	if envelope.Code != CodeError {
		t.Errorf("业务码 = %d, 期望 %d", envelope.Code, CodeError)
	}
	if envelope.Message != "内部错误" {
		t.Errorf("消息 = %q, 期望 %q", envelope.Message, "内部错误")
	}
	if envelope.Data != nil {
		t.Errorf("data 键 = %s, 期望被省略", envelope.Data)
	}
}

// 五个快捷函数必须映射到各自档位的响应码，这是 handler 层错误分档的唯一出口。
func TestErrorHelpersMapToDeclaredCodes(t *testing.T) {
	cases := []struct {
		name     string
		handler  gin.HandlerFunc
		expected int
	}{
		{"BadRequest", func(c *gin.Context) { BadRequest(c, "参数错误") }, CodeInvalid},
		{"InternalError", func(c *gin.Context) { InternalError(c, "内部错误") }, CodeError},
		{"NotFound", func(c *gin.Context) { NotFound(c, "未找到") }, CodeNotFound},
		{"Unauthorized", func(c *gin.Context) { Unauthorized(c, "未认证") }, CodeAuth},
		{"Forbidden", func(c *gin.Context) { Forbidden(c, "无权限") }, CodeForbid},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			envelope, status := invokeHandler(t, testCase.handler)

			if status != http.StatusOK {
				t.Errorf("HTTP 状态码 = %d, 期望 %d", status, http.StatusOK)
			}
			if envelope.Code != testCase.expected {
				t.Errorf("业务码 = %d, 期望 %d", envelope.Code, testCase.expected)
			}
			if envelope.Data != nil {
				t.Errorf("data 键 = %s, 期望被省略", envelope.Data)
			}
		})
	}
}

// 响应码取值是前后端共同依赖的契约，改动必须同时满足两侧，因此在此固定数值。
func TestDeclaredCodeValues(t *testing.T) {
	cases := []struct {
		name     string
		actual   int
		expected int
	}{
		{"CodeSuccess", CodeSuccess, 200},
		{"CodeInvalid", CodeInvalid, 400},
		{"CodeAuth", CodeAuth, 401},
		{"CodeForbid", CodeForbid, 403},
		{"CodeNotFound", CodeNotFound, 404},
		{"CodeError", CodeError, 500},
	}

	for _, testCase := range cases {
		if testCase.actual != testCase.expected {
			t.Errorf("%s = %d, 期望 %d", testCase.name, testCase.actual, testCase.expected)
		}
	}
}
