package daox

import (
	"reflect"
	"time"
)

type TableMeta struct {
	TableName       string
	StructType      reflect.Type
	Columns         []string
	PrimaryKey      string
	IsAutoIncrement bool
	CacheMeta       *CacheMeta
}

type CacheMeta struct {
	CacheKey       string
	Version        string
	ExpireTime     time.Duration
	CacheCleanTime time.Duration
}

func (meta *TableMeta) OmitColumns(omit ...string) []string {
	columnArr := make([]string, 0, len(meta.Columns))
	for _, column := range meta.Columns {
		for _, o := range omit {
			if column != o {
				columnArr = append(columnArr, column)
			}
		}
	}
	return columnArr
}
