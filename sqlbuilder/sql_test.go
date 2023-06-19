package sqlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	testCases := []struct {
		name     string
		selector *Selector
		wantSQL  string
		wantErr  error
	}{
		{
			name:     "select *",
			selector: New("user").Select(),
			wantSQL:  "SELECT * FROM `user`",
		},
		{
			name: "select columns",
			selector: New("user").Select().
				Columns("id", "username"),
			wantSQL: "SELECT `id`, `username` FROM `user`",
		},
		{
			name: "select by id",
			selector: New("user").Select().
				Columns("id", "username").
				Where(C().Where(true, "id = ?")),
			wantSQL: "SELECT `id`, `username` FROM `user` WHERE id = ?",
		},
		{
			name: "select where",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?"),
				),
			wantSQL: "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE age > ? AND sex = ?",
		},
		{
			name: "select where not meet",
			selector: New("user").Select().Columns("id", "username", "age", "sex").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?").
						And(false, "username like ?"),
				),
			wantSQL: "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE age > ? AND sex = ?",
		},
		{
			name: "select order by",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?"),
				).
				OrderBy(Desc("ctime")),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC",
		},
		{
			name: "select limit",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?"),
				).
				OrderBy(Desc("ctime")).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC LIMIT 10",
		},
		{
			name: "select offset limit",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?"),
				).
				OrderBy(Desc("ctime")).
				Offset(10).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC OFFSET 10 LIMIT 10",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sql, err := tc.selector.Sql()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantSQL, sql)
		})
	}
}

func TestInsert(t *testing.T) {
	testCases := []struct {
		name        string
		inserter    *Inserter
		wantSQL     string
		wantNameSQL string
		wantErr     error
	}{
		{
			name:        "insert",
			inserter:    New("user").Insert().Columns("username", "age", "sex"),
			wantSQL:     "INSERT INTO `user`(`username`, `age`, `sex`) VALUES (?, ?, ?)",
			wantNameSQL: "INSERT INTO `user`(`username`, `age`, `sex`) VALUES (:username, :age, :sex)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantSQL != "" {
				sql, err := tc.inserter.Sql()
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
				assert.Equal(t, tc.wantSQL, sql)
			}

			if tc.wantNameSQL != "" {
				sql, err := tc.inserter.NameSql()
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
				assert.Equal(t, tc.wantNameSQL, sql)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name        string
		updater     *Updater
		wantSQL     string
		wantNameSQL string
		wantErr     error
	}{
		{
			name:        "update",
			updater:     New("user").Update().Columns("username", "age"),
			wantSQL:     "UPDATE `user` SET `username` = ?, `age` = ?",
			wantNameSQL: "UPDATE `user` SET `username` = :username, `age` = :age",
		},
		{
			name: "update where",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(C().Where(true, "id = ?")),
			wantSQL: "UPDATE `user` SET `username` = ?, `age` = ? WHERE id = ?",
		},
		{
			name: "update name where",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(C().Where(true, "id = :id")),
			wantNameSQL: "UPDATE `user` SET `username` = :username, `age` = :age WHERE id = :id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantSQL != "" {
				sql, err := tc.updater.Sql()
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
				assert.Equal(t, tc.wantSQL, sql)
			}

			if tc.wantNameSQL != "" {
				sql, err := tc.updater.NameSql()
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
				assert.Equal(t, tc.wantNameSQL, sql)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name    string
		deleter *Deleter
		wantSQL string
		wantErr error
	}{
		{
			name:    "delete",
			deleter: New("user").Delete(),
			wantSQL: "DELETE FROM `user`",
			wantErr: SQLErrDeleteMissWhere,
		},
		{
			name: "delete",
			deleter: New("user").Delete().Where(
				C().Where(true, "id = ?"),
			),
			wantSQL: "DELETE FROM `user` WHERE id = ?",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sql, err := tc.deleter.Sql()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantSQL, sql)
		})
	}
}
