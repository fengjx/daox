package engine

import "context"

// Queryer 查询语句执行器
type Queryer interface {

	// Select 查询多条数据
	Select(dest any, query string, args ...any) error
	// SelectContext 查询多条数据
	SelectContext(ctx context.Context, dest any, query string, args ...any) error

	// Get 查询单条数据
	Get(dest any, query string, args ...any) error
	// GetContext 查询单条数据
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}