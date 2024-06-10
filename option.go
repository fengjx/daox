package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type Option func(*Dao)

// WithDBMaster 设置主库
func WithDBMaster(master *sqlx.DB) Option {
	return func(p *Dao) {
		p.masterDB = master
	}
}

// WithDBRead 设置从库
func WithDBRead(read *sqlx.DB) Option {
	return func(p *Dao) {
		p.readDB = read
	}
}

// IsAutoIncrement 是否自增主键
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

// WithIfNullVal 设置字段为null时的默认值
func WithIfNullVal(col string, val string) Option {
	return func(d *Dao) {
		d.initIfNullVal()
		d.ifNullVals[col] = val
	}
}

// WithIfNullVals 设置字段（多个）为null时的默认值
func WithIfNullVals(vals map[string]string) Option {
	return func(d *Dao) {
		d.initIfNullVal()
		for col, val := range vals {
			d.ifNullVals[col] = val
		}
	}
}

type InsertOptions struct {
	disableGlobalOmitColumns bool     // 禁用全局忽略字段
	omitColumns              []string // 当前 insert 忽略的字段
}

type InsertOption func(*InsertOptions)

// DisableGlobalInsertOmits insert 数据时，禁用全局忽略字段
func DisableGlobalInsertOmits(disable bool) InsertOption {
	return func(o *InsertOptions) {
		o.disableGlobalOmitColumns = disable
	}
}

// WithInsertOmits 当前 insert 时，忽略的字段
func WithInsertOmits(omits ...string) InsertOption {
	return func(o *InsertOptions) {
		o.omitColumns = append(o.omitColumns, omits...)
	}
}
