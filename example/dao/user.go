package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
	"github.com/fengjx/daox/v2/example/dao/schema/appuser"
)

// userDao 自动生成，但可以修改扩展其他方法

var UserDao *userDao

func init() {
	UserDao = newUserDao()
	if UserDao != nil {
		UserDao.Dao.RelFiller = UserDao.RelFiller
	}
}

type userDao struct {
	*daox.Dao[*schema.AppUser]
}

func newUserDao() *userDao {
	dao := daox.NewDao[*schema.AppUser](
		appuser.Meta,
	)
	inst := &userDao{
		Dao: dao,
	}
	dao.RelFiller = inst.RelFiller
	return inst
}
