# DateTime 包

`JSONDate` 包装 `time.Time`，数据库按完整 datetime 存储，只在 JSON 输出时收敛格式。

## 行为

- `MarshalJSON`：零值输出 `null`；当天零点输出 `2006-01-02`，其余时刻输出 `2006-01-02 15:04:05`。输出并非固定日期粒度。
- `UnmarshalJSON`：`null` 与 `""` 视为未设置；解析按 `2006-01-02`、`2006-01-02 15:04:05`、`time.RFC3339` 顺序逐个尝试，全部失败时返回错误。
- `String`：与序列化同规则，零值返回空串。
- `Value`、`Scan`：实现 `driver.Valuer` 与 `sql.Scanner`，落库与读库均使用完整时间，查询与排序能力不受输出格式影响。

## 构造

- `NowJSON()`：当前时间的 `JSONDate`。
- `NewJSONDate(t time.Time)`：从 `time.Time` 构造。
- `ParseJSONDate(dateStr string)`：解析字符串，失败返回错误。

## 用法

- 按天粒度的字段用 `JSONDate` 并声明 `gorm:"type:date"`，如 `domain.User.Birthday`，输出为 `YYYY-MM-DD`。
- 需要纳秒精度或 RFC3339 输出时用 `time.Time`；实体字段（如 `domain.User.CreatedAt`）转对外响应时可经 `datetime.NewJSONDate` 收敛，见 `internal/domain/user_response.go`。
