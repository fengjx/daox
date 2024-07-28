package engine

import (
	"context"
	"database/sql"
)

// Queryer 查询语句执行器
type Queryer interface {

	// SelectContext 查询多条数据
	SelectContext(ctx context.Context, dest any, query string, args ...any) error

	// GetContext 查询单条数据
	GetContext(ctx context.Context, dest any, query string, args ...any) error

	// QueryContext 查询多条数据，返回 sql.Rows
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}
