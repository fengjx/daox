package sqlbuilder

import (
	"strings"
)

type Inserter struct {
	sqlBuilder
	tableName                  string
	columns                    []string
	onDuplicateKeyUpdateString string
}

func NewInserter(tableName string) *Inserter {
	inserter := &Inserter{
		tableName: tableName,
	}
	return inserter
}

func (ins *Inserter) StructColumns(m interface{}, tagName string, omitColumns ...string) *Inserter {
	columns := GetColumnsByModel(GetMapperByTagName(tagName), m, omitColumns...)
	return ins.Columns(columns...)
}

func (ins *Inserter) Columns(columns ...string) *Inserter {
	ins.columns = columns
	return ins
}

func (ins *Inserter) OnDuplicateKeyUpdateString(updateString string) *Inserter {
	ins.onDuplicateKeyUpdateString = updateString
	return ins
}

func (ins *Inserter) NameSQL() (string, error) {
	if len(ins.columns) == 0 {
		return "", SQLErrColumnsRequire
	}
	ins.reset()
	ins.writeString("INSERT INTO ")
	ins.quote(strings.TrimSpace(ins.tableName))
	ins.writeByte('(')
	for i, column := range ins.columns {
		ins.quote(column)
		if i != len(ins.columns)-1 {
			ins.writeString(", ")
		}
	}
	ins.writeByte(')')
	ins.writeString(" VALUES (")
	for i, column := range ins.columns {
		ins.writeByte(':')
		ins.writeString(column)
		if i != len(ins.columns)-1 {
			ins.writeString(", ")
		}
	}
	ins.writeString(")")
	if ins.onDuplicateKeyUpdateString != "" {
		ins.writeString(" ON DUPLICATE KEY UPDATE ")
		ins.writeString(ins.onDuplicateKeyUpdateString)
	}
	ins.end()
	return ins.sb.String(), nil
}

func (ins *Inserter) SQL() (string, error) {
	if len(ins.columns) == 0 {
		return "", SQLErrColumnsRequire
	}
	ins.reset()
	ins.writeString("INSERT INTO ")
	ins.quote(ins.tableName)
	ins.writeByte('(')
	for i, column := range ins.columns {
		ins.quote(column)
		if i != len(ins.columns)-1 {
			ins.writeString(", ")
		}
	}
	ins.writeString(") VALUES (")
	for i := range ins.columns {
		ins.writeByte('?')
		if i != len(ins.columns)-1 {
			ins.writeString(", ")
		}
	}
	ins.writeString(")")
	if ins.onDuplicateKeyUpdateString != "" {
		ins.writeString(" ON DUPLICATE KEY UPDATE ")
		ins.writeString(ins.onDuplicateKeyUpdateString)
	}
	ins.end()
	return ins.sb.String(), nil
}
