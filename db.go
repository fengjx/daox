package daox

import (
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func NewDB(db *sqlx.DB) *DB {
	dbx := &DB{
		DB: db,
	}
	return dbx
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return NewDB(db), nil
}

func MustOpen(driverName, dataSourceName string) *DB {
	db := sqlx.MustOpen(driverName, dataSourceName)
	return NewDB(db)
}
