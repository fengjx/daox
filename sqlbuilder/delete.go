package sqlbuilder

import (
	"context"
	"strconv"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/engine"
)

// Deleter delete 语句构造器
type Deleter struct {
	sqlBuilder
	execer    engine.Execer
	tableName string
	where     ConditionBuilder
	limit     *int
}

// NewDeleter
// tableName 数据库表名
func NewDeleter(tableName string) *Deleter {
	return &Deleter{
		tableName: tableName,
	}
}

// Execer 设置Execer
func (d *Deleter) Execer(execer engine.Execer) *Deleter {
	d.execer = execer
	return d
}

// Where 条件
// condition 可以通过 sqlbuilder.C() 方法创建
func (d *Deleter) Where(where ConditionBuilder) *Deleter {
	d.where = where
	return d
}

// Limit 限制删除数量
func (d *Deleter) Limit(limit int) *Deleter {
	d.limit = &limit
	return d
}

// SQL 输出sql语句
func (d *Deleter) SQL() (string, error) {
	if d.where == nil || len(d.where.getPredicates()) == 0 {
		return "", ErrDeleteMissWhere
	}
	d.reset()
	d.writeString("DELETE FROM ")
	d.quote(d.tableName)
	d.whereSQL(d.where)
	if d.limit != nil {
		d.writeString(" LIMIT ")
		d.writeString(strconv.Itoa(*d.limit))
	}
	d.end()
	return d.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (d *Deleter) SQLArgs() (string, []any, error) {
	if d.where == nil || len(d.where.getPredicates()) == 0 {
		return "", nil, ErrDeleteMissWhere
	}
	execSQL, err := d.SQL()
	if err != nil {
		return "", nil, err
	}
	args, hasInSQL := d.whereArgs(d.where)
	if !hasInSQL {
		return execSQL, args, err
	}
	return sqlx.In(execSQL, args...)
}

// Exec 执行更新语句
func (d *Deleter) Exec() (int64, error) {
	return d.ExecContext(context.Background())
}

// ExecContext 执行更新语句
func (d *Deleter) ExecContext(ctx context.Context) (int64, error) {
	if d.execer == nil {
		return 0, ErrExecerNotSet
	}
	execSQL, args, err := d.SQLArgs()
	if err != nil {
		return 0, err
	}
	result, err := d.execer.ExecContext(ctx, execSQL, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
