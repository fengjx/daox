package daox

import (
	"context"
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
	defaultMasterDB *sqlx.DB
	// defaultReadDB 全局默认read数据库
	defaultReadDB *sqlx.DB

	metaMap     = map[string]TableMeta{}
	metaMapLock sync.Mutex
)

func UseDefaultMasterDB(master *sqlx.DB) {
	defaultMasterDB = master
}

func UseDefaultReadDB(read *sqlx.DB) {
	defaultMasterDB = read
}

type Dao struct {
	masterDB  *sqlx.DB
	readDB    *sqlx.DB
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
		masterDB: master,
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
func GetMeta(tableName string) (meta TableMeta, ok bool) {
	meta, ok = metaMap[tableName]
	return
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
func (dao *Dao) GetColumnsByModel(model any, omitColumns ...string) []string {
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
	return dao.SaveContext(context.Background(), dest, omitColumns...)
}

// SaveContext 插入数据，携带上下文
// omitColumns 不需要 insert 的字段
func (dao *Dao) SaveContext(ctx context.Context, dest Model, omitColumns ...string) (int64, error) {
	return dao.saveContext(ctx, nil, dest, omitColumns...)
}

// SaveTx 插入数据，支持事务
// omitColumns 不需要 insert 的字段
func (dao *Dao) SaveTx(tx *sqlx.Tx, dest Model, omitColumns ...string) (int64, error) {
	return dao.SaveTxContext(context.Background(), tx, dest, omitColumns...)
}

// SaveTxContext 插入数据，支持事务，携带上下文
// omitColumns 不需要 insert 的字段
func (dao *Dao) SaveTxContext(ctx context.Context, tx *sqlx.Tx, dest Model, omitColumns ...string) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.saveContext(ctx, tx, dest, omitColumns...)
}

func (dao *Dao) saveContext(ctx context.Context, tx *sqlx.Tx, dest Model, omitColumns ...string) (int64, error) {
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
		res, err = dao.GetMasterDB().NamedExecContext(ctx, execSql, dest)
	} else {
		res, err = tx.NamedExecContext(ctx, execSql, dest)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ReplaceInto replace into table
// omitColumns 不需要 insert 的字段
func (dao *Dao) ReplaceInto(dest Model, omitColumns ...string) (int64, error) {
	return dao.ReplaceIntoContext(context.Background(), dest, omitColumns...)
}

// ReplaceIntoContext replace into table，携带上下文
// omitColumns 不需要 insert 的字段
func (dao *Dao) ReplaceIntoContext(ctx context.Context, dest Model, omitColumns ...string) (int64, error) {
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
	res, err := dao.GetMasterDB().NamedExecContext(ctx, execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// IgnoreInto 使用 INSERT IGNORE INTO 如果记录已存在则忽略
// omitColumns 不需要 insert 的字段
func (dao *Dao) IgnoreInto(dest Model, omitColumns ...string) (int64, error) {
	return dao.IgnoreIntoContext(context.Background(), dest, omitColumns...)
}

// IgnoreIntoContext 使用 INSERT IGNORE INTO 如果记录已存在则忽略，携带上下文
// omitColumns 不需要 insert 的字段
func (dao *Dao) IgnoreIntoContext(ctx context.Context, dest Model, omitColumns ...string) (int64, error) {
	tableMeta := dao.TableMeta
	if tableMeta.IsAutoIncrement {
		omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	}
	columns := tableMeta.OmitColumns(omitColumns...)
	execSql, err := dao.SQLBuilder().Insert().
		IsIgnoreInto(true).
		Columns(columns...).
		NameSQL()
	if err != nil {
		return 0, err
	}
	res, err := dao.GetMasterDB().NamedExecContext(ctx, execSql, dest)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// BatchSave 批量新增，携带上下文
// omitColumns 不需要 insert 的字段
func (dao *Dao) BatchSave(models any, omitColumns ...string) (int64, error) {
	return dao.BatchSaveContext(context.Background(), models, omitColumns...)
}

// BatchSaveContext 批量新增
// omitColumns 不需要 insert 的字段
func (dao *Dao) BatchSaveContext(ctx context.Context, models any, omitColumns ...string) (int64, error) {
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
	res, err := dao.GetMasterDB().NamedExecContext(ctx, execSQL, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// BatchReplaceInto 批量新增，使用 replace into 方式
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (dao *Dao) BatchReplaceInto(models any, omitColumns ...string) (int64, error) {
	return dao.BatchReplaceIntoContext(context.Background(), models, omitColumns...)
}

// BatchReplaceIntoContext 批量新增，使用 replace into 方式，携带上下文
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (dao *Dao) BatchReplaceIntoContext(ctx context.Context, models any, omitColumns ...string) (int64, error) {
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
	res, err := dao.GetMasterDB().NamedExecContext(ctx, execSQL, models)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Get 根据查询条件查询单条记录
// dest 必须是一个指针
func (dao *Dao) Get(dest any, selector *sqlbuilder.Selector) (bool, error) {
	return dao.GetContext(context.Background(), dest, selector)
}

func (dao *Dao) GetContext(ctx context.Context, dest any, selector *sqlbuilder.Selector) (bool, error) {
	return dao.getContext(ctx, nil, dest, selector)
}

// GetTx 根据查询条件查询单条记录，支持事务
// dest 必须是一个指针
func (dao *Dao) GetTx(tx *sqlx.Tx, dest any, selector *sqlbuilder.Selector) (bool, error) {
	return dao.GetTxContext(context.Background(), tx, dest, selector)
}

// GetTxContext 根据查询条件查询单条记录，支持事务，携带上下文
// dest 必须是一个指针
func (dao *Dao) GetTxContext(ctx context.Context, tx *sqlx.Tx, dest any, selector *sqlbuilder.Selector) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.getContext(ctx, tx, dest, selector)
}

func (dao *Dao) getContext(ctx context.Context, tx *sqlx.Tx, dest any, selector *sqlbuilder.Selector) (bool, error) {
	querySQL, args, err := selector.SQLArgs()
	if err != nil {
		return false, err
	}
	if tx == nil {
		err = dao.GetReadDB().GetContext(ctx, dest, querySQL, args...)
	} else {
		err = tx.GetContext(ctx, dest, querySQL, args...)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Select 根据查询条件查询列表
// dest 必须是一个 slice 指针
func (dao *Dao) Select(dest any, selector *sqlbuilder.Selector) error {
	return dao.SelectContext(context.Background(), dest, selector)
}

// SelectContext 根据查询条件查询列表，携带上下文
// dest 必须是一个 slice 指针
func (dao *Dao) SelectContext(ctx context.Context, dest any, selector *sqlbuilder.Selector) error {
	return dao.selectContext(ctx, nil, dest, selector)
}

// SelectTx 根据查询条件查询列表
// dest 必须是一个 slice 指针
func (dao *Dao) SelectTx(tx *sqlx.Tx, dest any, selector *sqlbuilder.Selector) error {
	return dao.SelectTxContext(context.Background(), tx, dest, selector)
}

// SelectTxContext 根据查询条件查询列表，携带上下文
// dest 必须是一个 slice 指针
func (dao *Dao) SelectTxContext(ctx context.Context, tx *sqlx.Tx, dest any, selector *sqlbuilder.Selector) error {
	if err := dao.checkTxNil(tx); err != nil {
		return err
	}
	return dao.selectContext(ctx, tx, dest, selector)
}

func (dao *Dao) selectContext(ctx context.Context, tx *sqlx.Tx, dest any, selector *sqlbuilder.Selector) error {
	querySQL, args, err := selector.SQLArgs()
	if err != nil {
		return err
	}
	if tx == nil {
		err = dao.GetReadDB().SelectContext(ctx, dest, querySQL, args...)
	} else {
		err = tx.SelectContext(ctx, dest, querySQL, args...)
	}
	return err
}

// GetByColumn 按指定字段查询单条数据
// bool 数据是否存在
func (dao *Dao) GetByColumn(kv *KV, dest Model) (bool, error) {
	return dao.GetByColumnContext(context.Background(), kv, dest)
}

// GetByColumnContext 按指定字段查询单条数据，携带上下文
// bool 数据是否存在
func (dao *Dao) GetByColumnContext(ctx context.Context, kv *KV, dest Model) (bool, error) {
	return dao.getByColumnContext(ctx, nil, kv, dest)
}

// GetByColumnTx 按指定字段查询单条数据，支持事务
// bool 数据是否存在
func (dao *Dao) GetByColumnTx(tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	return dao.GetByColumnTxContext(context.Background(), tx, kv, dest)
}

// GetByColumnTxContext 按指定字段查询单条数据，支持事务，携带上下文
// bool 数据是否存在
func (dao *Dao) GetByColumnTxContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.getByColumnContext(ctx, tx, kv, dest)
}

func (dao *Dao) getByColumnContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	if kv == nil {
		return false, nil
	}
	selector := dao.Selector().
		Where(ql.C().And(ql.Col(kv.Key).EQ(kv.Value)))
	return dao.getContext(ctx, tx, dest, selector)
}

// ListByColumns 指定字段多个值查询多条数据
// dest: slice pointer
func (dao *Dao) ListByColumns(kvs *MultiKV, dest any) error {
	return dao.ListByColumnsContext(context.Background(), kvs, dest)
}

// ListByColumnsContext 指定字段多个值查询多条数据，携带上下文
// dest: slice pointer
func (dao *Dao) ListByColumnsContext(ctx context.Context, kvs *MultiKV, dest any) error {
	return dao.listByColumnsContext(ctx, nil, kvs, dest)
}

// ListByColumnsTx 指定字段多个值查询多条数据，支持事务
func (dao *Dao) ListByColumnsTx(tx *sqlx.Tx, kvs *MultiKV, dest any) error {
	return dao.ListByColumnsTxContext(context.Background(), tx, kvs, dest)
}

// ListByColumnsTxContext 指定字段多个值查询多条数据，支持事务，携带上下文
func (dao *Dao) ListByColumnsTxContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV, dest any) error {
	if err := dao.checkTxNil(tx); err != nil {
		return err
	}
	return dao.listByColumnsContext(ctx, tx, kvs, dest)
}

func (dao *Dao) listByColumnsContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV, dest any) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	selector := dao.Selector().
		Columns(dao.DBColumns()...).
		Where(ql.C().And(ql.Col(kvs.Key).In(kvs.Values...)))
	return dao.selectContext(ctx, tx, dest, selector)
}

// List 指定字段查询多条数据
func (dao *Dao) List(kv *KV, dest any) error {
	return dao.ListContext(context.Background(), kv, dest)
}

// ListContext 指定字段查询多条数据，携带上下文
func (dao *Dao) ListContext(ctx context.Context, kv *KV, dest any) error {
	return dao.listContext(ctx, nil, kv, dest)
}

// ListTx 指定字段查询多条数据，支持事务
func (dao *Dao) ListTx(tx *sqlx.Tx, kv *KV, dest any) error {
	return dao.ListTxContext(context.Background(), tx, kv, dest)
}

// ListTxContext 指定字段查询多条数据，支持事务，携带上下文
func (dao *Dao) ListTxContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest any) error {
	if err := dao.checkTxNil(tx); err != nil {
		return err
	}
	return dao.listContext(ctx, tx, kv, dest)
}

func (dao *Dao) listContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest any) error {
	if kv == nil {
		return nil
	}
	selector := dao.Selector().
		Columns(dao.DBColumns()...).
		Where(ql.C().And(ql.Col(kv.Key).EQ(kv.Value)))
	return dao.selectContext(ctx, tx, dest, selector)
}

// GetByID 根据 id 查询单条数据
func (dao *Dao) GetByID(id any, dest Model) (bool, error) {
	return dao.GetByIDContext(context.Background(), id, dest)
}

// GetByIDContext 根据 id 查询单条数据，携带上下文
func (dao *Dao) GetByIDContext(ctx context.Context, id any, dest Model) (bool, error) {
	tableMeta := dao.TableMeta
	return dao.GetByColumnContext(ctx, OfKv(tableMeta.PrimaryKey, id), dest)
}

// ListByIDs 根据 id 查询多条数据
func (dao *Dao) ListByIDs(dest any, ids ...any) error {
	return dao.ListByIDsContext(context.Background(), dest, ids...)
}

// ListByIDsContext 根据 id 查询多条数据，携带上下文
func (dao *Dao) ListByIDsContext(ctx context.Context, dest any, ids ...any) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumnsContext(ctx, OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// ListByIDsTx 查询多个id值，支持事务
func (dao *Dao) ListByIDsTx(tx *sqlx.Tx, dest any, ids ...any) error {
	return dao.ListByIDsTxContext(context.Background(), tx, dest, ids...)
}

// ListByIDsTxContext 查询多个id值，支持事务，携带上下文
func (dao *Dao) ListByIDsTxContext(ctx context.Context, tx *sqlx.Tx, dest any, ids ...any) error {
	tableMeta := dao.TableMeta
	return dao.ListByColumnsTxContext(ctx, tx, OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// UpdateByCond 根据条件更新字段
// attr 字段更新值
func (dao *Dao) UpdateByCond(attr map[string]any, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.UpdateByCondContext(context.Background(), attr, where)
}

// UpdateByCondContext 根据条件更新字段，携带上下文
// attr 字段更新值
func (dao *Dao) UpdateByCondContext(ctx context.Context, attr map[string]any, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.updateByCondContext(ctx, nil, attr, where)
}

// UpdateByCondTx 根据条件更新字段，支持事务
func (dao *Dao) UpdateByCondTx(tx *sqlx.Tx, attr map[string]any, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.UpdateByCondTxContext(context.Background(), tx, attr, where)
}

// UpdateByCondTxContext 根据条件更新字段，支持事务，携带上下文
func (dao *Dao) UpdateByCondTxContext(ctx context.Context, tx *sqlx.Tx, attr map[string]any, where sqlbuilder.ConditionBuilder) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.updateByCondContext(ctx, tx, attr, where)
}

func (dao *Dao) updateByCondContext(ctx context.Context, tx *sqlx.Tx, attr map[string]any, where sqlbuilder.ConditionBuilder) (int64, error) {
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
		res, err = dao.GetMasterDB().ExecContext(ctx, updateSQL, args...)
	} else {
		res, err = tx.ExecContext(ctx, updateSQL, args...)
	}
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateField 部分字段更新
func (dao *Dao) UpdateField(idValue any, fieldMap map[string]any) (bool, error) {
	return dao.UpdateFieldContext(context.Background(), idValue, fieldMap)
}

// UpdateFieldContext 部分字段更新，携带上下文
func (dao *Dao) UpdateFieldContext(ctx context.Context, idValue any, fieldMap map[string]any) (bool, error) {
	return dao.updateFieldContext(ctx, nil, idValue, fieldMap)
}

// UpdateFieldTx 部分字段更新，支持事务
func (dao *Dao) UpdateFieldTx(tx *sqlx.Tx, idValue any, fieldMap map[string]any) (bool, error) {
	return dao.UpdateFieldTxContext(context.Background(), tx, idValue, fieldMap)
}

// UpdateFieldTxContext 部分字段更新，支持事务，携带上下文
func (dao *Dao) UpdateFieldTxContext(ctx context.Context, tx *sqlx.Tx, idValue any, fieldMap map[string]any) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.updateFieldContext(ctx, tx, idValue, fieldMap)
}

func (dao *Dao) updateFieldContext(ctx context.Context, tx *sqlx.Tx, idValue any, fieldMap map[string]any) (bool, error) {
	if utils.IsIDEmpty(idValue) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := dao.TableMeta
	rows, err := dao.updateByCondContext(ctx, tx, fieldMap, ql.C().
		And(ql.Col(tableMeta.PrimaryKey).EQ(idValue)),
	)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// Update 全字段更新
func (dao *Dao) Update(m Model, omitColumns ...string) (bool, error) {
	return dao.UpdateContext(context.Background(), m, omitColumns...)
}

// UpdateContext 全字段更新，携带上下文
func (dao *Dao) UpdateContext(ctx context.Context, m Model, omitColumns ...string) (bool, error) {
	return dao.updateContext(ctx, nil, m, omitColumns...)
}

// UpdateTx 全字段更新，支持事务
func (dao *Dao) UpdateTx(tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
	return dao.UpdateTxContext(context.Background(), tx, m, omitColumns...)
}

// UpdateTxContext 全字段更新，支持事务，携带上下文
func (dao *Dao) UpdateTxContext(ctx context.Context, tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.updateContext(ctx, tx, m, omitColumns...)
}

func (dao *Dao) updateContext(ctx context.Context, tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
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
		res, err = dao.GetMasterDB().NamedExecContext(ctx, updateSQL, m)
	} else {
		res, err = tx.NamedExecContext(ctx, updateSQL, m)
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
	return dao.DeleteByCondContext(context.Background(), where)
}

// DeleteByCondContext 根据where条件删除，携带上下文
func (dao *Dao) DeleteByCondContext(ctx context.Context, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.deleteByCondContext(ctx, nil, where)
}

// DeleteByCondTx 根据where条件删除，支持事务
func (dao *Dao) DeleteByCondTx(tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	return dao.DeleteByCondTxContext(context.Background(), tx, where)
}

// DeleteByCondTxContext 根据where条件删除，支持事务，携带上下文
func (dao *Dao) DeleteByCondTxContext(ctx context.Context, tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.deleteByCondContext(ctx, tx, where)
}

func (dao *Dao) deleteByCondContext(ctx context.Context, tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	deleteSQL, args, err := dao.SQLBuilder().Delete().Where(where).SQLArgs()
	if err != nil {
		return 0, err
	}
	var res sql.Result
	if tx == nil {
		res, err = dao.GetMasterDB().ExecContext(ctx, deleteSQL, args...)
	} else {
		res, err = tx.ExecContext(ctx, deleteSQL, args...)
	}
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteByColumn 按字段名删除
func (dao *Dao) DeleteByColumn(kv *KV) (int64, error) {
	return dao.DeleteByColumnContext(context.Background(), kv)
}

// DeleteByColumnContext 按字段名删除，携带上下文
func (dao *Dao) DeleteByColumnContext(ctx context.Context, kv *KV) (int64, error) {
	return dao.deleteByColumnContext(ctx, nil, kv)
}

// DeleteByColumnTx 按字段名删除，支持事务
func (dao *Dao) DeleteByColumnTx(tx *sqlx.Tx, kv *KV) (int64, error) {
	return dao.DeleteByColumnTxContext(context.Background(), tx, kv)
}

// DeleteByColumnTxContext 按字段名删除，支持事务，携带上下文
func (dao *Dao) DeleteByColumnTxContext(ctx context.Context, tx *sqlx.Tx, kv *KV) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.deleteByColumnContext(ctx, tx, kv)
}

func (dao *Dao) deleteByColumnContext(ctx context.Context, tx *sqlx.Tx, kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
	return dao.deleteByCondContext(ctx, tx, ql.C().And(ql.Col(kv.Key).EQ(kv.Value)))
}

// DeleteByColumns 指定字段删除多个值
func (dao *Dao) DeleteByColumns(kvs *MultiKV) (int64, error) {
	return dao.DeleteByColumnsContext(context.Background(), kvs)
}

// DeleteByColumnsContext 指定字段删除多个值，携带上下文
func (dao *Dao) DeleteByColumnsContext(ctx context.Context, kvs *MultiKV) (int64, error) {
	return dao.deleteByColumnsContext(ctx, nil, kvs)
}

// DeleteByColumnsTx 指定字段多个值删除
func (dao *Dao) DeleteByColumnsTx(tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	return dao.DeleteByColumnsTxContext(context.Background(), tx, kvs)
}

// DeleteByColumnsTxContext 指定字段多个值删除，携带上下文
func (dao *Dao) DeleteByColumnsTxContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return 0, err
	}
	return dao.deleteByColumnsContext(ctx, tx, kvs)
}

func (dao *Dao) deleteByColumnsContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}
	return dao.deleteByCondContext(ctx, tx, ql.C().And(ql.Col(kvs.Key).In(kvs.Values...)))
}

// DeleteByID 根据id删除数据
func (dao *Dao) DeleteByID(id any) (bool, error) {
	return dao.DeleteByIDContext(context.Background(), id)
}

// DeleteByIDContext 根据id删除数据，携带上下文
func (dao *Dao) DeleteByIDContext(ctx context.Context, id any) (bool, error) {
	return dao.deleteByIDContext(ctx, nil, id)
}

// DeleteByIDTx 根据id删除数据，支持事务
func (dao *Dao) DeleteByIDTx(tx *sqlx.Tx, id any) (bool, error) {
	return dao.DeleteByIDTxContext(context.Background(), tx, id)
}

// DeleteByIDTxContext 根据id删除数据，支持事务，携带上下文
func (dao *Dao) DeleteByIDTxContext(ctx context.Context, tx *sqlx.Tx, id any) (bool, error) {
	if err := dao.checkTxNil(tx); err != nil {
		return false, err
	}
	return dao.deleteByIDContext(ctx, tx, id)
}

func (dao *Dao) deleteByIDContext(ctx context.Context, tx *sqlx.Tx, id any) (bool, error) {
	tableMeta := dao.TableMeta
	affected, err := dao.deleteByColumnContext(ctx, tx, OfKv(tableMeta.PrimaryKey, id))
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
		masterDB:  master,
		readDB:    read,
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

func (dao *Dao) GetMasterDB() *sqlx.DB {
	if dao.masterDB != nil {
		return dao.masterDB
	}
	return defaultMasterDB
}

func (dao *Dao) GetReadDB() *sqlx.DB {
	if dao.readDB != nil {
		return dao.readDB
	}
	if dao.masterDB != nil {
		return dao.masterDB
	}
	if defaultReadDB != nil {
		return defaultReadDB
	}
	return defaultMasterDB
}
