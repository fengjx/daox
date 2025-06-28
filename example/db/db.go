package db

import (
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

var DB *sqlx.DB

func init() {
	dsn := "root:123456@tcp(192.168.6.121:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	DB = db
}
