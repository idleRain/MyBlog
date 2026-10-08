package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestMediaAfterFindExposesNarrowedUploader 验证媒体文件的上传者经窄化视图输出且不含个人信息。
func TestMediaAfterFindExposesNarrowedUploader(t *testing.T) {
	media := &MediaFile{
		ID:       1,
		Filename: "cover.png",
		UploadIP: "203.0.113.8",
		Uploader: User{
			ID:       7,
			Username: "alice",
			Email:    "alice@example.com",
			Phone:    &[]string{"13800000000"}[0],
			Nickname: "爱丽丝",
			Location: "上海",
			Timezone: "Asia/Shanghai",
			Locale:   "zh-CN",
			Gender:   &[]int{1}[0],
		},
	}

	if err := media.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind 不应返回错误: %v", err)
	}
	if media.UploaderPublic == nil {
		t.Fatal("上传者关联存在时公开上传者视图不应为空")
	}

	data, err := json.Marshal(media)
	if err != nil {
		t.Fatalf("序列化媒体文件失败: %v", err)
	}
	output := string(data)

	// 上传者审计与个人信息字段一律不得随媒体响应输出。
	for _, forbidden := range []string{
		"email", "phone", "birthday", "location", "gender", "timezone", "locale", "uploadIP",
	} {
		if strings.Contains(output, `"`+forbidden+`"`) {
			t.Errorf("媒体公开输出不应包含字段 %q，实际为 %s", forbidden, output)
		}
	}
	if !strings.Contains(output, `"uploader"`) {
		t.Errorf("媒体公开输出应包含 uploader 窄化键，实际为 %s", output)
	}
	if !strings.Contains(output, `"nickname":"爱丽丝"`) {
		t.Errorf("媒体公开输出应包含窄化后的昵称，实际为 %s", output)
	}

	// 上传者输出键集须与窄化视图定义一致，确保白名单不被悄然扩大。
	var decoded struct {
		Uploader map[string]any `json:"uploader"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("反序列化媒体文件失败: %v", err)
	}
	allowedKeys := map[string]bool{"id": true, "username": true, "nickname": true, "avatar": true}
	for key := range decoded.Uploader {
		if !allowedKeys[key] {
			t.Errorf("上传者窄化视图不应输出未知字段 %q，实际键集为 %v", key, keysOf(decoded.Uploader))
		}
	}
}

// TestMediaAfterFindWithoutUploaderOmitsUploaderKey 验证上传者未预加载时 uploader 键整体省略。
func TestMediaAfterFindWithoutUploaderOmitsUploaderKey(t *testing.T) {
	media := &MediaFile{ID: 1, Filename: "cover.png"}

	if err := media.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind 不应返回错误: %v", err)
	}

	data, err := json.Marshal(media)
	if err != nil {
		t.Fatalf("序列化媒体文件失败: %v", err)
	}
	if strings.Contains(string(data), `"uploader"`) {
		t.Errorf("上传者未预加载时不应输出 uploader 键，实际为 %s", string(data))
	}
}

// TestMediaUploadIPHiddenFromJSON 验证上传者 IP 审计字段在未预加载关联时同样不随响应输出。
func TestMediaUploadIPHiddenFromJSON(t *testing.T) {
	media := &MediaFile{ID: 1, Filename: "cover.png", UploadIP: "203.0.113.8"}

	data, err := json.Marshal(media)
	if err != nil {
		t.Fatalf("序列化媒体文件失败: %v", err)
	}
	if strings.Contains(string(data), `"uploadIP"`) {
		t.Errorf("媒体公开输出不应包含审计字段 uploadIP，实际为 %s", string(data))
	}
}

// keysOf 列出 map 的键列表，仅用于测试失败信息展示。
func keysOf(source map[string]any) []string {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	return keys
}
