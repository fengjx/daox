package daox

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txCtxKey struct{}

// TxFun 事务处理函数
type TxFun func(txCtx context.Context, tx *sqlx.Tx) error

// TxManager 事务管理器
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager 创建事务管理器
func NewTxManager(db *sqlx.DB) *TxManager {
	m := &TxManager{
		db: db,
	}
	return m
}

func (m *TxManager) withTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func (m *TxManager) getTx(ctx context.Context) *sqlx.Tx {
	tx, ok := ctx.Value(txCtxKey{}).(*sqlx.Tx)
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
