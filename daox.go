package daox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/utils"
)

var (
	ErrUpdatePrimaryKeyRequire = errors.New("[daox] Primary key require for update")
)

var ctx = context.TODO()

type SliceToMapFun = func([]*Model) map[interface{}]*Model

type Dao struct {
	DBMaster      *DB
	DBRead        *DB
	RedisClient   *redis.Client
	TableMeta     *TableMeta
	CacheProvider *CacheProvider
}

func NewDAO(master *sqlx.DB, tableName string, primaryKey string, structType reflect.Type, opts ...Option) *Dao {
	dao := &Dao{
		DBMaster: NewDB(master),
	}
	columns := dao.GetColumnsByType(structType)
	dao.TableMeta = &TableMeta{
		TableName:  tableName,
		StructType: structType,
		PrimaryKey: primaryKey,
		Columns:    columns,
	}
	keyPrefix := fmt.Sprintf("data_%v", structType.Elem())
	dao.CacheProvider = NewCacheProvider(dao.RedisClient, keyPrefix, "v1", time.Minute*3)
	for _, opt := range opts {
		opt(dao)
	}
	if dao.DBRead == nil {
		dao.DBRead = dao.DBMaster
	}
	return dao
}

func (dao *Dao) SQLBuilder() *sqlbuilder.Builder {
	return sqlbuilder.New(dao.TableMeta.TableName)
}

func (dao *Dao) GetColumnsByModel(model interface{}, omitColumns ...string) []string {
	return dao.GetColumnsByType(reflect.TypeOf(model), omitColumns...)
}

// GetColumnsByType 通过字段 tag 解析数据库字段
func (dao *Dao) GetColumnsByType(typ reflect.Type, omitColumns ...string) []string {
	return sqlbuilder.GetColumnsByType(dao.DBMaster.Mapper, typ, omitColumns...)
}

func (dao *Dao) DBColumns(omitColumns ...string) []string {
	columns := make([]string, 0)
	for _, column := range dao.TableMeta.Columns {
		if utils.ContainsString(omitColumns, column) {
			continue
		}
		columns = append(columns, column)
	}
	return columns
}

func (dao *Dao) TableName() string {
	return dao.TableMeta.TableName
}

// Save
// omitColumns 不需要 insert 的字段
func (dao *Dao) Save(dest Model, omitColumns ...string) (int64, error) {
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

func (dao *Dao) BatchSave(models interface{}) (int64, error) {
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
	res, err := dao.DBMaster.NamedExec(execSql, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (dao *Dao) GetByColumn(kv *KV, dest Model) error {
	if kv == nil {
		return nil
	}
	querySql, err := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
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

func (dao *Dao) GetByColumnCache(kv *KV, dest Model) error {
	return dao.CacheProvider.Fetch(kv.Key, utils.ToString(kv.Value), dest, func() (interface{}, error) {
		err := dao.GetByColumn(kv, dest)
		if err != nil {
			return nil, err
		}
		return dest, nil
	})
}

func (dao *Dao) ListByColumns(kvs *MultiKV, dest interface{}) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	querySql, err := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s in (?)", kvs.Key))).
		Sql()
	if err != nil {
		return err
	}
	querySql, args, err := sqlx.In(querySql, kvs.Values)
	return dao.DBRead.Select(dest, querySql, args...)
}

func (dao *Dao) List(kv *KV, dest interface{}) error {
	if kv == nil {
		return nil
	}
	querySql, err := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", kv.Key))).
		Sql()
	if err != nil {
		return err
	}
	return dao.DBRead.Select(dest, querySql, kv.Value)
}

func (dao *Dao) GetByID(id interface{}, dest Model) error {
	tableMeta := dao.TableMeta
	return dao.GetByColumn(OfKv(tableMeta.PrimaryKey, id), dest)
}

func (dao *Dao) GetByIDCache(id interface{}, dest Model) error {
	primaryKey := dao.TableMeta.PrimaryKey
	return dao.CacheProvider.Fetch(primaryKey, id, dest, func() (interface{}, error) {
		return dest, dao.GetByID(id, dest)
	})
}

func (dao *Dao) ListByIds(dest interface{}, ids ...interface{}) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumns(OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

func (dao *Dao) UpdateField(idValue interface{}, fieldMap map[string]interface{}) (bool, error) {
	if utils.IsIDEmpty(idValue) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	columns := make([]string, 0, len(fieldMap))
	args := make([]interface{}, 0, len(fieldMap))
	for k, v := range fieldMap {
		columns = append(columns, k)
		args = append(args, v)
	}
	args = append(args, idValue)
	updateSQL, err := dao.SQLBuilder().Update().
		Columns(columns...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", tableMeta.PrimaryKey))).
		Sql()
	if err != nil {
		return false, err
	}
	res, err := dao.DBMaster.Exec(updateSQL, args...)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (dao *Dao) Update(m Model) (bool, error) {
	if utils.IsIDEmpty(m.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	updateSQL, err := dao.SQLBuilder().Update().
		Columns(dao.DBColumns(tableMeta.PrimaryKey)...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey))).
		NameSql()
	if err != nil {
		return false, err
	}
	res, err := dao.DBMaster.NamedExec(updateSQL, m)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (dao *Dao) DeleteByColumn(kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
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

func (dao *Dao) DeleteByColumns(kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}

	execSql, err := dao.SQLBuilder().Delete().
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s in (?)", kvs.Key))).
		Sql()
	if err != nil {
		return 0, err
	}
	execSql, args, err := sqlx.In(execSql, kvs.Values)
	if err != nil {
		return 0, err
	}
	res, err := dao.DBMaster.Exec(execSql, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (dao *Dao) DeleteById(id interface{}) (bool, error) {
	tableMeta := dao.TableMeta
	affected, err := dao.DeleteByColumn(OfKv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// Fetch query one row
func (dao *Dao) Fetch(field string, item string, dest interface{}, fun CreateDataFun) error {
	return dao.CacheProvider.Fetch(field, item, dest, fun)
}

// BatchFetch
// 注意不会按 items 顺序返回
func (dao *Dao) BatchFetch(field string, items []interface{}, dest interface{}, fun BatchCreateDataFun) error {
	return dao.CacheProvider.BatchFetch(field, items, dest, fun)
}

func (dao *Dao) DeleteCache(kv *KV) error {
	return dao.CacheProvider.Del(kv.Key, utils.ToString(kv.Value))
}

func (dao *Dao) BatchDeleteCache(field string, items []string) error {
	return dao.CacheProvider.BatchDel(field, items)
}

func ModelListToMap(src []Model) map[interface{}]Model {
	if len(src) == 0 {
		return make(map[interface{}]Model, 0)
	}
	resMap := make(map[interface{}]Model, 0)
	for _, m := range src {
		resMap[m.GetID()] = m
	}
	return resMap
}
