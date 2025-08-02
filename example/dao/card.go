package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
	"github.com/fengjx/daox/v2/example/dao/schema/appcard"
)

// cardDao 自动生成，但可以修改扩展其他方法

var CardDao *cardDao

func init() {
	CardDao = newCardDao()
	if CardDao != nil {
		CardDao.Dao.RelFiller = CardDao.RelFiller
	}
}

type cardDao struct {
	*daox.Dao[*schema.AppCard]
}

func newCardDao() *cardDao {
	dao := daox.NewDao[*schema.AppCard](
		appcard.Meta,
	)
	inst := &cardDao{
		Dao: dao,
	}
	dao.RelFiller = inst.RelFiller
	return inst
}
