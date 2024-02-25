package daox

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/sqlbuilder"
)

type InsertRecord struct {
	TableName string         `json:"table_name"` // 表名
	Row       map[string]any `json:"row"`        // 行数据
}

func Insert(ctx context.Context, dbx *sqlx.DB, record InsertRecord) (int64, error) {
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

type UpdateRecord struct {
	TableName  string         `json:"table_name"` // 表名
	Fields     map[string]any `json:"fields"`     // 修改的字段
	Conditions []Condition    `json:"conditions"` // 条件字段
}

func Update(ctx context.Context, dbx *sqlx.DB, record UpdateRecord) (int64, error) {
	updater := sqlbuilder.NewUpdater(record.TableName)
	for col, val := range record.Fields {
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

type DeleteRecord struct {
	TableName  string      `json:"table_name"` // 表名
	Conditions []Condition `json:"conditions"` // 条件字段
}

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
