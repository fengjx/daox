package daox

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/v2/engine"
)

type Tx struct {
	*sqlx.Tx
	hook engine.Hook
}

// NamedExecContext 使用命名参数执行sql
func (t *Tx) NamedExecContext(ctx context.Context, execSQL string, arg any) (sql.Result, error) {
	return doNamedExec(ctx, t.Tx, execSQL, arg, t.hook)
}

// ExecContext 使用数组参数执行sql
func (t *Tx) ExecContext(ctx context.Context, execSQL string, args ...any) (sql.Result, error) {
	return doExec(ctx, t.Tx, execSQL, args, t.hook)
}

// SelectContext 查询多条数据
func (t *Tx) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return doSelect(ctx, t.Tx, dest, query, args, t.hook)
}

// GetContext 查询单条数据
func (t *Tx) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return doGet(ctx, t.Tx, dest, query, args, t.hook)
}

// QueryContext 查询多条数据，返回 sql.Rows
func (t *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return doQueryContext(ctx, t.Tx, query, args, t.hook)
}

type txCtxKey struct{}

// TxFun 事务处理函数
type TxFun func(txCtx context.Context, executor engine.Executor) error

// TxManager 事务管理器
type TxManager struct {
	db *DB
}

// NewTxManager 创建事务管理器
func NewTxManager(db *sqlx.DB) *TxManager {
	m := &TxManager{
		db: NewDb(db),
	}
	return m
}

func (m *TxManager) withTx(ctx context.Context, tx *Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func (m *TxManager) getTx(ctx context.Context) *Tx {
	val := ctx.Value(txCtxKey{})
	if val == nil {
		return nil
	}
	tx, ok := val.(*Tx)
	if !ok {
		return nil
	}
	return tx
}

// ExecTx 事务处理
func (m *TxManager) ExecTx(ctx context.Context, fn TxFun) (err error) {
	tx := m.getTx(ctx)
	if tx != nil {
		return fn(ctx, tx)
	}
	tx, err = m.db.Beginx()
	if err != nil {
		return err
	}
	ctx = m.withTx(ctx, tx)
	defer func() {
		if perr := recover(); perr != nil {
			_ = tx.Rollback()
			// 对外抛出 panic
			panic(perr)
		}
		if err != nil {
			// 这里不给 err 赋值，因为希望返回回滚前的原始 err
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	err = fn(ctx, tx)
	return
}
