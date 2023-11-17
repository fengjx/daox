package sqlbuilder

type op struct {
	symbol string
	text   string
}

var (
	OpAnd = &op{symbol: "AND", text: " AND "}
	OpOr  = &op{symbol: "OR", text: " OR "}
)

type Predicate struct {
	Op      *op
	Express string
	Args    []interface{}
}

type ConditionBuilder interface {
	getPredicates() []*Predicate
}

type Condition struct {
	predicates []*Predicate
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
	c.predicates = append(c.predicates, &Predicate{
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
	c.predicates = append(c.predicates, &Predicate{
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
	c.predicates = append(c.predicates, &Predicate{
		Op:      OpOr,
		Express: express,
		Args:    args,
	})
	return c
}

func (c *Condition) getPredicates() []*Predicate {
	return c.predicates
}

// SimpleCondition 简单 where 语句
type SimpleCondition struct {
	predicates []*Predicate
}

// SC 简单 where 条件
func SC() *SimpleCondition {
	return new(SimpleCondition)
}

func (c *SimpleCondition) Predicates() []*Predicate {
	return c.predicates
}

// Where
// express where 表达式
func (c *SimpleCondition) Where(express string, args ...interface{}) *SimpleCondition {
	c.predicates = append(c.predicates, &Predicate{
		Express: express,
		Args:    args,
	})
	return c
}

// And and 语句
// express where 表达式
func (c *SimpleCondition) And(express string, args ...interface{}) *SimpleCondition {
	c.predicates = append(c.predicates, &Predicate{
		Op:      OpAnd,
		Express: express,
		Args:    args,
	})
	return c
}

// Or or 语句
// express where 表达式
func (c *SimpleCondition) Or(express string, args ...interface{}) *SimpleCondition {
	c.predicates = append(c.predicates, &Predicate{
		Op:      OpOr,
		Express: express,
		Args:    args,
	})
	return c
}

func (c *SimpleCondition) getPredicates() []*Predicate {
	return c.predicates
}
