package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
	"github.com/fengjx/daox/v2/example/dao/schema/apporder"
	"github.com/fengjx/daox/v2/example/db"
)

var OrderDao *orderDao

func init() {
	OrderDao = newOrderDao()
}

type orderDao struct {
	*daox.Dao[*schema.AppOrder]
}

func newOrderDao() *orderDao {
	dao := daox.NewDao[*schema.AppOrder](
		apporder.Meta,
		daox.WithDBMaster(db.DB),
	)
	inst := &orderDao{
		Dao: dao,
	}
	return inst
}
