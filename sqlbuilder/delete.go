package sqlbuilder

type Deleter struct {
	sqlBuilder
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
	d.reset()
	d.writeString("DELETE FROM ")
	d.quote(d.tableName)
	d.whereSql(d.where)
	return d.sb.String(), nil
}
