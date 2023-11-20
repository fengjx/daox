package sqlbuilder

// Predicate where 断言
type Predicate struct {
	Op       Op
	Express  string
	Args     []interface{}
	HasInSQL bool
}

// ConditionBuilder 条件构造器
type ConditionBuilder interface {
	getPredicates() []Predicate
}

// Condition 条件构造器实现
type Condition struct {
	predicates []Predicate
}

// C where 条件
func C() *Condition {
	return new(Condition)
}

// Where
// meet 判断是否需要拼接这个where表达式
// express where 表达式
func (c *Condition) Where(meet bool, express string, args ...interface{}) *Condition {
	if !meet {
		return c
	}
	c.predicates = append(c.predicates, Predicate{
		Op:      emptyOp,
		Express: express,
		Args:    args,
	})
	return c
}

// And and 语句
// meet 判断是否需要拼接这个where表达式
// express where 表达式
func (c *Condition) And(meet bool, express string, args ...interface{}) *Condition {
	if !meet {
		return c
	}
	c.predicates = append(c.predicates, Predicate{
		Op:      OpAnd,
		Express: express,
		Args:    args,
	})
	return c
}

// Or or 语句
// meet 判断是否需要拼接这个where表达式
// express where 表达式
func (c *Condition) Or(meet bool, express string, args ...interface{}) *Condition {
	if !meet {
		return c
	}
	c.predicates = append(c.predicates, Predicate{
		Op:      OpOr,
		Express: express,
		Args:    args,
	})
	return c
}

func (c *Condition) getPredicates() []Predicate {
	return c.predicates
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

// Where
// express where 表达式
func (c *SimpleCondition) Where(express string, args ...interface{}) *SimpleCondition {
	c.predicates = append(c.predicates, Predicate{
		Op:      emptyOp,
		Express: express,
		Args:    args,
	})
	return c
}

// And and 语句
// express where 表达式
func (c *SimpleCondition) And(express string, args ...interface{}) *SimpleCondition {
	c.predicates = append(c.predicates, Predicate{
		Op:      OpAnd,
		Express: express,
		Args:    args,
	})
	return c
}

// Or or 语句
// express where 表达式
func (c *SimpleCondition) Or(express string, args ...interface{}) *SimpleCondition {
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

// ExpressCondition 基于表达式的条件构造
type ExpressCondition struct {
	predicates []Predicate
}

func (e *ExpressCondition) getPredicates() []Predicate {
	return e.predicates
}

// Where 增加 and 条件
func (e *ExpressCondition) Where(cols ...Column) *ExpressCondition {
	for _, c := range cols {
		if !c.isUse {
			continue
		}
		e.predicates = append(e.predicates, Predicate{
			Op:       OpAnd,
			Express:  c.Express(),
			Args:     []interface{}{c.arg},
			HasInSQL: c.HasInSQL(),
		})
	}
	return e
}

// And 增加 and 条件
func (e *ExpressCondition) And(cols ...Column) *ExpressCondition {
	for _, c := range cols {
		if !c.isUse {
			continue
		}
		e.predicates = append(e.predicates, Predicate{
			Op:       OpAnd,
			Express:  c.Express(),
			Args:     []interface{}{c.arg},
			HasInSQL: c.HasInSQL(),
		})
	}
	return e
}

// Or 增加 and 条件
func (e *ExpressCondition) Or(c Column) *ExpressCondition {
	if !c.isUse {
		return e
	}
	e.predicates = append(e.predicates, Predicate{
		Op:       OpOr,
		Express:  c.Express(),
		Args:     []interface{}{c.arg},
		HasInSQL: c.HasInSQL(),
	})
	return e
}

// EC 创建一个使用?占位符的 ExpressCondition 条件构造器
func EC() *ExpressCondition {
	ec := &ExpressCondition{}
	return ec
}
