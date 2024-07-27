package engine

// Executor sql 语句执行器。Execer 和 Queryer 的组合
type Executor interface {
	Execer
	Queryer
}
