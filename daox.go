package daox

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
	"github.com/fengjx/daox/utils"
)

var (
	ErrUpdatePrimaryKeyRequire = errors.New("[daox] Primary key require for update")
)

type SliceToMapFun = func([]*Model) map[interface{}]*Model

type Dao struct {
	DBMaster  *DB
	DBRead    *DB
	TableMeta *TableMeta
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

func (dao *Dao) Selector(columns ...string) *sqlbuilder.Selector {
	if len(columns) == 0 {
		columns = dao.DBColumns()
	}
	return sqlbuilder.New(dao.TableMeta.TableName).Select(columns...)
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
		return 0, err
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
		return 0, err
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
		return 0, err
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
		return 0, err
	}
	res, err := dao.DBMaster.NamedExec(execSQL, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetByColumn 按指定字段查询单条数据
// bool 数据是否存在
func (dao *Dao) GetByColumn(kv *KV, dest Model) (bool, error) {
	if kv == nil {
		return false, nil
	}
	selector := dao.SQLBuilder().Select().
		Where(ql.EC().Where(ql.Col(kv.Key).EQ(kv.Value)))
	return dao.Get(dest, selector)
}

// ListByColumns 指定字段多个值查询多条数据
// dest: slice pointer
func (dao *Dao) ListByColumns(kvs *MultiKV, dest interface{}) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	selector := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
		Where(ql.EC().Where(ql.Col(kvs.Key).In(kvs.Values...)))
	return dao.Select(dest, selector)
}

// List 指定字段查询多条数据
func (dao *Dao) List(kv *KV, dest interface{}) error {
	if kv == nil {
		return nil
	}
	selector := dao.SQLBuilder().Select().
		Columns(dao.DBColumns()...).
		Where(ql.EC().Where(ql.Col(kv.Key).EQ(kv.Value)))
	return dao.Select(dest, selector)
}

func (dao *Dao) GetByID(id interface{}, dest Model) (bool, error) {
	tableMeta := dao.TableMeta
	return dao.GetByColumn(OfKv(tableMeta.PrimaryKey, id), dest)
}

func (dao *Dao) ListByIDs(dest interface{}, ids ...interface{}) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumns(OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// Select 根据查询条件查询列表
// dest 必须是一个 slice 指针
func (dao *Dao) Select(dest interface{}, selector *sqlbuilder.Selector) error {
	querySQL, args, err := selector.SQLArgs()
	if err != nil {
		return err
	}
	return dao.DBRead.Select(dest, querySQL, args...)
}

// Get 根据查询条件查询单条记录
// dest 必须是一个指针
func (dao *Dao) Get(dest interface{}, selector *sqlbuilder.Selector) (bool, error) {
	querySQL, args, err := selector.SQLArgs()
	if err != nil {
		return false, err
	}
	err = dao.DBRead.Get(dest, querySQL, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateByCond 根据条件更新字段
func (dao *Dao) UpdateByCond(attr map[string]interface{}, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.UpdateByCondTX(nil, attr, where)
}

// UpdateByCondTX 根据条件更新字段，支持事务
func (dao *Dao) UpdateByCondTX(tx *sqlx.Tx, attr map[string]interface{}, where sqlbuilder.ConditionBuilder) (int64, error) {
	updater := dao.SQLBuilder().Update()
	for col, val := range attr {
		updater.Set(col, val)
	}
	updater.Where(where)
	updateSQL, args, err := updater.SQLArgs()
	if err != nil {
		return 0, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.DBMaster.Exec(updateSQL, args...)
	} else {
		res, err = tx.Exec(updateSQL, args...)
	}
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateField 部分字段更新
func (dao *Dao) UpdateField(idValue interface{}, fieldMap map[string]interface{}) (bool, error) {
	return dao.UpdateFieldTx(nil, idValue, fieldMap)
}

// UpdateFieldTx 部分字段更新，支持事务
func (dao *Dao) UpdateFieldTx(tx *sqlx.Tx, idValue interface{}, attr map[string]interface{}) (bool, error) {
	if utils.IsIDEmpty(idValue) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	rows, err := dao.UpdateByCondTX(tx, attr, ql.EC().
		Where(ql.Col(tableMeta.PrimaryKey).EQ(idValue)),
	)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// Update 全字段更新
func (dao *Dao) Update(m Model) (bool, error) {
	return dao.UpdateTx(nil, m)
}

// UpdateTx 全字段更新，支持事务
func (dao *Dao) UpdateTx(tx *sqlx.Tx, m Model) (bool, error) {
	if utils.IsIDEmpty(m.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	updateSQL, err := dao.SQLBuilder().Update().
		Columns(dao.DBColumns(tableMeta.PrimaryKey)...).
		Where(
			ql.SC().Where(fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey)),
		).NameSQL()
	if err != nil {
		return false, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.DBMaster.NamedExec(updateSQL, m)
	} else {
		res, err = tx.NamedExec(updateSQL, m)
	}
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// DeleteByCond 根据where条件删除
func (dao *Dao) DeleteByCond(where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.DeleteByCondTX(nil, where)
}

// DeleteByCondTX 根据where条件删除，支持事务
func (dao *Dao) DeleteByCondTX(tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	deleteSQL, args, err := dao.SQLBuilder().Delete().Where(where).SQLArgs()
	if err != nil {
		return 0, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.DBMaster.Exec(deleteSQL, args...)
	} else {
		res, err = tx.Exec(deleteSQL, args...)
	}
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteByColumn 按字段名删除
func (dao *Dao) DeleteByColumn(kv *KV) (int64, error) {
	return dao.DeleteByColumnTx(nil, kv)
}

// DeleteByColumnTx 按字段名删除，支持事务
func (dao *Dao) DeleteByColumnTx(tx *sqlx.Tx, kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
	return dao.DeleteByCondTX(tx, ql.EC().Where(ql.Col(kv.Key).EQ(kv.Value)))
}

// DeleteByColumns 指定字段删除多个值
func (dao *Dao) DeleteByColumns(kvs *MultiKV) (int64, error) {
	return dao.DeleteByColumnsTx(nil, kvs)
}

// DeleteByColumnsTx 指定字段多个值删除
func (dao *Dao) DeleteByColumnsTx(tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}
	return dao.DeleteByCondTX(tx, ql.EC().Where(ql.Col(kvs.Key).In(kvs.Values...)))
}

// DeleteByID 根据id删除数据
func (dao *Dao) DeleteByID(id interface{}) (bool, error) {
	return dao.DeleteByIDTx(nil, id)
}

// DeleteByIDTx 根据id删除数据，支持事务
func (dao *Dao) DeleteByIDTx(tx *sqlx.Tx, id interface{}) (bool, error) {
	tableMeta := dao.TableMeta
	affected, err := dao.DeleteByColumnTx(tx, OfKv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}
