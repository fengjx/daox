package daox

import (
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/engine"
)

var global *globalConfig

func init() {
	global = &globalConfig{
		metaMap: make(map[string]TableMeta),
	}
}

// global 全局配置
type globalConfig struct {
	mux sync.Mutex
	// defaultMasterDB 全局默认master数据库
	defaultMasterDB *DB
	// defaultReadDB 全局默认read数据库
	defaultReadDB *DB
	// 所有表元信息
	metaMap map[string]TableMeta
	// 保存时默认忽略的字段，全局生效
	// 一般用户统一的开发规范
	saveOmitColumns []string
}

func (g *globalConfig) setDefaultMasterDB(db *sqlx.DB, middlewares ...engine.Middleware) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.defaultMasterDB = NewDb(db, middlewares...)
}

func (g *globalConfig) setDefaultReadDB(db *sqlx.DB, middlewares ...engine.Middleware) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.defaultReadDB = NewDb(db, middlewares...)
}

func (g *globalConfig) registerMeta(meta TableMeta) {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.metaMap[meta.TableName] = meta
}

// UseDefaultMasterDB 默认主库
func UseDefaultMasterDB(master *sqlx.DB, middlewares ...engine.Middleware) {
	global.setDefaultMasterDB(master, middlewares...)
}

// UseDefaultReadDB 默认从库
func UseDefaultReadDB(read *sqlx.DB, middlewares ...engine.Middleware) {
	global.setDefaultReadDB(read, middlewares...)
}

// UseSaveOmits 设置保存时全局默认忽略的字段
func UseSaveOmits(omits ...string) {
	global.saveOmitColumns = append(global.saveOmitColumns, omits...)
}

// GetMetaInfo 根据表名获得元信息
func GetMetaInfo(tableName string) (meta TableMeta, ok bool) {
	meta, ok = global.metaMap[tableName]
	return
}
