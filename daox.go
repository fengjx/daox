package daox

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
	"github.com/fengjx/daox/utils"
)

var (
	ErrUpdatePrimaryKeyRequire = errors.New("[daox] Primary key require for update")
	ErrTxNil                   = errors.New("[daox] Tx is nil")

	// defaultMasterDB 全局默认master数据库
	defaultMasterDB *DB
	// defaultReadDB 全局默认read数据库
	defaultReadDB *DB

	metaMap     = map[string]TableMeta{}
	metaMapLock sync.Mutex
)

func UseDefaultMasterDB(master *sqlx.DB) {
	defaultMasterDB = NewDB(master)
}

func UseDefaultReadDB(read *sqlx.DB) {
	defaultMasterDB = NewDB(read)
}

type Dao struct {
	masterDB  *DB
	ReadDB    *DB
	Mapper    *reflectx.Mapper
	TableMeta TableMeta
}

// CreateDAO 函数用于创建一个新的Dao对象
// tableName 参数表示表名
// primaryKey 参数表示主键
// structType 参数表示数据结构类型
// opts 参数表示可选的选项
// 返回值为创建的Dao对象指针
func CreateDAO(tableName string, primaryKey string, structType reflect.Type, opts ...Option) *Dao {
	dao := &Dao{}
	columns := dao.GetColumnsByType(structType)
	dao.TableMeta = TableMeta{
		TableName:  tableName,
		StructType: structType,
		PrimaryKey: primaryKey,
		Columns:    columns,
	}
	for _, opt := range opts {
		opt(dao)
	}
	registerMeta(dao.TableMeta)
	return dao
}

// NewDAO 函数用于创建一个新的Dao对象
// master 参数用于连接数据库
// tableName 参数表示表名
// primaryKey 参数表示主键
// structType 参数表示数据结构类型
// opts 参数表示可选的选项
// 返回值为创建的Dao对象指针
func NewDAO(master *sqlx.DB, tableName string, primaryKey string, structType reflect.Type, opts ...Option) *Dao {
	dao := &Dao{
		masterDB: NewDB(master),
	}
	columns := dao.GetColumnsByType(structType)
	dao.TableMeta = TableMeta{
		TableName:  tableName,
		StructType: structType,
		PrimaryKey: primaryKey,
		Columns:    columns,
	}
	for _, opt := range opts {
		opt(dao)
	}
	registerMeta(dao.TableMeta)
	return dao
}

func registerMeta(meta TableMeta) {
	metaMapLock.Lock()
	defer metaMapLock.Unlock()
	metaMap[meta.TableName] = meta
}

// GetMeta 根据表名获得元信息
func GetMeta(tableName string) TableMeta {
	return metaMap[tableName]
}

// SQLBuilder 创建当前表的 sqlbuilder
func (dao *Dao) SQLBuilder() *sqlbuilder.Builder {
	return sqlbuilder.New(dao.TableMeta.TableName)
}

// Selector 创建当前表的 selector
// columns 是查询指定字段，为空则是全部字段
func (dao *Dao) Selector(columns ...string) *sqlbuilder.Selector {
	if len(columns) == 0 {
		columns = dao.DBColumns()
	}
	return sqlbuilder.New(dao.TableMeta.TableName).Select(columns...)
}

// GetColumnsByModel 根据 model 结构获取数据库字段
// omitColumns 表示需要忽略的字段
func (dao *Dao) GetColumnsByModel(model interface{}, omitColumns ...string) []string {
	return dao.GetColumnsByType(reflect.TypeOf(model), omitColumns...)
}

// GetColumnsByType 通过字段 tag 解析数据库字段
func (dao *Dao) GetColumnsByType(typ reflect.Type, omitColumns ...string) []string {
	return sqlbuilder.GetColumnsByType(dao.getMapper(), typ, omitColumns...)
}

// DBColumns 获取当前表数据库字段
// omitColumns 表示需要忽略的字段
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

// TableName 获取当前表名
func (dao *Dao) TableName() string {
	return dao.TableMeta.TableName
}

// Save 插入数据
// omitColumns 不需要 insert 的字段
func (dao *Dao) Save(dest Model, omitColumns ...string) (int64, error) {
	return dao.saveTx(nil, dest, omitColumns...)
}

// SaveTx 插入数据，支持事务
// omitColumns 不需要 insert 的字段
func (dao *Dao) SaveTx(tx *sqlx.Tx, dest Model, omitColumns ...string) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.saveTx(tx, dest, omitColumns...)
}

func (dao *Dao) saveTx(tx *sqlx.Tx, dest Model, omitColumns ...string) (int64, error) {
	tableMeta := dao.TableMeta
	if tableMeta.IsAutoIncrement {
		omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	}
	columns := tableMeta.OmitColumns(omitColumns...)
	execSql, err := dao.SQLBuilder().Insert().Columns(columns...).NameSQL()
	if err != nil {
		return 0, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.GetMasterDB().NamedExec(execSql, dest)
	} else {
		res, err = tx.NamedExec(execSql, dest)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ReplaceInto replace into table
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
	res, err := dao.GetMasterDB().NamedExec(execSql, dest)
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
	res, err := dao.GetMasterDB().NamedExec(execSQL, models)
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
	res, err := dao.GetMasterDB().NamedExec(execSQL, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetTx 根据查询条件查询单条记录，支持事务
// dest 必须是一个指针
func (dao *Dao) GetTx(tx *sqlx.Tx, dest interface{}, selector *sqlbuilder.Selector) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.getTx(tx, dest, selector)
}

func (dao *Dao) getTx(tx *sqlx.Tx, dest interface{}, selector *sqlbuilder.Selector) (bool, error) {
	querySQL, args, err := selector.SQLArgs()
	if err != nil {
		return false, err
	}
	if tx == nil {
		err = dao.GetReadDB().Get(dest, querySQL, args...)
	} else {
		err = tx.Get(dest, querySQL, args...)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// SelectTx 根据查询条件查询列表
// dest 必须是一个 slice 指针
func (dao *Dao) SelectTx(tx *sqlx.Tx, dest interface{}, selector *sqlbuilder.Selector) error {
	if err := dao.checkTxNil(tx); err != nil {
		return err
	}
	return dao.selectTx(tx, dest, selector)
}

func (dao *Dao) selectTx(tx *sqlx.Tx, dest interface{}, selector *sqlbuilder.Selector) error {
	querySQL, args, err := selector.SQLArgs()
	if err != nil {
		return err
	}
	if tx == nil {
		err = dao.GetReadDB().Select(dest, querySQL, args...)
	} else {
		err = tx.Select(dest, querySQL, args...)
	}
	return err
}

// GetByColumn 按指定字段查询单条数据
// bool 数据是否存在
func (dao *Dao) GetByColumn(kv *KV, dest Model) (bool, error) {
	return dao.getByColumnTx(nil, kv, dest)
}

// GetByColumnTx 按指定字段查询单条数据，支持事务
// bool 数据是否存在
func (dao *Dao) GetByColumnTx(tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.getByColumnTx(tx, kv, dest)
}

func (dao *Dao) getByColumnTx(tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	if kv == nil {
		return false, nil
	}
	selector := dao.Selector().
		Where(ql.C().And(ql.Col(kv.Key).EQ(kv.Value)))
	return dao.getTx(tx, dest, selector)
}

// ListByColumns 指定字段多个值查询多条数据
// dest: slice pointer
func (dao *Dao) ListByColumns(kvs *MultiKV, dest interface{}) error {
	return dao.listByColumnsTx(nil, kvs, dest)
}

// ListByColumnsTx 指定字段多个值查询多条数据，支持事务
func (dao *Dao) ListByColumnsTx(tx *sqlx.Tx, kvs *MultiKV, dest interface{}) error {
	if err := dao.checkTxNil(tx); err != nil {
		return err
	}
	return dao.listByColumnsTx(tx, kvs, dest)
}

func (dao *Dao) listByColumnsTx(tx *sqlx.Tx, kvs *MultiKV, dest interface{}) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	selector := dao.Selector().
		Columns(dao.DBColumns()...).
		Where(ql.C().And(ql.Col(kvs.Key).In(kvs.Values...)))
	return dao.selectTx(tx, dest, selector)
}

// List 指定字段查询多条数据
func (dao *Dao) List(kv *KV, dest interface{}) error {
	return dao.listTx(nil, kv, dest)
}

// ListTx 指定字段查询多条数据，支持事务
func (dao *Dao) ListTx(tx *sqlx.Tx, kv *KV, dest interface{}) error {
	if err := dao.checkTxNil(tx); err != nil {
		return err
	}
	return dao.listTx(tx, kv, dest)
}

func (dao *Dao) listTx(tx *sqlx.Tx, kv *KV, dest interface{}) error {
	if kv == nil {
		return nil
	}
	selector := dao.Selector().
		Columns(dao.DBColumns()...).
		Where(ql.C().And(ql.Col(kv.Key).EQ(kv.Value)))
	return dao.SelectTx(tx, dest, selector)
}

// GetByID 根据 id 查询单条数据
func (dao *Dao) GetByID(id interface{}, dest Model) (bool, error) {
	tableMeta := dao.TableMeta
	return dao.GetByColumn(OfKv(tableMeta.PrimaryKey, id), dest)
}

// ListByIDs 根据 id 查询多条数据
func (dao *Dao) ListByIDs(dest interface{}, ids ...interface{}) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumns(OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// ListByIdsTx 查询多个id值
func (dao *Dao) ListByIdsTx(tx *sqlx.Tx, dest interface{}, ids ...interface{}) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumnsTx(tx, OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// Select 根据查询条件查询列表
// dest 必须是一个 slice 指针
func (dao *Dao) Select(dest interface{}, selector *sqlbuilder.Selector) error {
	return dao.selectTx(nil, dest, selector)
}

// Get 根据查询条件查询单条记录
// dest 必须是一个指针
func (dao *Dao) Get(dest interface{}, selector *sqlbuilder.Selector) (bool, error) {
	return dao.getTx(nil, dest, selector)
}

// UpdateByCond 根据条件更新字段
// attr 字段更新值
func (dao *Dao) UpdateByCond(attr map[string]interface{}, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.updateByCondTx(nil, attr, where)
}

// UpdateByCondTx 根据条件更新字段，支持事务
func (dao *Dao) UpdateByCondTx(tx *sqlx.Tx, attr map[string]interface{}, where sqlbuilder.ConditionBuilder) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.updateByCondTx(tx, attr, where)
}

func (dao *Dao) updateByCondTx(tx *sqlx.Tx, attr map[string]interface{}, where sqlbuilder.ConditionBuilder) (int64, error) {
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
		res, err = dao.GetMasterDB().Exec(updateSQL, args...)
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
	return dao.updateFieldTx(nil, idValue, fieldMap)
}

// UpdateFieldTx 部分字段更新，支持事务
func (dao *Dao) UpdateFieldTx(tx *sqlx.Tx, idValue interface{}, fieldMap map[string]interface{}) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.updateFieldTx(tx, idValue, fieldMap)
}

func (dao *Dao) updateFieldTx(tx *sqlx.Tx, idValue interface{}, attr map[string]interface{}) (bool, error) {
	if utils.IsIDEmpty(idValue) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	rows, err := dao.updateByCondTx(tx, attr, ql.C().
		And(ql.Col(tableMeta.PrimaryKey).EQ(idValue)),
	)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// Update 全字段更新
func (dao *Dao) Update(m Model, omitColumns ...string) (bool, error) {
	return dao.updateTx(nil, m, omitColumns...)
}

// UpdateTx 全字段更新，支持事务
func (dao *Dao) UpdateTx(tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.updateTx(tx, m, omitColumns...)
}

func (dao *Dao) updateTx(tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
	if utils.IsIDEmpty(m.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	updateSQL, err := dao.SQLBuilder().Update().
		Columns(dao.DBColumns(omitColumns...)...).
		Where(
			ql.SC().And(fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey)),
		).NameSQL()
	if err != nil {
		return false, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.GetMasterDB().NamedExec(updateSQL, m)
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
	return dao.deleteByCondTx(nil, where)
}

// DeleteByCondTx 根据where条件删除，支持事务
func (dao *Dao) DeleteByCondTx(tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.deleteByCondTx(tx, where)
}

func (dao *Dao) deleteByCondTx(tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	deleteSQL, args, err := dao.SQLBuilder().Delete().Where(where).SQLArgs()
	if err != nil {
		return 0, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.GetMasterDB().Exec(deleteSQL, args...)
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
	return dao.deleteByColumnTx(nil, kv)
}

// DeleteByColumnTx 按字段名删除，支持事务
func (dao *Dao) DeleteByColumnTx(tx *sqlx.Tx, kv *KV) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.deleteByColumnTx(tx, kv)
}

func (dao *Dao) deleteByColumnTx(tx *sqlx.Tx, kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
	return dao.deleteByCondTx(tx, ql.C().And(ql.Col(kv.Key).EQ(kv.Value)))
}

// DeleteByColumns 指定字段删除多个值
func (dao *Dao) DeleteByColumns(kvs *MultiKV) (int64, error) {
	return dao.deleteByColumnsTx(nil, kvs)
}

// DeleteByColumnsTx 指定字段多个值删除
func (dao *Dao) DeleteByColumnsTx(tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.deleteByColumnsTx(tx, kvs)
}

func (dao *Dao) deleteByColumnsTx(tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}
	return dao.deleteByCondTx(tx, ql.C().And(ql.Col(kvs.Key).In(kvs.Values...)))
}

// DeleteByID 根据id删除数据
func (dao *Dao) DeleteByID(id interface{}) (bool, error) {
	return dao.deleteByIDTx(nil, id)
}

// DeleteByIDTx 根据id删除数据，支持事务
func (dao *Dao) DeleteByIDTx(tx *sqlx.Tx, id interface{}) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.deleteByIDTx(tx, id)
}

func (dao *Dao) deleteByIDTx(tx *sqlx.Tx, id interface{}) (bool, error) {
	tableMeta := dao.TableMeta
	affected, err := dao.deleteByColumnTx(tx, OfKv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (dao *Dao) checkTxNil(tx *sqlx.Tx) error {
	if tx == nil {
		return ErrTxNil
	}
	return nil
}

// With 使用新的数据库连接创建 Dao
func (dao *Dao) With(master, read *sqlx.DB, opts ...Option) *Dao {
	newDao := &Dao{
		masterDB:  NewDB(master),
		ReadDB:    NewDB(read),
		TableMeta: dao.TableMeta,
	}
	for _, opt := range opts {
		opt(newDao)
	}
	return newDao
}

func (dao *Dao) getMapper() *reflectx.Mapper {
	if dao.Mapper != nil {
		return dao.Mapper
	}
	return dao.GetMasterDB().Mapper
}

func (dao *Dao) GetMasterDB() *DB {
	if dao.masterDB != nil {
		return dao.masterDB
	}
	return defaultMasterDB
}

func (dao *Dao) GetReadDB() *DB {
	if dao.ReadDB != nil {
		return dao.ReadDB
	}
	if dao.masterDB != nil {
		return dao.masterDB
	}
	if defaultReadDB != nil {
		return defaultReadDB
	}
	return defaultMasterDB
}
