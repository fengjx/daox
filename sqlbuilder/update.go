package sqlbuilder

import "github.com/jmoiron/sqlx"

type Updater struct {
	sqlBuilder
	tableName string
	columns   []string
	vals      []interface{}
	where     ConditionBuilder
}

func NewUpdater(tableName string) *Updater {
	return &Updater{
		tableName: tableName,
	}
}

// Set 设置字段值
func (u *Updater) Set(column string, val interface{}) *Updater {
	u.columns = append(u.columns, column)
	u.vals = append(u.vals, val)
	return u
}

// Columns update 的数据库字段
func (u *Updater) Columns(columns ...string) *Updater {
	u.columns = columns
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
	if len(u.columns) == 0 {
		return "", ErrColumnsRequire
	}
	u.reset()
	u.writeString("UPDATE ")
	u.quote(u.tableName)
	u.writeString(" SET ")
	for i, column := range u.columns {
		u.quote(column)
		u.writeString(" = ?")
		if i != len(u.columns)-1 {
			u.writeString(", ")
		}
	}
	u.whereSQL(u.where)
	u.end()
	return u.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (u *Updater) SQLArgs() (string, []interface{}, error) {
	sql, err := u.SQL()
	args := u.vals
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
	if len(u.columns) == 0 {
		return "", ErrColumnsRequire
	}
	u.reset()
	u.writeString("UPDATE ")
	u.quote(u.tableName)
	u.writeString(" SET ")
	for i, column := range u.columns {
		u.quote(column)
		u.writeString(" = :")
		u.writeString(column)
		if i != len(u.columns)-1 {
			u.writeString(", ")
		}
	}
	u.whereSQL(u.where)
	u.end()
	return u.sb.String(), nil
}
