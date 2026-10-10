# 字典 API 文档

## 概述

字典是**读多写少的配置数据**，用于承载可动态管理的状态类枚举：字典类型（`DictType`）按业务语义聚合一组字典项（`DictItem`），业务侧按**字典码**取值，不硬编码枚举字面量。

- **公开读取**：`/api/dicts/*` 只输出**已生效**（`status = 1`）的类型与字典项，无需认证，供前台与 app 端取用。
- **管理端维护**：`/api/admin/dicts/*` 需要 `dict:manage` 权限。
- **不变式**：字典码与字典项值在创建后**不可变更**（更新接口不接受这两个字段），保证业务侧按码取值的契约稳定；字典为硬删除，删除类型时级联清理字典项与翻译行。
- **多语言**：类型与字典项均支持内容多语言，规则与文章/分类/标签一致，见 [`contracts/i18n-protocol.md`](../../../contracts/i18n-protocol.md)。

## 数据形状

### 字典类型

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | integer | 字典类型 ID |
| code | string | 字典码，业务侧唯一定位标识，如 `tag_status` |
| name | string | 类型名称，按 `Accept-Language` 本地化 |
| description | string | 类型描述，按 `Accept-Language` 本地化 |
| status | integer | 生效状态：`1` 生效、`0` 停用 |
| sortOrder | integer | 排序权重，数值小的靠前 |
| extra | object | 预留扩展字段，存放颜色、图标等展示元数据 |
| createdAt / updatedAt | string | 创建与更新时间，`datetime(3)` |
| translations | array | 翻译行数组，仅 `Accept-Language: *` 时返回 |

### 字典项

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | integer | 字典项 ID |
| typeId | integer | 所属字典类型 ID |
| value | string | 字典项值，同一类型内唯一 |
| label | string | 显示名，按 `Accept-Language` 本地化 |
| description | string | 描述，按 `Accept-Language` 本地化 |
| status | integer | 生效状态：`1` 生效、`0` 停用 |
| sortOrder | integer | 排序权重，数值小的靠前 |
| extra | object | 预留扩展字段 |
| createdAt / updatedAt | string | 创建与更新时间，`datetime(3)` |
| translations | array | 翻译行数组，仅 `Accept-Language: *` 时返回 |

## 公开接口

### 1. 全量已生效字典

返回全部已生效字典类型及其已生效字典项，按类型的排序权重组织。

#### 请求信息

- **接口地址**: `/api/dicts/all`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/dicts/all
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200 表示成功 |
| message | string | 是 | 响应消息 |
| data | object | 是 | 响应数据 |
| data.dicts | array | 是 | 已生效字典分组集合，元素为字典类型字段与 `items` 平铺的同一层级对象 |

每个分组的 `items` 为该类型下的已生效字典项；类型下没有已生效字典项时 `items` 为**空数组**，前端可直接迭代。响应头 `Content-Language` 标注实际输出语言。

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "dicts": [
      {
        "id": 1,
        "code": "tag_status",
        "name": "标签状态",
        "description": "标签的启用状态",
        "status": 1,
        "sortOrder": 0,
        "extra": null,
        "createdAt": "2026-01-01 10:00:00",
        "updatedAt": "2026-01-01 10:00:00",
        "items": [
          {
            "id": 1,
            "typeId": 1,
            "value": "enabled",
            "label": "启用",
            "description": "",
            "status": 1,
            "sortOrder": 0,
            "extra": null,
            "createdAt": "2026-01-01 10:00:00",
            "updatedAt": "2026-01-01 10:00:00"
          }
        ]
      }
    ]
  }
}
```

### 2. 按字典码查询单个字典

#### 请求信息

- **接口地址**: `/api/dicts/:type`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **路径参数**: `type` 为字典码

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/dicts/tag_status
```

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "code": "tag_status",
    "name": "标签状态",
    "description": "标签的启用状态",
    "status": 1,
    "sortOrder": 0,
    "extra": null,
    "createdAt": "2026-01-01 10:00:00",
    "updatedAt": "2026-01-01 10:00:00",
    "items": []
  }
}
```

#### 错误响应

类型不存在或已停用时统一返回 `字典类型不存在`，不区分两种情况，避免经错误信息探测未生效的字典码。

| 业务码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 字典码不能为空 | 路径参数为空 |
| 404 | 字典类型不存在 | 字典码未登记或该类型已停用 |

## 管理端接口

以下接口均在 `/api/admin/dicts` 分组注册，统一要求 `dict:manage` 权限；权限不足返回业务码 403。

### 3. 创建字典类型

- **接口地址**: `/api/admin/dicts/types/create`

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| code | string | 是 | 字典码 | 长度 2–50，须匹配 `^[a-z][a-z0-9_]*$`，全局唯一 |
| name | string | 是 | 类型名称 | 长度 1–50 |
| description | string | 否 | 类型描述 | 最大 200 |
| status | integer | 否 | 生效状态 | `0` 或 `1`，缺省 `1` |
| sortOrder | integer | 否 | 排序权重 | 0–9999，缺省 `0` |
| extra | object | 否 | 扩展字段 | 必须是合法 JSON |
| i18n | object | 否 | 翻译补丁，键为语言标识 | 见多语言契约 |

### 4. 更新字典类型

- **接口地址**: `/api/admin/dicts/types/update`
- **请求参数**: `id` 必填；`name`、`description`、`status`、`sortOrder`、`extra`、`i18n` 均可选，未提供的字段保留原值。
- **不接受的字段**: `code`，字典码创建后不可变更。
- **翻译合并语义**: `i18n` 以既有翻译行为底合并，提供即更新。

### 5. 删除字典类型

- **接口地址**: `/api/admin/dicts/types/delete`
- **请求参数**: `id` 必填。
- **级联语义**: 删除类型时在同一事务内清理其字典项与全部翻译行。

### 6. 字典类型列表

- **接口地址**: `/api/admin/dicts/types/list`

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| page | integer | 否 | 页码 | 最小 1，缺省 1 |
| pageSize | integer | 否 | 每页条数 | 1–100，缺省 10 |
| status | integer | 否 | 按生效状态过滤 | `0` 或 `1`，缺省不过滤 |
| search | string | 否 | 关键词，匹配字典码与名称 | — |

响应 `data` 为 `{ types, total, page, pageSize }`。

### 7. 创建字典项

- **接口地址**: `/api/admin/dicts/items/create`

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| typeId | integer | 是 | 所属字典类型 ID | 类型必须存在，停用类型仍允许维护字典项 |
| value | string | 是 | 字典项值 | 最大 50，同类型内唯一 |
| label | string | 是 | 显示名 | 长度 1–50 |
| description | string | 否 | 描述 | 最大 200 |
| status | integer | 否 | 生效状态 | `0` 或 `1`，缺省 `1` |
| sortOrder | integer | 否 | 排序权重 | 0–9999，缺省 `0` |
| extra | object | 否 | 扩展字段 | 必须是合法 JSON |
| i18n | object | 否 | 翻译补丁 | 见多语言契约 |

### 8. 更新字典项

- **接口地址**: `/api/admin/dicts/items/update`
- **请求参数**: `id` 必填；`label`、`description`、`status`、`sortOrder`、`extra`、`i18n` 均可选。
- **不接受的字段**: `typeId` 与 `value`，字典项值创建后不可变更，跨类型搬迁同样不支持。

### 9. 删除字典项

- **接口地址**: `/api/admin/dicts/items/delete`
- **请求参数**: `id` 必填。
- **级联语义**: 在同一事务内清理该字典项的翻译行。

### 10. 字典项列表

- **接口地址**: `/api/admin/dicts/items/list`

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| typeId | integer | 是 | 所属字典类型 ID | 必填 |
| page | integer | 否 | 页码 | 最小 1，缺省 1 |
| pageSize | integer | 否 | 每页条数 | 1–100，缺省 10 |
| status | integer | 否 | 按生效状态过滤 | `0` 或 `1`，缺省不过滤 |
| search | string | 否 | 关键词，匹配字典项值与显示名 | — |

响应 `data` 为 `{ items, total, page, pageSize }`。

## 错误响应

业务错误经 `pkg/response` 信封返回，**HTTP 状态码恒为 200**，错误语义只体现在响应体 `code` 字段。

| 业务码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 请求参数错误: {具体原因} | `binding` 校验失败，含字段名与约束 |
| 400 | 字典码需以小写字母开头，仅含小写字母、数字与下划线 | 字典码格式不符 |
| 400 | 字典码已存在 | 创建或更新时字典码被其他类型占用 |
| 400 | 字典项值已存在 | 同类型内字典项值重复 |
| 400 | 扩展字段必须是合法 JSON | `extra` 不是合法 JSON |
| 401 | 未提供认证令牌 / 无效的认证令牌 / 用户不存在 | 管理端接口缺少有效访问令牌 |
| 403 | 权限不足，无法访问该资源 | 当前角色不含 `dict:manage` |
| 403 | 用户已被禁用 / 用户角色无效 | 令牌有效但账号状态或角色不合法 |
| 404 | 字典类型不存在 / 字典项不存在 | 目标资源不存在 |

## 多语言说明

- 请求头 `Accept-Language` 决定 `name`/`description`/`label` 的输出语言，缺省中文；白名单外的语言回退中文，响应头 `Content-Language` 标注实际输出语言。
- `Accept-Language: *` 返回全量翻译行数组 `translations`，供管理端编辑使用；指定语言时该字段不返回。
- 创建与更新请求体的 `i18n` 键为语言标识，缺省语言键被拒绝并返回 400。

## 权限说明

| 能力 | 权限标识 |
|------|----------|
| 读取公开字典 | 无需权限 |
| 管理字典类型与字典项 | `dict:manage` |

`dict:manage` 的持有角色由后端 `configs/config.yaml` 的 `rbac` 节唯一权威定义，前端不得复刻该映射。
