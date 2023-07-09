package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/fengjx/daox"
)

type User struct {
	Id    int64  `json:"id"`
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Sex   string `json:"sex"`
	Utime int64  `json:"utime"`
	Ctime int64  `json:"ctime"`
}

func memoryDB() *sql.DB {
	db, err := sql.Open("sqlite3", "file:.cache/test.db?cache=shared&mode=memory")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
		CREATE TABLE user (
		  id integer primary key autoincrement,
		  name text,
		  sex text,
		  utime integer,
		  ctime integer
		);	
	`)
	if err != nil {
		panic(err)
	}
	return db
}

func insertUser(db *sqlx.DB) {
	sql := "insert into user(name, sex, utime, ctime) values ('fenjgx', '1', 0, 0)"
	now := time.Now().Unix()
	user := &User{
		Name:  "fengjx",
		Sex:   "1",
		Utime: now,
		Ctime: now,
	}
	res, err := db.NamedExec(sql, user)
	if err != nil {
		panic(err)
	}
	id, _ := res.LastInsertId()
	log.Printf("id: %d", id)
}

func selectUser(db *sqlx.DB) {
	sql := "select * from user where id = ?"
	u1 := &User{}
	err := db.Get(u1, sql, 1)
	if err != nil {
		panic(err)
	}
	json, _ := json.Marshal(u1)
	log.Println(string(json))
}

func main() {
	var db *sqlx.DB
	db = sqlx.MustOpen("sqlite3", "file:.cache/example.db?cache=shared&mode=memory")
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
	dao := daox.NewDAO(db, "user", "id", reflect.TypeOf(&User{}), daox.IsAutoIncrement())
	for _, col := range dao.TableMeta.Columns {
		fmt.Println(col)
	}
}
