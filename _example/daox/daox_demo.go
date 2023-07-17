package main

import (
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

// 查询单条记录
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

// 查询多条记录
func queryList(dao *daox.Dao) {
	var list []*User
	err := dao.List(daox.OfKv("sex", 0), &list)
	if err != nil {
		log.Panic(err)
	}
	log.Println("query by sex")
	for _, user := range list {
		log.Println(user)
	}

	log.Println("ListByColumns")
	var list2 []User
	err = dao.ListByColumns(daox.OfMultiKv("uid", 10000, 10001), &list2)
	if err != nil {
		log.Panic(err)
	}
	for _, user := range list2 {
		log.Println(user)
	}

	log.Println("ListByIds")
	var list3 []User
	err = dao.ListByIds(&list3, 10, 11)
	if err != nil {
		log.Panic(err)
	}
	for _, user := range list3 {
		log.Println(user)
	}
}

func updateUser(dao *daox.Dao) {
	log.Println("=========== update ============")
	user := new(User)
	err := dao.GetByID(10, user)
	if err != nil {
		log.Fatal(err)
	}
	user.Nickname = "update-name-10"
	// 全字段更新
	ok, err := dao.Update(user)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("update res - %v", ok)

	// 部分字段更新
	ok, err = dao.UpdateField(11, map[string]interface{}{
		"nickname": "update-name-11",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("update res - %v", ok)
	var list []*User
	err = dao.ListByIds(&list, 10, 11)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range list {
		log.Println(u)
	}
}

func deleteUSer(dao *daox.Dao) {
	log.Println("=========== delete ============")
	// 按 id 删除
	ok, err := dao.DeleteById(21)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("delete res - %v", ok)
	user := new(User)
	err = dao.GetByID(21, user)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("delete by id res - %v", user.Id)

	// 按指定字段删除
	affected, err := dao.DeleteByColumn(daox.OfKv("uid", 101))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("delete by column res - %v", affected)

	// 按字段删除多条记录
	affected, err = dao.DeleteByColumns(daox.OfMultiKv("uid", 102, 103))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("multiple delete by column res - %v", affected)
}

func main() {
	db := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3306)/demo")
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
	dao := daox.NewDAO(db, "user_info", "id", reflect.TypeOf(&User{}), daox.IsAutoIncrement())
	log.Printf("columns: %v\n", dao.TableMeta.Columns)
	// insertUser(dao)
	// batchInsertUser(dao)
	selectUser(dao)
	queryList(dao)
	updateUser(dao)
	deleteUSer(dao)
}
