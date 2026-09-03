package domain

import (
	"MyBlog/pkg/datetime"
)

// CreateUserRequest 创建用户请求，承载参数校验。
// binding 校验 tag 在 handler 层 DTO 分离完成后剥离，见契约 C1。
type CreateUserRequest struct {
	Username string            `json:"username" binding:"required,min=1,max=50"`
	Email    string            `json:"email" binding:"required,email"`
	Password string            `json:"password" binding:"required,min=6,max=100"`
	Nickname string            `json:"nickname" binding:"max=50"`
	Role     string            `json:"role" binding:"omitempty,oneof=user editor admin superadmin"`
	Birthday datetime.JSONDate `json:"birthday" binding:"omitempty"`
}

// UpdateUserRequest 更新用户请求，承载参数校验。
type UpdateUserRequest struct {
	ID       uint              `json:"id" binding:"required"`
	Username string            `json:"username" binding:"required,min=1,max=50"`
	Email    string            `json:"email" binding:"required,email"`
	Password string            `json:"password" binding:"omitempty,min=6,max=100"` // 可选，留空则不更新
	Nickname string            `json:"nickname" binding:"max=50"`
	Role     string            `json:"role" binding:"omitempty,oneof=user editor admin superadmin"`
	Birthday datetime.JSONDate `json:"birthday" binding:"omitempty"`
	Status   int               `json:"status" binding:"omitempty,oneof=0 1"` // 用户状态，可选
}
