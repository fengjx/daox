package dao

import (
	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
	"github.com/fengjx/daox/v2/example/dao/schema/appcard"
	"github.com/fengjx/daox/v2/example/db"
)

var CardDao *cardDao

func init() {
	CardDao = newCardDao()
}

type cardDao struct {
	*daox.Dao[*schema.AppCard]
}

func newCardDao() *cardDao {
	dao := daox.NewDao[*schema.AppCard](
		appcard.Meta,
		daox.WithDBMaster(db.DB),
	)
	inst := &cardDao{
		Dao: dao,
	}
	return inst
}
