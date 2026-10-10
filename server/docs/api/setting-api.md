# 系统设置 API 文档

## 概述

系统设置模块提供站点的键值化全局配置管理，支持公开项读取与后台批量更新。敏感配置项（如密码、密钥）输出时自动脱敏。

## 配置说明

- 配置以键值对存储，键名使用点分命名空间，如 `site_name`、`seo_title`
- 公开项（`isPublic=true`）对前端可见，私有项仅管理端可见；公开列表按 `sortOrder ASC, id ASC` 排列，管理端列表按 `groupName ASC, sortOrder ASC, id ASC` 排列
- 只读项（`isReadonly=true`）禁止通过接口修改
- 敏感项（`isSensitive=true` 或键名含 `password`/`secret`/`key`/`token`，大小写不敏感）输出为 `********`；原值为空串时输出空串

### 设置键的生效口径

接口只负责键值的存储与读取，不保证任一设置键对业务产生运行时效果。当前被后端业务逻辑实际消费的键仅两个：

- `allow_guest_comment`：关闭后游客评论通道被拒绝（`internal/service/comment.go`）
- `comment_auto_approve`：开启后新评论直接进入已审核状态（同上）

其余键仅为存储值，不产生运行时效果；安全分组（`enable_rate_limit`、`session_timeout` 等）的真实生效渠道是 `server/configs/config.yaml` 的 `security` 节与安全中间件，不是本接口。后台设置页以 `apps/admin/src/lib/constants/setting.ts` 的 `EFFECTIVE_SETTING_KEYS` 标注未生效项。

## 权限说明

| 操作 | 所需权限 | 角色要求 |
|------|----------|----------|
| 读取公开设置 | 无 | 无需认证 |
| 设置项列表 / 批量更新 | `system:config` | 仅 superadmin 持有（`server/configs/config.yaml` 的 `rbac` 节中 admin 角色不含该权限） |

## 公开接口（无需认证）

### 1. 获取公开设置

返回全部 `isPublic=true` 的设置项，敏感项已脱敏。

#### 请求信息

- **接口地址**: `/api/settings/public`
- **请求方式**: `POST`
- **权限要求**: 无需认证
- **Content-Type**: `application/json`

#### 请求参数

无需参数，请求体不参与解析。

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
| data.settings | array | 是 | 公开设置项列表，按 `sortOrder ASC, id ASC` 排列 |
| data.settings[].id | integer | 是 | 设置ID |
| data.settings[].keyName | string | 是 | 配置键名 |
| data.settings[].label | string | 是 | 设置项显示名称，无值时为空串 |
| data.settings[].value | string | 是 | 配置值，敏感项输出掩码 |
| data.settings[].defaultValue | string | 是 | 默认值 |
| data.settings[].description | string | 是 | 配置描述 |
| data.settings[].type | string | 是 | 值类型：string/number/boolean/json/array |
| data.settings[].groupName | string | 是 | 配置分组 |
| data.settings[].isPublic | boolean | 是 | 是否公开，公开接口恒为 true |
| data.settings[].isReadonly | boolean | 是 | 是否只读 |
| data.settings[].isSensitive | boolean | 是 | 是否敏感项 |
| data.settings[].validationRule | string | 是 | 校验规则说明，当前无消费方 |
| data.settings[].sortOrder | integer | 是 | 排序权重 |
| data.settings[].updatedBy | integer | 否 | 最后更新用户ID |
| data.settings[].createdAt | string | 是 | 创建时间 |
| data.settings[].updatedAt | string | 是 | 更新时间 |

> 脱敏只替换 `value`，键名、分组等元信息照常输出；`updatedByUser` 关联未预加载，响应中省略该键。

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
        "value": "闲雨小筑",
        "type": "string",
        "groupName": "general"
      }
    ]
  }
}
```

## 管理接口（需要 system:config 权限）

管理接口需携带有效访问令牌：`Authorization: Bearer {accessToken}` 头优先，其次读取会话 Cookie `mb_access_token`；所需权限 `system:config` 在 `config.yaml` 中仅 superadmin 持有。

### 2. 设置项列表

返回全部设置项（含私有项），敏感项已脱敏。

#### 请求信息

- **接口地址**: `/api/admin/settings/list`
- **请求方式**: `POST`
- **权限要求**: `system:config`（仅 superadmin）
- **Content-Type**: `application/json`

#### 请求参数

无需参数，请求体不参与解析。

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/settings/list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{}'
```

#### 响应参数

响应结构与公开设置接口一致，字段为设置实体全量字段，区别在于：

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| data.settings | array | 是 | 全部设置项，按 `groupName ASC, sortOrder ASC, id ASC` 排列 |
| data.settings[].isPublic | boolean | 是 | 是否公开 |
| data.settings[].isReadonly | boolean | 是 | 是否只读 |
| data.settings[].isSensitive | boolean | 是 | 是否敏感项 |
| data.settings[].updatedBy | integer | 否 | 最后更新用户ID |

### 3. 批量更新设置

按 `items` 逐项校验后，在单个事务内写回；任一项不存在或为只读即整批拒绝，不产生部分写入。

#### 请求信息

- **接口地址**: `/api/admin/settings/update`
- **请求方式**: `POST`
- **权限要求**: `system:config`（仅 superadmin）
- **Content-Type**: `application/json`

#### 请求参数

| 字段名 | 类型 | 必填 | 说明 | 验证规则 |
|--------|------|------|------|----------|
| items | array | 是 | 待更新的设置项数组 | 至少1项，逐项校验 |
| items[].keyName | string | 是 | 配置键名 | 必填，最大100字符 |
| items[].value | string | 是 | 配置值 | 无校验，可为空串 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/admin/settings/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {accessToken}" \
  -d '{
    "items": [
      {
        "keyName": "site_name",
        "value": "闲雨小筑 新站名"
      }
    ]
  }'
```

#### 响应示例

`data.settings` 为更新后的全部设置项，等同再次调用设置项列表的结果，含未提交的条目且敏感项已脱敏，不是仅提交的条目；示例仅摘录其中一项。

```json
{
  "code": 200,
  "message": "设置更新成功",
  "data": {
    "settings": [
      {
        "keyName": "site_name",
        "value": "闲雨小筑 新站名",
        "type": "string"
      }
    ]
  }
}
```

#### 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数校验失败（`items` 缺失或为空、`items[].keyName` 缺失） |
| 403 | 缺少 `system:config` 权限 |
| 500 | 设置项不存在、设置为只读不可修改（服务层返回未包装哨兵的普通错误，落入统一内部错误路径） |
