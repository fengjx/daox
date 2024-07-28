package engine

import (
	"context"
	"database/sql"
)

// Execer 更新语句执行器
type Execer interface {
	// NamedExecContext 使用命名参数执行sql
	NamedExecContext(ctx context.Context, execSQL string, arg any) (sql.Result, error)

	// ExecContext 使用数组参数执行sql
	ExecContext(ctx context.Context, execSQL string, args ...any) (sql.Result, error)
}
