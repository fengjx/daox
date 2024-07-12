package adapter

import (
	"context"
	"database/sql"
)

// Execer exec sql
type Execer interface {
	NamedExecContext(ctx context.Context, execSQL string, arg interface{}) (sql.Result, error)

	ExecContext(ctx context.Context, execSQL string, args ...any) (sql.Result, error)
}
