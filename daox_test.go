package daox

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func sqliteDB() (*sql.DB, error) {
	return sql.Open("sqlite3", "file:.cache/test.db?cache=shared&mode=memory")
}

// lint:ignore U1000 Ignore unused function temporarily
func mysqlDB() (*sql.DB, error) {
	return sql.Open("mysql", "root:1234@tcp(192.168.1.200:3306)/fjx?charset=utf8mb4,utf8&tls=false&timeout=10s")
}

func createMockRedisClient(t *testing.T) *redis.Client {
	serv := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{
		Addr:     serv.Addr(),
		Password: "",
		DB:       0,
	})
}

func createRedisClient(t *testing.T) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
}

func newDb() (*sqlx.DB, error) {
	db, err := mysqlDB()
	if err != nil {
		return nil, err
	}
	dbx := sqlx.NewDb(db, "mysql")
	dbx.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return dbx, nil
}

func newSqliteDb() (*sqlx.DB, error) {
	db, err := sqliteDB()
	if err != nil {
		return nil, err
	}
	dbx := sqlx.NewDb(db, "sqlite3")
	dbx.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return dbx, nil
}

var once sync.Once

func Init() {
	once.Do(func() {
		log.Println("before...")
		db, err := newSqliteDb()
		if err != nil {
			panic(err)
		}
		_, err = db.Exec(`
		CREATE TABLE user (
		  id integer primary key autoincrement,
		  uid integer,
		  name text,
		  sex text,
		  login_time integer,
		  utime integer,
		  ctime integer
		);	
	`)
		if err != nil {
			panic(err)
		}
		dao := NewDAO(db, "user", "id", reflect.TypeOf(&user{}), IsAutoIncrement())
		for i := 0; i < 10; i++ {
			id, err := dao.Save(&user{
				Uid:       int64(100 + i),
				Name:      fmt.Sprintf("u-%d", i),
				Sex:       "male",
				LoginTime: time.Now().Unix(),
				Utime:     time.Now().Unix(),
			}, "ctime")
			if err != nil {
				panic(err)
				continue
			}
			log.Printf("save id - %d \n", id)
		}
	})
}

type user struct {
	Id        int64  `json:"id"`
	Uid       int64  `json:"uid"`
	Name      string `json:"name"`
	Sex       string `json:"sex"`
	LoginTime int64  `json:"login_time"`
	Utime     int64  `json:"utime"`
	Ctime     int64  `json:"ctime"`
}

func (u *user) GetId() interface{} {
	return u.Id
}

func TestCreate(t *testing.T) {
	DBMaster, err := newSqliteDb()
	if err != nil {
		log.Panic(err)
	}
	redisClient := createRedisClient(t)
	dao := NewDAO(
		DBMaster,
		"user",
		"id",
		reflect.TypeOf(&user{}),
		IsAutoIncrement(),
		WithCache(redisClient),
		WithCacheVersion("v1"),
	)
	assert.Equal(t, len(dao.TableMeta.Columns), 7)
	assert.Equal(t, dao.TableMeta.PrimaryKey, "id")
	for _, column := range dao.TableMeta.Columns {
		t.Log(column)
	}
}

func TestCrud(t *testing.T) {
	Init()
	DBMaster, err := newSqliteDb()
	if err != nil {
		log.Panic(err)
	}
	dao := NewDAO(DBMaster, "user", "id", reflect.TypeOf(&user{}), IsAutoIncrement())
	u1 := &user{
		Uid:       10000,
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
	u2 := &user{}
	err = dao.GetByID(id, u2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, u1.Uid, u2.Uid)
}

func TestBatchSave(t *testing.T) {
	Init()
	// DBMaster := newDb(t)
	DBMaster, err := newSqliteDb()
	if err != nil {
		log.Panic(err)
	}
	dao := NewDAO(DBMaster, "user", "id", reflect.TypeOf(&user{}), IsAutoIncrement())
	nowUnix := time.Now().Unix()
	users := []*user{
		{
			Uid:       1000,
			Name:      "fengjx0",
			Sex:       "1",
			LoginTime: nowUnix,
			Utime:     nowUnix,
			Ctime:     nowUnix,
		},
		{
			Uid:       1001,
			Name:      "fengjx1",
			Sex:       "2",
			LoginTime: nowUnix,
			Utime:     nowUnix,
			Ctime:     nowUnix,
		},
	}
	affected, err := dao.BatchSave(users)
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, int64(2), affected)
	u := &user{}
	err = dao.GetByColumn(OfKv("uid", 1000), u)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "fengjx0", u.Name)
}
