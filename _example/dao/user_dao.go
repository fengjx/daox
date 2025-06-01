package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/_example/dao/schema"
	"github.com/fengjx/daox/v2/_example/dao/schema/sysuser"
)

var SysUserDAO *sysUserDAO

func init() {
	SysUserDAO = newUserDAO()
}

type sysUserDAO struct {
	*daox.Dao[*schema.SysUser]
}

func newUserDAO() *sysUserDAO {
	dao := daox.NewDao[*schema.SysUser](sysuser.Meta)
	return &sysUserDAO{
		Dao: dao,
	}
}
