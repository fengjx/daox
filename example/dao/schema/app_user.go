package schema

import (
	"time"

	"github.com/fengjx/daox/v2"
)

// AppUser 用户信息表
type AppUser struct {
	ID       int64     `json:"id"`       // -
	Username string    `json:"username"` // 用户名
	Pwd      string    `json:"pwd"`      // 密码
	Salt     string    `json:"salt"`     // 密码盐
	Email    string    `json:"email"`    // 邮箱
	Nickname string    `json:"nickname"` // 昵称
	Avatar   string    `json:"avatar"`   // 头像
	Phone    string    `json:"phone"`    // 手机号
	Status   string    `json:"status"`   // 用户状态
	Remark   string    `json:"remark"`   // 备注
	Utime    time.Time `json:"utime"`    // 更新时间
	Ctime    time.Time `json:"ctime"`    // 创建时间
	// relations
	Orders []*AppOrder `json:"orders"`
	Card   *AppCard    `json:"card"`
}

func (m *AppUser) GetID() any {
	return m.ID
}

func (m *AppUser) New() daox.Model {
	return &AppUser{}
}
