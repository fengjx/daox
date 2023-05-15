package daox

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Dao struct {
	DBMaster  *sqlx.DB
	DBRead    *sqlx.DB
	Redis     *redis.Client
	TableMeta *TableMeta
	CacheMeta *CacheMeta
}

func Create(master *sqlx.DB, tableName string, primaryKey string, structType reflect.Type, opts ...Option) *Dao {
	structMap := master.Mapper.TypeMap(structType)
	columns := make([]string, 0, len(structMap.Names))
	for _, column := range structMap.Names {
		columns = append(columns, column.Name)
	}
	dao := &Dao{
		TableMeta: &TableMeta{
			TableName:  tableName,
			StructType: structType,
			PrimaryKey: primaryKey,
			Columns:    columns,
		},
		DBMaster: master,
	}
	for _, opt := range opts {
		opt(dao)
	}
	if dao.DBRead == nil {
		dao.DBRead = dao.DBMaster
	}
	return dao
}

func (dao *Dao) Save(dest interface{}) (int64, error) {
	tableMeta := dao.TableMeta
	var columns []string
	if tableMeta.IsAutoIncrement {
		columns = tableMeta.OmitColumns(tableMeta.PrimaryKey)
	} else {
		columns = tableMeta.OmitColumns()
	}
	execSql, err := sqlbuilder.New(tableMeta.TableName).Insert().Columns(columns...).NameSql()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (dao *Dao) GetByPrimaryKey(primaryKey interface{}, dest interface{}) error {
	tableMeta := dao.TableMeta
	querySql, err := sqlbuilder.New(tableMeta.TableName).Select().Columns(tableMeta.OmitColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", tableMeta.PrimaryKey))).
		Sql()
	if err != nil {
		return err
	}
	err = dao.DBRead.Get(dest, querySql, primaryKey)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
