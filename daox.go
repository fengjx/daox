package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Dao struct {
	DBMaster *sqlx.DB
	DBRead   *sqlx.DB
	Redis    *redis.Client
}

func Create(db *sqlx.DB, opts ...Option) *Dao {
	dao := &Dao{
		DBMaster: db,
	}
	for _, opt := range opts {
		opt(dao)
	}
	return dao
}
