package sqlbuilder

type Deleter struct {
	sqlBuilder
	tableName string
	where     ConditionBuilder
}

// NewDeleter
// tableName 数据库表名
func NewDeleter(tableName string) *Deleter {
	return &Deleter{
		tableName: tableName,
	}
}

// Where 条件
// condition 可以通过 sqlbuilder.C() 方法创建
func (d *Deleter) Where(where ConditionBuilder) *Deleter {
	d.where = where
	return d
}

// SQL 输出sql语句
func (d *Deleter) SQL() (string, error) {
	if d.where == nil || len(d.where.getPredicates()) == 0 {
		return "", SQLErrDeleteMissWhere
	}
	d.reset()
	d.writeString("DELETE FROM ")
	d.quote(d.tableName)
	d.whereSQL(d.where)
	d.end()
	return d.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (d *Deleter) SQLArgs() (string, []interface{}, error) {
	sql, err := d.SQL()
	args := d.whereArgs(d.where)
	return sql, args, err
}
