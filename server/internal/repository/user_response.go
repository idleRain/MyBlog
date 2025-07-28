package repository

import (
	"MyBlog/pkg/datetime"
)

// UserResponse 用户响应结构体 - 专门用于API响应
type UserResponse struct {
	ID        uint              `json:"id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Nickname  string            `json:"nickname"`
	Avatar    string            `json:"avatar"`
	Birthday  datetime.JSONDate `json:"birthday"`
	Role      string            `json:"role"`
	Status    int               `json:"status"`
	CreatedAt datetime.JSONDate `json:"createdAt"`
	UpdatedAt datetime.JSONDate `json:"updatedAt"`
}

// ToResponse 将 User 模型转换为响应格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Birthday:  u.Birthday,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: datetime.NewJSONDate(u.CreatedAt),
		UpdatedAt: datetime.NewJSONDate(u.UpdatedAt),
	}
}

// ToResponseList 将用户列表转换为响应格式
func ToResponseList(users []*User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}
	return responses
}
