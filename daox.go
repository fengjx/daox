package daox

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	"github.com/fengjx/daox/engine"
	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
	"github.com/fengjx/daox/utils"
)

var (
	ErrUpdatePrimaryKeyRequire = errors.New("[daox] Primary key require for update")
	ErrTxNil                   = errors.New("[daox] Tx is nil")
)

// Dao 数据访问
type Dao struct {
	masterDB    *sqlx.DB
	readDB      *sqlx.DB
	Mapper      *reflectx.Mapper
	TableMeta   TableMeta
	ifNullVals  map[string]string
	omitColumns []string
}

// NewDao 创建一个新的 dao 对象
// tableName 参数表示表名
// primaryKey 参数表示主键
// structType 参数表示数据结构类型
// opts 参数表示可选的选项
// 返回值为创建的Dao对象指针
func NewDao[T Model](tableName string, primaryKey string, opts ...Option) *Dao {
	dao := &Dao{}
	dao.TableMeta = TableMeta{
		TableName:  tableName,
		PrimaryKey: primaryKey,
	}
	for _, opt := range opts {
		opt(dao)
	}
	structType := reflect.TypeFor[T]()
	columns := dao.GetColumnsByType(structType, dao.omitColumns...)
	dao.TableMeta.Columns = columns
	global.registerMeta(dao.TableMeta)
	return dao
}

// NewDaoByMeta 根据 meta 接口创建 dao 对象
func NewDaoByMeta(m Meta, opts ...Option) *Dao {
	dao := &Dao{}
	dao.TableMeta = TableMeta{
		TableName:       m.TableName(),
		PrimaryKey:      m.PrimaryKey(),
		Columns:         m.Columns(),
		IsAutoIncrement: m.IsAutoIncrement(),
	}
	for _, opt := range opts {
		opt(dao)
	}
	global.registerMeta(dao.TableMeta)
	return dao
}

// SQLBuilder 创建当前表的 sqlbuilder
func (d *Dao) SQLBuilder() *sqlbuilder.Builder {
	return sqlbuilder.New(d.TableMeta.TableName)
}

// Selector 创建当前表的 Selector
// columns 是查询指定字段，为空则是全部字段
func (d *Dao) Selector(columns ...string) *sqlbuilder.Selector {
	if len(columns) == 0 {
		columns = d.DBColumns()
	}
	selector := sqlbuilder.New(d.TableMeta.TableName).Select(columns...)
	if len(d.ifNullVals) > 0 {
		selector.IfNullVals(d.ifNullVals)
	}
	selector.Queryer(d.GetReadDB())
	return selector
}

// Updater 创建当前表的 Updater
func (d *Dao) Updater() *sqlbuilder.Updater {
	return d.SQLBuilder().Update().Execer(d.GetMasterDB())
}

// Deleter 创建当前表的 Deleter
func (d *Dao) Deleter() *sqlbuilder.Deleter {
	return d.SQLBuilder().Delete().Execer(d.GetMasterDB())
}

// Inserter 创建当前表的 updater
func (d *Dao) Inserter() *sqlbuilder.Inserter {
	return d.SQLBuilder().Insert().Execer(d.GetMasterDB())
}

// GetColumnsByModel 根据 model 结构获取数据库字段
// omitColumns 表示需要忽略的字段
func (d *Dao) GetColumnsByModel(model any, omitColumns ...string) []string {
	return d.GetColumnsByType(reflect.TypeOf(model), omitColumns...)
}

// GetColumnsByType 通过字段 tag 解析数据库字段
func (d *Dao) GetColumnsByType(typ reflect.Type, omitColumns ...string) []string {
	return sqlbuilder.GetColumnsByType(d.getMapper(), typ, omitColumns...)
}

// DBColumns 获取当前表数据库字段
// omitColumns 表示需要忽略的字段
func (d *Dao) DBColumns(omitColumns ...string) []string {
	columns := make([]string, 0)
	for _, column := range d.TableMeta.Columns {
		if utils.ContainsString(omitColumns, column) {
			continue
		}
		columns = append(columns, column)
	}
	return columns
}

// TableName 获取当前表名
func (d *Dao) TableName() string {
	return d.TableMeta.TableName
}

// Save 插入数据
// omitColumns 不需要 insert 的字段
func (d *Dao) Save(dest Model, opts ...InsertOption) (int64, error) {
	return d.SaveContext(context.Background(), dest, opts...)
}

// SaveContext 插入数据，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao) SaveContext(ctx context.Context, dest Model, opts ...InsertOption) (int64, error) {
	return d.saveContext(ctx, nil, dest, opts...)
}

// SaveTx 插入数据，支持事务
// omitColumns 不需要 insert 的字段
func (d *Dao) SaveTx(tx *sqlx.Tx, dest Model, opts ...InsertOption) (int64, error) {
	return d.SaveTxContext(context.Background(), tx, dest, opts...)
}

// SaveTxContext 插入数据，支持事务，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao) SaveTxContext(ctx context.Context, tx *sqlx.Tx, dest Model, opts ...InsertOption) (int64, error) {
	if err := d.checkTxNil(tx); err != nil {
		return 0, err
	}
	return d.saveContext(ctx, tx, dest, opts...)
}

func (d *Dao) saveContext(ctx context.Context, tx *sqlx.Tx, model Model, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}

	var execer engine.Execer = d.GetMasterDB()
	if tx != nil {
		execer = tx
	}
	id, _, err := d.SQLBuilder().Insert().Execer(execer).
		Columns(d.getSaveColumns(opt)...).
		NamedExecContext(ctx, model)
	return id, err
}

// ReplaceInto replace into table
// omitColumns 不需要 insert 的字段
func (d *Dao) ReplaceInto(dest Model, opts ...InsertOption) (int64, error) {
	return d.ReplaceIntoContext(context.Background(), dest, opts...)
}

// ReplaceIntoContext replace into table，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao) ReplaceIntoContext(ctx context.Context, model Model, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	id, _, err := d.Inserter().
		Columns(d.getSaveColumns(opt)...).
		IsReplaceInto(true).
		NamedExecContext(ctx, model)
	return id, err
}

// IgnoreInto 使用 INSERT IGNORE INTO 如果记录已存在则忽略
// omitColumns 不需要 insert 的字段
func (d *Dao) IgnoreInto(model Model, opts ...InsertOption) (int64, error) {
	return d.IgnoreIntoContext(context.Background(), model, opts...)
}

// IgnoreIntoContext 使用 INSERT IGNORE INTO 如果记录已存在则忽略，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao) IgnoreIntoContext(ctx context.Context, model Model, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	id, _, err := d.Inserter().
		Columns(d.getSaveColumns(opt)...).
		IsIgnoreInto(true).
		NamedExecContext(ctx, model)
	return id, err
}

// BatchSave 批量新增，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao) BatchSave(models any, opts ...InsertOption) (int64, error) {
	return d.BatchSaveContext(context.Background(), models, opts...)
}

// BatchSaveContext 批量新增
// omitColumns 不需要 insert 的字段
// models 是一个批量 insert 的 slice
func (d *Dao) BatchSaveContext(ctx context.Context, models any, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	_, affected, err := d.Inserter().
		Columns(d.getSaveColumns(opt)...).
		NamedExecContext(ctx, models)
	return affected, err
}

// BatchReplaceInto 批量新增，使用 replace into 方式
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (d *Dao) BatchReplaceInto(models any, opts ...InsertOption) (int64, error) {
	return d.BatchReplaceIntoContext(context.Background(), models, opts...)
}

// BatchReplaceIntoContext 批量新增，使用 replace into 方式，携带上下文
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (d *Dao) BatchReplaceIntoContext(ctx context.Context, models any, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	_, affected, err := d.Inserter().
		Columns(d.getSaveColumns(opt)...).
		IsReplaceInto(true).
		NamedExecContext(ctx, models)
	return affected, err
}

func (d *Dao) getSaveColumns(opt *InsertOptions) []string {
	meta := d.TableMeta
	var omits []string
	if meta.IsAutoIncrement {
		omits = append(omits, meta.PrimaryKey)
	}
	if len(opt.omitColumns) > 0 {
		omits = append(omits, opt.omitColumns...)
	}
	if !opt.disableGlobalOmitColumns && len(global.saveOmitColumns) > 0 {
		omits = append(omits, global.saveOmitColumns...)
	}
	return meta.OmitColumns(omits...)
}

// GetByColumn 按指定字段查询单条数据
// bool 数据是否存在
func (d *Dao) GetByColumn(kv *KV, dest Model) (bool, error) {
	return d.GetByColumnContext(context.Background(), kv, dest)
}

// GetByColumnContext 按指定字段查询单条数据，携带上下文
// bool 数据是否存在
func (d *Dao) GetByColumnContext(ctx context.Context, kv *KV, dest Model) (bool, error) {
	return d.getByColumnContext(ctx, nil, kv, dest)
}

// GetByColumnTx 按指定字段查询单条数据，支持事务
// bool 数据是否存在
func (d *Dao) GetByColumnTx(tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	return d.GetByColumnTxContext(context.Background(), tx, kv, dest)
}

// GetByColumnTxContext 按指定字段查询单条数据，支持事务，携带上下文
// bool 数据是否存在
func (d *Dao) GetByColumnTxContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	if err := d.checkTxNil(tx); err != nil {
		return false, err
	}
	return d.getByColumnContext(ctx, tx, kv, dest)
}

func (d *Dao) getByColumnContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest Model) (bool, error) {
	if kv == nil {
		return false, nil
	}

	var queryer engine.Queryer = d.GetReadDB()
	if tx != nil {
		queryer = tx
	}

	return d.Selector().Queryer(queryer).
		Where(ql.C(ql.Col(kv.Key).EQ(kv.Value))).
		GetContext(ctx, dest)

}

// ListByColumns 指定字段多个值查询多条数据
// dest: slice pointer
func (d *Dao) ListByColumns(kvs *MultiKV, dest any) error {
	return d.ListByColumnsContext(context.Background(), kvs, dest)
}

// ListByColumnsContext 指定字段多个值查询多条数据，携带上下文
// dest: slice pointer
func (d *Dao) ListByColumnsContext(ctx context.Context, kvs *MultiKV, dest any) error {
	return d.listByColumnsContext(ctx, nil, kvs, dest)
}

// ListByColumnsTx 指定字段多个值查询多条数据，支持事务
func (d *Dao) ListByColumnsTx(tx *sqlx.Tx, kvs *MultiKV, dest any) error {
	return d.ListByColumnsTxContext(context.Background(), tx, kvs, dest)
}

// ListByColumnsTxContext 指定字段多个值查询多条数据，支持事务，携带上下文
func (d *Dao) ListByColumnsTxContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV, dest any) error {
	if err := d.checkTxNil(tx); err != nil {
		return err
	}
	return d.listByColumnsContext(ctx, tx, kvs, dest)
}

func (d *Dao) listByColumnsContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV, dest any) error {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}

	var queryer engine.Queryer = d.GetReadDB()
	if tx != nil {
		queryer = tx
	}

	return d.Selector().Queryer(queryer).
		Columns(d.DBColumns()...).
		Where(ql.C(ql.Col(kvs.Key).In(kvs.Values...))).
		SelectContext(ctx, dest)
}

// List 指定字段查询多条数据
func (d *Dao) List(kv *KV, dest any) error {
	return d.ListContext(context.Background(), kv, dest)
}

// ListContext 指定字段查询多条数据，携带上下文
func (d *Dao) ListContext(ctx context.Context, kv *KV, dest any) error {
	return d.listContext(ctx, nil, kv, dest)
}

// ListTx 指定字段查询多条数据，支持事务
func (d *Dao) ListTx(tx *sqlx.Tx, kv *KV, dest any) error {
	return d.ListTxContext(context.Background(), tx, kv, dest)
}

// ListTxContext 指定字段查询多条数据，支持事务，携带上下文
func (d *Dao) ListTxContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest any) error {
	if err := d.checkTxNil(tx); err != nil {
		return err
	}
	return d.listContext(ctx, tx, kv, dest)
}

func (d *Dao) listContext(ctx context.Context, tx *sqlx.Tx, kv *KV, dest any) error {
	if kv == nil {
		return nil
	}

	var queryer engine.Queryer = d.GetReadDB()
	if tx != nil {
		queryer = tx
	}

	return d.Selector().Queryer(queryer).
		Columns(d.DBColumns()...).
		Where(ql.C(ql.Col(kv.Key).EQ(kv.Value))).
		SelectContext(ctx, dest)
}

// GetByID 根据 id 查询单条数据
func (d *Dao) GetByID(id any, dest Model) (bool, error) {
	return d.GetByIDContext(context.Background(), id, dest)
}

// GetByIDContext 根据 id 查询单条数据，携带上下文
func (d *Dao) GetByIDContext(ctx context.Context, id any, dest Model) (bool, error) {
	tableMeta := d.TableMeta
	return d.GetByColumnContext(ctx, OfKv(tableMeta.PrimaryKey, id), dest)
}

// ListByIDs 根据 id 查询多条数据
func (d *Dao) ListByIDs(dest any, ids ...any) error {
	return d.ListByIDsContext(context.Background(), dest, ids...)
}

// ListByIDsContext 根据 id 查询多条数据，携带上下文
func (d *Dao) ListByIDsContext(ctx context.Context, dest any, ids ...any) error {
	tableMeta := d.TableMeta
	return d.ListByColumnsContext(ctx, OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// ListByIDsTx 查询多个id值，支持事务
func (d *Dao) ListByIDsTx(tx *sqlx.Tx, dest any, ids ...any) error {
	return d.ListByIDsTxContext(context.Background(), tx, dest, ids...)
}

// ListByIDsTxContext 查询多个id值，支持事务，携带上下文
func (d *Dao) ListByIDsTxContext(ctx context.Context, tx *sqlx.Tx, dest any, ids ...any) error {
	tableMeta := d.TableMeta
	return d.ListByColumnsTxContext(ctx, tx, OfMultiKv(tableMeta.PrimaryKey, ids...), dest)
}

// UpdateField 部分字段更新
func (d *Dao) UpdateField(idValue any, fieldMap map[string]any) (bool, error) {
	return d.UpdateFieldContext(context.Background(), idValue, fieldMap)
}

// UpdateFieldContext 部分字段更新，携带上下文
func (d *Dao) UpdateFieldContext(ctx context.Context, idValue any, fieldMap map[string]any) (bool, error) {
	return d.updateFieldContext(ctx, nil, idValue, fieldMap)
}

// UpdateFieldTx 部分字段更新，支持事务
func (d *Dao) UpdateFieldTx(tx *sqlx.Tx, idValue any, fieldMap map[string]any) (bool, error) {
	return d.UpdateFieldTxContext(context.Background(), tx, idValue, fieldMap)
}

// UpdateFieldTxContext 部分字段更新，支持事务，携带上下文
func (d *Dao) UpdateFieldTxContext(ctx context.Context, tx *sqlx.Tx, idValue any, fieldMap map[string]any) (bool, error) {
	if err := d.checkTxNil(tx); err != nil {
		return false, err
	}
	return d.updateFieldContext(ctx, tx, idValue, fieldMap)
}

func (d *Dao) updateFieldContext(ctx context.Context, tx *sqlx.Tx, idValue any, fieldMap map[string]any) (bool, error) {
	if utils.IsIDEmpty(idValue) {
		return false, ErrUpdatePrimaryKeyRequire
	}

	tableMeta := d.TableMeta
	rows, err := d.updateByCondContext(ctx, tx, fieldMap, ql.C(ql.Col(tableMeta.PrimaryKey).EQ(idValue)))
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (d *Dao) updateByCondContext(ctx context.Context, tx *sqlx.Tx, attr map[string]any, where sqlbuilder.ConditionBuilder) (int64, error) {
	var execer engine.Execer = d.GetMasterDB()
	if tx != nil {
		execer = tx
	}

	updater := d.Updater().Execer(execer)
	for col, val := range attr {
		updater.Set(col, val)
	}
	updater.Where(where)
	return updater.ExecContext(ctx)
}

// Update 全字段更新
func (d *Dao) Update(m Model, omitColumns ...string) (bool, error) {
	return d.UpdateContext(context.Background(), m, omitColumns...)
}

// UpdateContext 全字段更新，携带上下文
func (d *Dao) UpdateContext(ctx context.Context, m Model, omitColumns ...string) (bool, error) {
	return d.updateContext(ctx, nil, m, omitColumns...)
}

// UpdateTx 全字段更新，支持事务
func (d *Dao) UpdateTx(tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
	return d.UpdateTxContext(context.Background(), tx, m, omitColumns...)
}

// UpdateTxContext 全字段更新，支持事务，携带上下文
func (d *Dao) UpdateTxContext(ctx context.Context, tx *sqlx.Tx, m Model, omitColumns ...string) (bool, error) {
	if err := d.checkTxNil(tx); err != nil {
		return false, err
	}
	return d.updateContext(ctx, tx, m, omitColumns...)
}

func (d *Dao) updateContext(ctx context.Context, tx *sqlx.Tx, model Model, omitColumns ...string) (bool, error) {
	if utils.IsIDEmpty(model.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}

	var execer engine.Execer = d.GetMasterDB()
	if tx != nil {
		execer = tx
	}

	tableMeta := d.TableMeta
	omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	updater := d.Updater().Execer(execer).
		Columns(d.DBColumns(omitColumns...)...).
		Where(
			ql.SC().And(fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey)),
		)
	affected, err := updater.NamedExecContext(ctx, model)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (d *Dao) deleteByCondContext(ctx context.Context, tx *sqlx.Tx, where sqlbuilder.ConditionBuilder) (int64, error) {
	var execer engine.Execer = d.GetMasterDB()
	if tx != nil {
		execer = tx
	}

	return d.Deleter().Execer(execer).Where(where).ExecContext(ctx)
}

// DeleteByColumn 按字段名删除
func (d *Dao) DeleteByColumn(kv *KV) (int64, error) {
	return d.DeleteByColumnContext(context.Background(), kv)
}

// DeleteByColumnContext 按字段名删除，携带上下文
func (d *Dao) DeleteByColumnContext(ctx context.Context, kv *KV) (int64, error) {
	return d.deleteByColumnContext(ctx, nil, kv)
}

// DeleteByColumnTx 按字段名删除，支持事务
func (d *Dao) DeleteByColumnTx(tx *sqlx.Tx, kv *KV) (int64, error) {
	return d.DeleteByColumnTxContext(context.Background(), tx, kv)
}

// DeleteByColumnTxContext 按字段名删除，支持事务，携带上下文
func (d *Dao) DeleteByColumnTxContext(ctx context.Context, tx *sqlx.Tx, kv *KV) (int64, error) {
	if err := d.checkTxNil(tx); err != nil {
		return 0, err
	}
	return d.deleteByColumnContext(ctx, tx, kv)
}

func (d *Dao) deleteByColumnContext(ctx context.Context, tx *sqlx.Tx, kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
	return d.deleteByCondContext(ctx, tx, ql.C(ql.Col(kv.Key).EQ(kv.Value)))
}

// DeleteByColumns 指定字段删除多个值
func (d *Dao) DeleteByColumns(kvs *MultiKV) (int64, error) {
	return d.DeleteByColumnsContext(context.Background(), kvs)
}

// DeleteByColumnsContext 指定字段删除多个值，携带上下文
func (d *Dao) DeleteByColumnsContext(ctx context.Context, kvs *MultiKV) (int64, error) {
	return d.deleteByColumnsContext(ctx, nil, kvs)
}

// DeleteByColumnsTx 指定字段多个值删除
func (d *Dao) DeleteByColumnsTx(tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	return d.DeleteByColumnsTxContext(context.Background(), tx, kvs)
}

// DeleteByColumnsTxContext 指定字段多个值删除，携带上下文
func (d *Dao) DeleteByColumnsTxContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if err := d.checkTxNil(tx); err != nil {
		return 0, err
	}
	return d.deleteByColumnsContext(ctx, tx, kvs)
}

func (d *Dao) deleteByColumnsContext(ctx context.Context, tx *sqlx.Tx, kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}
	return d.deleteByCondContext(ctx, tx, ql.C(ql.Col(kvs.Key).In(kvs.Values...)))
}

// DeleteByID 根据id删除数据
func (d *Dao) DeleteByID(id any) (bool, error) {
	return d.DeleteByIDContext(context.Background(), id)
}

// DeleteByIDContext 根据id删除数据，携带上下文
func (d *Dao) DeleteByIDContext(ctx context.Context, id any) (bool, error) {
	return d.deleteByIDContext(ctx, nil, id)
}

// DeleteByIDTx 根据id删除数据，支持事务
func (d *Dao) DeleteByIDTx(tx *sqlx.Tx, id any) (bool, error) {
	return d.DeleteByIDTxContext(context.Background(), tx, id)
}

// DeleteByIDTxContext 根据id删除数据，支持事务，携带上下文
func (d *Dao) DeleteByIDTxContext(ctx context.Context, tx *sqlx.Tx, id any) (bool, error) {
	if err := d.checkTxNil(tx); err != nil {
		return false, err
	}
	return d.deleteByIDContext(ctx, tx, id)
}

func (d *Dao) deleteByIDContext(ctx context.Context, tx *sqlx.Tx, id any) (bool, error) {
	tableMeta := d.TableMeta
	affected, err := d.deleteByColumnContext(ctx, tx, OfKv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (d *Dao) checkTxNil(tx *sqlx.Tx) error {
	if tx == nil {
		return ErrTxNil
	}
	return nil
}

// With 使用新的数据库连接创建 Dao
func (d *Dao) With(master, read *sqlx.DB, opts ...Option) *Dao {
	newDao := &Dao{
		masterDB:   master,
		readDB:     read,
		TableMeta:  d.TableMeta,
		Mapper:     d.Mapper,
		ifNullVals: d.ifNullVals,
	}
	for _, opt := range opts {
		opt(newDao)
	}
	return newDao
}

// WithTableName 使用新的数据库连接创建 Dao
func (d *Dao) WithTableName(tableName string) *Dao {
	newDao := &Dao{
		masterDB:   d.masterDB,
		readDB:     d.readDB,
		TableMeta:  d.TableMeta.WithTableName(tableName),
		Mapper:     d.Mapper,
		ifNullVals: d.ifNullVals,
	}
	return newDao
}

func (d *Dao) getMapper() *reflectx.Mapper {
	if d.Mapper != nil {
		return d.Mapper
	}
	return d.GetMasterDB().Mapper
}

func (d *Dao) GetMasterDB() *sqlx.DB {
	if d.masterDB != nil {
		return d.masterDB
	}
	return global.defaultMasterDB
}

func (d *Dao) GetReadDB() *sqlx.DB {
	if d.readDB != nil {
		return d.readDB
	}
	if d.masterDB != nil {
		return d.masterDB
	}
	if global.defaultReadDB != nil {
		return global.defaultReadDB
	}
	return global.defaultMasterDB
}

func (d *Dao) initIfNullVal() {
	if d.ifNullVals == nil {
		d.ifNullVals = make(map[string]string)
	}
}
