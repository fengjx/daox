package sqlbuilder

import (
	"strings"
)

type Inserter struct {
	tableName string
	columns   []string
}

func NewInserter(tableName string) *Inserter {
	inserter := &Inserter{
		tableName: tableName,
	}
	return inserter
}

func (ins *Inserter) Columns(columns ...string) *Inserter {
	ins.columns = columns
	return ins
}

func (ins *Inserter) NameSql() (string, error) {
	if len(ins.columns) == 0 {
		return "", SQLErrColumnsRequire
	}
	sb := &strings.Builder{}
	sb.WriteString("INSERT INTO ")
	warpQuote(sb, strings.TrimSpace(ins.tableName))
	sb.WriteString("(")
	for i, column := range ins.columns {
		warpQuote(sb, strings.TrimSpace(column))
		if i != len(ins.columns)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	sb.WriteString(" VALUES (")
	for i, column := range ins.columns {
		sb.WriteString(":")
		sb.WriteString(column)
		if i != len(ins.columns)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	return sb.String(), nil
}

func (ins *Inserter) Sql() (string, error) {
	if len(ins.columns) == 0 {
		return "", SQLErrColumnsRequire
	}
	sb := &strings.Builder{}
	sb.WriteString("INSERT INTO ")
	warpQuote(sb, strings.TrimSpace(ins.tableName))
	sb.WriteString("(")
	for i, column := range ins.columns {
		warpQuote(sb, strings.TrimSpace(column))
		if i != len(ins.columns)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(") VALUES (")
	for i := range ins.columns {
		sb.WriteString("?")
		if i != len(ins.columns)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	return sb.String(), nil
}
