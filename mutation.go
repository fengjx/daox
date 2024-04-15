package daox

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/sqlbuilder"
)

// InsertRecord 插入记录
type InsertRecord struct {
	TableName string         `json:"table_name"` // 表名
	Row       map[string]any `json:"row"`        // 行数据
}

// Insert 通用 insert 操作
func Insert(ctx context.Context, dbx *sqlx.DB, record InsertRecord, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, option := range opts {
		option(opt)
	}
	if opt.DataWrapper != nil {
		record.Row = opt.DataWrapper(ctx, record.Row)
	}
	if opt.FieldsFilter != nil {
		for _, disableField := range opt.FieldsFilter(ctx) {
			delete(record.Row, disableField)
		}
	}
	inserter := sqlbuilder.NewInserter(record.TableName)
	var columns []string
	for col := range record.Row {
		columns = append(columns, col)
	}
	inserter.Columns(columns...)
	sql, err := inserter.NameSQL()
	if err != nil {
		return 0, err
	}
	result, err := dbx.NamedExecContext(ctx, sql, record.Row)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateRecord 更新记录
type UpdateRecord struct {
	TableName  string         `json:"table_name"` // 表名
	Row        map[string]any `json:"row"`        // 要修改的行记录
	Conditions []Condition    `json:"conditions"` // 条件字段
}

// Update 通用 update 操作
func Update(ctx context.Context, dbx *sqlx.DB, record UpdateRecord, opts ...UpdateOption) (int64, error) {
	opt := &UpdateOptions{}
	for _, option := range opts {
		option(opt)
	}
	if opt.DataWrapper != nil {
		record.Row = opt.DataWrapper(ctx, record.Row)
	}
	if opt.FieldsFilter != nil {
		for _, disableField := range opt.FieldsFilter(ctx) {
			delete(record.Row, disableField)
		}
	}
	updater := sqlbuilder.NewUpdater(record.TableName)
	for col, val := range record.Row {
		updater.Set(col, val)
	}
	updater.Where(buildCondition(record.Conditions))
	sql, args, err := updater.SQLArgs()
	if err != nil {
		return 0, err
	}
	result, err := dbx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteRecord 删除记录
type DeleteRecord struct {
	TableName  string      `json:"table_name"` // 表名
	Conditions []Condition `json:"conditions"` // 条件字段
}

// Delete 通用 delete 操作
func Delete(ctx context.Context, dbx *sqlx.DB, record DeleteRecord) (int64, error) {
	deleter := sqlbuilder.NewDeleter(record.TableName)
	deleter.Where(buildCondition(record.Conditions))
	sql, args, err := deleter.SQLArgs()
	if err != nil {
		return 0, err
	}
	result, err := dbx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
