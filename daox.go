package daox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var ctx = context.TODO()

type Dao struct {
	DBMaster      *sqlx.DB
	DBRead        *sqlx.DB
	RedisClient   *redis.Client
	TableMeta     *TableMeta
	cacheProvider *CacheProvider
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
	keyPrefix := fmt.Sprintf("data_{%v}_%s", structType.Elem(), dao.TableMeta.CacheVersion)
	dao.TableMeta.cachePrefix = keyPrefix
	dao.cacheProvider = NewCacheProvider(dao.RedisClient, dao.TableMeta.CacheExpireTime)
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

func (dao *Dao) BatchSave(dest interface{}) (int64, error) {
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
	if kv == nil {
		return nil
	}
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

func (dao *Dao) ListByColumns(kvs *MultiKV, dest interface{}) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	tableMeta := dao.TableMeta
	querySql, err := dao.SQLBuilder().Select().
		Columns(tableMeta.OmitColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s in ?", kvs.Key))).
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
	tableMeta := dao.TableMeta
	querySql, err := dao.SQLBuilder().Select().
		Columns(tableMeta.OmitColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", kv.Key))).
		Sql()
	if err != nil {
		return err
	}
	return dao.DBRead.Select(dest, querySql, kv.Value)
}

func (dao *Dao) GetById(id interface{}, dest interface{}) error {
	tableMeta := dao.TableMeta
	return dao.GetByColumn(Kv(tableMeta.PrimaryKey, id), dest)
}

func (dao *Dao) GetByIds(ids []interface{}, dest interface{}) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumns(MultiKv(tableMeta.PrimaryKey, ids), dest)
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
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s in ?", kvs.Key))).
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
	affected, err := dao.DeleteByColumn(Kv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (dao *Dao) Fetch(kv *KV, dest interface{}, fun FillDataFun) error {
	return dao.cacheProvider.Fetch(dao.KeyPrefix(kv.Key), toString(kv.Value), dest, fun)
}

// BatchFetch
// 注意不会按 items 顺序返回
func (dao *Dao) BatchFetch(field string, items []string, dest interface{}, fun BatchCreateDataFun) error {
	return dao.cacheProvider.BatchFetch(field, items, dest, fun)
}

func (dao *Dao) DeleteCache(kv *KV) error {
	return dao.cacheProvider.Del(kv.Key, toString(kv.Value))
}

func (dao *Dao) BatchDeleteCache(field string, items []string) error {
	return dao.cacheProvider.BatchDel(field, items)
}

func (dao *Dao) KeyPrefix(field string) string {
	return fmt.Sprintf("%s_%s", dao.TableMeta.cachePrefix, field)
}

func containsString(collection []string, element string) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if bs, ok := v.([]byte); ok {
		return string(bs)
	}

	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.String:
		if s, ok := v.(string); ok {
			return s
		}
		return reflect.ValueOf(v).String()
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return fmt.Sprintf("%d", v)
	case reflect.Float32, reflect.Float64:
		bs, _ := json.Marshal(v)
		return string(bs)
	case reflect.Bool:
		if b, ok := v.(bool); ok && b {
			return "true"
		} else {
			return "false"
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}
