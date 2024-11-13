package sqlbuilder

import (
	"strings"
)

// Column 表字段
type Column struct {
	name  string
	op    Op
	arg   any
	isUse bool
}

// Col 表字段
func Col(c string) Column {
	return Column{
		name:  c,
		isUse: true,
	}
}

// Use 是否使用
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

// IsNull -> IS NULL
func (c Column) IsNull() Column {
	c.op = OpIsNull
	return c
}

// IsNotNull -> IS NOT NULL
func (c Column) IsNotNull() Column {
	c.op = OpIsNotNull
	return c
}

func (c Column) getArgs() []any {
	if c.arg == nil {
		return nil
	}
	return []any{c.arg}
}

// Express 输出 sql 表达式
func (c Column) Express() string {
	sb := strings.Builder{}
	sb.WriteByte('`')
	sb.WriteString(c.name)
	sb.WriteByte('`')
	sb.WriteString(c.op.Text)
	if c.HasInSQL() {
		sb.WriteString("(?)")
	} else if c.arg != nil {
		sb.WriteString("?")
	}
	return sb.String()
}

// HasInSQL 是否有 in 语句
func (c Column) HasInSQL() bool {
	return c.op == OpIn || c.op == OpNotIN
}

// Field 表更新字段
type Field struct {
	isUse   bool   // 是否启用
	col     string // 字段名
	val     any    // 字段值
	incrVal *int64 // 递增值，eg: set a = a + 1
}

// F 创建更新字段
func F(col string) Field {
	return Field{
		isUse: true,
		col:   col,
	}
}

// Val 设置字段值
func (f Field) Val(val any) Field {
	f.val = val
	return f
}

// Incr 设置字段增加值
func (f Field) Incr(n int64) Field {
	f.incrVal = &n
	return f
}

// Use 是否启用
func (f Field) Use(use bool) Field {
	f.isUse = use
	return f
}
