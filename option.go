package daox

import (
	"github.com/jmoiron/sqlx"
)

type Option func(*Dao)

func WithDBRead(read *sqlx.DB) Option {
	return func(p *Dao) {
		p.DBRead = NewDB(read)
	}
}

func IsAutoIncrement() Option {
	return func(dao *Dao) {
		dao.TableMeta.IsAutoIncrement = true
	}
}
