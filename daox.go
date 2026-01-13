package daox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/fengjx/daox/v2/engine"
	"github.com/fengjx/daox/v2/sqlbuilder"
	"github.com/fengjx/daox/v2/sqlbuilder/ql"
	"github.com/fengjx/daox/v2/utils"
)

var (
	// ErrUpdatePrimaryKeyRequire 更新操作必须提供主键值
	ErrUpdatePrimaryKeyRequire = errors.New("[daox] Primary key require for update")
	ErrInsertPrimaryKeyRequire = errors.New("[daox] Primary key require for delete")
)

// Dao 数据访问对象，封装了数据库操作的基础方法
type Dao[T Model] struct {
	TableMapper      *TableMapper      // 表映射器
	currentTableName string            // 当前使用的表名，优先于 TableMapper.Meta.TableName
	options          *Options          // 配置选项
	masterDB         engine.Executor   // 主库连接
	readDB           engine.Executor   // 从库连接
	ifNullVals       map[string]string // NULL值替换配置
	omitColumns      []string          // 忽略的字段列表
	executor         engine.Executor   // SQL执行器，用于事务等场景
	preloads         *PreloadNode      // 预加载节点
	RelFiller        RelFiller[T]      // 关联数据填充器，用于填充关联数据
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
			read = global.defaultReadDB
		} else if options.master != nil {
			read = master
		}
	}
	// 检查泛型必须是指针
	modelType, err := baseType(reflect.TypeOf(new(T)), reflect.Ptr)
	if err != nil {
		panic(err)
	}
	tableMapper := &TableMapper{
		Meta:          meta,
		Mapper:        options.mapper,
		ModelType:     modelType,
		ModelTypeName: modelType.Name(),
	}
	dao := &Dao[T]{
		masterDB:    NewDb(master, hooks...),
		readDB:      NewDb(read, hooks...),
		TableMapper: tableMapper,
		ifNullVals:  options.ifNullVals,
		omitColumns: options.omitColumns,
		options:     options,
	}
	regDao(dao)
	return dao
}

// SQLBuilder 创建当前表的 SQL 构建器
// 返回值: SQL构建器对象
func (d *Dao[T]) SQLBuilder() *sqlbuilder.Builder {
	return sqlbuilder.New(d.TableName())
}

// Selector 创建当前表的查询构建器
// columns: 查询的字段列表，为空则查询全部字段
// 返回值: 查询构建器对象
func (d *Dao[T]) Selector(columns ...string) *sqlbuilder.Selector {
	if len(columns) == 0 {
		columns = d.DBColumns()
	}
	selector := d.SQLBuilder().Select(columns...)
	if len(d.ifNullVals) > 0 {
		selector.IfNullVals(d.ifNullVals)
	}
	selector.Queryer(d.getQueryer()).Preload(d.selectorPreload)
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

// DBColumns 获取当前表数据库字段
// omitColumns: 需要忽略的字段列表
// 返回值: 字段名列表
func (d *Dao[T]) DBColumns(omitColumns ...string) []string {
	columns := make([]string, 0)
	for _, column := range d.TableMapper.Meta.Columns {
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
	if d.currentTableName != "" {
		return d.currentTableName
	}
	return d.TableMapper.Meta.TableName
}

// Save 插入数据
// dest: 要插入的数据对象
// opts: 插入选项
// 返回值: 插入ID，错误信息
func (d *Dao[T]) Save(dest T, opts ...InsertOption) (int64, error) {
	return d.SaveContext(context.Background(), dest, opts...)
}

// SaveContext 插入数据，传递 context
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

// ReplaceIntoContext replace into table，传递 context
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

// IgnoreIntoContext 使用 INSERT IGNORE INTO 如果记录已存在则忽略，传递 context
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) IgnoreIntoContext(ctx context.Context, model T, opts ...InsertOption) (sql.Result, error) {
	return d.Inserter(opts...).
		IsIgnoreInto(true).
		NamedExecContext(ctx, model)
}

// BatchSave 批量新增，传递 context
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

// BatchReplaceIntoContext 批量新增，使用 replace into 方式，传递 context
// models 是一个 slice
// omitColumns 不需要 insert 的字段
func (d *Dao[T]) BatchReplaceIntoContext(ctx context.Context, models []T, opts ...InsertOption) (sql.Result, error) {
	return d.Inserter(opts...).
		IsReplaceInto(true).
		NamedExecContext(ctx, models)
}

func (d *Dao[T]) getSaveColumns(opt *InsertOptions) []string {
	meta := d.TableMapper.Meta
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

func (d *Dao[T]) getByCond(ctx context.Context, cond sqlbuilder.ConditionBuilder) (T, error) {
	dest := d.newModel()
	exist, err := d.Selector().Where(cond).OneContext(ctx, dest)
	if err != nil {
		return d.emptyModel(), err
	}
	if !exist {
		return d.emptyModel(), nil
	}
	return dest, nil
}

func (d *Dao[T]) listByCond(ctx context.Context, cond sqlbuilder.ConditionBuilder) ([]T, error) {
	var dest []T
	err := d.Selector().Where(cond).ListContext(ctx, &dest)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

// GetByCondContext 根据条件查询单条数据
func (d *Dao[T]) GetByCondContext(ctx context.Context, whereCol ...sqlbuilder.Column) (T, error) {
	return d.getByCond(ctx, ql.C(whereCol...))
}

// ListByCondContext 根据条件查询多条数据
func (d *Dao[T]) ListByCondContext(ctx context.Context, whereCol ...sqlbuilder.Column) ([]T, error) {
	return d.listByCond(ctx, ql.C(whereCol...))
}

// GetByID 根据 id 查询单条数据
func (d *Dao[T]) GetByID(id any) (T, error) {
	return d.GetByIDContext(context.Background(), id)
}

// GetByIDContext 根据 id 查询单条数据，传递 context
func (d *Dao[T]) GetByIDContext(ctx context.Context, id any) (T, error) {
	tableMeta := d.TableMapper.Meta
	return d.getByCond(ctx, ql.C(ql.Col(tableMeta.PrimaryKey).EQ(id)))
}

// Page 分页查询
// offset: 偏移量
// limit: 限制数量
// whereCol: 条件
// 返回值: 总数量, 数据列表, 错误
func (d *Dao[T]) Page(offset, limit int64, whereCol ...sqlbuilder.Column) (total int64, items []T, err error) {
	return d.PageContext(context.Background(), offset, limit, whereCol...)
}

// PageContext 分页查询，传递 context
// ctx: 上下文
// offset: 偏移量
// limit: 限制数量
// whereCol: 条件
// 返回值: 总数量, 数据列表, 错误
func (d *Dao[T]) PageContext(ctx context.Context, offset, limit int64, whereCol ...sqlbuilder.Column) (total int64, items []T, err error) {
	selector := d.Selector().Offset(offset).Limit(limit)
	if len(whereCol) > 0 {
		selector.Where(ql.C(whereCol...))
	}
	total, err = selector.GetCountContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	err = selector.ListContext(ctx, &items)
	if err != nil {
		return 0, nil, err
	}
	return total, items, nil
}

// ListByIDs 根据 id 查询多条数据
// ids 需要传一个 slice
func (d *Dao[T]) ListByIDs(ids any) ([]T, error) {
	return d.ListByIDsContext(context.Background(), ids)
}

// ListByIDsContext 根据 id 查询多条数据，传递 context
// ids 需要传一个 slice
func (d *Dao[T]) ListByIDsContext(ctx context.Context, ids any) ([]T, error) {
	tableMeta := d.TableMapper.Meta
	return d.listByCond(ctx, ql.C(ql.Col(tableMeta.PrimaryKey).InSlice(ids)))
}

// UpdateMapFields 部分字段更新
func (d *Dao[T]) UpdateMapFields(idValue any, fieldMap map[string]any) (int64, error) {
	return d.UpdateMapFieldsContext(context.Background(), idValue, fieldMap)
}

// UpdateMapFieldsContext 部分字段更新，传递 context
func (d *Dao[T]) UpdateMapFieldsContext(ctx context.Context, idValue any, attr map[string]any) (int64, error) {
	if utils.IsIDEmpty(idValue) {
		return 0, ErrUpdatePrimaryKeyRequire
	}
	cond := ql.C(ql.Col(d.TableMapper.Meta.PrimaryKey).EQ(idValue))
	affected, err := d.updateByCondContext(ctx, attr, cond, 0)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// UpdateFieldsByID 部分字段更新，传递 context
func (d *Dao[T]) UpdateFieldsByID(id any, fields ...sqlbuilder.Field) (int64, error) {
	return d.UpdateFieldsByIDContext(context.Background(), id, fields...)
}

// UpdateFieldsByIDContext 部分字段更新，传递 context
func (d *Dao[T]) UpdateFieldsByIDContext(ctx context.Context, id any, fields ...sqlbuilder.Field) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	return d.Updater().Fields(
		fields...,
	).WhereC(ql.Col(d.TableMapper.Meta.PrimaryKey).EQ(id)).ExecContext(ctx)
}

// Update 根据 ID 全字段更新
func (d *Dao[T]) Update(model T, omitColumns ...string) (bool, error) {
	return d.UpdateContext(context.Background(), model, omitColumns...)
}

// UpdateContext 根据 ID 全字段更新，传递 context
func (d *Dao[T]) UpdateContext(ctx context.Context, model T, omitColumns ...string) (bool, error) {
	if utils.IsIDEmpty(model.GetID()) {
		return false, ErrUpdatePrimaryKeyRequire
	}
	tableMeta := d.TableMapper.Meta
	omitColumns = append(omitColumns, tableMeta.PrimaryKey)
	cond := ql.SC().And(fmt.Sprintf("%[1]s = :%[1]s", tableMeta.PrimaryKey))
	updater := d.Updater().Execer(d.getExecer()).
		Columns(d.DBColumns(omitColumns...)...).
		Where(cond)
	affected, err := updater.NamedExecContext(ctx, model)
	if err != nil {
		return false, err
	}
	return affected > 0, err
}

// UpdateByCond 按条件更新全部字段
func (d *Dao[T]) UpdateByCond(attr map[string]any, whereCol ...sqlbuilder.Column) (int64, error) {
	return d.UpdateByCondContext(context.Background(), attr, whereCol...)
}

// UpdateByCondContext 按条件更新全部字段
func (d *Dao[T]) UpdateByCondContext(ctx context.Context, attr map[string]any, whereCol ...sqlbuilder.Column) (int64, error) {
	return d.updateByCondContext(ctx, attr, ql.C(whereCol...), 0)
}

func (d *Dao[T]) updateByCondContext(ctx context.Context, attr map[string]any, where sqlbuilder.ConditionBuilder, limit int) (int64, error) {
	updater := d.Updater().Execer(d.getExecer()).
		SetMap(attr).
		Where(where)
	if limit > 0 {
		updater.Limit(limit)
	}
	affected, err := updater.ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (d *Dao[T]) deleteByCondContext(ctx context.Context, where sqlbuilder.ConditionBuilder, limit int) (int64, error) {
	deleter := d.Deleter().Execer(d.getExecer()).Where(where)
	if limit > 0 {
		deleter.Limit(limit)
	}
	return deleter.ExecContext(ctx)
}

// DeleteByID 根据id删除数据
func (d *Dao[T]) DeleteByID(id any) (bool, error) {
	return d.DeleteByIDContext(context.Background(), id)
}

// DeleteByIDContext 根据id删除数据，传递 context
func (d *Dao[T]) DeleteByIDContext(ctx context.Context, id any) (bool, error) {
	if utils.IsIDEmpty(id) {
		return false, ErrInsertPrimaryKeyRequire
	}
	tableMeta := d.TableMapper.Meta
	affected, err := d.deleteByCondContext(ctx, ql.C(
		ql.Col(tableMeta.PrimaryKey).EQ(id),
	), 0)
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (d *Dao[T]) DeleteByCondContext(ctx context.Context, whereCol ...sqlbuilder.Column) (int64, error) {
	return d.deleteByCondContext(ctx, ql.C(
		whereCol...,
	), 0)
}

// relFill 填充关联数据
func (d *Dao[T]) relFill(ctx context.Context, items []T) error {
	if d.RelFiller != nil {
		return d.RelFiller(ctx, items, d.preloads)
	}
	return nil
}

func (d *Dao[T]) selectorPreload(ctx context.Context, dest any) error {
	if d.RelFiller == nil {
		return nil
	}
	var items []T
	if v, ok := dest.(T); ok {
		items = []T{v}
	} else if v, ok := dest.(*[]T); ok {
		items = *v
	} else if v, ok := dest.([]T); ok {
		items = v
	}
	return d.relFill(ctx, items)
}

// Preload 预加载
func (d *Dao[T]) Preload(paths ...string) *Dao[T] {
	newDao := d.copy()
	preloads := parsePreloadPath(paths...)
	newDao.preloads = preloads
	return newDao
}

// WithPreloadNode 设置关联节点
// 一般是代码自动生成调用的 api，大多数情况不需要手动调用这个方法
func (d *Dao[T]) WithPreloadNode(p *PreloadNode) *Dao[T] {
	newDao := d.copy()
	newDao.preloads = p
	return newDao
}

// WithTableName 使用新的表名创建 Dao
func (d *Dao[T]) WithTableName(tableName string) *Dao[T] {
	newDao := d.copy()
	newDao.currentTableName = tableName
	return newDao
}

// WithExecutor 使用新的执行器创建 Dao
func (d *Dao[T]) WithExecutor(executor engine.Executor) *Dao[T] {
	newDao := d.copy()
	newDao.executor = executor
	return newDao
}

// WithMaster 使用主库执行器创建 Dao
func (d *Dao[T]) WithMaster() *Dao[T] {
	newDao := d.copy()
	newDao.executor = d.masterDB
	return newDao
}

// WithRead 使用从库执行器创建 Dao
func (d *Dao[T]) WithRead() *Dao[T] {
	newDao := d.copy()
	newDao.executor = d.readDB
	return newDao
}

func (d *Dao[T]) copy() *Dao[T] {
	newDao := new(Dao[T])
	*newDao = *d
	return newDao
}

// GetMasterDB 返回主库连接
// 返回值: 主库连接对象
func (d *Dao[T]) GetMasterDB() engine.Execer {
	return d.masterDB
}

// GetReadDB 返回从库连接
// 返回值: 从库连接对象
func (d *Dao[T]) GetReadDB() engine.Queryer {
	return d.readDB
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
	var dest T
	return dest.New().(T)
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
