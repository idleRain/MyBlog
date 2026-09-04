package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// fixtureContractPath 契约金样本目录相对路径，基于本测试包所在目录计算。
const fixtureContractPath = "../../../contracts/fixtures"

// fakeLoginUserService 登录场景测试替身，仅覆盖 Login 方法，其余方法沿接口零值。
type fakeLoginUserService struct {
	service.UserService
	loginErr error
}

func (f *fakeLoginUserService) Login(username, password string) (*service.LoginResponse, error) {
	return nil, f.loginErr
}

// assertResponseMatchesFixture 断言 gin 响应体与契约金样本在语义上逐字节一致。
// JSON 空白不参与比对，避免金样本的人读缩进干扰字节级校验。
func assertResponseMatchesFixture(t *testing.T, recorder *httptest.ResponseRecorder, fixtureName string) {
	t.Helper()

	fixturePath := fixtureContractPath + "/" + fixtureName
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("读取契约金样本失败: %v", err)
	}

	var fixtureCompact bytes.Buffer
	if err := json.Compact(&fixtureCompact, fixtureBytes); err != nil {
		t.Fatalf("解析契约金样本失败: %v", err)
	}

	if recorder.Body.String() != fixtureCompact.String() {
		t.Errorf("响应与契约金样本不一致\n实际: %s\n期望: %s", recorder.Body.String(), fixtureCompact.String())
	}
}

// TestLoginWrongPasswordMatchesFixture 验证登录失败响应与契约金样本一致，构成契约锁 ①。
func TestLoginWrongPasswordMatchesFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewUserHandler(&fakeLoginUserService{loginErr: errors.New("密码错误")})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/users/login",
		strings.NewReader(`{"username":"admin","password":"wrong-password"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Login(ctx)

	assertResponseMatchesFixture(t, recorder, "login.wrong-password.json")
}
