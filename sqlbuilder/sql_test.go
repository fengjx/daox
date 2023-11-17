package sqlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tm struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Age   string `json:"age"`
	Ctime int64  `json:"ctime"`
}

func TestSelect(t *testing.T) {
	testCases := []struct {
		name     string
		selector *Selector
		wantSQL  string
		wantErr  error
		wantArgs []interface{}
	}{
		{
			name:     "select *",
			selector: New("user").Select(),
			wantSQL:  "SELECT * FROM `user`;",
		},
		{
			name: "select columns",
			selector: New("user").Select().
				Columns("id", "username"),
			wantSQL: "SELECT `id`, `username` FROM `user`;",
		},
		{
			name: "select by id",
			selector: New("user").Select().
				Columns("id", "username").
				Where(SC().Where("id = ?")),
			wantSQL: "SELECT `id`, `username` FROM `user` WHERE id = ?;",
		},
		{
			name: "select where",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex").
				Where(
					SC().Where("age > ?").
						And("sex = ?"),
				),
			wantSQL: "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE age > ? AND sex = ?;",
		},
		{
			name: "select where not meet",
			selector: New("user").Select().Columns("id", "username", "age", "sex").
				Where(
					C().Where(true, "age > ?").
						And(true, "sex = ?").
						And(false, "username like ?"),
				),
			wantSQL: "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE age > ? AND sex = ?;",
		},
		{
			name: "select order by",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					SC().Where("age > ?").
						And("sex = ?"),
				).
				OrderBy(Desc("ctime")),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC;",
		},
		{
			name: "select limit",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					SC().Where("age > ?").
						And("sex = ?"),
				).
				OrderBy(Desc("ctime")).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC LIMIT 10;",
		},
		{
			name: "select limit offset",
			selector: New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					SC().Where("age > ?").
						And("sex = ?"),
				).
				OrderBy(Desc("ctime")).
				Offset(10).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC LIMIT 10 OFFSET 10;",
		},
		{
			name: "select model",
			selector: New("user").Select().
				StructColumns(&tm{}, "json"),
			wantSQL: "SELECT `id`, `name`, `age`, `ctime` FROM `user`;",
		},
		{
			name: "select model omit",
			selector: New("user").Select().
				StructColumns(&tm{}, "json", "ctime"),
			wantSQL: "SELECT `id`, `name`, `age` FROM `user`;",
		},
		{
			name: "select count",
			selector: New("user").Select().
				QueryString("count(*)"),
			wantSQL: "SELECT count(*) FROM `user`;",
		},
		{
			name: "select group by",
			selector: New("user").Select().
				QueryString("sum(coin) as total, uid, type").
				GroupBy("uid", "type"),
			wantSQL: "SELECT sum(coin) as total, uid, type FROM `user` GROUP BY `uid`, `type`;",
		},
		{
			name: "select for update",
			selector: New("user").Select().
				Columns("uid", "nickname").
				Where(SC().Where("id = 1")).
				ForUpdate(true),
			wantSQL: "SELECT `uid`, `nickname` FROM `user` WHERE id = 1 FOR UPDATE ;",
		},
		{
			name: "select * with args",
			selector: New("user").Select().
				Where(SC().Where("id = ?", 100)),
			wantSQL:  "SELECT * FROM `user` WHERE id = ?;",
			wantArgs: []interface{}{100},
		},
		{
			name: "select * with multiple args",
			selector: New("user").Select().
				Where(SC().Where("id in (?, ?)", 100, 101)),
			wantSQL:  "SELECT * FROM `user` WHERE id in (?, ?);",
			wantArgs: []interface{}{100, 101},
		},
		{
			name: "select where not meet with args",
			selector: New("user").Select().Columns("id", "username", "age", "sex").
				Where(
					C().Where(true, "age > ?", 18).
						And(true, "sex = ?", 1).
						And(false, "username like ?", "%hello%"),
				),
			wantSQL:  "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE age > ? AND sex = ?;",
			wantArgs: []interface{}{18, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sql string
			var err error
			var args []interface{}
			if tc.wantArgs != nil {
				sql, args, err = tc.selector.SQLArgs()
				EqualArgs(t, tc.wantArgs, args)
			} else {
				sql, err = tc.selector.SQL()
			}
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
			wantSQL:     "INSERT INTO `user`(`username`, `age`, `sex`) VALUES (?, ?, ?);",
			wantNameSQL: "INSERT INTO `user`(`username`, `age`, `sex`) VALUES (:username, :age, :sex);",
		},
		{
			name:        "insert model",
			inserter:    New("user").Insert().StructColumns(&tm{}, "json", "id"),
			wantSQL:     "INSERT INTO `user`(`name`, `age`, `ctime`) VALUES (?, ?, ?);",
			wantNameSQL: "INSERT INTO `user`(`name`, `age`, `ctime`) VALUES (:name, :age, :ctime);",
		},
		{
			name:        "insert on duplicate key update",
			inserter:    New("user").Insert().Columns("username", "age", "sex", "version").OnDuplicateKeyUpdateString("version = version + 1"),
			wantSQL:     "INSERT INTO `user`(`username`, `age`, `sex`, `version`) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE version = version + 1;",
			wantNameSQL: "INSERT INTO `user`(`username`, `age`, `sex`, `version`) VALUES (:username, :age, :sex, :version) ON DUPLICATE KEY UPDATE version = version + 1;",
		},
		{
			name: "replace into",
			inserter: New("user").Insert().
				Columns("username", "age", "sex").
				IsReplaceInto(true),
			wantSQL:     "REPLACE INTO `user`(`username`, `age`, `sex`) VALUES (?, ?, ?);",
			wantNameSQL: "REPLACE INTO `user`(`username`, `age`, `sex`) VALUES (:username, :age, :sex);",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantSQL != "" {
				sql, err := tc.inserter.SQL()
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
				assert.Equal(t, tc.wantSQL, sql)
			}

			if tc.wantNameSQL != "" {
				sql, err := tc.inserter.NameSQL()
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
		wantArgs    []interface{}
	}{
		{
			name:        "update",
			updater:     New("user").Update().Columns("username", "age"),
			wantSQL:     "UPDATE `user` SET `username` = ?, `age` = ?;",
			wantNameSQL: "UPDATE `user` SET `username` = :username, `age` = :age;",
		},
		{
			name: "update where",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(SC().Where("id = ?")),
			wantSQL: "UPDATE `user` SET `username` = ?, `age` = ? WHERE id = ?;",
		},
		{
			name: "update where with args",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(SC().Where("id = ?", 100)),
			wantSQL:  "UPDATE `user` SET `username` = ?, `age` = ? WHERE id = ?;",
			wantArgs: []interface{}{100},
		},
		{
			name: "update where with multiple args",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(SC().Where("id in (?, ?)", 100, 101)),
			wantSQL:  "UPDATE `user` SET `username` = ?, `age` = ? WHERE id in (?, ?);",
			wantArgs: []interface{}{100, 101},
		},
		{
			name: "update where not meet with args",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(
					C().Where(true, "id = ?", 100).
						And(false, "status = ?", 1),
				),
			wantSQL:  "UPDATE `user` SET `username` = ?, `age` = ? WHERE id = ?;",
			wantArgs: []interface{}{100},
		},
		{
			name: "update name where",
			updater: New("user").
				Update().
				Columns("username", "age").
				Where(SC().Where("id = :id")),
			wantNameSQL: "UPDATE `user` SET `username` = :username, `age` = :age WHERE id = :id;",
		},
		{
			name: "update model",
			updater: New("user").
				Update().
				StructColumns(&tm{}, "json", "id", "ctime").
				Where(SC().Where("id = :id")),
			wantNameSQL: "UPDATE `user` SET `name` = :name, `age` = :age WHERE id = :id;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sql string
			var err error
			var args []interface{}
			if tc.wantArgs != nil {
				sql, args, err = tc.updater.SQLArgs()
				EqualArgs(t, tc.wantArgs, args)
			}
			if tc.wantSQL != "" {
				sql, err = tc.updater.SQL()
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
				assert.Equal(t, tc.wantSQL, sql)
			}

			if tc.wantNameSQL != "" {
				sql, err = tc.updater.NameSQL()
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
		name     string
		deleter  *Deleter
		wantSQL  string
		wantErr  error
		wantArgs []interface{}
	}{
		{
			name:    "delete",
			deleter: New("user").Delete(),
			wantSQL: "DELETE FROM `user`;",
			wantErr: SQLErrDeleteMissWhere,
		},
		{
			name: "delete by id",
			deleter: New("user").Delete().Where(
				C().Where(true, "id = ?"),
			),
			wantSQL: "DELETE FROM `user` WHERE id = ?;",
		},
		{
			name: "delete by id with args",
			deleter: New("user").Delete().Where(
				C().Where(true, "id = ?", 100),
			),
			wantSQL:  "DELETE FROM `user` WHERE id = ?;",
			wantArgs: []interface{}{100},
		},
		{
			name: "delete by where not meet with args",
			deleter: New("user").Delete().Where(
				C().Where(true, "id = ?", 100).And(false, "status = ?", 1),
			),
			wantSQL:  "DELETE FROM `user` WHERE id = ?;",
			wantArgs: []interface{}{100},
		},
		{
			name: "delete by id with multiple args",
			deleter: New("user").Delete().Where(
				C().Where(true, "id in (?, ?)", 100, 101),
			),
			wantSQL:  "DELETE FROM `user` WHERE id in (?, ?);",
			wantArgs: []interface{}{100, 101},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sql string
			var err error
			var args []interface{}
			if tc.wantArgs != nil {
				sql, args, err = tc.deleter.SQLArgs()
				EqualArgs(t, tc.wantArgs, args)
			} else {
				sql, err = tc.deleter.SQL()
			}
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantSQL, sql)
		})
	}
}

func EqualArgs(t *testing.T, wantArgs []interface{}, args []interface{}) {
	for i, wantArg := range wantArgs {
		assert.Equal(t, wantArg, args[i])
	}
}
