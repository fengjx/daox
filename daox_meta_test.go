package daox_test

import (
	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
)

const createMySQLTableSQL = `
create table if not exists demo_info
(
    id         bigint auto_increment,
    uid        bigint,
    name       varchar(32) default '',
    sex        varchar(12) default '',
    login_time bigint      default 0,
    utime      bigint      default 0,
    ctime      bigint      default 0,
    primary key pk (id),
    unique uni_uid (uid)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_bin;
`

const createSqliteTableSQL = `
CREATE TABLE %s (
  id integer primary key autoincrement,
  uid integer,
  name text,
  sex text,
  login_time integer,
  utime integer,
  ctime integer
);
`

// DemoInfo
type DemoInfo struct {
	ID        int64  `json:"id"`         // -
	UID       int64  `json:"uid"`        // -
	Name      string `json:"name"`       // -
	Sex       string `json:"sex"`        // -
	LoginTime int64  `json:"login_time"` // -
	Utime     int64  `json:"utime"`      // -
	Ctime     int64  `json:"ctime"`      // -
}

func (m *DemoInfo) GetID() any {
	return m.ID
}

// DemoInfoM
type DemoInfoM struct {
	ID        string
	UID       string
	Name      string
	Sex       string
	LoginTime string
	Utime     string
	Ctime     string
}

func (m DemoInfoM) TableName() string {
	return "demo_info"
}

func (m DemoInfoM) PrimaryKey() string {
	return "id"
}

func (m DemoInfoM) IsAutoIncrement() bool {
	return true
}

func (m DemoInfoM) Columns() []string {
	return []string{
		"id",
		"uid",
		"name",
		"sex",
		"login_time",
		"utime",
		"ctime",
	}
}

var DemoInfoMeta = DemoInfoM{
	ID:        "id",
	UID:       "uid",
	Name:      "name",
	Sex:       "sex",
	LoginTime: "login_time",
	Utime:     "utime",
	Ctime:     "ctime",
}

func (m DemoInfoM) IdIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.ID).In(args...)
}

func (m DemoInfoM) IdNotIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.ID).NotIn(args...)
}

func (m DemoInfoM) IdEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).EQ(val)
}

func (m DemoInfoM) IdNotEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).NotEQ(val)
}

func (m DemoInfoM) IdLT(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).LT(val)
}

func (m DemoInfoM) IdLTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).LTEQ(val)
}

func (m DemoInfoM) IdGT(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).GT(val)
}

func (m DemoInfoM) IdGTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).GTEQ(val)
}

func (m DemoInfoM) IdLike(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).Like(val)
}

func (m DemoInfoM) IdNotLike(val int64) sqlbuilder.Column {
	return ql.Col(m.ID).NotLike(val)
}

func (m DemoInfoM) IdDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.ID)
}

func (m DemoInfoM) IdAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.ID)
}

func (m DemoInfoM) UidIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.UID).In(args...)
}

func (m DemoInfoM) UidNotIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.UID).NotIn(args...)
}

func (m DemoInfoM) UidEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).EQ(val)
}

func (m DemoInfoM) UidNotEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).NotEQ(val)
}

func (m DemoInfoM) UidLT(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).LT(val)
}

func (m DemoInfoM) UidLTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).LTEQ(val)
}

func (m DemoInfoM) UidGT(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).GT(val)
}

func (m DemoInfoM) UidGTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).GTEQ(val)
}

func (m DemoInfoM) UidLike(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).Like(val)
}

func (m DemoInfoM) UidNotLike(val int64) sqlbuilder.Column {
	return ql.Col(m.UID).NotLike(val)
}

func (m DemoInfoM) UidDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.UID)
}

func (m DemoInfoM) UidAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.UID)
}

func (m DemoInfoM) NameIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Name).In(args...)
}

func (m DemoInfoM) NameNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Name).NotIn(args...)
}

func (m DemoInfoM) NameEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Name).EQ(val)
}

func (m DemoInfoM) NameNotEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Name).NotEQ(val)
}

func (m DemoInfoM) NameLT(val string) sqlbuilder.Column {
	return ql.Col(m.Name).LT(val)
}

func (m DemoInfoM) NameLTEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Name).LTEQ(val)
}

func (m DemoInfoM) NameGT(val string) sqlbuilder.Column {
	return ql.Col(m.Name).GT(val)
}

func (m DemoInfoM) NameGTEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Name).GTEQ(val)
}

func (m DemoInfoM) NameLike(val string) sqlbuilder.Column {
	return ql.Col(m.Name).Like(val)
}

func (m DemoInfoM) NameNotLike(val string) sqlbuilder.Column {
	return ql.Col(m.Name).NotLike(val)
}

func (m DemoInfoM) NameDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.Name)
}

func (m DemoInfoM) NameAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.Name)
}

func (m DemoInfoM) SexIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Sex).In(args...)
}

func (m DemoInfoM) SexNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Sex).NotIn(args...)
}

func (m DemoInfoM) SexEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).EQ(val)
}

func (m DemoInfoM) SexNotEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).NotEQ(val)
}

func (m DemoInfoM) SexLT(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).LT(val)
}

func (m DemoInfoM) SexLTEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).LTEQ(val)
}

func (m DemoInfoM) SexGT(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).GT(val)
}

func (m DemoInfoM) SexGTEQ(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).GTEQ(val)
}

func (m DemoInfoM) SexLike(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).Like(val)
}

func (m DemoInfoM) SexNotLike(val string) sqlbuilder.Column {
	return ql.Col(m.Sex).NotLike(val)
}

func (m DemoInfoM) SexDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.Sex)
}

func (m DemoInfoM) SexAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.Sex)
}

func (m DemoInfoM) LoginTimeIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.LoginTime).In(args...)
}

func (m DemoInfoM) LoginTimeNotIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.LoginTime).NotIn(args...)
}

func (m DemoInfoM) LoginTimeEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).EQ(val)
}

func (m DemoInfoM) LoginTimeNotEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).NotEQ(val)
}

func (m DemoInfoM) LoginTimeLT(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).LT(val)
}

func (m DemoInfoM) LoginTimeLTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).LTEQ(val)
}

func (m DemoInfoM) LoginTimeGT(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).GT(val)
}

func (m DemoInfoM) LoginTimeGTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).GTEQ(val)
}

func (m DemoInfoM) LoginTimeLike(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).Like(val)
}

func (m DemoInfoM) LoginTimeNotLike(val int64) sqlbuilder.Column {
	return ql.Col(m.LoginTime).NotLike(val)
}

func (m DemoInfoM) LoginTimeDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.LoginTime)
}

func (m DemoInfoM) LoginTimeAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.LoginTime)
}

func (m DemoInfoM) UtimeIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Utime).In(args...)
}

func (m DemoInfoM) UtimeNotIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Utime).NotIn(args...)
}

func (m DemoInfoM) UtimeEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).EQ(val)
}

func (m DemoInfoM) UtimeNotEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).NotEQ(val)
}

func (m DemoInfoM) UtimeLT(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).LT(val)
}

func (m DemoInfoM) UtimeLTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).LTEQ(val)
}

func (m DemoInfoM) UtimeGT(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).GT(val)
}

func (m DemoInfoM) UtimeGTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).GTEQ(val)
}

func (m DemoInfoM) UtimeLike(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).Like(val)
}

func (m DemoInfoM) UtimeNotLike(val int64) sqlbuilder.Column {
	return ql.Col(m.Utime).NotLike(val)
}

func (m DemoInfoM) UtimeDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.Utime)
}

func (m DemoInfoM) UtimeAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.Utime)
}

func (m DemoInfoM) CtimeIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Ctime).In(args...)
}

func (m DemoInfoM) CtimeNotIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(m.Ctime).NotIn(args...)
}

func (m DemoInfoM) CtimeEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).EQ(val)
}

func (m DemoInfoM) CtimeNotEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).NotEQ(val)
}

func (m DemoInfoM) CtimeLT(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).LT(val)
}

func (m DemoInfoM) CtimeLTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).LTEQ(val)
}

func (m DemoInfoM) CtimeGT(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).GT(val)
}

func (m DemoInfoM) CtimeGTEQ(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).GTEQ(val)
}

func (m DemoInfoM) CtimeLike(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).Like(val)
}

func (m DemoInfoM) CtimeNotLike(val int64) sqlbuilder.Column {
	return ql.Col(m.Ctime).NotLike(val)
}

func (m DemoInfoM) CtimeDesc() sqlbuilder.OrderBy {
	return ql.Desc(m.Ctime)
}

func (m DemoInfoM) CtimeAsc() sqlbuilder.OrderBy {
	return ql.Asc(m.Ctime)
}
