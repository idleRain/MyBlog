// 后端统一响应业务码，与 server/pkg/response 的预定义响应码一一对应。
// 常量名与后端 CodeXxx 的对应关系在 packages/api 的契约锚定测试中登记，改动须两端同步。

export const RESPONSE_CODE_SUCCESS = 200
export const RESPONSE_CODE_ERROR = 500
export const RESPONSE_CODE_INVALID = 400
export const RESPONSE_CODE_UNAUTHORIZED = 401
export const RESPONSE_CODE_FORBIDDEN = 403
export const RESPONSE_CODE_NOT_FOUND = 404
