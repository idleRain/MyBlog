// Package service 业务逻辑层
package service

import "errors"

// ErrPermissionDenied 权限不足的哨兵错误，handler 层经 errors.Is 识别后映射为 403。
// 业务侧的具体动作经 fmt.Errorf 以 %w 包装本错误追加，禁止再以裸 errors.New 产生权限错误。
var ErrPermissionDenied = errors.New("没有执行此操作的权限")

// ErrInvalidRequest 业务规则校验失败的哨兵错误，handler 层经 errors.Is 识别后映射为 400。
// binding 层无法覆盖的业务校验失败统一包装本错误，避免散落的私有映射分支。
var ErrInvalidRequest = errors.New("请求参数不合法")
