// Package datetime 提供JSON日期格式化
package datetime

import (
  "database/sql/driver"
  "fmt"
  "time"
)

const JSONDateFormat = "2006-01-02 15:04:05"

// JSONDate 包装 time.Time，只在JSON序列化时使用 YYYY-MM-DD HH:mm:ss 格式
// 数据库存储仍使用标准的 datetime 格式
type JSONDate struct {
  time.Time
}

// NewJSONDate 创建新的JSONDate
func NewJSONDate(t time.Time) JSONDate {
  return JSONDate{Time: t}
}

// NowJSON 获取当前时间的JSONDate
func NowJSON() JSONDate {
  return JSONDate{Time: time.Now()}
}

// ParseJSONDate 解析日期字符串为JSONDate
func ParseJSONDate(dateStr string) (JSONDate, error) {
  t, err := time.Parse(JSONDateFormat, dateStr)
  if err != nil {
    return JSONDate{}, fmt.Errorf("解析日期失败: %w", err)
  }
  return JSONDate{Time: t}, nil
}

// MarshalJSON 实现 JSON 序列化 - 只影响API响应格式
func (jd JSONDate) MarshalJSON() ([]byte, error) {
  if jd.Time.IsZero() {
    return []byte("null"), nil
  }
  return []byte(`"` + jd.Time.Format(JSONDateFormat) + `"`), nil
}

// UnmarshalJSON 实现 JSON 反序列化 - 只影响API请求解析
func (jd *JSONDate) UnmarshalJSON(data []byte) error {
  if string(data) == "null" {
    jd.Time = time.Time{}
    return nil
  }

  // 去掉引号
  dateStr := string(data[1 : len(data)-1])

  // 尝试解析 YYYY-MM-DD 格式
  if t, err := time.Parse(JSONDateFormat, dateStr); err == nil {
    jd.Time = t
    return nil
  }

  // 如果上面失败，尝试解析完整时间格式（兼容性）
  if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
    jd.Time = t
    return nil
  }

  return fmt.Errorf("无法解析日期格式: %s", dateStr)
}

// String 实现 Stringer 接口
func (jd JSONDate) String() string {
  if jd.Time.IsZero() {
    return ""
  }
  return jd.Time.Format(JSONDateFormat)
}

// Value 实现 driver.Valuer 接口 - 数据库存储使用完整时间
func (jd JSONDate) Value() (driver.Value, error) {
  if jd.Time.IsZero() {
    return nil, nil
  }
  // 返回完整的 time.Time，让数据库按标准格式存储
  return jd.Time, nil
}

// Scan 实现 sql.Scanner 接口 - 从数据库读取完整时间
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
    // 尝试解析各种可能的时间格式
    formats := []string{
      time.RFC3339,
      "2006-01-02 15:04:05",
      JSONDateFormat,
    }

    for _, format := range formats {
      if t, err := time.Parse(format, v); err == nil {
        jd.Time = t
        return nil
      }
    }
    return fmt.Errorf("无法解析时间格式: %s", v)
  default:
    return fmt.Errorf("无法将 %T 类型转换为 JSONDate", value)
  }
}
