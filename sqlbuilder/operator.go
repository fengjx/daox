package sqlbuilder

type Op struct {
	Symbol string
	Text   string
}

var emptyOp = Op{}

var (
	OpLT      = Op{Symbol: "<", Text: " < "}
	OpLTEQ    = Op{Symbol: "<=", Text: " <= "}
	OpGT      = Op{Symbol: ">", Text: " > "}
	OpGTEQ    = Op{Symbol: ">=", Text: " >= "}
	OpEQ      = Op{Symbol: "=", Text: " = "}
	OpNEQ     = Op{Symbol: "!=", Text: " != "}
	OpAnd     = Op{Symbol: "AND", Text: " AND "}
	OpOr      = Op{Symbol: "OR", Text: " OR "}
	OpNot     = Op{Symbol: "NOT", Text: " NOT "}
	OpIn      = Op{Symbol: "IN", Text: " IN "}
	OpNotIN   = Op{Symbol: "NOT IN", Text: " NOT IN "}
	OpLike    = Op{Symbol: "LIKE", Text: " LIKE "}
	OpNotLike = Op{Symbol: "NOT LIKE", Text: " NOT LIKE "}
	OpExist   = Op{Symbol: "EXIST", Text: " EXIST "}
)
