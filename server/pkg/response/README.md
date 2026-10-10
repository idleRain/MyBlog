# Response 模块

统一响应格式包。响应体为 `{code, message, data}`，`data` 为空时不出现在 JSON；业务响应统一使用 HTTP 200，结果由 `code` 区分（未匹配路由的 404 与就绪探针的 503 由 router 直接输出）。

## 预定义响应码

| 常量 | 值 | 语义 |
|---|---|---|
| `CodeSuccess` | 200 | 成功 |
| `CodeInvalid` | 400 | 请求参数错误 |
| `CodeAuth` | 401 | 认证失败 |
| `CodeForbid` | 403 | 权限不足 |
| `CodeNotFound` | 404 | 接口或资源不存在 |
| `CodeError` | 500 | 服务器内部错误 |

## 方法

- `Success(c, data)`、`SuccessWithMessage(c, message, data)`：成功响应，前者使用默认文案「操作成功」。
- `Error(c, code, message)`：按指定业务码返回。
- `BadRequest`、`Unauthorized`、`Forbidden`、`NotFound`、`InternalError`：分别对应 400、401、403、404、500 的便捷封装。

## 约定

- 具体错误信息放在 `message` 字段，前端直接展示；`Response` 无 `error` 字段。
- handler 不把内部错误原文写入 `message`：service 错误统一经 `handler.HandleServiceError` 分档，未分类错误脱敏为「服务器内部错误」。
- 前端响应类型唯一来源为 `@myblog/shared` 的 `ApiResponse`；响应码常量见 `AGENTS.md` 第 5 节「统一响应」。
