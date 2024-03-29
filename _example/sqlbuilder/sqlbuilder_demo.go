package main

import (
	"log"

	"github.com/fengjx/daox/sqlbuilder"
)

func main() {
	querySQL, err := sqlbuilder.New("user_info").Select().
		Columns("id", "username", "age", "sex", "ctime").
		Where(
			sqlbuilder.C().
				Where(true, "age > ?").
				And(true, "sex = ?"),
		).
		OrderBy(sqlbuilder.Desc("ctime")).
		Offset(10).
		Limit(10).Sql()
	if err != nil {
		log.Panic(err)
	}
	log.Println(querySQL)

	inserter := sqlbuilder.New("user_info").Insert().
		Columns("username", "age", "sex")

	sql, err := inserter.Sql()
	log.Println(sql)

	nameSql, err := inserter.NameSql()
	log.Println(nameSql)

	updateSQL, err := sqlbuilder.New("user_info").
		Update().
		Columns("username", "age").
		Where(
			sqlbuilder.C().
				Where(true, "id = ?")).
		Sql()
	log.Println(updateSQL)

	deleteSQL, err := sqlbuilder.New("user_info").Delete().
		Where(sqlbuilder.C().Where(true, "id = ?")).
		Sql()
	log.Println(deleteSQL)

}
