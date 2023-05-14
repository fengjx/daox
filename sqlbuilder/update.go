package sqlbuilder

import "strings"

type Updater struct {
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
	sb := &strings.Builder{}
	sb.WriteString("UPDATE ")
	warpQuote(sb, strings.TrimSpace(u.tableName))
	sb.WriteString(" SET ")
	for i, column := range u.columns {
		warpQuote(sb, strings.TrimSpace(column))
		sb.WriteString(" = ?")
		if i != len(u.columns)-1 {
			sb.WriteString(", ")
		}
	}
	buildWhereSql(sb, u.where)
	return sb.String(), nil
}

func (u *Updater) NameSql() (string, error) {
	if len(u.columns) == 0 {
		return "", SQLErrColumnsRequire
	}
	sb := &strings.Builder{}
	sb.WriteString("UPDATE ")
	warpQuote(sb, strings.TrimSpace(u.tableName))
	sb.WriteString(" SET ")
	for i, column := range u.columns {
		warpQuote(sb, strings.TrimSpace(column))
		sb.WriteString(" = :")
		sb.WriteString(column)
		if i != len(u.columns)-1 {
			sb.WriteString(", ")
		}
	}
	buildWhereSql(sb, u.where)
	return sb.String(), nil
}
