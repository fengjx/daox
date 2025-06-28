package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
	"github.com/fengjx/daox/v2/example/dao/schema/appuser"
	"github.com/fengjx/daox/v2/example/db"
)

var UserDao *userDao

func init() {
	UserDao = newUserDao()
}

type userDao struct {
	*daox.Dao[*schema.AppUser]
}

func newUserDao() *userDao {
	inst := &userDao{}
	dao := daox.NewDao[*schema.AppUser](
		appuser.Meta,
		daox.WithDBMaster(db.DB),
	)
	dao.RelFiller = inst.RelFiller
	inst.Dao = dao
	return inst
}
