package daox

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func sqliteDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file:.cache/test.db?cache=shared&mode=memory")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func createRedisClient(t *testing.T) *redis.Client {
	serv := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{
		Addr:     serv.Addr(),
		Password: "",
		DB:       0,
	})
}

func newDb(t *testing.T) *sqlx.DB {
	db := sqlx.NewDb(sqliteDB(t), "sqlite3")
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return db
}

func before(t *testing.T) {
	t.Log("before...")
	db := newDb(t)
	_, err := db.Exec(`
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
	dao := Create(db, "user", "id", reflect.TypeOf(&user{}), IsAutoIncrement())
	for i := 0; i < 10; i++ {
		id, err := dao.Save(&user{
			Uid:       int64(1000 + i),
			Name:      fmt.Sprintf("u-%d", i),
			Sex:       "male",
			LoginTime: time.Now().Unix(),
			Utime:     time.Now().Unix(),
		}, "ctime")
		if err != nil {
			t.Log(err.Error())
			continue
		}
		t.Logf("save id - %d", id)
	}

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

func TestCreate(t *testing.T) {
	DBMaster := newDb(t)
	cacheMeta := &CacheMeta{
		Version:    "v1",
		CacheKey:   "uid",
		CacheTime:  time.Minute * 3,
		ExpireTime: time.Minute * 10,
	}
	redisClient := createRedisClient(t)
	dao := Create(DBMaster, "user", "id", reflect.TypeOf(&user{}), IsAutoIncrement(), WithCache(redisClient, cacheMeta))
	assert.Equal(t, len(dao.TableMeta.Columns), 7)
	assert.Equal(t, dao.TableMeta.PrimaryKey, "id")
	for _, column := range dao.TableMeta.Columns {
		t.Log(column)
	}
	assert.Equal(t, cacheMeta, dao.CacheMeta)
}

func TestSave(t *testing.T) {
	before(t)
	DBMaster := newDb(t)
	dao := Create(DBMaster, "user", "id", reflect.TypeOf(&user{}), IsAutoIncrement())
	id, err := dao.Save(&user{
		Uid:       1000,
		Name:      "fengjx",
		Sex:       "1",
		LoginTime: time.Now().Unix(),
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("id: %d", id)

	u := &user{}
	err = dao.GetById(id, u)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u.Name)
}
