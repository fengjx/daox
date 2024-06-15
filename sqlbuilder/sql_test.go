package sqlbuilder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
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
		selector *sqlbuilder.Selector
		wantSQL  string
		wantErr  error
		wantArgs []any
	}{
		{
			name:     "select *",
			selector: sqlbuilder.New("user").Select(),
			wantSQL:  "SELECT * FROM `user`;",
		},
		{
			name: "select columns",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username"),
			wantSQL: "SELECT `id`, `username` FROM `user`;",
		},
		{
			name: "select columns if null",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "email").
				IfNullVal("email", ""),
			wantSQL: "SELECT `id`, `username`, IFNULL(`email`, '') as `email` FROM `user`;",
		},
		{
			name: "select columns if null use map",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "email").
				IfNullVals(map[string]string{
					"email": "",
				}),
			wantSQL: "SELECT `id`, `username`, IFNULL(`email`, '') as `email` FROM `user`;",
		},
		{
			name: "select by id",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username").
				Where(ql.SC().And("`id` = ?")),
			wantSQL: "SELECT `id`, `username` FROM `user` WHERE `id` = ?;",
		},
		{
			name: "select by id use ExpressCondition",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username").
				Where(ql.C().And(ql.Col("id").EQ(1000))),
			wantSQL:  "SELECT `id`, `username` FROM `user` WHERE `id` = ?;",
			wantArgs: []any{1000},
		},
		{
			name: "select where",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex").
				Where(
					ql.SC().And("`age` > ?").
						And("`sex` = ?"),
				),
			wantSQL: "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE `age` > ? AND `sex` = ?;",
		},
		{
			name: "select where use ExpressCondition",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex").
				Where(
					ql.C().And(
						ql.Col("age").GT(20),
						ql.Col("sex").EQ("male"),
					),
				),
			wantSQL:  "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE `age` > ? AND `sex` = ?;",
			wantArgs: []any{20, "male"},
		},
		{
			name: "select where use ExpressCondition with use check",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex").
				Where(
					ql.C().And(
						ql.Col("age").GT(20),
						ql.Col("sex").EQ("male").Use(false),
					),
				),
			wantSQL:  "SELECT `id`, `username`, `age`, `sex` FROM `user` WHERE `age` > ?;",
			wantArgs: []any{20},
		},
		{
			name: "select where use ExpressCondition with no use",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex").
				Where(
					ql.C().And(
						ql.Col("age").GT(20).Use(false),
						ql.Col("sex").EQ("male").Use(false),
					),
				),
			wantSQL: "SELECT `id`, `username`, `age`, `sex` FROM `user`;",
		},
		{
			name: "select order by",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					ql.SC().And("`age` > ?").
						And("`sex` = ?"),
				).
				OrderBy(ql.Desc("ctime")),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE `age` > ? AND `sex` = ? ORDER BY `ctime` DESC;",
		},
		{
			name: "select limit",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					ql.SC().And("`age` > ?").
						And("`sex` = ?"),
				).
				OrderBy(ql.Desc("ctime")).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE `age` > ? AND `sex` = ? ORDER BY `ctime` DESC LIMIT 10;",
		},
		{
			name: "select limit offset",
			selector: sqlbuilder.New("user").Select().
				Columns("id", "username", "age", "sex", "ctime").
				Where(
					ql.SC().And("`age` > ?").
						And("`sex` = ?"),
				).
				OrderBy(ql.Desc("ctime")).
				Offset(10).
				Limit(10),
			wantSQL: "SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user` WHERE `age` > ? AND `sex` = ? ORDER BY `ctime` DESC LIMIT 10 OFFSET 10;",
		},
		{
			name: "select model",
			selector: sqlbuilder.New("user").Select().
				StructColumns(&tm{}, "json"),
			wantSQL: "SELECT `id`, `name`, `age`, `ctime` FROM `user`;",
		},
		{
			name: "select model omit",
			selector: sqlbuilder.New("user").Select().
				StructColumns(&tm{}, "json", "ctime"),
			wantSQL: "SELECT `id`, `name`, `age` FROM `user`;",
		},
		{
			name: "select count",
			selector: sqlbuilder.New("user").Select().
				QueryString("count(*)"),
			wantSQL: "SELECT count(*) FROM `user`;",
		},
		{
			name: "select group by",
			selector: sqlbuilder.New("user").Select().
				QueryString("sum(`coin`) as total, `uid`, `type`").
				GroupBy("uid", "type"),
			wantSQL: "SELECT sum(`coin`) as total, `uid`, `type` FROM `user` GROUP BY `uid`, `type`;",
		},
		{
			name: "select for update",
			selector: sqlbuilder.New("user").Select().
				Columns("uid", "nickname").
				Where(ql.SC().And("`id` = 1")).
				ForUpdate(true),
			wantSQL: "SELECT `uid`, `nickname` FROM `user` WHERE `id` = 1 FOR UPDATE ;",
		},
		{
			name: "select * with args",
			selector: sqlbuilder.New("user").Select().
				Where(ql.SC().And("`id` = ?", 100)),
			wantSQL:  "SELECT * FROM `user` WHERE `id` = ?;",
			wantArgs: []any{100},
		},
		{
			name: "select * with multiple args",
			selector: sqlbuilder.New("user").Select().
				Where(ql.SC().And("`id` IN (?, ?)", 100, 101)),
			wantSQL:  "SELECT * FROM `user` WHERE `id` IN (?, ?);",
			wantArgs: []any{100, 101},
		},
		{
			name: "select * with multiple args use ExpressCondition",
			selector: sqlbuilder.New("user").Select().
				Where(
					ql.C().And(
						ql.Col("id").In(100, 101),
						ql.Col("age").GT(20),
					),
				),
			wantSQL:  "SELECT * FROM `user` WHERE `id` IN (?, ?) AND `age` > ?;",
			wantArgs: []any{100, 101, 20},
		},
		{
			name: "select * with not in args use ExpressCondition",
			selector: sqlbuilder.New("user").Select().
				Where(
					ql.C().And(
						ql.Col("id").NotIn(100, 101),
					),
				),
			wantSQL:  "SELECT * FROM `user` WHERE `id` NOT IN (?, ?);",
			wantArgs: []any{100, 101},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sql string
			var err error
			var args []any
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
		inserter    *sqlbuilder.Inserter
		wantSQL     string
		wantNameSQL string
		wantErr     error
	}{
		{
			name:        "insert",
			inserter:    sqlbuilder.New("user").Insert().Columns("username", "age", "sex"),
			wantSQL:     "INSERT INTO `user`(`username`, `age`, `sex`) VALUES (?, ?, ?);",
			wantNameSQL: "INSERT INTO `user`(`username`, `age`, `sex`) VALUES (:username, :age, :sex);",
		},
		{
			name:        "insert model",
			inserter:    sqlbuilder.New("user").Insert().StructColumns(&tm{}, "json", "id"),
			wantSQL:     "INSERT INTO `user`(`name`, `age`, `ctime`) VALUES (?, ?, ?);",
			wantNameSQL: "INSERT INTO `user`(`name`, `age`, `ctime`) VALUES (:name, :age, :ctime);",
		},
		{
			name:        "insert on duplicate key update",
			inserter:    sqlbuilder.New("user").Insert().Columns("username", "age", "sex", "version").OnDuplicateKeyUpdateString("version = version + 1"),
			wantSQL:     "INSERT INTO `user`(`username`, `age`, `sex`, `version`) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE version = version + 1;",
			wantNameSQL: "INSERT INTO `user`(`username`, `age`, `sex`, `version`) VALUES (:username, :age, :sex, :version) ON DUPLICATE KEY UPDATE version = version + 1;",
		},
		{
			name: "replace into",
			inserter: sqlbuilder.New("user").Insert().
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
		updater     *sqlbuilder.Updater
		wantSQL     string
		wantNameSQL string
		wantErr     error
		wantArgs    []any
	}{
		{
			name:        "update",
			updater:     sqlbuilder.New("user").Update().Columns("username", "age"),
			wantSQL:     "UPDATE `user` SET `username` = ?, `age` = ?;",
			wantNameSQL: "UPDATE `user` SET `username` = :username, `age` = :age;",
		},
		{
			name: "update where",
			updater: sqlbuilder.New("user").
				Update().
				Columns("username", "age").
				Where(ql.SC().And("`id` = ?")),
			wantSQL: "UPDATE `user` SET `username` = ?, `age` = ? WHERE `id` = ?;",
		},
		{
			name: "update where use ExpressCondition",
			updater: sqlbuilder.New("user").
				Update().
				Columns("username", "age").
				Where(ql.C().And(ql.Col("id").EQ(100))),
			wantSQL:  "UPDATE `user` SET `username` = ?, `age` = ? WHERE `id` = ?;",
			wantArgs: []any{100},
		},
		{
			name: "update where with args",
			updater: sqlbuilder.New("user").
				Update().
				Columns("username", "age").
				Where(ql.SC().And("`id` = ?", 100)),
			wantSQL:  "UPDATE `user` SET `username` = ?, `age` = ? WHERE `id` = ?;",
			wantArgs: []any{100},
		},
		{
			name: "update where with multiple args",
			updater: sqlbuilder.New("user").
				Update().
				Columns("username", "age").
				Where(ql.SC().And("`id` in (?, ?)", 100, 101)),
			wantSQL:  "UPDATE `user` SET `username` = ?, `age` = ? WHERE `id` in (?, ?);",
			wantArgs: []any{100, 101},
		},
		{
			name: "update name where",
			updater: sqlbuilder.New("user").
				Update().
				Columns("username", "age").
				Where(ql.SC().And("`id` = :id")),
			wantNameSQL: "UPDATE `user` SET `username` = :username, `age` = :age WHERE `id` = :id;",
		},
		{
			name: "update with set",
			updater: sqlbuilder.New("user").
				Update().
				Set("name", "fengjx").
				Set("age", 20).
				Where(ql.C().And(ql.Col("id").EQ(1000))),
			wantSQL:  "UPDATE `user` SET `name` = ?, `age` = ? WHERE `id` = ?;",
			wantArgs: []any{"fengjx", 20, 1000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sql string
			var err error
			var args []any
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
		deleter  *sqlbuilder.Deleter
		wantSQL  string
		wantErr  error
		wantArgs []any
	}{
		{
			name:    "delete",
			deleter: sqlbuilder.New("user").Delete(),
			wantSQL: "DELETE FROM `user`;",
			wantErr: sqlbuilder.ErrDeleteMissWhere,
		},
		{
			name: "delete by id",
			deleter: sqlbuilder.New("user").Delete().Where(
				ql.SC().And("`id` = ?"),
			),
			wantSQL: "DELETE FROM `user` WHERE `id` = ?;",
		},
		{
			name: "delete by id with args",
			deleter: sqlbuilder.New("user").Delete().Where(
				ql.SC().And("`id` = ?", 100),
			),
			wantSQL:  "DELETE FROM `user` WHERE `id` = ?;",
			wantArgs: []any{100},
		},
		{
			name: "delete by where not meet with args",
			deleter: sqlbuilder.New("user").Delete().Where(
				ql.C().And(
					ql.Col("id").EQ(1000),
					ql.Col("status").EQ(1).Use(false),
				),
			),
			wantSQL:  "DELETE FROM `user` WHERE `id` = ?;",
			wantArgs: []any{1000},
		},
		{
			name: "delete by id with multiple args",
			deleter: sqlbuilder.New("user").Delete().Where(
				ql.SC().And("`id` in (?, ?)", 100, 101),
			),
			wantSQL:  "DELETE FROM `user` WHERE `id` in (?, ?);",
			wantArgs: []any{100, 101},
		},
		{
			name: "delete with limit",
			deleter: sqlbuilder.New("user").Delete().Where(
				ql.SC().And("`id` = ?"),
			).Limit(10),
			wantSQL: "DELETE FROM `user` WHERE `id` = ? LIMIT 10;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sql string
			var err error
			var args []any
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

func EqualArgs(t *testing.T, wantArgs []any, args []any) {
	for i, wantArg := range wantArgs {
		assert.Equal(t, wantArg, args[i])
	}
}
