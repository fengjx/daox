package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
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

func WithMapper(Mapper *reflectx.Mapper) Option {
	return func(d *Dao) {
		d.Mapper = Mapper
	}
}

func WithTableName(tableName string) Option {
	return func(d *Dao) {
		d.TableMeta.TableName = tableName
	}
}
