package daox

import (
	"context"

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

type DataWrapper[S any, T any] func(context.Context, S) T

type FieldsFilter func(context.Context) []string

type InsertOptions struct {
	DataWrapper  DataWrapper[map[string]any, map[string]any]
	FieldsFilter FieldsFilter
}

type InsertOption func(*InsertOptions)

// WithInsertDataWrapper 数据转换
func WithInsertDataWrapper(dataWrapper DataWrapper[map[string]any, map[string]any]) InsertOption {
	return func(o *InsertOptions) {
		o.DataWrapper = dataWrapper
	}
}

// WithInsertFieldsFilter 过滤 insert 字段
func WithInsertFieldsFilter(fieldsFilter FieldsFilter) InsertOption {
	return func(o *InsertOptions) {
		o.FieldsFilter = fieldsFilter
	}
}

type SelectOptions struct {
	FieldsFilter  FieldsFilter
	ResultWrapper DataWrapper[any, any]
}

type SelectOption func(*SelectOptions)

// WithSelectFieldsFilter 过滤 select 字段
func WithSelectFieldsFilter(fieldsFilter FieldsFilter) SelectOption {
	return func(o *SelectOptions) {
		o.FieldsFilter = fieldsFilter
	}
}

// WithSelectDataWrapper 返回结果转换
func WithSelectDataWrapper(resultWrapper DataWrapper[any, any]) SelectOption {
	return func(o *SelectOptions) {
		o.ResultWrapper = resultWrapper
	}
}

type UpdateOptions struct {
	DataWrapper  DataWrapper[map[string]any, map[string]any]
	FieldsFilter FieldsFilter
}

type UpdateOption func(*UpdateOptions)

// WithUpdateFieldsFilter 过滤 update 字段
func WithUpdateFieldsFilter(fieldsFilter FieldsFilter) UpdateOption {
	return func(o *UpdateOptions) {
		o.FieldsFilter = fieldsFilter
	}
}

// WithUpdateDataWrapper 数据转换
func WithUpdateDataWrapper(dataWrapper DataWrapper[map[string]any, map[string]any]) UpdateOption {
	return func(o *UpdateOptions) {
		o.DataWrapper = dataWrapper
	}
}
