package daox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx/reflectx"

	"github.com/fengjx/daox/engine"
	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
	"github.com/fengjx/daox/utils"
)

var (
	// ErrUpdatePrimaryKeyRequire 更新操作必须提供主键值
	ErrUpdatePrimaryKeyRequire = errors.New("[daox] Primary key require for update")
	// ErrTxNil 事务对象为空
	ErrTxNil = errors.New("[daox] Tx is nil")
)

// Dao 数据访问对象，封装了数据库操作的基础方法
type Dao[T Model] struct {
	options     *Options          // 配置选项
	masterDB    engine.Executor   // 主库连接
	readDB      engine.Queryer    // 从库连接
	mapper      *reflectx.Mapper  // 字段映射器
	modelType   reflect.Type      // 模型类型
	TableMeta   *TableMeta        // 表元数据
	ifNullVals  map[string]string // NULL值替换配置
	omitColumns []string          // 忽略的字段列表
	executor    engine.Executor   // SQL执行器，用于事务等场景
}

// NewDao 根据 meta 接口创建 dao 对象
// m: 表元数据接口
// opts: 可选配置项
// 返回值: 创建的Dao对象指针
func NewDao[T Model](m Meta, opts ...Option) *Dao[T] {
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
	// 检查泛型必须是指针
	modelType, err := baseType(reflect.TypeOf(new(T)), reflect.Ptr)
	if err != nil {
		panic(err)
	}
	dao := &Dao[T]{
		masterDB:    NewDb(master, hooks...),
		readDB:      NewDb(read, hooks...),
		mapper:      options.mapper,
		TableMeta:   meta,
		ifNullVals:  options.ifNullVals,
		omitColumns: options.omitColumns,
		options:     options,
		modelType:   modelType,
	}
	global.registerMeta(dao.TableMeta)
	return dao
}

// SQLBuilder 创建当前表的 SQL 构建器
// 返回值: SQL构建器对象
func (d *Dao[T]) SQLBuilder() *sqlbuilder.Builder {
	return sqlbuilder.New(d.TableMeta.TableName)
}

// Selector 创建当前表的查询构建器
// columns: 查询的字段列表，为空则查询全部字段
// 返回值: 查询构建器对象
func (d *Dao[T]) Selector(columns ...string) *sqlbuilder.Selector {
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

// Updater 创建当前表的更新构建器
// 返回值: 更新构建器对象
func (d *Dao[T]) Updater() *sqlbuilder.Updater {
	return d.SQLBuilder().Update().Execer(d.getExecer())
}

// Deleter 创建当前表的删除构建器
// 返回值: 删除构建器对象
func (d *Dao[T]) Deleter() *sqlbuilder.Deleter {
	return d.SQLBuilder().Delete().Execer(d.getExecer())
}

// Inserter 创建当前表的插入构建器
// opts: 插入选项，如忽略字段等
// 返回值: 插入构建器对象
func (d *Dao[T]) Inserter(opts ...InsertOption) *sqlbuilder.Inserter {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	return d.SQLBuilder().Insert(d.getSaveColumns(opt)...).Execer(d.getExecer())
}

// GetColumnsByModel 根据 model 结构获取数据库字段
// model: 模型结构体
// omitColumns: 需要忽略的字段列表
// 返回值: 字段名列表
func (d *Dao[T]) GetColumnsByModel(model any, omitColumns ...string) []string {
	return d.GetColumnsByType(reflect.TypeOf(model), omitColumns...)
}

// GetColumnsByType 通过字段 tag 解析数据库字段
// typ: 结构体类型
// omitColumns: 需要忽略的字段列表
// 返回值: 字段名列表
func (d *Dao[T]) GetColumnsByType(typ reflect.Type, omitColumns ...string) []string {
	return sqlbuilder.GetColumnsByType(d.mapper, typ, omitColumns...)
}

// DBColumns 获取当前表数据库字段
// omitColumns: 需要忽略的字段列表
// 返回值: 字段名列表
func (d *Dao[T]) DBColumns(omitColumns ...string) []string {
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
// 返回值: 表名
func (d *Dao[T]) TableName() string {
	return d.TableMeta.TableName
}

// Save 插入数据
// dest: 要插入的数据对象
// opts: 插入选项
// 返回值: 插入ID，错误信息
func (d *Dao[T]) Save(dest T, opts ...InsertOption) (int64, error) {
	return d.SaveContext(context.Background(), dest, opts...)
}

// SaveContext 插入数据，携带上下文
// ctx: 上下文
// dest: 要插入的数据对象
// opts: 插入选项
// 返回值: 插入ID，错误信息
func (d *Dao[T]) SaveContext(ctx context.Context, dest T, opts ...InsertOption) (int64, error) {
	opt := &InsertOptions{}
	for _, o := range opts {
		o(opt)
	}
	result, err := d.SQLBuilder().Insert().Execer(d.getExecer()).
		Columns(d.getSaveColumns(opt)...).
		NamedExecContext(ctx, dest)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ReplaceInto replace into table
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) ReplaceInto(dest T, opts ...InsertOption) (sql.Result, error) {
	return d.ReplaceIntoContext(context.Background(), dest, opts...)
}

// ReplaceIntoContext replace into table，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) ReplaceIntoContext(ctx context.Context, model T, opts ...InsertOption) (sql.Result, error) {
	return d.Inserter(opts...).
		IsReplaceInto(true).
		NamedExecContext(ctx, model)
}

// IgnoreInto 使用 INSERT IGNORE INTO 如果记录已存在则忽略
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) IgnoreInto(model T, opts ...InsertOption) (sql.Result, error) {
	return d.IgnoreIntoContext(context.Background(), model, opts...)
}

// IgnoreIntoContext 使用 INSERT IGNORE INTO 如果记录已存在则忽略，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) IgnoreIntoContext(ctx context.Context, model T, opts ...InsertOption) (sql.Result, error) {
	return d.Inserter(opts...).
		IsIgnoreInto(true).
		NamedExecContext(ctx, model)
}

// BatchSave 批量新增，携带上下文
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) BatchSave(models []T, opts ...InsertOption) (sql.Result, error) {
	return d.BatchSaveContext(context.Background(), models, opts...)
}

// BatchSaveContext 批量新增
// omitColumns 不需要 insert 的字段
// models 是一个批量 insert 的 slice
func (d *Dao[T]) BatchSaveContext(ctx context.Context, models []T, opts ...InsertOption) (sql.Result, error) {
	return d.Inserter(opts...).
		NamedExecContext(ctx, models)
}

// BatchReplaceInto 批量新增，使用 replace into 方式
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) BatchReplaceInto(models []T, opts ...InsertOption) (sql.Result, error) {
	return d.BatchReplaceIntoContext(context.Background(), models, opts...)
}

// BatchReplaceIntoContext 批量新增，使用 replace into 方式，携带上下文
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) BatchReplaceIntoContext(ctx context.Context, models []T, opts ...InsertOption) (sql.Result, error) {
	return d.Inserter(opts...).
		IsReplaceInto(true).
		NamedExecContext(ctx, models)
}

func (d *Dao[T]) getSaveColumns(opt *InsertOptions) []string {
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
func (d *Dao[T]) GetByColumn(kv *KV) (T, error) {
	return d.GetByColumnContext(context.Background(), kv)
}

// GetByColumnContext 按指定字段查询单条数据，携带上下文
// bool 数据是否存在
func (d *Dao[T]) GetByColumnContext(ctx context.Context, kv *KV) (T, error) {
	if kv == nil {
		return d.emptyModel(), nil
	}
	dest := d.newModel()
	exist, err := d.Selector().Queryer(d.getQueryer()).
		Where(ql.C(ql.Col(kv.Key).EQ(kv.Value))).
		OneContext(ctx, dest)
	if err != nil {
		return d.emptyModel(), err
	}
	if !exist {
		return d.emptyModel(), nil
	}
	return dest, nil
}

// ListByColumns 指定字段多个值查询多条数据
// dest: slice pointer
func (d *Dao[T]) ListByColumns(kvs *MultiKV) ([]T, error) {
	return d.ListByColumnsContext(context.Background(), kvs)
}

// ListByColumnsContext 指定字段多个值查询多条数据，携带上下文
// dest: slice pointer
func (d *Dao[T]) ListByColumnsContext(ctx context.Context, kvs *MultiKV) ([]T, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return nil, nil
	}
	var dest []T
	err := d.Selector().Queryer(d.getQueryer()).
		Columns(d.DBColumns()...).
		Where(ql.C(ql.Col(kvs.Key).In(kvs.Values...))).
		ListContext(ctx, &dest)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

// List 指定字段查询多条数据
func (d *Dao[T]) List(kv *KV) ([]T, error) {
	return d.ListContext(context.Background(), kv)
}

// ListContext 指定字段查询多条数据，携带上下文
func (d *Dao[T]) ListContext(ctx context.Context, kv *KV) ([]T, error) {
	var dest []T
	err := d.Selector().Queryer(d.getQueryer()).
		Columns(d.DBColumns()...).
		Where(ql.C(ql.Col(kv.Key).EQ(kv.Value))).
		ListContext(ctx, dest)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

// GetByID 根据 id 查询单条数据
func (d *Dao[T]) GetByID(id any) (T, error) {
	return d.GetByIDContext(context.Background(), id)
}

// GetByIDContext 根据 id 查询单条数据，携带上下文
func (d *Dao[T]) GetByIDContext(ctx context.Context, id any) (T, error) {
	tableMeta := d.TableMeta
	return d.GetByColumnContext(ctx, OfKv(tableMeta.PrimaryKey, id))
}

// ListByIDs 根据 id 查询多条数据
func (d *Dao[T]) ListByIDs(ids ...any) ([]T, error) {
	return d.ListByIDsContext(context.Background(), ids...)
}

// ListByIDsContext 根据 id 查询多条数据，携带上下文
func (d *Dao[T]) ListByIDsContext(ctx context.Context, ids ...any) ([]T, error) {
	tableMeta := d.TableMeta
	return d.ListByColumnsContext(ctx, OfMultiKv(tableMeta.PrimaryKey, ids...))
}

// UpdateField 部分字段更新
func (d *Dao[T]) UpdateField(idValue any, fieldMap map[string]any) (int64, error) {
	return d.UpdateFieldContext(context.Background(), idValue, fieldMap)
}

// UpdateFieldContext 部分字段更新，携带上下文
func (d *Dao[T]) UpdateFieldContext(ctx context.Context, idValue any, fieldMap map[string]any) (int64, error) {
	if utils.IsIDEmpty(idValue) {
		return 0, ErrUpdatePrimaryKeyRequire
	}

	updater := d.Updater().Execer(d.getExecer())
	for col, val := range fieldMap {
		updater.Set(col, val)
	}
	updater.Where(ql.C(ql.Col(d.TableMeta.PrimaryKey).EQ(idValue)))
	affected, err := updater.ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// Update 根据 ID 全字段更新
func (d *Dao[T]) Update(model T, omitColumns ...string) (bool, error) {
	return d.UpdateContext(context.Background(), model, omitColumns...)
}

// UpdateContext 根据 ID 全字段更新，携带上下文
func (d *Dao[T]) UpdateContext(ctx context.Context, model T, omitColumns ...string) (bool, error) {
	if utils.IsIDEmpty(model.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := d.TableMeta
	affected, err := d.UpdateByCondContext(ctx, model, ql.SC().And(fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey)), tableMeta.PrimaryKey)
	return affected > 0, err
}

// UpdateByCond 按条件更新全部字段
func (d *Dao[T]) UpdateByCond(model T, where sqlbuilder.ConditionBuilder, omitColumns ...string) (int64, error) {
	return d.UpdateByCondContext(context.Background(), model, where, omitColumns...)
}

// UpdateByCondContext 按条件更新全部字段
func (d *Dao[T]) UpdateByCondContext(ctx context.Context, model T, where sqlbuilder.ConditionBuilder, omitColumns ...string) (int64, error) {
	omitColumns = append(omitColumns, d.TableMeta.PrimaryKey)
	if len(global.omitColumns) > 0 {
		omitColumns = append(omitColumns, global.omitColumns...)
	}
	updater := d.Updater().Execer(d.getExecer()).
		Columns(d.DBColumns(omitColumns...)...).
		Where(where)
	affected, err := updater.NamedExecContext(ctx, model)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (d *Dao[T]) deleteByCondContext(ctx context.Context, where sqlbuilder.ConditionBuilder) (int64, error) {
	return d.Deleter().Execer(d.getExecer()).Where(where).ExecContext(ctx)
}

// DeleteByColumn 按字段名删除
func (d *Dao[T]) DeleteByColumn(kv *KV) (int64, error) {
	return d.DeleteByColumnContext(context.Background(), kv)
}

// DeleteByColumnContext 按字段名删除，携带上下文
func (d *Dao[T]) DeleteByColumnContext(ctx context.Context, kv *KV) (int64, error) {
	if kv == nil {
		return 0, nil
	}
	return d.deleteByCondContext(ctx, ql.C(ql.Col(kv.Key).EQ(kv.Value)))
}

// DeleteByColumns 指定字段删除多个值
func (d *Dao[T]) DeleteByColumns(kvs *MultiKV) (int64, error) {
	return d.DeleteByColumnsContext(context.Background(), kvs)
}

// DeleteByColumnsContext 指定字段删除多个值，携带上下文
func (d *Dao[T]) DeleteByColumnsContext(ctx context.Context, kvs *MultiKV) (int64, error) {
	if kvs == nil || len(kvs.Values) == 0 {
		return 0, nil
	}
	return d.deleteByCondContext(ctx, ql.C(ql.Col(kvs.Key).In(kvs.Values...)))
}

// DeleteByID 根据id删除数据
func (d *Dao[T]) DeleteByID(id any) (bool, error) {
	return d.DeleteByIDContext(context.Background(), id)
}

// DeleteByIDContext 根据id删除数据，携带上下文
func (d *Dao[T]) DeleteByIDContext(ctx context.Context, id any) (bool, error) {
	tableMeta := d.TableMeta
	affected, err := d.DeleteByColumnContext(ctx, OfKv(tableMeta.PrimaryKey, id))
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// WithTableName 使用新的数据库连接创建 Dao
func (d *Dao[T]) WithTableName(tableName string) *Dao[T] {
	newDao := new(Dao[T])
	*newDao = *d
	newDao.TableMeta = newDao.TableMeta.WithTableName(tableName)
	return newDao
}

func (d *Dao[T]) WithExecutor(executor engine.Executor) *Dao[T] {
	newDao := new(Dao[T])
	*newDao = *d
	newDao.executor = executor
	return newDao
}

func (d *Dao[T]) initIfNullVal() {
	if d.ifNullVals == nil {
		d.ifNullVals = make(map[string]string)
	}
}

// GetMasterDB 返回主库连接
// 返回值: 主库连接对象
func (d *Dao[T]) GetMasterDB() engine.Execer {
	if d.masterDB != nil {
		return d.masterDB
	}
	if global.defaultMasterDB == nil {
		return nil
	}
	// double check
	if d.masterDB != nil {
		return d.masterDB
	}
	hooks := mergeHooks(d.options)
	d.masterDB = NewDb(global.defaultMasterDB, hooks...)
	return d.masterDB
}

// GetReadDB 返回从库连接
// 返回值: 从库连接对象
func (d *Dao[T]) GetReadDB() engine.Queryer {
	if d.readDB != nil {
		return d.readDB
	}
	if global.defaultReadDB == nil && d.GetMasterDB() == nil {
		return nil
	}
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

// getQueryer 获取查询执行器
// 返回值: 查询执行器接口
func (d *Dao[T]) getQueryer() engine.Queryer {
	if d.executor != nil {
		return d.executor
	}
	return d.GetReadDB()
}

// getExecer 获取更新执行器
// 返回值: 更新执行器接口
func (d *Dao[T]) getExecer() engine.Execer {
	if d.executor != nil {
		return d.executor
	}
	return d.GetMasterDB()
}

func (d *Dao[T]) newModel() T {
	return reflect.New(d.modelType.Elem()).Interface().(T)
}

func (d *Dao[T]) emptyModel() T {
	var dest T
	return dest
}

// mergeHooks 合并 hooks
// options: 配置选项
// 返回值: 合并后的 hooks 列表
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
