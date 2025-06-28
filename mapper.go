package daox

import (
	"reflect"

	"github.com/jmoiron/sqlx/reflectx"
)

// TableMapper 表映射器，用于将数据库表和模型进行映射
type TableMapper struct {
	Meta          *TableMeta       // 表元信息
	Mapper        *reflectx.Mapper // 字段映射器
	ModelType     reflect.Type     // 模型类型
	ModelTypeName string           // 模型类型名称
}
