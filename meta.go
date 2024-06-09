package daox

import (
	"github.com/fengjx/daox/utils"
)

// TableMeta 数据库表元信息
type TableMeta struct {
	TableName       string
	Columns         []string
	PrimaryKey      string
	IsAutoIncrement bool
}

// OmitColumns 数据库表字段
// omit 包含的字段
func (meta *TableMeta) OmitColumns(omit ...string) []string {
	if len(omit) == 0 {
		return meta.Columns
	}
	columnArr := make([]string, 0, len(meta.Columns))
	for _, column := range meta.Columns {
		if !utils.ContainsString(omit, column) {
			columnArr = append(columnArr, column)
		}
	}
	return columnArr
}

// Model 数据库 model 定义
type Model interface {
	GetID() any
}

// Meta 数据库表元信息定义接口
type Meta interface {
	TableName() string
	PrimaryKey() string
	IsAutoIncrement() bool
	Columns() []string
}
