package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	"github.com/fengjx/daox/engine"
)

type Options struct {
	tableName     string
	master        *sqlx.DB
	read          *sqlx.DB
	omitColumns   []string
	autoIncrement bool
	mapper        *reflectx.Mapper
	ifNullVals    map[string]string
	middlewares   []engine.Middleware
}

type Option func(*Options)

// WithDBMaster 设置主库
func WithDBMaster(master *sqlx.DB) Option {
	return func(p *Options) {
		p.master = master
	}
}

// WithDBRead 设置从库
func WithDBRead(read *sqlx.DB) Option {
	return func(p *Options) {
		p.read = read
	}
}

// IsAutoIncrement 是否自增主键
func IsAutoIncrement() Option {
	return func(dao *Options) {
		dao.autoIncrement = true
	}
}

// WithMapper 设置字段映射
func WithMapper(mapper *reflectx.Mapper) Option {
	return func(d *Options) {
		d.mapper = mapper
	}
}

// WithTableName 设置表名
func WithTableName(tableName string) Option {
	return func(d *Options) {
		d.tableName = tableName
	}
}

// WithIfNullVal 设置字段为null时的默认值
func WithIfNullVal(col string, val string) Option {
	return func(d *Options) {
		if d.ifNullVals == nil {
			d.ifNullVals = make(map[string]string)
		}
		d.ifNullVals[col] = val
	}
}

// WithIfNullVals 设置字段（多个）为null时的默认值
func WithIfNullVals(vals map[string]string) Option {
	return func(d *Options) {
		if d.ifNullVals == nil {
			d.ifNullVals = make(map[string]string)
		}
		for col, val := range vals {
			d.ifNullVals[col] = val
		}
	}
}

// WithOmitColumns 设置忽略字段
func WithOmitColumns(omitColumns ...string) Option {
	return func(d *Options) {
		d.omitColumns = omitColumns
	}
}

// WithMiddleware 设置中间件
func WithMiddleware(middlewares ...engine.Middleware) Option {
	return func(d *Options) {
		d.middlewares = middlewares
	}
}

// InsertOptions insert 选项
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
