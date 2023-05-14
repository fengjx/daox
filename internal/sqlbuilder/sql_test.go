package sqlbuilder

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
				OrderBy("ctime desc"),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY ctime desc",
		},
		{
			name: "select limit",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?"),
				).
				OrderBy("ctime desc").
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY ctime desc LIMIT 10",
		},
		{
			name: "select offset limit",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?"),
				).
				OrderBy("ctime desc").
				Offset(10).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY ctime desc OFFSET 10 LIMIT 10",
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
			sql, err := tc.inserter.Sql()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantSQL, sql)

			sql, err = tc.inserter.NameSql()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantNameSQL, sql)
		})
	}
}
