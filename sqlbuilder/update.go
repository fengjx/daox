package sqlbuilder

import (
	"context"
	"strconv"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/internal/adapter"
)

// Updater 语句构造器
type Updater struct {
	execer adapter.Execer
	sqlBuilder
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

// DB 设置数据库
func (u *Updater) DB(execer adapter.Execer) *Updater {
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
	for i, f := range u.fields {
		u.quote(f.col)
		u.writeString(" = ")
		if f.incrVal != nil {
			u.quote(f.col)
			u.writeString(" + ")
			u.writeString(strconv.FormatInt(*f.incrVal, 10))
		} else {
			u.writeString("?")
		}
		if i != len(u.fields)-1 {
			u.writeString(", ")
		}
	}
	u.whereSQL(u.where)
	u.end()
	return u.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (u *Updater) SQLArgs() (string, []any, error) {
	sql, err := u.SQL()
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
		return sql, args, err
	}
	return sqlx.In(sql, args...)
}

func (u *Updater) NameSQL() (string, error) {
	if len(u.fields) == 0 {
		return "", ErrColumnsRequire
	}
	u.reset()
	u.writeString("UPDATE ")
	u.quote(u.tableName)
	u.writeString(" SET ")
	for i, f := range u.fields {
		u.quote(f.col)
		u.writeString(" = ")
		if f.incrVal != nil {
			u.quote(f.col)
			u.writeString(" + ")
			u.writeString(strconv.FormatInt(*f.incrVal, 10))
		} else {
			u.writeString(":")
			u.writeString(f.col)
		}
		if i != len(u.fields)-1 {
			u.writeString(", ")
		}
	}
	u.whereSQL(u.where)
	u.end()
	return u.sb.String(), nil
}

// Exec 执行更新语句
func (u *Updater) Exec(ctx context.Context) (int64, error) {
	if u.execer == nil {
		return 0, ErrExecerNotSet
	}
	sql, args, err := u.SQLArgs()
	if err != nil {
		return 0, err
	}
	result, err := u.execer.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// NamedExec 通过 NameSQL 执行更新语句，参数通过 data 填充
func (u *Updater) NamedExec(ctx context.Context, data any) (int64, error) {
	if u.execer == nil {
		return 0, ErrExecerNotSet
	}
	sql, err := u.NameSQL()
	if err != nil {
		return 0, err
	}
	result, err := u.execer.NamedExecContext(ctx, sql, data)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
