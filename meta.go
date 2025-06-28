package daox

import (
	"reflect"
	"strings"

	"github.com/fengjx/daox/v2/utils"
)

// TableMeta 数据库表元信息，包含表的基本信息和字段定义
type TableMeta struct {
	TableName       string   // 表名
	Columns         []string // 表字段列表
	PrimaryKey      string   // 主键字段名
	IsAutoIncrement bool     // 主键是否自增
}

// OmitColumns 获取排除指定字段后的字段列表
// omit: 需要排除的字段列表
// 返回值: 排除指定字段后的字段列表
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
// tableName: 新的表名
func (meta TableMeta) WithTableName(tableName string) *TableMeta {
	return &TableMeta{
		TableName:       tableName,
		Columns:         meta.Columns,
		PrimaryKey:      meta.PrimaryKey,
		IsAutoIncrement: meta.IsAutoIncrement,
	}
}

// Model 数据库模型接口，所有数据库模型结构体都需要实现此接口
type Model interface {
	// GetID 获取模型的主键值
	GetID() any

	New() Model
}

// Meta 数据库表元信息定义接口，用于自定义表元信息的获取方式
type Meta interface {
	// TableName 获取表名
	TableName() string
	// PrimaryKey 获取主键字段名
	PrimaryKey() string
	// IsAutoIncrement 判断主键是否自增
	IsAutoIncrement() bool
	// Columns 获取表的所有字段
	Columns() []string
}

type RelationType string

const (
	// RelO2O 一对一
	RelO2O RelationType = "o2o"
	// RelO2M 一对多
	RelO2M RelationType = "o2m"
	// RelM2M 多对多
	RelM2M RelationType = "m2m"
)

// RelationInfo 关系定义
type RelationInfo struct {
	RelType RelationType // 关系类型
	FromKey string       // 主表用于关联的字段
	Ref     string       // 关联表模型名称
	JoinKey string       // 关联表用于关联的字段
}

// ParseRelation 解析 field 中的 daox 关联信息
func ParseRelation(field reflect.StructField) (*RelationInfo, bool) {
	tag := field.Tag.Get("daox")
	if tag == "" {
		return nil, false
	}
	info := &RelationInfo{}
	for _, part := range strings.Split(tag, ",") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			switch kv[0] {
			case "relType":
				info.RelType = RelationType(kv[1])
			case "fromKey":
				info.FromKey = kv[1]
			case "ref":
				info.Ref = kv[1]
			case "joinKey":
				info.JoinKey = kv[1]
			}
		}
	}
	return info, true
}

// ParseRelations 解析 type 中所有带 daox 关联的字段
func ParseRelations(typ reflect.Type) map[string]*RelationInfo {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	res := make(map[string]*RelationInfo)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if info, ok := ParseRelation(field); ok {
			res[field.Name] = info
		}
	}
	return res
}
