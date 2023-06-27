package daox

import (
	"reflect"
)

type TableMeta struct {
	TableName       string
	StructType      reflect.Type
	Columns         []string
	PrimaryKey      string
	IsAutoIncrement bool
}

func (meta *TableMeta) OmitColumns(omit ...string) []string {
	columnArr := make([]string, 0, len(meta.Columns))
	for _, column := range meta.Columns {
		if !containsString(omit, column) {
			columnArr = append(columnArr, column)
		}
	}
	return columnArr
}

type Model interface {
	GetID() interface{}
}
