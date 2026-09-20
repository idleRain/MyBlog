package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// fakeDictService 字典服务的测试替身，返回可配置结果。
type fakeDictService struct {
	service.DictServiceInterface
	groups    []*service.EnabledDictGroup
	groupErr  error
	byCode    *service.EnabledDictGroup
	byCodeErr error
	createErr error
}

// ListEnabledDicts 显式覆写全量查询，返回配置的分组集合。
func (f *fakeDictService) ListEnabledDicts() ([]*service.EnabledDictGroup, error) {
	return f.groups, f.groupErr
}

// ListEnabledItemsByTypeCode 显式覆写按码查询，返回配置的分组。
func (f *fakeDictService) ListEnabledItemsByTypeCode(code string) (*service.EnabledDictGroup, error) {
	return f.byCode, f.byCodeErr
}

// CreateDictType 显式覆写类型创建，返回配置的业务错误。
func (f *fakeDictService) CreateDictType(req *service.CreateDictTypeRequest) (*model.DictType, error) {
	return nil, f.createErr
}

// newDictProbeGroup 构造带英文翻译行的探针字典分组。
func newDictProbeGroup(code string) *service.EnabledDictGroup {
	return &service.EnabledDictGroup{
		DictType: model.DictType{
			ID:   1,
			Code: code,
			Name: "标签状态",
			Translations: []model.DictTypeTranslation{
				{Locale: "en", Name: "Tag Status"},
			},
		},
		Items: []*model.DictItem{
			{
				ID:     1,
				TypeID: 1,
				Value:  "1",
				Label:  "启用",
				Translations: []model.DictItemTranslation{
					{Locale: "en", Label: "Enabled"},
				},
			},
		},
	}
}

// newDictProbeRouter 构造注入语言上下文与字典处理器的测试路由。
// 同时注册 /all 与 /:type 两条路由，验证静态段与参数段可共存。
func newDictProbeRouter(dictHandler DictHandlerInterface, language domain.Language) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(domain.LanguageContextKey, language)
		c.Next()
	})
	router.POST("/api/dicts/all", dictHandler.GetAllDicts)
	router.POST("/api/dicts/:type", dictHandler.GetDictByType)
	router.POST("/api/admin/dicts/types/create", dictHandler.CreateDictType)
	return router
}

// parseDictResponseCode 解析响应体中的业务码。
func parseDictResponseCode(t *testing.T, recorder *httptest.ResponseRecorder) float64 {
	t.Helper()

	var body struct {
		Code float64 `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return body.Code
}

// TestGetAllDictsLocalizesAndWrapsGroups 验证全量字典按语言本地化输出并包裹 dicts 形状。
func TestGetAllDictsLocalizesAndWrapsGroups(t *testing.T) {
	dictHandler := NewDictHandler(&fakeDictService{
		groups: []*service.EnabledDictGroup{newDictProbeGroup("tag_status")},
	})
	router := newDictProbeRouter(dictHandler, domain.LanguageEnglish)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/dicts/all", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际 %d", recorder.Code)
	}
	if header := recorder.Header().Get(ContentLanguageHeader); header != "en" {
		t.Errorf("Content-Language = %q, 期望 en", header)
	}

	var body struct {
		Code float64 `json:"code"`
		Data struct {
			Dicts []struct {
				Code  string `json:"code"`
				Name  string `json:"name"`
				Items []struct {
					Label string `json:"label"`
				} `json:"items"`
			} `json:"dicts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if body.Code != response.CodeSuccess {
		t.Errorf("业务码 = %v, 期望 %v", body.Code, response.CodeSuccess)
	}
	if len(body.Data.Dicts) != 1 || body.Data.Dicts[0].Code != "tag_status" {
		t.Fatalf("期望返回 1 个 tag_status 分组, 实际 %v", body.Data.Dicts)
	}
	if body.Data.Dicts[0].Name != "Tag Status" {
		t.Errorf("分组名称 = %q, 期望 Tag Status", body.Data.Dicts[0].Name)
	}
	if len(body.Data.Dicts[0].Items) != 1 || body.Data.Dicts[0].Items[0].Label != "Enabled" {
		t.Errorf("字典项标签 = %v, 期望 Enabled", body.Data.Dicts[0].Items)
	}
}

// TestGetDictByTypeMapsUnknownTypeToNotFound 验证字典类型不存在时业务码映射 404。
func TestGetDictByTypeMapsUnknownTypeToNotFound(t *testing.T) {
	dictHandler := NewDictHandler(&fakeDictService{
		byCodeErr: repository.ErrDictTypeNotFound,
	})
	router := newDictProbeRouter(dictHandler, domain.DefaultLanguage)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/dicts/tag_status", nil))

	if code := parseDictResponseCode(t, recorder); code != response.CodeNotFound {
		t.Errorf("业务码 = %v, 期望 %v", code, response.CodeNotFound)
	}
}

// TestGetDictByTypeReturnsLocalizedGroup 验证单字典接口按语言本地化输出。
func TestGetDictByTypeReturnsLocalizedGroup(t *testing.T) {
	dictHandler := NewDictHandler(&fakeDictService{
		byCode: newDictProbeGroup("tag_status"),
	})
	router := newDictProbeRouter(dictHandler, domain.LanguageEnglish)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/dicts/tag_status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际 %d", recorder.Code)
	}

	var body struct {
		Data struct {
			Code  string `json:"code"`
			Items []struct {
				Value string `json:"value"`
				Label string `json:"label"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if body.Data.Code != "tag_status" {
		t.Errorf("分组字典码 = %q, 期望 tag_status", body.Data.Code)
	}
	if len(body.Data.Items) != 1 || body.Data.Items[0].Label != "Enabled" {
		t.Errorf("字典项标签 = %v, 期望 Enabled", body.Data.Items)
	}
}

// TestCreateDictTypeMapsBusinessErrors 验证管理端创建请求的错误分档。
func TestCreateDictTypeMapsBusinessErrors(t *testing.T) {
	cases := []struct {
		name         string
		requestBody  string
		serviceError error
		expectedCode int
	}{
		{
			name:         "绑定失败映射 400",
			requestBody:  `{"code": 1}`,
			expectedCode: response.CodeInvalid,
		},
		{
			name:         "业务校验失败映射 400",
			requestBody:  `{"code":"TagStatus","name":"标签状态"}`,
			serviceError: service.ErrInvalidRequest,
			expectedCode: response.CodeInvalid,
		},
		{
			name:         "类型不存在映射 404",
			requestBody:  `{"code":"tag_status","name":"标签状态"}`,
			serviceError: repository.ErrDictTypeNotFound,
			expectedCode: response.CodeNotFound,
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			dictHandler := NewDictHandler(&fakeDictService{createErr: item.serviceError})
			router := newDictProbeRouter(dictHandler, domain.DefaultLanguage)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/admin/dicts/types/create", strings.NewReader(item.requestBody))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)

			if code := parseDictResponseCode(t, recorder); code != float64(item.expectedCode) {
				t.Errorf("业务码 = %v, 期望 %v", code, item.expectedCode)
			}
		})
	}
}

// TestGetAllDictsInternalError 验证服务层内部错误业务码映射 500。
func TestGetAllDictsInternalError(t *testing.T) {
	dictHandler := NewDictHandler(&fakeDictService{
		groupErr: errors.New("数据库错误"),
	})
	router := newDictProbeRouter(dictHandler, domain.DefaultLanguage)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/dicts/all", nil))

	if code := parseDictResponseCode(t, recorder); code != response.CodeError {
		t.Errorf("业务码 = %v, 期望 %v", code, response.CodeError)
	}
}
