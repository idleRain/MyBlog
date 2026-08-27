// Package datetime 提供 JSON 日期格式化相关单元的单元测试
package datetime

import (
	"encoding/json"
	"testing"
	"time"
)

// TestParseJSONDate 验证各类受支持格式的日期字符串解析。
func TestParseJSONDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "纯日期格式", input: "2026-07-31", want: "2026-07-31 00:00:00"},
		{name: "日期时间格式", input: "2026-07-31 12:30:45", want: "2026-07-31 12:30:45"},
		{name: "RFC3339格式", input: "2026-07-31T12:30:45Z", want: "2026-07-31 12:30:45"},
		{name: "非法格式报错", input: "not-a-date", wantErr: true},
		{name: "空字符串报错", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSONDate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseJSONDate(%q) 期望报错，实际无错误", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseJSONDate(%q) 返回错误: %v", tt.input, err)
			}

			wantTime, _ := time.Parse(JSONDateFormat, tt.want)
			if !got.Time.Equal(wantTime) {
				t.Errorf("ParseJSONDate(%q) = %v，期望 %v", tt.input, got.Time, wantTime)
			}
		})
	}
}

// TestUnmarshalJSON 验证 JSON 反序列化对可选日期的容错处理。
func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		isZero  bool
	}{
		{name: "纯日期格式", input: `"2026-07-31"`},
		{name: "日期时间格式", input: `"2026-07-31 12:30:45"`},
		{name: "RFC3339格式", input: `"2026-07-31T12:30:45Z"`},
		{name: "空字符串视为未设置", input: `""`, isZero: true},
		{name: "null视为未设置", input: `null`, isZero: true},
		{name: "非法格式报错", input: `"invalid"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jd JSONDate
			err := json.Unmarshal([]byte(tt.input), &jd)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Unmarshal(%q) 期望报错，实际无错误", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unmarshal(%q) 返回错误: %v", tt.input, err)
			}

			if tt.isZero && !jd.Time.IsZero() {
				t.Errorf("Unmarshal(%q) 期望零值时间，实际为 %v", tt.input, jd.Time)
			}
			if !tt.isZero && jd.Time.IsZero() {
				t.Errorf("Unmarshal(%q) 期望非零时间，实际为零值", tt.input)
			}
		})
	}
}

// TestMarshalJSON 验证零值时间输出为 null，当天零点输出纯日期。
func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{name: "零值时间输出null", input: time.Time{}, want: "null"},
		{name: "当天零点输出纯日期", input: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC), want: `"2026-07-31"`},
		{name: "非零点输出完整时间", input: time.Date(2026, 7, 31, 12, 30, 45, 0, time.UTC), want: `"2026-07-31 12:30:45"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(JSONDate{Time: tt.input})
			if err != nil {
				t.Fatalf("Marshal(%v) 返回错误: %v", tt.input, err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal(%v) = %s，期望 %s", tt.input, got, tt.want)
			}
		})
	}
}
