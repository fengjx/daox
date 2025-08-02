package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
	"github.com/fengjx/daox/v2/example/dao/schema/apporder"
)

// orderDao 自动生成，但可以修改扩展其他方法

var OrderDao *orderDao

func init() {
	OrderDao = newOrderDao()
	if OrderDao != nil {
		OrderDao.Dao.RelFiller = OrderDao.RelFiller
	}
}

type orderDao struct {
	*daox.Dao[*schema.AppOrder]
}

func newOrderDao() *orderDao {
	dao := daox.NewDao[*schema.AppOrder](
		apporder.Meta,
	)
	inst := &orderDao{
		Dao: dao,
	}
	dao.RelFiller = inst.RelFiller
	return inst
}
