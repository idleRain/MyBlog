# DateTime 包

自定义日期时间类型，用于JSON序列化格式化，不影响数据库存储。

## 设计原则

- **数据库存储**: 保持标准的 datetime/timestamp 格式
- **JSON序列化**: 在API响应中转换为 YYYY-MM-DD 格式
- **兼容性**: 完全兼容标准 time.Time 类型

## JSONDate 类型

### 核心特性

- 包装 `time.Time`，不改变其本质
- 数据库存储使用完整的日期时间格式
- JSON序列化时只显示 YYYY-MM-DD 格式
- 支持多种输入格式的解析

### 使用示例

#### 模型定义

```go
type User struct {
    ID        uint              `json:"id" gorm:"primaryKey"`
    Username  string            `json:"username"`
    CreatedAt datetime.JSONDate `json:"createdAt"`  // 数据库: datetime, JSON: "2024-01-01"
    UpdatedAt datetime.JSONDate `json:"updatedAt"`  // 数据库: datetime, JSON: "2024-01-01"
}
```

#### 创建实例

```go
import "MyBlog/pkg/datetime"

// 当前时间
now := datetime.NowJSON()

// 从 time.Time 转换
date := datetime.NewJSONDate(time.Now())

// 解析字符串
date, err := datetime.ParseJSONDate("2024-01-01")
```

#### JSON 输入/输出

```go
// 输入 (支持多种格式)
{
    "birthday": "2024-01-01",           // YYYY-MM-DD
    "createdAt": "2024-01-01T10:00:00Z" // ISO格式(兼容)
}

// 输出 (统一格式)
{
    "birthday": "2024-01-01",
    "createdAt": "2024-01-01",
    "updatedAt": "2024-01-01"
}
```

#### 数据库存储

```sql
-- 实际存储在数据库中的格式
CREATE TABLE users (
    id INT PRIMARY KEY,
    username VARCHAR(50),
    created_at DATETIME,    -- 2024-01-01 10:30:45
    updated_at DATETIME     -- 2024-01-01 10:30:45
);
```

### 方法说明

#### 构造方法

- `NowJSON()`: 获取当前时间的JSONDate
- `NewJSONDate(t time.Time)`: 从time.Time创建JSONDate
- `ParseJSONDate(dateStr string)`: 解析日期字符串

#### 序列化方法

- `MarshalJSON()`: JSON序列化为 "YYYY-MM-DD"
- `UnmarshalJSON()`: 支持多种格式的JSON反序列化
- `String()`: 字符串表示为 "YYYY-MM-DD"

#### 数据库方法

- `Value()`: 数据库存储时返回完整的time.Time
- `Scan()`: 从数据库读取时解析为time.Time

### 重要说明

**数据库存储格式**: JSONDate 在数据库中存储的是完整的 datetime 格式（如：2024-01-01 14:30:
45），只是在JSON序列化时显示为日期格式（如：2024-01-01）。

这样设计的好处：

1. 保留了时间的完整信息
2. 数据库查询和排序功能完整
3. 只在API响应时简化显示格式
4. 兼容现有的时间处理逻辑

### 最佳实践

1. **CreatedAt/UpdatedAt**: 使用JSONDate，数据库存完整时间，API显示日期
2. **Birthday等日期**: 使用JSONDate + gorm:"type:date"
3. **时间戳需求**: 如需要精确时间，请使用标准time.Time

### 注意事项

1. **数据库实际存储**: 仍然是完整的datetime格式
2. **时区处理**: 内部仍使用time.Time，支持时区操作
3. **精度**: 数据库存储保持完整精度，JSON显示只到日期
4. **兼容性**: 可以与现有time.Time代码无缝集成
