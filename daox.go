package daox

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var ctx = context.TODO()

type Dao struct {
	DBMaster  *sqlx.DB
	DBRead    *sqlx.DB
	Redis     *redis.Client
	TableMeta *TableMeta
}

func NewDAO(master *sqlx.DB, tableName string, primaryKey string, structType reflect.Type, opts ...Option) *Dao {
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
	if dao.TableMeta.CacheVersion == "" {
		dao.TableMeta.CacheVersion = "v1"
	}
	if dao.TableMeta.CacheExpireTime == 0 {
		dao.TableMeta.CacheExpireTime = time.Minute * 3
	}
	keyPrefix := fmt.Sprintf("{%v}-%s-", structType.Elem(), dao.TableMeta.CacheVersion)
	dao.TableMeta.cachePrefix = keyPrefix
	return dao
}

func (dao *Dao) SQLBuilder() *sqlbuilder.Builder {
	return sqlbuilder.New(dao.TableMeta.TableName)
}

// Save
// omitColumns 不需要 insert 的字段
func (dao *Dao) Save(dest interface{}, omitColumns ...string) (int64, error) {
	tableMeta := dao.TableMeta
	if tableMeta.IsAutoIncrement {
		omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	}
	columns := tableMeta.OmitColumns(omitColumns...)
	execSql, err := dao.SQLBuilder().Insert().Columns(columns...).NameSql()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (dao *Dao) BatchSave(dest []interface{}) (int64, error) {
	tableMeta := dao.TableMeta
	var columns []string
	if tableMeta.IsAutoIncrement {
		columns = tableMeta.OmitColumns(tableMeta.PrimaryKey)
	} else {
		columns = tableMeta.OmitColumns()
	}
	execSql, err := dao.SQLBuilder().Insert().Columns(columns...).NameSql()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (dao *Dao) GetByColumn(kv *KV, dest interface{}) error {

	tableMeta := dao.TableMeta
	querySql, err := dao.SQLBuilder().Select().
		Columns(tableMeta.OmitColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", kv.Key))).
		Sql()
	if err != nil {
		return err
	}
	err = dao.DBRead.Get(dest, querySql, kv.Value)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (dao *Dao) GetById(IdValue interface{}, dest interface{}) error {
	tableMeta := dao.TableMeta
	return dao.GetByColumn(Kv(tableMeta.PrimaryKey, IdValue), dest)
}

func (dao *Dao) UpdateById(idValue interface{}, dict map[string]interface{}) (int64, error) {
	tableMeta := dao.TableMeta
	columns := make([]string, 0, len(dict))
	args := make([]interface{}, 0, len(dict))
	for k, v := range dict {
		columns = append(columns, k)
		args = append(args, v)
	}
	args = append(args, idValue)
	execSql, err := dao.SQLBuilder().Update().
		Columns(columns...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", tableMeta.PrimaryKey))).
		Sql()
	if err != nil {
		return 0, err
	}
	res, err := dao.DBMaster.Exec(execSql, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (dao *Dao) DeleteByColumn(kv *KV) (int64, error) {
	execSql, err := dao.SQLBuilder().Delete().
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", kv.Key))).
		Sql()
	if err != nil {
		return 0, err
	}
	res, err := dao.DBMaster.Exec(execSql, kv.Value)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (dao *Dao) DeleteInColumn(mkv *MultiKV) (int64, error) {
	execSql, err := dao.SQLBuilder().Delete().
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s in ?", mkv.Key))).
		Sql()
	if err != nil {
		return 0, err
	}
	execSql, args, err := sqlx.In(execSql, mkv.Values)
	if err != nil {
		return 0, err
	}
	res, err := dao.DBMaster.Exec(execSql, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (dao *Dao) DeleteById(idValue interface{}) (bool, error) {
	tableMeta := dao.TableMeta
	affected, err := dao.DeleteByColumn(Kv(tableMeta.PrimaryKey, idValue))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (dao *Dao) DeleteByIds(idValue interface{}) (bool, error) {
	tableMeta := dao.TableMeta
	affected, err := dao.DeleteByColumn(Kv(tableMeta.PrimaryKey, idValue))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func containsString(collection []string, element string) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}
