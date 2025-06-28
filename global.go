package daox

import (
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/v2/engine"
)

var global *globalConfig

func init() {
	global = &globalConfig{
		tableMap: make(map[string]*TableMapper),
		daoMap:   make(map[string]any),
	}
}

// global 全局配置
type globalConfig struct {
	mux sync.Mutex
	// defaultMasterDB 全局默认master数据库
	defaultMasterDB *sqlx.DB
	// defaultReadDB 全局默认read数据库
	defaultReadDB *sqlx.DB
	// 所有表元信息
	tableMap map[string]*TableMapper
	// 所有dao
	daoMap map[string]any
	// 保存时默认忽略的字段，全局生效
	// 一般用户统一的开发规范
	omitColumns []string
	// 全局中间件
	hooks []engine.Hook
	// 打印sql
	printSQL engine.AfterHandler
}

func (g *globalConfig) setDefaultMasterDB(db *sqlx.DB) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.defaultMasterDB = db
}

func (g *globalConfig) setDefaultReadDB(db *sqlx.DB) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.defaultReadDB = db
}

func (g *globalConfig) regTable(tm *TableMapper) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.tableMap[tm.Meta.TableName] = tm
	g.tableMap[tm.ModelTypeName] = tm
}

func (g *globalConfig) regDao(table string, dao any) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.daoMap[table] = dao
}

func regDao[T Model](d *Dao[T]) {
	tm := d.TableMapper
	global.regTable(tm)
	global.regDao(tm.Meta.TableName, d)
}

// GetDao 根据表名获取dao
func GetDao[T Model](table string) *Dao[T] {
	d, ok := global.daoMap[table]
	if !ok {
		return nil
	}
	return d.(*Dao[T])
}

// UseDefaultMasterDB 默认主库
func UseDefaultMasterDB(master *sqlx.DB) {
	global.setDefaultMasterDB(master)
}

// UseDefaultReadDB 默认从库
func UseDefaultReadDB(read *sqlx.DB) {
	global.setDefaultReadDB(read)
}

// UseOmits 设置保存时全局默认忽略的字段
func UseOmits(omits ...string) {
	global.omitColumns = append(global.omitColumns, omits...)
}

// GetMetaInfo 根据表名获得元信息
func GetMetaInfo(tableName string) (*TableMeta, bool) {
	tm, ok := global.tableMap[tableName]
	if ok {
		return tm.Meta, true
	}
	return nil, false
}

// UseHooks 使用全局 hook
func UseHooks(hooks ...engine.Hook) {
	global.hooks = append(global.hooks, hooks...)
}

// PrintSQL 打印sql处理
func PrintSQL(p engine.AfterHandler) {
	global.printSQL = p
}
