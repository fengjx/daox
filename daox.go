package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"reflect"
)

type Dao struct {
	DBMaster  *sqlx.DB
	DBRead    *sqlx.DB
	Redis     *redis.Client
	TableMeta TableMeta
}

func Create(tableName, t reflect.Type, db *sqlx.DB, opts ...Option) *Dao {
	dao := &Dao{
		DBMaster: db,
	}
	for _, opt := range opts {
		opt(dao)
	}
	return dao
}
