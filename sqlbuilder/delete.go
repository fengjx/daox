package sqlbuilder

import "strings"

type Deleter struct {
	tableName string
	where     *condition
}

func NewDeleter(tableName string) *Deleter {
	return &Deleter{
		tableName: tableName,
	}
}

func (d *Deleter) Where(condition *condition) *Deleter {
	d.where = condition
	return d
}

func (d *Deleter) Sql() (string, error) {
	if d.where == nil || len(d.where.predicates) == 0 {
		return "", SQLErrDeleteMissWhere
	}
	sb := &strings.Builder{}
	sb.WriteString("DELETE FROM ")
	warpQuote(sb, strings.TrimSpace(d.tableName))
	buildWhereSql(sb, d.where)
	return sb.String(), nil
}
