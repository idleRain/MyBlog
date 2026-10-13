// Package service 业务逻辑层
package service

import "errors"

// ErrPermissionDenied 权限不足的哨兵错误，handler 层经 errors.Is 识别后映射为 403。
// 业务侧的具体动作经 fmt.Errorf 以 %w 包装本错误追加，禁止再以裸 errors.New 产生权限错误。
var ErrPermissionDenied = errors.New("没有执行此操作的权限")

// ErrUnauthenticated 认证失败的哨兵错误，handler 层经 errors.Is 识别后映射为 401。
// 登录凭据错误、账号状态异常与令牌校验失败统一包装本错误，
// 使认证失败只经 handler 的唯一映射点分档，不再由各入口的私有分支产出 401。
var ErrUnauthenticated = errors.New("认证失败")

// ErrInvalidRequest 业务规则校验失败的哨兵错误，handler 层经 errors.Is 识别后映射为 400。
// binding 层无法覆盖的业务校验失败统一包装本错误，避免散落的私有映射分支。
var ErrInvalidRequest = errors.New("请求参数不合法")

// ErrUsernameTaken 用户名已被占用的哨兵错误，handler 层经 errors.Is 识别后映射为 400。
var ErrUsernameTaken = errors.New("用户名已存在")

// ErrEmailTaken 邮箱已被其他账号占用的哨兵错误，handler 层经 errors.Is 识别后映射为 400。
var ErrEmailTaken = errors.New("邮箱已被其他用户使用")
