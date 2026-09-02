# 系统设置 API 文档

## 概述

系统设置模块提供站点的键值化全局配置管理，支持公开项读取与后台批量更新。敏感配置项（如密码、密钥）输出时自动脱敏。

## 配置说明

- 配置以键值对存储，键名使用点分命名空间，如 `site_name`、`seo_title`
- 公开项（`isPublic=true`）对前端可见，私有项仅管理端可见
- 只读项（`isReadonly=true`）禁止通过接口修改
- 敏感项（`isSensitive=true` 或键名含 password/secret/key/token）输出为 `********`

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 读取公开设置 | 无 | 无 |
| 设置项列表 / 批量更新 | `system:config` | admin及以上 |

## 公开接口（无需认证）

### 1. 获取公开设置

返回全部公开设置项，敏感项已脱敏。

#### 请求信息

- **接口地址**: `/api/settings/public`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/settings/public \
  -H "Content-Type: application/json" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | integer | 是 | 状态码，200表示成功 |
| data.settings | array | 是 | 公开设置项列表 |
| data.settings[].keyName | string | 是 | 配置键名 |
| data.settings[].label | string | 否 | 设置项显示名称 |
| data.settings[].value | string | 是 | 配置值，敏感项输出掩码 |
| data.settings[].type | string | 是 | 值类型：string/number/boolean/json/array |
| data.settings[].groupName | string | 是 | 配置分组 |

#### 响应示例

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "settings": [
      {
        "keyName": "site_name",
        "label": "站点名称",
        "value": "MyBlog",
        "type": "string",
        "groupName": "general"
      }
    ]
  }
}
```

## 管理接口（需要 system:config 权限）

管理接口需在请求头携带 `Authorization: Bearer {accessToken}`，操作者角色为 admin 及以上。

### 2. 设置项列表

返回全部设置项（含私有项），敏感项已脱敏。

#### 请求信息

- **接口地址**: `/api/admin/settings/list`
- **请求方式**: `POST`
- **权限要求**: `system:config`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

无需参数。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/settings/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{}'
```

#### 响应参数

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| data.settings | array | 是 | 全部设置项，按分组与排序权重排列 |
| data.settings[].isPublic | boolean | 是 | 是否公开 |
| data.settings[].isReadonly | boolean | 是 | 是否只读 |
| data.settings[].isSensitive | boolean | 是 | 是否敏感项 |
| data.settings[].updatedBy | integer | 否 | 最后更新用户ID |

### 3. 批量更新设置

#### 请求信息

- **接口地址**: `/api/admin/settings/update`
- **请求方式**: `POST`
- **权限要求**: `system:config`（admin及以上）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| items | array | 是 | 待更新的设置项数组 | 至少1项 |
| items[].keyName | string | 是 | 配置键名 | 最大100字符 |
| items[].value | string | 是 | 配置值 | 字符串 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/settings/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "items": [
      {
        "keyName": "site_name",
        "value": "MyBlog 新站名"
      }
    ]
  }'
```

#### 响应示例

```json
{
  "code": 200,
  "message": "设置更新成功",
  "data": {
    "settings": [
      {
        "keyName": "site_name",
        "value": "MyBlog 新站名",
        "type": "string"
      }
    ]
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 更新项为空、设置项不存在、设置为只读不可修改 |
