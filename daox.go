package daox

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

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
	lock        sync.Mutex
	options     *Options
	masterDB    *DB
	readDB      *DB
	mapper      *reflectx.Mapper
	TableMeta   *TableMeta
	ifNullVals  map[string]string
	omitColumns []string
	executor    engine.Executor
}

// NewDao 创建一个新的 dao 对象
// tableName 参数表示表名
// primaryKey 参数表示主键
// structType 参数表示数据结构类型
// opts 参数表示可选的选项
// 返回值为创建的Dao对象指针
func NewDao[T Model](tableName string, primaryKey string, opts ...Option) *Dao {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if options.mapper == nil {
		if options.master != nil {
			options.mapper = options.master.Mapper
		} else {
			options.mapper = sqlbuilder.GetMapperByTagName("json")
		}
	}
	options.tableName = tableName
	structType := reflect.TypeFor[T]()
	columns := sqlbuilder.GetColumnsByType(options.mapper, structType, options.omitColumns...)
	meta := &TableMeta{
		TableName:       tableName,
		PrimaryKey:      primaryKey,
		IsAutoIncrement: options.autoIncrement,
		Columns:         columns,
	}
	hooks := mergeHooks(options)
	master := options.master
	if options.master == nil {
		master = global.defaultMasterDB
	}
	read := options.read
	if read == nil {
		if global.defaultReadDB != nil {
			read = global.defaultMasterDB
		} else if options.master != nil {
			read = master
		}
	}
	dao := &Dao{
		masterDB:    NewDb(master, hooks...),
		readDB:      NewDb(read, hooks...),
		mapper:      options.mapper,
		TableMeta:   meta,
		ifNullVals:  options.ifNullVals,
		omitColumns: options.omitColumns,
		options:     options,
	}
	global.registerMeta(dao.TableMeta)
	return dao
}

// NewDaoByMeta 根据 meta 接口创建 dao 对象
func NewDaoByMeta(m Meta, opts ...Option) *Dao {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if options.mapper == nil {
		if options.master != nil {
			options.mapper = options.master.Mapper
		} else {
			options.mapper = sqlbuilder.GetMapperByTagName("json")
		}
	}
	meta := &TableMeta{
		TableName:       m.TableName(),
		PrimaryKey:      m.PrimaryKey(),
		Columns:         m.Columns(),
		IsAutoIncrement: m.IsAutoIncrement(),
	}
	hooks := mergeHooks(options)
	master := options.master
	if options.master == nil {
		master = global.defaultMasterDB
	}
	read := options.read
	if read == nil {
		if global.defaultReadDB != nil {
			read = global.defaultMasterDB
		} else if options.master != nil {
			read = master
		}
	}
	dao := &Dao{
		masterDB:    NewDb(master, hooks...),
		readDB:      NewDb(read, hooks...),
		mapper:      options.mapper,
		TableMeta:   meta,
		ifNullVals:  options.ifNullVals,
		omitColumns: options.omitColumns,
		options:     options,
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
	selector.Queryer(d.getQueryer())
	return selector
}

// Updater 创建当前表的 Updater
func (d *Dao) Updater() *sqlbuilder.Updater {
	return d.SQLBuilder().Update().Execer(d.getExecer())
}

// Deleter 创建当前表的 Deleter
func (d *Dao) Deleter() *sqlbuilder.Deleter {
	return d.SQLBuilder().Delete().Execer(d.getExecer())
}

// Inserter 创建当前表的 updater
func (d *Dao) Inserter() *sqlbuilder.Inserter {
	return d.SQLBuilder().Insert().Execer(d.getExecer())
}

// GetColumnsByModel 根据 model 结构获取数据库字段
// omitColumns 表示需要忽略的字段
func (d *Dao) GetColumnsByModel(model any, omitColumns ...string) []string {
	return d.GetColumnsByType(reflect.TypeOf(model), omitColumns...)
}

// GetColumnsByType 通过字段 tag 解析数据库字段
func (d *Dao) GetColumnsByType(typ reflect.Type, omitColumns ...string) []string {
	return sqlbuilder.GetColumnsByType(d.mapper, typ, omitColumns...)
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
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	id, _, err := d.SQLBuilder().Insert().Execer(d.getExecer()).
		Columns(d.getSaveColumns(opt)...).
		NamedExecContext(ctx, dest)
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
	if !opt.disableGlobalOmitColumns && len(global.omitColumns) > 0 {
		omits = append(omits, global.omitColumns...)
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
	if kv == nil {
		return false, nil
	}
	return d.Selector().Queryer(d.getQueryer()).
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
	if kvs == nil || len(kvs.Values) == 0 {
		return nil
	}
	return d.Selector().Queryer(d.getQueryer()).
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
	return d.Selector().Queryer(d.getQueryer()).
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

// UpdateField 部分字段更新
func (d *Dao) UpdateField(idValue any, fieldMap map[string]any) (bool, error) {
	return d.UpdateFieldContext(context.Background(), idValue, fieldMap)
}

// UpdateFieldContext 部分字段更新，携带上下文
func (d *Dao) UpdateFieldContext(ctx context.Context, idValue any, fieldMap map[string]any) (bool, error) {
	if utils.IsIDEmpty(idValue) {
		return false, ErrUpdatePrimaryKeyRequire
	}

	updater := d.Updater().Execer(d.getExecer())
	for col, val := range fieldMap {
		updater.Set(col, val)
	}
	updater.Where(ql.C(ql.Col(d.TableMeta.PrimaryKey).EQ(idValue)))
	rows, err := updater.ExecContext(ctx)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// Update 全字段更新
func (d *Dao) Update(m Model, omitColumns ...string) (bool, error) {
	return d.UpdateContext(context.Background(), m, omitColumns...)
}

// UpdateContext 全字段更新，携带上下文
func (d *Dao) UpdateContext(ctx context.Context, model Model, omitColumns ...string) (bool, error) {
	if utils.IsIDEmpty(model.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := d.TableMeta
	return d.UpdateByCondContext(ctx, model, ql.SC().And(fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey)), tableMeta.PrimaryKey)
}

// UpdateByCond 按条件更新全部字段
func (d *Dao) UpdateByCond(model Model, where sqlbuilder.ConditionBuilder, omitColumns ...string) (bool, error) {
	return d.UpdateByCondContext(context.Background(), model, where, omitColumns...)
}

// UpdateByCondContext 按条件更新全部字段
func (d *Dao) UpdateByCondContext(ctx context.Context, model Model, where sqlbuilder.ConditionBuilder, omitColumns ...string) (bool, error) {
	if len(global.omitColumns) > 0 {
		omitColumns = append(omitColumns, global.omitColumns...)
	}
	updater := d.Updater().Execer(d.getExecer()).
		Columns(d.DBColumns(omitColumns...)...).
		Where(where)
	affected, err := updater.NamedExecContext(ctx, model)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (d *Dao) deleteByCondContext(ctx context.Context, where sqlbuilder.ConditionBuilder) (int64, error) {
	return d.Deleter().Execer(d.getExecer()).Where(where).ExecContext(ctx)
}

// DeleteByColumn 按字段名删除
func (d *Dao) DeleteByColumn(kv *KV) (int64, error) {
	return d.DeleteByColumnContext(context.Background(), kv)
}

// DeleteByColumnContext 按字段名删除，携带上下文
func (d *Dao) DeleteByColumnContext(ctx context.Context, kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
	return d.deleteByCondContext(ctx, ql.C(ql.Col(kv.Key).EQ(kv.Value)))
}

// DeleteByColumns 指定字段删除多个值
func (d *Dao) DeleteByColumns(kvs *MultiKV) (int64, error) {
	return d.DeleteByColumnsContext(context.Background(), kvs)
}

// DeleteByColumnsContext 指定字段删除多个值，携带上下文
func (d *Dao) DeleteByColumnsContext(ctx context.Context, kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}
	return d.deleteByCondContext(ctx, ql.C(ql.Col(kvs.Key).In(kvs.Values...)))
}

// DeleteByID 根据id删除数据
func (d *Dao) DeleteByID(id any) (bool, error) {
	return d.DeleteByIDContext(context.Background(), id)
}

// DeleteByIDContext 根据id删除数据，携带上下文
func (d *Dao) DeleteByIDContext(ctx context.Context, id any) (bool, error) {
	tableMeta := d.TableMeta
	affected, err := d.DeleteByColumnContext(ctx, OfKv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// With 使用新的数据库连接创建 Dao
func (d *Dao) With(master, read *sqlx.DB) *Dao {
	hooks := mergeHooks(d.options)
	newDao := &Dao{
		masterDB:   NewDb(master, hooks...),
		readDB:     NewDb(read, hooks...),
		TableMeta:  d.TableMeta,
		mapper:     d.mapper,
		ifNullVals: d.ifNullVals,
		options:    d.options,
	}
	return newDao
}

// WithTableName 使用新的数据库连接创建 Dao
func (d *Dao) WithTableName(tableName string) *Dao {
	newDao := &Dao{
		masterDB:   d.masterDB,
		readDB:     d.readDB,
		TableMeta:  d.TableMeta.WithTableName(tableName),
		mapper:     d.mapper,
		ifNullVals: d.ifNullVals,
		options:    d.options,
	}
	return newDao
}

func (d *Dao) WithExecutor(executor engine.Executor) *Dao {
	newDao := &Dao{
		masterDB:   d.masterDB,
		readDB:     d.readDB,
		TableMeta:  d.TableMeta,
		mapper:     d.mapper,
		ifNullVals: d.ifNullVals,
		options:    d.options,
		executor:   executor,
	}
	return newDao
}

func (d *Dao) initIfNullVal() {
	if d.ifNullVals == nil {
		d.ifNullVals = make(map[string]string)
	}
}

// GetMasterDB 返回主库
func (d *Dao) GetMasterDB() *DB {
	if d.masterDB != nil {
		return d.masterDB
	}
	if global.defaultMasterDB == nil {
		return nil
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	// double check
	if d.masterDB != nil {
		return d.masterDB
	}
	hooks := mergeHooks(d.options)
	d.masterDB = NewDb(global.defaultMasterDB, hooks...)
	return d.masterDB
}

// GetReadDB 返回从库
func (d *Dao) GetReadDB() *DB {
	if d.readDB != nil {
		return d.readDB
	}
	if global.defaultReadDB == nil && d.GetMasterDB() == nil {
		return nil
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	// double check
	if d.readDB != nil {
		return d.readDB
	}
	hooks := mergeHooks(d.options)
	if global.defaultReadDB != nil {
		d.readDB = NewDb(global.defaultReadDB, hooks...)
	} else if d.masterDB != nil {
		d.readDB = d.masterDB
	} else if global.defaultMasterDB == nil {
		d.readDB = NewDb(global.defaultMasterDB, hooks...)
	}
	return d.masterDB
}

func (d *Dao) getQueryer() engine.Queryer {
	if d.executor != nil {
		return d.executor
	}
	return d.GetReadDB()
}

func (d *Dao) getExecer() engine.Execer {
	if d.executor != nil {
		return d.executor
	}
	return d.GetMasterDB()
}

func mergeHooks(options *Options) []engine.Hook {
	hooks := global.hooks
	if len(options.hooks) > 0 {
		hooks = append(hooks, options.hooks...)
	}
	if options.printSQL != nil {
		hooks = append(hooks, NewLogHook(options.printSQL))
	} else if global.printSQL != nil {
		hooks = append(hooks, NewLogHook(global.printSQL))
	}
	return hooks
}
