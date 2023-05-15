package daox

import "time"

type Column struct {
	ColumnName   string
	IsPrimaryKey bool
}

type TableMeta struct {
	TableName       string
	Columns         []*Column
	PrimaryKey      string
	IsAutoIncrement bool
	CacheMeta       *CacheMeta
}

type CacheMeta struct {
	CacheKey   string
	Version    string
	CacheTime  time.Duration
	ExpireTime time.Duration
}
