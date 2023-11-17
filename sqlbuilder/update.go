package sqlbuilder

type Updater struct {
	sqlBuilder
	tableName string
	columns   []string
	where     ConditionBuilder
}

func NewUpdater(tableName string) *Updater {
	return &Updater{
		tableName: tableName,
	}
}

// StructColumns 通过任意model解析出表字段
// tagName 解析数据库字段的 tag-name
// omitColumns 排除哪些字段
func (u *Updater) StructColumns(model interface{}, tagName string, omitColumns ...string) *Updater {
	columns := GetColumnsByModel(GetMapperByTagName(tagName), model, omitColumns...)
	return u.Columns(columns...)
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

// SQL 拼接sql语句
func (u *Updater) SQL() (string, error) {
	if len(u.columns) == 0 {
		return "", SQLErrColumnsRequire
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
	args := u.whereArgs(u.where)
	return sql, args, err
}

func (u *Updater) NameSQL() (string, error) {
	if len(u.columns) == 0 {
		return "", SQLErrColumnsRequire
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
