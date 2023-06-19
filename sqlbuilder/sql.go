package sqlbuilder

import (
	"errors"
	"strings"
)

var (
	SQLErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
	SQLErrColumnsRequire   = errors.New("[sqlbuilder] columns requires")
	SQLErrDeleteMissWhere  = errors.New("[sqlbuilder] delete sql miss where")
)

type Builder struct {
	tableName string
}

func New(tableName string) *Builder {
	builder := &Builder{
		tableName: tableName,
	}
	return builder
}

func (b *Builder) Select() *Selector {
	return NewSelector(b.tableName)
}

func (b *Builder) Insert() *Inserter {
	return NewInserter(b.tableName)
}

func (b *Builder) Update() *Updater {
	return NewUpdater(b.tableName)
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
