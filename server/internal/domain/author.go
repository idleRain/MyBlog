// Package domain 全系统唯一类型语言层。
package domain

// AuthorPublic 作者公开信息窄化视图。
// 公开端点的响应不得直接序列化 User 实体，须经本视图输出白名单字段，
// 防止 email 等个人信息随文章作者与评论者关联泄露，见债务 D15。
type AuthorPublic struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Website  string `json:"website"`
}

// NewAuthorPublic 从用户实体构建公开作者视图，用户为空时返回空值。
func NewAuthorPublic(user *User) *AuthorPublic {
	if user == nil {
		return nil
	}
	return &AuthorPublic{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Bio:      user.Bio,
		Website:  user.Website,
	}
}
