package sqlbuilder

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/engine"
)

// Updater update 语句构造器
type Updater struct {
	sqlBuilder
	execer    engine.Execer
	tableName string
	fields    []Field
	where     ConditionBuilder
}

// NewUpdater 创建一个 update 语句构造器
func NewUpdater(tableName string) *Updater {
	return &Updater{
		tableName: tableName,
	}
}

// Execer 设置Execer
func (u *Updater) Execer(execer engine.Execer) *Updater {
	u.execer = execer
	return u
}

// Fields 设置字段值
func (u *Updater) Fields(fields ...Field) *Updater {
	for _, field := range fields {
		if !field.isUse {
			continue
		}
		u.fields = append(u.fields, field)
	}
	return u
}

// Set 设置字段值
func (u *Updater) Set(column string, val any) *Updater {
	u.fields = append(u.fields, F(column).Val(val))
	return u
}

// Columns update 的数据库字段
func (u *Updater) Columns(columns ...string) *Updater {
	for _, col := range columns {
		u.fields = append(u.fields, F(col))
	}
	return u
}

// Incr 数值增加，eg: set a = a + 1
func (u *Updater) Incr(column string, n int64) *Updater {
	u.fields = append(u.fields, F(column).Incr(n))
	return u
}

// Where 条件
// condition 可以通过 sqlbuilder.C() 方法创建
func (u *Updater) Where(where ConditionBuilder) *Updater {
	u.where = where
	return u
}

// SQL 输出sql语句
func (u *Updater) SQL() (string, error) {
	if len(u.fields) == 0 {
		return "", ErrColumnsRequire
	}
	u.reset()
	u.writeString("UPDATE ")
	u.quote(u.tableName)
	u.writeString(" SET ")
	u.setFields(u.fields)
	u.whereSQL(u.where)
	u.end()
	return u.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (u *Updater) SQLArgs() (string, []any, error) {
	if len(u.fields) == 0 {
		return "", nil, ErrColumnsRequire
	}
	if u.where != nil && len(u.where.getPredicates()) == 0 {
		return "", nil, ErrUpdateMissWhere
	}
	execSQL, err := u.SQL()
	var args []any
	for _, f := range u.fields {
		if f.val != nil {
			args = append(args, f.val)
		}
	}
	wargs, hasInSQL := u.whereArgs(u.where)
	if len(wargs) > 0 {
		args = append(args, wargs...)
	}
	if !hasInSQL {
		return execSQL, args, err
	}
	return sqlx.In(execSQL, args...)
}

func (u *Updater) NameSQL() (string, error) {
	if len(u.fields) == 0 {
		return "", ErrColumnsRequire
	}
	if u.where != nil && len(u.where.getPredicates()) == 0 {
		return "", ErrUpdateMissWhere
	}
	u.reset()
	u.writeString("UPDATE ")
	u.quote(u.tableName)
	u.writeString(" SET ")
	u.setNameFields(u.fields)
	u.whereSQL(u.where)
	u.end()
	return u.sb.String(), nil
}

// Exec 执行更新语句
func (u *Updater) Exec() (int64, error) {
	return u.ExecContext(context.Background())
}

// ExecContext 执行更新语句
func (u *Updater) ExecContext(ctx context.Context) (int64, error) {
	if u.execer == nil {
		return 0, ErrExecerNotSet
	}
	execSQL, args, err := u.SQLArgs()
	if err != nil {
		return 0, err
	}
	result, err := u.execer.ExecContext(ctx, execSQL, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// NamedExec 通过 NameSQL 执行更新语句，参数通过 data 填充
// where 条件也必须是 name 风格
func (u *Updater) NamedExec(data any) (int64, error) {
	return u.NamedExecContext(context.Background(), data)
}

// NamedExecContext 通过 NameSQL 执行更新语句，参数通过 data 填充
// where 条件也必须是 name 风格
func (u *Updater) NamedExecContext(ctx context.Context, data any) (int64, error) {
	if u.execer == nil {
		return 0, ErrExecerNotSet
	}
	execSQL, err := u.NameSQL()
	if err != nil {
		return 0, err
	}
	result, err := u.execer.NamedExecContext(ctx, execSQL, data)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
