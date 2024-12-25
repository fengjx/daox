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
func (meta TableMeta) OmitColumns(omit ...string) []string {
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

// WithTableName 设置表名，一般用在分表的场景，设置实际物理表名
func (meta TableMeta) WithTableName(tableName string) *TableMeta {
	return &TableMeta{
		TableName:       tableName,
		Columns:         meta.Columns,
		PrimaryKey:      meta.PrimaryKey,
		IsAutoIncrement: meta.IsAutoIncrement,
	}
}

// Model 数据库 model 定义
type Model interface {
	GetID() any
}

// Meta 定义数据库表的元数据信息接口
type Meta interface {
	// TableName 获取表名
	TableName() string
	// PrimaryKey 获取主键字段名
	PrimaryKey() string
	// IsAutoIncrement 判断主键是否为自增类型
	IsAutoIncrement() bool
	// Columns 获取所有数据库字段名
	Columns() []string
}
