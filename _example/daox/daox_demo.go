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
	"github.com/redis/go-redis/v9"

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

var createTableSQL = `
create table if not exists demo_user
(
    id         bigint auto_increment,
    uid        bigint,
    name       varchar(32) default '',
    sex        tinyint  default 0,
    utime      bigint      default 0,
    ctime      bigint      default 0,
    primary key pk (id),
    unique uni_uid (uid)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_bin;
`

type User struct {
	Id    int64  `json:"id"`
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Sex   int32  `json:"sex"`
	Utime int64  `json:"utime"`
	Ctime int64  `json:"ctime"`
}

func (u *User) GetID() interface{} {
	return u.Id
}

func insertUser(dao *daox.Dao) {
	for i := 0; i < 20; i++ {
		sec := time.Now().Unix()
		user := &User{
			Uid:   100 + int64(i),
			Name:  randString(6),
			Sex:   int32(i) % 2,
			Utime: sec,
			Ctime: sec,
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
			Uid:   10000 + int64(i),
			Name:  randString(6),
			Sex:   int32(i) % 2,
			Utime: sec,
			Ctime: sec,
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
	id, err := dao.GetByID(1, user)
	if err != nil {
		log.Panic(err)
	}
	log.Println("id", id)
	log.Println(user)

	user2 := new(User)
	exist, err := dao.GetByColumn(daox.OfKv("uid", 10000), user2)
	if err != nil {
		log.Panic(err)
	}
	log.Println("exist", exist)
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
	err = dao.ListByIDs(&list3, 10, 11)
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
	_, err := dao.GetByID(10, user)
	if err != nil {
		log.Fatal(err)
	}
	user.Name = "update-name-10"
	// 全字段更新
	ok, err := dao.Update(user)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("update res - %v", ok)

	// 部分字段更新
	ok, err = dao.UpdateField(11, map[string]interface{}{
		"name": "update-name-11",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("update res - %v", ok)
	var list []*User
	err = dao.ListByIDs(&list, 10, 11)
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
	ok, err := dao.DeleteByID(21)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("delete res - %v", ok)
	user := new(User)
	exist, err := dao.GetByID(21, user)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exist", exist)
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

func cache(dao *daox.Dao) {
	user := new(User)
	// 按id查询并缓存
	exist, err := dao.GetByIDCache(10, user)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exist", exist)
	log.Printf("get by id with cache - %v", user)

	// 删除缓存
	err = dao.DeleteCache("id", 10)
	if err != nil {
		log.Fatal(err)
	}

	// 按指定字段查询并缓存
	cacheUser := new(User)
	exist, err = dao.GetByColumnCache(daox.OfKv("uid", 10001), cacheUser)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("exist", exist)
	log.Printf("get by uid with cache - %v", cacheUser)
}

func main() {
	redisCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	db := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3306)/demo")
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	dao := daox.NewDAO(
		db,
		"demo_user",
		"id",
		reflect.TypeOf(&User{}),
		daox.IsAutoIncrement(),
		daox.WithCache(redisCli),
		daox.WithCacheVersion("v1"),
	)
	log.Printf("columns: %v\n", dao.TableMeta.Columns)
	insertUser(dao)
	batchInsertUser(dao)
	selectUser(dao)
	queryList(dao)
	updateUser(dao)
	deleteUSer(dao)
	cache(dao)
	_, err = db.Exec("drop table demo_user")
	if err != nil {
		log.Fatal(err)
	}
}
