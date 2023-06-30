package sqlbuilder

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx/reflectx"
)

var (
	SQLErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
	SQLErrColumnsRequire   = errors.New("[sqlbuilder] columns requires")
	SQLErrDeleteMissWhere  = errors.New("[sqlbuilder] delete sql miss where")
)

var MapperMap = map[string]*reflectx.Mapper{}

func init() {
	MapperMap["json"] = reflectx.NewMapperFunc("json", strings.ToTitle)
	MapperMap["db"] = reflectx.NewMapperFunc("db", strings.ToTitle)
}

var createMapperLock sync.Mutex

func GetMapperByTagName(tagName string) *reflectx.Mapper {
	if mapper, ok := MapperMap[tagName]; ok {
		return mapper
	}
	createMapperLock.Lock()
	mapper := reflectx.NewMapperFunc(tagName, strings.ToTitle)
	MapperMap[tagName] = mapper
	createMapperLock.Unlock()
	return mapper
}

func GetColumnsByModel(mapper *reflectx.Mapper, model interface{}, omitColumns ...string) []string {
	return GetColumnsByType(mapper, reflect.TypeOf(model), omitColumns...)
}

// GetColumnsByType 通过字段 tag 解析数据库字段
func GetColumnsByType(mapper *reflectx.Mapper, typ reflect.Type, omitColumns ...string) []string {
	structMap := mapper.TypeMap(typ)
	columns := make([]string, 0)
	for _, fieldInfo := range structMap.Tree.Children {
		if fieldInfo == nil || fieldInfo.Name == "" || ContainsString(omitColumns, fieldInfo.Name) {
			continue
		}
		columns = append(columns, fieldInfo.Name)
	}
	return columns
}

type Builder struct {
	tableName string
}

func New(tableName string) *Builder {
	builder := &Builder{
		tableName: tableName,
	}
	return builder
}

func (b *Builder) Select(columns ...string) *Selector {
	return NewSelector(b.tableName).Columns(columns...)
}

func (b *Builder) Insert(columns ...string) *Inserter {
	return NewInserter(b.tableName).Columns(columns...)
}

func (b *Builder) Update(columns ...string) *Updater {
	return NewUpdater(b.tableName).Columns(columns...)
}

func (b *Builder) Delete() *Deleter {
	return NewDeleter(b.tableName)
}

type sqlBuilder struct {
	sb strings.Builder
}

func (b *sqlBuilder) reset() {
	b.sb.Reset()
}

func (b *sqlBuilder) writeString(val string) {
	_, _ = b.sb.WriteString(val)
}

func (b *sqlBuilder) writeByte(c byte) {
	_ = b.sb.WriteByte(c)
}

func (b *sqlBuilder) quote(val string) {
	b.writeByte('`')
	b.writeString(strings.TrimSpace(val))
	b.writeByte('`')
}

func (b *sqlBuilder) space() {
	b.writeByte(' ')
}

func (b *sqlBuilder) end() {
	b.writeByte(';')
}

func (b *sqlBuilder) comma() {
	b.writeByte(',')
}

func (b *sqlBuilder) whereSQL(condition *condition) {
	if condition != nil && len(condition.predicates) > 0 {
		b.writeString(" WHERE ")
		for _, predicate := range condition.predicates {
			if predicate.op != nil {
				b.writeString(predicate.op.text)
			}
			b.writeString(predicate.express)
		}
	}
}

func ContainsString(collection []string, element string) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}
