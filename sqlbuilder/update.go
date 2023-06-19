package sqlbuilder

type Updater struct {
	sqlBuilder
	tableName string
	columns   []string
	where     *condition
}

func NewUpdater(tableName string) *Updater {
	return &Updater{
		tableName: tableName,
	}
}

func (u *Updater) Columns(columns ...string) *Updater {
	u.columns = columns
	return u
}

func (u *Updater) Where(condition *condition) *Updater {
	u.where = condition
	return u
}

func (u *Updater) Sql() (string, error) {
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
	u.whereSql(u.where)
	return u.sb.String(), nil
}

func (u *Updater) NameSql() (string, error) {
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
	u.whereSql(u.where)
	return u.sb.String(), nil
}
