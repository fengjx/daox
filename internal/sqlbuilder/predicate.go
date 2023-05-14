package sqlbuilder

type op struct {
	symbol string
	text   string
}

var (
	opAnd = &op{symbol: "AND", text: " AND "}
	opOr  = &op{symbol: "OR", text: " OR "}
)

type Predicate struct {
	op      *op
	express string
}

type condition struct {
	predicates []*Predicate
}

func C() *condition {
	return new(condition)
}

func (c *condition) Predicates() []*Predicate {
	return c.predicates
}

func (c *condition) Where(meet bool, express string) *condition {
	if !meet {
		return c
	}
	c.predicates = append(c.predicates, &Predicate{
		express: express,
	})
	return c
}

func (c *condition) And(meet bool, express string) *condition {
	if !meet {
		return c
	}
	c.predicates = append(c.predicates, &Predicate{
		op:      opAnd,
		express: express,
	})
	return c
}

func (c *condition) Or(meet bool, express string) *condition {
	if !meet {
		return c
	}
	c.predicates = append(c.predicates, &Predicate{
		op:      opOr,
		express: express,
	})
	return c
}
