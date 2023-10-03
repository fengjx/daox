package sqlbuilder

import (
	"strconv"
)

type Selector struct {
	sqlBuilder
	tableName   string
	queryString string
	distinct    bool
	columns     []string
	where       ConditionBuilder
	orderBy     []OrderBy
	groupBy     []string
	limit       *int
	offset      *int
	IsForUpdate bool
}

func NewSelector(tableName string) *Selector {
	selector := &Selector{
		tableName: tableName,
	}
	return selector
}

// QueryString 自定义select字段，sql原样输出
func (s *Selector) QueryString(queryString string) *Selector {
	s.queryString = queryString
	return s
}

// StructColumns 通过任意model解析出表字段
// tagName 解析数据库字段的 tag-name
// omitColumns 排除哪些字段
func (s *Selector) StructColumns(model interface{}, tagName string, omitColumns ...string) *Selector {
	columns := GetColumnsByModel(GetMapperByTagName(tagName), model, omitColumns...)
	return s.Columns(columns...)
}

// Columns select 的数据库字段
func (s *Selector) Columns(columns ...string) *Selector {
	s.columns = columns
	return s
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
	s.IsForUpdate = isForUpdate
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
func (s *Selector) Limit(limit int) *Selector {
	s.limit = &limit
	return s
}

// Offset 分页 offset
func (s *Selector) Offset(offset int) *Selector {
	s.offset = &offset
	return s
}

// SQL 拼接sql语句
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
			for i, column := range s.columns {
				s.quote(column)
				if i != len(s.columns)-1 {
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
		s.writeString(strconv.Itoa(*s.limit))
	}
	if s.offset != nil {
		s.writeString(" OFFSET ")
		s.writeString(strconv.Itoa(*s.offset))
	}
	if s.IsForUpdate {
		s.writeString(" FOR UPDATE ")
	}
	s.end()
	return s.sb.String(), nil
}

// SQLArgs 构造 sql 并返回对应参数
func (s *Selector) SQLArgs() (string, []interface{}, error) {
	sql, err := s.SQL()
	args := s.whereArgs(s.where)
	return sql, args, err
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
