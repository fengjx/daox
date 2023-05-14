package sqlbuilder

import (
	"errors"
	"strings"
)

var (
	SQLErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
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

type condition struct {
	meet    bool
	express string
}

func Condition(meet bool, express string) *condition {
	return &condition{
		meet:    meet,
		express: express,
	}
}

func warpQuote(sb *strings.Builder, s string) {
	sb.WriteString("`")
	sb.WriteString(s)
	sb.WriteString("`")
}
