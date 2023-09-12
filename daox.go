package daox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"

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
	dao.CacheProvider = NewCacheProvider(nil, keyPrefix, "v1", time.Minute*3)
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
	execSql, err := dao.SQLBuilder().Insert().Columns(columns...).NameSQL()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ReplaceInto
// omitColumns 不需要 insert 的字段
func (dao *Dao) ReplaceInto(dest Model, omitColumns ...string) (int64, error) {
	tableMeta := dao.TableMeta
	if tableMeta.IsAutoIncrement {
		omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	}
	columns := tableMeta.OmitColumns(omitColumns...)
	execSql, err := dao.SQLBuilder().Insert().
		Columns(columns...).
		IsReplaceInto(true).
		NameSQL()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// BatchSave 批量新增
// omitColumns 不需要 insert 的字段
func (dao *Dao) BatchSave(models interface{}, omitColumns ...string) (int64, error) {
	tableMeta := dao.TableMeta
	var columns []string
	if tableMeta.IsAutoIncrement {
		omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	}
	columns = tableMeta.OmitColumns(omitColumns...)
	execSQL, err := dao.SQLBuilder().Insert().Columns(columns...).NameSQL()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSQL, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// BatchReplaceInto 批量新增，使用 replace into 方式
// omitColumns 不需要 insert 的字段
func (dao *Dao) BatchReplaceInto(models interface{}, omitColumns ...string) (int64, error) {
	tableMeta := dao.TableMeta
	var columns []string
	if tableMeta.IsAutoIncrement {
		omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	}
	columns = tableMeta.OmitColumns(omitColumns...)
	execSQL, err := dao.SQLBuilder().Insert().
		Columns(columns...).
		IsReplaceInto(true).
		NameSQL()
	if err != nil {
		return 0, nil
	}
	res, err := dao.DBMaster.NamedExec(execSQL, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetByColumn get one row
// bool: exist or not
func (dao *Dao) GetByColumn(kv *KV, dest Model) (bool, error) {
	if kv == nil {
		return false, nil
	}
	querySql, err := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s = ?", kv.Key))).
		SQL()
	if err != nil {
		return false, err
	}
	err = dao.DBRead.Get(dest, querySql, kv.Value)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, nil
}

func (dao *Dao) GetByColumnCache(kv *KV, dest Model) (bool, error) {
	exist := true
	err := dao.CacheProvider.Fetch(kv.Key, utils.ToString(kv.Value), dest, func() (interface{}, error) {
		var err error
		exist, err = dao.GetByColumn(kv, dest)
		if err != nil {
			return nil, err
		}
		return dest, nil
	})
	return exist, err
}

func (dao *Dao) ListByColumns(kvs *MultiKV, dest interface{}) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	querySql, err := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
		Where(sqlbuilder.C().Where(true, fmt.Sprintf("%s in (?)", kvs.Key))).
		SQL()
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
		SQL()
	if err != nil {
		return err
	}
	return dao.DBRead.Select(dest, querySql, kv.Value)
}

func (dao *Dao) GetByID(id interface{}, dest Model) (bool, error) {
	tableMeta := dao.TableMeta
	return dao.GetByColumn(OfKv(tableMeta.PrimaryKey, id), dest)
}

func (dao *Dao) GetByIDCache(id interface{}, dest Model) (bool, error) {
	primaryKey := dao.TableMeta.PrimaryKey
	exist := true
	err := dao.CacheProvider.Fetch(primaryKey, id, dest, func() (interface{}, error) {
		var err error
		exist, err = dao.GetByID(id, dest)
		if err != nil {
			return nil, err
		}
		return dest, nil
	})
	return exist, err
}

func (dao *Dao) ListByIDs(dest interface{}, ids ...interface{}) error {
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
		SQL()
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
		NameSQL()
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
		SQL()
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
		SQL()
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

func (dao *Dao) DeleteByID(id interface{}) (bool, error) {
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

func (dao *Dao) DeleteCache(field string, values ...interface{}) error {
	items := make([]string, len(values))
	for i, item := range values {
		items[i] = utils.ToString(item)
	}
	return dao.CacheProvider.BatchDel(field, items...)
}
