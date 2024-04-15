package sqlbuilder

// Predicate where 断言
type Predicate struct {
	Op       Op
	Express  string
	Args     []any
	HasInSQL bool
}

// ConditionBuilder 条件构造器
type ConditionBuilder interface {
	getPredicates() []Predicate
}

// SimpleCondition 简单 where 条件构造
type SimpleCondition struct {
	predicates []Predicate
}

// SC 简单 where 条件
func SC() *SimpleCondition {
	return new(SimpleCondition)
}

func (c *SimpleCondition) Predicates() []Predicate {
	return c.predicates
}

// And and 语句
// express where 表达式
func (c *SimpleCondition) And(express string, args ...any) *SimpleCondition {
	c.predicates = append(c.predicates, Predicate{
		Op:      OpAnd,
		Express: express,
		Args:    args,
	})
	return c
}

// Or or 语句
// express where 表达式
func (c *SimpleCondition) Or(express string, args ...any) *SimpleCondition {
	c.predicates = append(c.predicates, Predicate{
		Op:      OpOr,
		Express: express,
		Args:    args,
	})
	return c
}

func (c *SimpleCondition) getPredicates() []Predicate {
	return c.predicates
}

// Condition 条件构造器实现
type Condition struct {
	predicates []Predicate
}

func (e *Condition) getPredicates() []Predicate {
	return e.predicates
}

// And 增加 and 条件
func (e *Condition) And(cols ...Column) *Condition {
	for _, c := range cols {
		if !c.isUse {
			continue
		}
		e.predicates = append(e.predicates, Predicate{
			Op:       OpAnd,
			Express:  c.Express(),
			Args:     []any{c.arg},
			HasInSQL: c.HasInSQL(),
		})
	}
	return e
}

// Or 增加 and 条件
func (e *Condition) Or(c Column) *Condition {
	if !c.isUse {
		return e
	}
	e.predicates = append(e.predicates, Predicate{
		Op:       OpOr,
		Express:  c.Express(),
		Args:     []any{c.arg},
		HasInSQL: c.HasInSQL(),
	})
	return e
}

// C 创建 Condition 条件构造器
func C(cols ...Column) *Condition {
	ec := &Condition{}
	ec.And(cols...)
	return ec
}
