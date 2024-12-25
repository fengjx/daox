package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	"github.com/fengjx/daox/engine"
)

// Options Dao的配置选项
type Options struct {
	// 表名
	tableName string
	// 主库连接
	master *sqlx.DB
	// 从库连接
	read *sqlx.DB
	// 忽略的字段列表
	omitColumns []string
	// 是否自增主键
	autoIncrement bool
	// 字段映射器
	mapper *reflectx.Mapper
	// IFNULL默认值配置
	ifNullVals map[string]string
	// 回调处理函数
	hooks []engine.Hook
	// SQL打印处理器
	printSQL engine.AfterHandler
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

// WithHooks 设置中间件
func WithHooks(hooks ...engine.Hook) Option {
	return func(d *Options) {
		d.hooks = hooks
	}
}

// WithPrintSQL 打印 sql 回调
func WithPrintSQL(printSQL engine.AfterHandler) Option {
	return func(d *Options) {
		d.printSQL = printSQL
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
