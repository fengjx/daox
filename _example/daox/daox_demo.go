package main

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	"github.com/fengjx/daox"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type User struct {
	Id       int64  `json:"id"`
	Uid      int64  `json:"uid"`
	Nickname string `json:"nickname"`
	Sex      int32  `json:"sex"`
	Utime    int64  `json:"utime"`
	Ctime    int64  `json:"ctime"`
}

func (receiver User) GetID() interface{} {
	return receiver.Id
}

func insertUser(dao *daox.Dao) {
	for i := 0; i < 20; i++ {
		sec := time.Now().Unix()
		user := &User{
			Uid:      100 + int64(i),
			Nickname: randString(6),
			Sex:      int32(i) % 2,
			Utime:    sec,
			Ctime:    sec,
		}
		id, err := dao.Save(user)
		if err != nil {
			log.Panic(err)
		}
		log.Println(id)
	}
}

func batchInsertUser(dao *daox.Dao) {
	var users []*User
	for i := 0; i < 20; i++ {
		sec := time.Now().Unix()
		user := &User{
			Uid:      10000 + int64(i),
			Nickname: randString(6),
			Sex:      int32(i) % 2,
			Utime:    sec,
			Ctime:    sec,
		}
		users = append(users, user)
	}
	count, err := dao.BatchSave(users)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("save count: %d", count)
}

func selectUser(dao *daox.Dao) {
	user := new(User)
	err := dao.GetByID(1, user)
	if err != nil {
		log.Panic(err)
	}
	log.Println(user)

	user2 := new(User)
	err2 := dao.GetByColumn(daox.OfKv("uid", 10000), user2)
	if err2 != nil {
		log.Panic(err2)
	}
	log.Println(user2)
}

func main() {
	db := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3306)/demo")
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
	dao := daox.NewDAO(db, "user_info", "id", reflect.TypeOf(&User{}), daox.IsAutoIncrement())
	for _, col := range dao.TableMeta.Columns {
		fmt.Println(col)
	}
	// insertUser(dao)
	// batchInsertUser(dao)
	selectUser(dao)
}
