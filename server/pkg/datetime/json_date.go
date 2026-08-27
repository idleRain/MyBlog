// Package datetime 提供JSON日期格式化。
package datetime

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// 支持的日期时间字符串格式常量，按从简到繁的顺序依次尝试解析。
const (
	// JSONDateOnlyFormat 仅包含日期部分，用于生日等按天粒度的字段。
	JSONDateOnlyFormat = "2006-01-02"
	// JSONDateFormat 包含日期与时间部分，用于时间戳类字段。
	JSONDateFormat = "2006-01-02 15:04:05"
)

// jsonDateFormats 保存 JSONDate 支持的字符串解析格式列表。
var jsonDateFormats = []string{
	JSONDateOnlyFormat,
	JSONDateFormat,
	time.RFC3339,
}

// parseDateString 按格式列表依次尝试解析日期字符串，全部格式均失败时返回错误。
func parseDateString(dateStr string) (time.Time, error) {
	for _, format := range jsonDateFormats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			return parsedTime, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}

// JSONDate 包装 time.Time，在 JSON 序列化时使用 YYYY-MM-DD HH:mm:ss 格式。
// 数据库存储仍使用标准的 datetime 格式。
type JSONDate struct {
	time.Time
}

// NewJSONDate 创建新的JSONDate。
func NewJSONDate(t time.Time) JSONDate {
	return JSONDate{Time: t}
}

// NowJSON 获取当前时间的JSONDate。
func NowJSON() JSONDate {
	return JSONDate{Time: time.Now()}
}

// ParseJSONDate 解析日期字符串为JSONDate。
func ParseJSONDate(dateStr string) (JSONDate, error) {
	parsedTime, err := parseDateString(dateStr)
	if err != nil {
		return JSONDate{}, fmt.Errorf("解析日期失败: %w", err)
	}
	return JSONDate{Time: parsedTime}, nil
}

// isMidnight 判断时间是否处于当天零点，用于决定序列化输出为日期还是日期时间。
func isMidnight(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
}

// MarshalJSON 实现 JSON 序列化方法，零值时间输出为 null。
// 当天零点的时间按纯日期输出，其余时间输出完整日期时间。
func (jd JSONDate) MarshalJSON() ([]byte, error) {
	if jd.Time.IsZero() {
		return []byte("null"), nil
	}

	outputFormat := JSONDateFormat
	if isMidnight(jd.Time) {
		outputFormat = JSONDateOnlyFormat
	}
	return []byte(`"` + jd.Time.Format(outputFormat) + `"`), nil
}

// UnmarshalJSON 实现 JSON 反序列化方法，空字符串与 null 均视为未设置。
func (jd *JSONDate) UnmarshalJSON(data []byte) error {
	// JSON 空值 null 视为未设置，保持零值时间。
	if string(data) == "null" {
		jd.Time = time.Time{}
		return nil
	}

	// 空字符串视为未设置，确保可选日期字段在前端留空时仍可提交。
	if string(data) == `""` {
		jd.Time = time.Time{}
		return nil
	}

	// 去掉引号。
	dateStr := string(data[1 : len(data)-1])

	parsedTime, err := parseDateString(dateStr)
	if err != nil {
		return err
	}
	jd.Time = parsedTime
	return nil
}

// String 实现 Stringer 接口，当天零点的时间按纯日期输出。
func (jd JSONDate) String() string {
	if jd.Time.IsZero() {
		return ""
	}

	outputFormat := JSONDateFormat
	if isMidnight(jd.Time) {
		outputFormat = JSONDateOnlyFormat
	}
	return jd.Time.Format(outputFormat)
}

// Value 实现 driver.Valuer 接口，数据库存储使用完整时间。
func (jd JSONDate) Value() (driver.Value, error) {
	if jd.Time.IsZero() {
		return nil, nil
	}
	// 返回完整的 time.Time，让数据库按标准格式存储。
	return jd.Time, nil
}

// Scan 实现 sql.Scanner 接口，从数据库读取完整时间。
func (jd *JSONDate) Scan(value interface{}) error {
	if value == nil {
		jd.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		jd.Time = v
		return nil
	case string:
		parsedTime, err := parseDateString(v)
		if err != nil {
			return fmt.Errorf("无法解析时间格式: %s", v)
		}
		jd.Time = parsedTime
		return nil
	default:
		return fmt.Errorf("无法将 %T 类型转换为 JSONDate", value)
	}
}
