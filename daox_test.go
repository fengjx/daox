package daox_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/engine"
	"github.com/fengjx/daox/v2/sqlbuilder/ql"
)

func sqliteDB() (*sql.DB, error) {
	err := os.MkdirAll(".db", 0755)
	if err != nil {
		panic(err)
	}
	return sql.Open("sqlite3", "./.db/test.db")
}

func newMySQLDb() *sqlx.DB {
	dbx := sqlx.MustOpen("mysql", "root:1234@tcp(192.168.1.200:3306)/fjx?charset=utf8mb4,utf8&tls=false&timeout=10s")
	dbx.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return dbx
}

func newSqliteDb() *sqlx.DB {
	db, err := sqliteDB()
	if err != nil {
		panic(err)
	}
	dbx := sqlx.NewDb(db, "sqlite3")
	dbx.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return dbx
}

func newDb() *sqlx.DB {
	return newSqliteDb()
}

func newMockDB() (dbx *sqlx.DB, mock sqlmock.Sqlmock, err error) {
	var db *sql.DB
	db, mock, err = sqlmock.New()
	if err != nil {
		return
	}
	dbx = sqlx.NewDb(db, "mysql")
	return
}

// Blog meta 定义
// 用于泛型 Dao
// -------------------
type BlogM struct {
	Id         string
	Uid        string
	Title      string
	Content    string
	CreateTime string
}

func (m BlogM) TableName() string     { return "blog" }
func (m BlogM) PrimaryKey() string    { return "id" }
func (m BlogM) IsAutoIncrement() bool { return true }
func (m BlogM) Columns() []string {
	return []string{"id", "uid", "title", "content", "create_time"}
}

var BlogMeta = BlogM{
	Id: "id", Uid: "uid", Title: "title", Content: "content", CreateTime: "create_time",
}

func before(t *testing.T, tableName string) {
	after(t, tableName)
	t.Log("before...", tableName)
	db := newDb()
	_, err := db.Exec(fmt.Sprintf(createSqliteTableSQL, tableName))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("create table success", tableName)
	daox.UseDefaultMasterDB(db)
	// 这里 DemoInfoMeta 适配泛型 Dao
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta).WithTableName(tableName)
	for i := 0; i < 10; i++ {
		nowSec := time.Now().Unix()
		data := &DemoInfo{
			UID:       int64(100 + i),
			Name:      fmt.Sprintf("u-%d", i),
			Sex:       "male",
			LoginTime: nowSec,
			Utime:     nowSec,
			Ctime:     nowSec,
		}
		id, err := dao.Save(data, daox.DisableGlobalInsertOmits(true))
		if err != nil {
			t.Log("save err", data, err)
			panic(err)
		}
		t.Logf("save id - %d \r", id)
	}
}

func after(t *testing.T, tableName string) {
	db := newDb()
	t.Log("drop table", tableName)
	_, err := db.Exec(fmt.Sprintf("drop table if exists %s", tableName))
	if err != nil {
		t.Fatal(err)
	}
}

func testCreate(t *testing.T) {
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	assert.Equal(t, len(dao.TableMeta.Columns), 7)
	assert.Equal(t, dao.TableMeta.PrimaryKey, "id")
	for _, column := range dao.TableMeta.Columns {
		t.Log(column)
	}
}

func testCrud(t *testing.T) {
	DBMaster := newDb()
	table := "demo_info_crud"
	before(t, table)
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster)).WithTableName(table)
	u1 := &DemoInfo{
		UID:       10000,
		Name:      "fengjx",
		Sex:       "1",
		LoginTime: time.Now().Unix(),
		Utime:     time.Now().Unix(),
		Ctime:     time.Now().Unix(),
	}
	id, err := dao.Save(u1)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("id: %d", id)
	u2, err := dao.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if u2 == nil {
		t.Fatal("GetByID not exist")
	}
	assert.Equal(t, u1.UID, u2.UID)

	updateName := "fengjx_2023"
	affected, err := dao.UpdateField(id, map[string]any{
		"name": updateName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if affected == 0 {
		t.Fatal("update affected is 0")
	}
	u2, err = dao.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if u2 == nil {
		t.Fatal("GetByID not exist")
	}
	assert.Equal(t, updateName, u2.Name)
	ok, err := dao.DeleteByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("delete by id fail")
	}
	u2, err = dao.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, u2)
}

func testSelect(t *testing.T) {
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	var list []*DemoInfo
	err := dao.Selector().WhereC(
		DemoInfoMeta.UidIn(101, 102),
		DemoInfoMeta.SexEQ("male"),
	).List(&list)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		t.Log(item.UID)
	}
}

func testGet(t *testing.T) {
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	var demoInfo DemoInfo
	exist, err := dao.Selector().Where(
		ql.C(DemoInfoMeta.UidGT(100)),
	).Limit(1).OrderBy(ql.Asc(DemoInfoMeta.UID)).One(&demoInfo)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, exist)
	assert.NotNil(t, demoInfo)
	assert.Equal(t, int64(101), demoInfo.UID)
	t.Log("demoInfo[uid]", demoInfo.UID)
}

func testDeleteByColumns(t *testing.T) {
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	num, err := dao.DeleteByColumns(daox.OfMultiKv(DemoInfoMeta.UID, 100, 101))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(2), num)
}

type blog struct {
	Id         int64      `json:"id,string"`
	Uid        int64      `json:"uid,string"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	CreateTime int64      `json:"create_time"`
	Ctime      *time.Time `json:"-"`
	Utime      *time.Time `json:"-"`
}

func (b blog) GetID() any {
	return b.Id
}

func TestIgnoreField(t *testing.T) {
	tableName := "TestIgnoreField"
	DBMaster := newDb()
	dao := daox.NewDao[*blog](BlogMeta, daox.WithDBMaster(DBMaster)).WithTableName(tableName)
	t.Log(strings.Join(dao.TableMeta.Columns, ","))
	assert.Equal(t, "id,uid,title,content,create_time", strings.Join(dao.TableMeta.Columns, ","))
}

func testPage(t *testing.T) {
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	var list []*DemoInfo
	err := dao.Selector().Limit(10).Offset(5).List(&list)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		t.Log(item)
	}
}

func TestGetDaoByMeta(t *testing.T) {
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta)
	assert.Equal(t, DemoInfoMeta.TableName(), dao.TableName())
}

func TestSelectIfNull(t *testing.T) {
	tb := "demo_info_if_null"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta,
		daox.WithDBMaster(DBMaster),
		daox.WithIfNullVals(map[string]string{"name": "''"}),
	).WithTableName(tb)
	var list []*DemoInfo
	err := dao.Selector().Where(ql.C(
		DemoInfoMeta.UidIn(101, 102),
		DemoInfoMeta.SexEQ("male"),
	)).List(&list)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		t.Log(item.UID)
	}
	after(t, tb)
}

func TestDaox(t *testing.T) {
	before(t, "demo_info")
	t.Run("testCreate", testCreate)
	t.Run("testCrud", testCrud)
	t.Run("testSelect", testSelect)
	t.Run("testGet", testGet)
	t.Run("testPage", testPage)
	t.Run("testDeleteByColumns", testDeleteByColumns)
}

func TestBatchSave(t *testing.T) {
	tb := "demo_info_batch"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	nowUnix := time.Now().Unix()
	users := []*DemoInfo{
		{
			UID:       1000,
			Name:      "fengjx0",
			Sex:       "1",
			LoginTime: nowUnix,
			Utime:     nowUnix,
			Ctime:     nowUnix,
		},
		{
			UID:       1001,
			Name:      "fengjx1",
			Sex:       "2",
			LoginTime: nowUnix,
			Utime:     nowUnix,
			Ctime:     nowUnix,
		},
	}
	result, err := dao.BatchSave(users)
	if err != nil {
		t.Fatal(err)
	}
	affected, _ := result.RowsAffected()
	assert.Equal(t, int64(2), affected)
	u, err := dao.GetByColumn(daox.OfKv("uid", 1000))
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("GetByColumn not exist")
	}
	assert.Equal(t, "fengjx0", u.Name)
}

func TestDao_Update(t *testing.T) {
	tb := "demo_info_update"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster))
	u1 := &DemoInfo{
		UID:       20000,
		Name:      "fengjx",
		Sex:       "1",
		LoginTime: time.Now().Unix(),
		Utime:     time.Now().Unix(),
		Ctime:     time.Now().Unix(),
	}
	id, err := dao.Save(u1)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("save id: %d", id)
	u1.ID = id
	u1.Name = "fjx"
	u1.Utime = time.Now().Unix()
	ok, err := dao.Update(u1, "ctime")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("dao update not success")
	}
	u2, err := dao.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if u2 == nil {
		t.Fatal("GetByID not exist")
	}
	assert.Equal(t, u1.Name, u2.Name)
}

func TestDisableGlobalOmitColumns(t *testing.T) {
	daox.UseOmits("ctime", "utime")
	tb := "demo_info_global_disable_omit"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](
		DemoInfoMeta,
		daox.WithDBMaster(DBMaster),
		daox.WithIfNullVals(map[string]string{"utime": "10"}),
	).WithTableName(tb)
	nowSec := time.Now().Unix()
	u1 := &DemoInfo{
		UID:       10000,
		Name:      "fengjx",
		Sex:       "1",
		LoginTime: nowSec,
		Utime:     nowSec,
		Ctime:     nowSec,
	}
	id, err := dao.Save(u1, daox.DisableGlobalInsertOmits(true), daox.WithInsertOmits("utime"))
	assert.NoError(t, err)
	t.Log("info id", id)
	u, err := dao.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, nowSec, u.Ctime)
	assert.Equal(t, int64(10), u.Utime)
}

func TestWithTableName(t *testing.T) {
	tb := "demo_info_with_table"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](
		DemoInfoMeta,
		daox.WithDBMaster(DBMaster),
	).WithTableName(tb)
	tb2 := "demo_info_with_table_2"
	dao2 := dao.WithTableName(tb2)
	assert.Equal(t, tb, dao.TableName())
	assert.Equal(t, tb2, dao2.TableName())
}

func TestUpdater_Exec(t *testing.T) {
	tb := "demo_info_updater"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster)).WithTableName(tb)
	affected, err := dao.Updater().
		Fields(
			ql.F("name").Val("fengjx-1024"),
			ql.F("login_time").Incr(60*60),
		).
		Where(ql.C(DemoInfoMeta.UidEQ(100))).
		Exec()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), affected)
	m1, err := dao.GetByColumn(daox.OfKv("uid", 100))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "fengjx-1024", m1.Name)
}

func TestUpdater_NamedExec(t *testing.T) {
	tb := "demo_info_updater"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](DemoInfoMeta, daox.WithDBMaster(DBMaster)).WithTableName(tb)
	u := &DemoInfo{
		UID:  100,
		Name: "fengjx-1024",
	}
	affected, err := dao.Updater().
		Columns("name").
		Incr("login_time", 60*60).
		Where(ql.SC().And("uid = :uid")).
		NamedExec(u)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), affected)
	m1, err := dao.GetByColumn(daox.OfKv("uid", 100))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "fengjx-1024", m1.Name)
}

func TestDao_Hook(t *testing.T) {
	tb := "demo_info_hook"
	before(t, tb)
	DBMaster := newDb()
	dao := daox.NewDao[*DemoInfo](
		DemoInfoMeta,
		daox.WithDBMaster(DBMaster),
		daox.WithIfNullVals(map[string]string{"utime": "10", "ctime": "10"}),
		daox.WithHooks(daox.NewLogHook(func(ctx context.Context, ec *engine.ExecutorContext, er *engine.ExecutorResult) {
			t.Log("sql_type", ec.Type, "sql:", ec.SQL, "args:", ec.Args, "rows:", er.QueryRows, "duration:", er.Duration, "err:", er.Err)
		})),
	).WithTableName(tb)
	nowSec := time.Now().Unix()
	u1 := &DemoInfo{
		UID:       10000,
		Name:      "fengjx",
		Sex:       "1",
		LoginTime: nowSec,
		Utime:     nowSec,
		Ctime:     nowSec,
	}
	id, err := dao.Save(u1)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("id: %d", id)
	u2, err := dao.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if u2 == nil {
		t.Fatal("GetByID not exist")
	}
	assert.Equal(t, u1.UID, u2.UID)
	list, err := dao.ListByIDs(1, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(list))
	assert.NoError(t, err)
}
