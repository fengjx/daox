package sqlbuilder

import (
	"errors"
	"strings"
)

var (
	SQLErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
	SQLErrColumnsRequire   = errors.New("[sqlbuilder] columns requires")
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

func warpQuote(sb *strings.Builder, s string) {
	sb.WriteString("`")
	sb.WriteString(s)
	sb.WriteString("`")
}

func buildWhereSql(sb *strings.Builder, condition *condition) {
	if condition != nil && len(condition.predicates) > 0 {
		sb.WriteString(" WHERE ")
		for _, predicate := range condition.predicates {
			if predicate.op != nil {
				sb.WriteString(predicate.op.text)
			}
			sb.WriteString(predicate.express)
		}
	}
}
