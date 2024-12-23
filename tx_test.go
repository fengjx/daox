package daox_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/fengjx/daox"
	"github.com/fengjx/daox/engine"
	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
)

func TestTxManager_ExecTx(t *testing.T) {
	type testCase struct {
		name        string
		mockHandler func(mock sqlmock.Sqlmock)
		sourceFunc  func(t *testing.T, db *sql.DB) *sqlx.DB
		txFun       daox.TxFun
		wantErr     error
	}
	testCases := []testCase{
		{
			name: "commit",
			mockHandler: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `blog`").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO `blog_viewer`").WithArgs(100, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			sourceFunc: func(t *testing.T, db *sql.DB) *sqlx.DB {
				dbx := sqlx.NewDb(db, "mysql")
				return dbx
			},
			txFun: func(txCtx context.Context, executor engine.Executor) error {
				_, err := sqlbuilder.NewUpdater("blog").Execer(executor).
					Set("views", 100).
					Where(ql.C(ql.Col("id").EQ(1))).
					ExecContext(txCtx)
				if err != nil {
					t.Fatal(err)
				}
				id, affected, err := sqlbuilder.NewInserter("blog_viewer").Execer(executor).
					Columns("user_id", "blog_id").
					NamedExecContext(txCtx, map[string]any{
						"user_id": 100,
						"blog_id": 1,
					})
				if err != nil {
					t.Fatal(err)
				}
				t.Log(id, affected)
				return nil
			},
			wantErr: nil,
		},
		{
			name: "rollback",
			mockHandler: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			sourceFunc: func(t *testing.T, db *sql.DB) *sqlx.DB {
				dbx := sqlx.NewDb(db, "mysql")
				return dbx
			},
			txFun: func(txCtx context.Context, executor engine.Executor) error {
				return errors.New("rollback")
			},
			wantErr: errors.New("rollback"),
		},
	}

	doTC := func(tc testCase) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer func(db *sql.DB) { _ = db.Close() }(mockDB)
		db := tc.sourceFunc(t, mockDB)
		tc.mockHandler(mock)

		ctx := context.Background()
		manager := daox.NewTxManager(db)
		err = manager.ExecTx(ctx, func(txCtx context.Context, executor engine.Executor) error {
			return tc.txFun(txCtx, executor)
		})
		assert.Equal(t, tc.wantErr, err)
		if err != nil {
			return
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
	for _, tc := range testCases {
		doTC(tc)
	}
}
