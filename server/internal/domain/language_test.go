package domain

import "testing"

// TestParseLanguage 覆盖语言解析的归一化、白名单命中与回退规则。
func TestParseLanguage(t *testing.T) {
	supported := []Language{LanguageChinese, LanguageEnglish}

	cases := []struct {
		name     string
		raw      string
		expected Language
	}{
		{name: "空请求头回退缺省语言", raw: "", expected: LanguageChinese},
		{name: "纯空白请求头回退缺省语言", raw: "   ", expected: LanguageChinese},
		{name: "精确命中中文", raw: "zh", expected: LanguageChinese},
		{name: "精确命中英文", raw: "en", expected: LanguageEnglish},
		{name: "地区变体折叠为主子标签", raw: "zh-CN", expected: LanguageChinese},
		{name: "书写变体折叠为主子标签", raw: "zh-Hans", expected: LanguageChinese},
		{name: "英文地区变体折叠为主子标签", raw: "en-US", expected: LanguageEnglish},
		{name: "大写标签归一化后命中", raw: "EN", expected: LanguageEnglish},
		{name: "白名单外语言回退缺省语言", raw: "fr", expected: LanguageChinese},
		{name: "携带质量参数时剥离后命中", raw: "en;q=0.9", expected: LanguageEnglish},
		{name: "多标签按出现顺序取首个命中", raw: "fr-FR, en;q=0.8", expected: LanguageEnglish},
		{name: "通配符整头返回全量包标识", raw: "*", expected: LanguageAll},
		{name: "列表内通配符跳过后继续匹配", raw: "*, zh;q=0.5", expected: LanguageChinese},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			actual := ParseLanguage(item.raw, supported, LanguageChinese)
			if actual != item.expected {
				t.Errorf("ParseLanguage(%q) = %q, 期望 %q", item.raw, actual, item.expected)
			}
		})
	}
}

// TestParseLanguages 覆盖配置语言列表的归一化规则。
func TestParseLanguages(t *testing.T) {
	actual := ParseLanguages([]string{"zh-CN", " EN ", "", "*", "en;q=0.9"})
	expected := []Language{LanguageChinese, LanguageEnglish}

	if len(actual) != len(expected) {
		t.Fatalf("ParseLanguages 结果长度 = %d, 期望 %d, 实际值 %v", len(actual), len(expected), actual)
	}
	for index, item := range expected {
		if actual[index] != item {
			t.Errorf("ParseLanguages 第 %d 项 = %q, 期望 %q", index, actual[index], item)
		}
	}
}

// TestParseLanguagesDropsInvalidEntries 验证通配符与空值不会进入白名单。
func TestParseLanguagesDropsInvalidEntries(t *testing.T) {
	actual := ParseLanguages([]string{"*", "  "})
	if len(actual) != 0 {
		t.Errorf("无效条目应全部丢弃, 实际结果 %v", actual)
	}
}
