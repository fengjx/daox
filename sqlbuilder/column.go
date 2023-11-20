package sqlbuilder

import (
	"strings"
)

type Column struct {
	name  string
	op    Op
	arg   interface{}
	isUse bool
}

// Col 表字段
func Col(c string) Column {
	return Column{
		name:  c,
		isUse: true,
	}
}

func (c Column) Use(use bool) Column {
	c.isUse = use
	return c
}

// EQ =
func (c Column) EQ(val any) Column {
	c.op = OpEQ
	c.arg = val
	return c
}

// NotEQ !=
func (c Column) NotEQ(val any) Column {
	c.op = OpNEQ
	c.arg = val
	return c
}

// LT <
func (c Column) LT(val any) Column {
	c.op = OpLT
	c.arg = val
	return c
}

// LTEQ <=
func (c Column) LTEQ(val any) Column {
	c.op = OpLTEQ
	c.arg = val
	return c
}

// GT >
func (c Column) GT(val any) Column {
	c.op = OpGT
	c.arg = val
	return c
}

// GTEQ >=
func (c Column) GTEQ(val any) Column {
	c.op = OpGTEQ
	c.arg = val
	return c
}

// Like -> LIKE %XXX
func (c Column) Like(val any) Column {
	c.op = OpLike
	c.arg = val
	return c
}

// NotLike -> NOT LIKE %XXX 、_x_ 、xx[xx-xx] 、xx[^xx-xx]
func (c Column) NotLike(val any) Column {
	c.op = OpNotLike
	c.arg = val
	return c
}

// In -> in ()
func (c Column) In(vals ...any) Column {
	c.op = OpIn
	c.arg = vals
	return c
}

// NotIn -> not in ()
func (c Column) NotIn(vals ...any) Column {
	c.op = OpNotIN
	c.arg = vals
	return c
}

func (c Column) Express() string {
	sb := strings.Builder{}
	sb.WriteByte('`')
	sb.WriteString(c.name)
	sb.WriteByte('`')
	sb.WriteString(c.op.Text)
	if c.HasInSQL() {
		sb.WriteString("(?)")
	} else {
		sb.WriteString("?")
	}
	return sb.String()
}

func (c Column) HasInSQL() bool {
	return c.op == OpIn || c.op == OpNotIN
}
