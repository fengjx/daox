package daox

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/engine"
	"github.com/fengjx/daox/utils"
)

// DB 包装 sqlx.DB
type DB struct {
	*sqlx.DB
	middleware engine.Middleware
}

// NewDb 创建 DB
func NewDb(db *sqlx.DB, middlewares ...engine.Middleware) *DB {
	m := engine.NewChain(middlewares...)
	ndb := &DB{
		DB:         db,
		middleware: m,
	}
	return ndb
}

// NamedExecContext 使用命名参数执行sql
func (d *DB) NamedExecContext(ctx context.Context, execSQL string, arg any) (sql.Result, error) {
	return doNamedExec(ctx, d.DB, execSQL, arg, d.middleware)
}

// ExecContext 使用数组参数执行sql
func (d *DB) ExecContext(ctx context.Context, execSQL string, args ...any) (sql.Result, error) {
	return doExec(ctx, d.DB, execSQL, args, d.middleware)
}

// SelectContext 查询多条数据
func (d *DB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return doSelect(ctx, d.DB, dest, query, args, d.middleware)
}

// GetContext 查询单条数据
func (d *DB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return doGet(ctx, d.DB, dest, query, args, d.middleware)
}

// Beginx 打开一个事务
func (d *DB) Beginx() (*Tx, error) {
	tx, err := d.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return &Tx{
		Tx:         tx,
		middleware: d.middleware,
	}, nil
}

func doNamedExec(ctx context.Context, execer engine.Execer, execSQL string, arg any, middleware engine.Middleware) (sql.Result, error) {
	if middleware == nil {
		return execer.NamedExecContext(ctx, execSQL, arg)
	}
	ec := engine.ExecutorCtx(ctx)
	if ec == nil {
		ec = &engine.ExecutorContext{
			Type:      engine.ParseSQLType(execSQL),
			TableName: engine.ParseTableName(execSQL),
			SQL:       execSQL,
			NameArgs:  arg,
			Start:     time.Now(),
		}
	}
	err := middleware.Before(ctx, ec)
	if err != nil {
		return nil, err
	}
	result, err := execer.NamedExecContext(ctx, execSQL, arg)
	er := &engine.ExecutorResult{
		Err:      err,
		Duration: time.Since(ec.Start),
	}
	if result != nil {
		affected, _ := result.RowsAffected()
		er.Affected = affected
	}
	middleware.After(ctx, ec, er)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func doExec(ctx context.Context, execer engine.Execer, execSQL string, args []any, middleware engine.Middleware) (sql.Result, error) {
	if middleware == nil {
		return execer.ExecContext(ctx, execSQL, args...)
	}
	ec := engine.ExecutorCtx(ctx)
	if ec == nil {
		ec = &engine.ExecutorContext{
			Type:      engine.ParseSQLType(execSQL),
			TableName: engine.ParseTableName(execSQL),
			SQL:       execSQL,
			Args:      args,
			Start:     time.Now(),
		}
	}
	err := middleware.Before(ctx, ec)
	if err != nil {
		return nil, err
	}
	result, err := execer.ExecContext(ctx, execSQL, args...)
	er := &engine.ExecutorResult{
		Err:      err,
		Duration: time.Since(ec.Start),
	}
	if result != nil {
		affected, _ := result.RowsAffected()
		er.Affected = affected
	}
	middleware.After(ctx, ec, er)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func doSelect(ctx context.Context, queryer engine.Queryer, dest any, query string, args []any, middleware engine.Middleware) error {
	if middleware == nil {
		return queryer.SelectContext(ctx, dest, query, args...)
	}
	ec := engine.ExecutorCtx(ctx)
	if ec == nil {
		ec = &engine.ExecutorContext{
			Type:      engine.ParseSQLType(query),
			TableName: engine.ParseTableName(query),
			SQL:       query,
			Args:      args,
			Start:     time.Now(),
		}
	}
	err := middleware.Before(ctx, ec)
	if err != nil {
		return err
	}
	err = queryer.SelectContext(ctx, dest, query, args...)
	er := &engine.ExecutorResult{
		Err:      err,
		Duration: time.Since(ec.Start),
	}
	if err == nil {
		er.QueryRows = int64(utils.GetLength(dest))
	}
	middleware.After(ctx, ec, er)
	return err
}

func doGet(ctx context.Context, queryer engine.Queryer, dest any, query string, args []any, middleware engine.Middleware) error {
	if middleware == nil {
		return queryer.GetContext(ctx, dest, query, args...)
	}
	ec := engine.ExecutorCtx(ctx)
	if ec == nil {
		ec = &engine.ExecutorContext{
			Type:      engine.ParseSQLType(query),
			TableName: engine.ParseTableName(query),
			SQL:       query,
			Args:      args,
			Start:     time.Now(),
		}
	}
	err := middleware.Before(ctx, ec)
	if err != nil {
		return err
	}
	err = queryer.GetContext(ctx, dest, query, args...)
	er := &engine.ExecutorResult{
		Err:      err,
		Duration: time.Since(ec.Start),
	}
	if err == nil {
		er.QueryRows = 1
	}
	middleware.After(ctx, ec, er)
	return err
}
