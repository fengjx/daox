package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/engine"
)

// Selector select 语句构造器
type Selector struct {
	sqlBuilder
	queryer     engine.Queryer
	tableName   string
	queryString string
	distinct    bool
	columns     []string
	where       ConditionBuilder
	orderBy     []OrderBy
	groupBy     []string
	limit       *int64
	offset      *int64
	isForUpdate bool
	ifNullVals  map[string]string
}

// NewSelector 创建一个selector
func NewSelector(tableName string) *Selector {
	selector := &Selector{
		tableName: tableName,
	}
	return selector
}

// Queryer 设置查询器
func (s *Selector) Queryer(queryer engine.Queryer) *Selector {
	s.queryer = queryer
	return s
}

// QueryString 自定义select字段，sql原样输出
func (s *Selector) QueryString(queryString string) *Selector {
	s.queryString = queryString
	return s
}

// StructColumns 通过任意model解析出表字段
// tagName 解析数据库字段的 tag-name
// omitColumns 排除哪些字段
func (s *Selector) StructColumns(model any, tagName string, omitColumns ...string) *Selector {
	columns := GetColumnsByModel(GetMapperByTagName(tagName), model, omitColumns...)
	return s.Columns(columns...)
}

// Columns select 的数据库字段
func (s *Selector) Columns(columns ...string) *Selector {
	s.columns = columns
	return s
}

// IfNullVal 设置字段为空时，返回的值
func (s *Selector) IfNullVal(col string, val string) *Selector {
	s.initIfNullVal()
	s.ifNullVals[col] = val
	return s
}

// IfNullVals 设置字段为空时，返回的值
// key 为数据库表字段名
// value 为默认值表达式，如：空字符串为 "”"
func (s *Selector) IfNullVals(vals map[string]string) *Selector {
	s.initIfNullVal()
	for col, val := range vals {
		s.ifNullVals[col] = val
	}
	return s
}

func (s *Selector) initIfNullVal() {
	if s.ifNullVals == nil {
		s.ifNullVals = make(map[string]string)
	}
}

// Distinct select distinct
func (s *Selector) Distinct() *Selector {
	s.distinct = true
	return s
}

// Where 条件
// condition 可以通过 sqlbuilder.C() 方法创建
func (s *Selector) Where(where ConditionBuilder) *Selector {
	s.where = where
	return s
}

// ForUpdate select for update
func (s *Selector) ForUpdate(isForUpdate bool) *Selector {
	s.isForUpdate = isForUpdate
	return s
}

// GroupBy group by
func (s *Selector) GroupBy(columns ...string) *Selector {
	s.groupBy = columns
	return s
}

// OrderBy order by
// orderBy sqlbuilder.Desc("col")
func (s *Selector) OrderBy(orderBy ...OrderBy) *Selector {
	s.orderBy = orderBy
	return s
}

// Limit 分页 limit
func (s *Selector) Limit(limit int64) *Selector {
	s.limit = &limit
	return s
}

// Offset 分页 offset
func (s *Selector) Offset(offset int64) *Selector {
	s.offset = &offset
	return s
}

// SQL 输出sql语句
func (s *Selector) SQL() (string, error) {
	s.reset()
	s.writeString("SELECT ")
	if s.queryString != "" {
		// 使用原始语句，例如 count, max 之类的函数
		s.writeString(s.queryString)
	} else {
		if s.distinct {
			s.writeString("DISTINCT ")
		}
		if len(s.columns) == 0 {
			s.writeByte('*')
		} else {
			n := len(s.columns)
			for i, column := range s.columns {
				if defVal, ok := s.ifNullVals[column]; ok {
					s.ifNullCol(column, defVal)
				} else {
					s.quote(column)
				}
				if i != n-1 {
					s.writeString(", ")
				}
			}
		}
	}
	s.writeString(" FROM ")
	s.quote(s.tableName)
	s.whereSQL(s.where)

	if len(s.groupBy) > 0 {
		s.writeString(" GROUP BY ")
		for i, column := range s.groupBy {
			if i > 0 {
				s.comma()
				s.space()
			}
			s.quote(column)
		}
	}

	// order by
	if len(s.orderBy) > 0 {
		s.writeString(" ORDER BY ")
		for i, ob := range s.orderBy {
			if i > 0 {
				s.comma()
			}
			for _, c := range ob.columns {
				s.quote(c)
			}
			s.space()
			s.writeString(ob.orderType)
		}
	}

	if s.limit != nil {
		s.writeString(" LIMIT ")
		s.writeString(strconv.FormatInt(*s.limit, 10))
	}
	if s.offset != nil {
		s.writeString(" OFFSET ")
		s.writeString(strconv.FormatInt(*s.offset, 10))
	}
	if s.isForUpdate {
		s.writeString(" FOR UPDATE ")
	}
	s.end()
	return s.sb.String(), nil
}

// CountSQL 构造 count 查询 sql
func (s *Selector) CountSQL() (string, error) {
	s.reset()
	s.writeString("SELECT COUNT(*)")
	s.writeString(" FROM ")
	s.quote(s.tableName)
	s.whereSQL(s.where)

	if len(s.groupBy) > 0 {
		s.writeString(" GROUP BY ")
		for i, column := range s.groupBy {
			if i > 0 {
				s.comma()
				s.space()
			}
			s.quote(column)
		}
	}
	s.end()
	return s.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (s *Selector) SQLArgs() (string, []any, error) {
	querySQL, err := s.SQL()
	if err != nil {
		return "", nil, err
	}
	args, hasInSQL := s.whereArgs(s.where)
	if !hasInSQL {
		return querySQL, args, err
	}
	return sqlx.In(querySQL, args...)
}

// CountSQLArgs 构造 count 查询 sql 并返回对应参数
func (s *Selector) CountSQLArgs() (string, []any, error) {
	querySQL, err := s.CountSQL()
	args, hasInSQL := s.whereArgs(s.where)
	if !hasInSQL {
		return querySQL, args, err
	}
	return sqlx.In(querySQL, args...)
}

// Select 查询多条数据
func (s *Selector) Select(dest any) error {
	return s.SelectContext(context.Background(), dest)
}

// SelectContext 查询多条数据
func (s *Selector) SelectContext(ctx context.Context, dest any) error {
	if s.queryer == nil {
		return ErrQueryerNotSet
	}
	querySQL, args, err := s.SQLArgs()
	if err != nil {
		return err
	}
	ec := &engine.ExecutorContext{
		Type:      engine.SELECT,
		SQL:       querySQL,
		TableName: s.tableName,
		Start:     time.Now(),
		Args:      args,
	}
	ctx = engine.WithExecutorCtx(ctx, ec)
	err = s.queryer.SelectContext(ctx, dest, querySQL, args...)
	if err != nil {
		return err
	}
	return nil
}

// Get 查询单条数据
func (s *Selector) Get(dest any) (exist bool, err error) {
	return s.GetContext(context.Background(), dest)
}

// GetContext 查询单条数据
func (s *Selector) GetContext(ctx context.Context, dest any) (exist bool, err error) {
	if s.queryer == nil {
		return false, ErrQueryerNotSet
	}
	querySQL, args, err := s.SQLArgs()
	if err != nil {
		return false, err
	}
	ec := &engine.ExecutorContext{
		Type:      engine.SELECT,
		SQL:       querySQL,
		TableName: s.tableName,
		Start:     time.Now(),
		Args:      args,
	}
	ctx = engine.WithExecutorCtx(ctx, ec)
	err = s.queryer.GetContext(ctx, dest, querySQL, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type OrderType string

const (
	ASC  OrderType = "ASC"
	DESC OrderType = "DESC"
)

type OrderBy struct {
	columns   []string
	orderType string
}

func Asc(columns ...string) OrderBy {
	return OrderBy{
		columns:   columns,
		orderType: string(ASC),
	}
}

func Desc(columns ...string) OrderBy {
	return OrderBy{
		columns:   columns,
		orderType: string(DESC),
	}
}
