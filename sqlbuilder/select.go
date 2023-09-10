package sqlbuilder

import (
	"strconv"
)

type Selector struct {
	sqlBuilder
	tableName   string
	distinct    bool
	columns     []string
	queryString string
	where       *condition
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

func (s *Selector) StructColumns(m interface{}, tagName string, omitColumns ...string) *Selector {
	columns := GetColumnsByModel(GetMapperByTagName(tagName), m, omitColumns...)
	return s.Columns(columns...)
}

func (s *Selector) QueryString(queryString string) *Selector {
	s.queryString = queryString
	return s
}

func (s *Selector) Columns(columns ...string) *Selector {
	s.columns = columns
	return s
}

func (s *Selector) Distinct() *Selector {
	s.distinct = true
	return s
}

func (s *Selector) Where(condition *condition) *Selector {
	s.where = condition
	return s
}

func (s *Selector) ForUpdate(isForUpdate bool) *Selector {
	s.IsForUpdate = isForUpdate
	return s
}

func (s *Selector) GroupBy(columns ...string) *Selector {
	s.groupBy = columns
	return s
}

func (s *Selector) OrderBy(orderBy ...OrderBy) *Selector {
	s.orderBy = orderBy
	return s
}

func (s *Selector) Limit(limit int) *Selector {
	s.limit = &limit
	return s
}

func (s *Selector) Offset(offset int) *Selector {
	s.offset = &offset
	return s
}

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
