package sqlbuilder

import (
	"strconv"
	"strings"
)

type Selector struct {
	tableName    string
	distinct     bool
	columns      []string
	wheres       []*condition
	orderExpress string
	limit        *int
	offset       *int
}

func NewSelector(tableName string) *Selector {
	selector := &Selector{
		tableName: tableName,
	}
	return selector
}

func (s *Selector) Columns(columns ...string) *Selector {
	s.columns = columns
	return s
}

func (s *Selector) Where(conditions ...*condition) *Selector {
	s.wheres = conditions
	return s
}

func (s *Selector) OrderBy(orderExpress string) *Selector {
	s.orderExpress = orderExpress
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

func (s *Selector) Sql() (string, error) {
	sb := &strings.Builder{}
	sb.WriteString("SELECT ")
	if s.distinct {
		sb.WriteString("DISTINCT ")
	}
	if len(s.columns) == 0 {
		sb.WriteString("*")
	} else {
		for i, column := range s.columns {
			warpQuote(sb, strings.TrimSpace(column))
			if i != len(s.columns)-1 {
				sb.WriteString(", ")
			}
		}
	}
	sb.WriteString(" FROM ")
	warpQuote(sb, strings.TrimSpace(s.tableName))
	if len(s.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, cond := range s.wheres {
			if cond.meet {
				if i != 0 {
					sb.WriteString(" AND ")
				}
				sb.WriteString(cond.express)
			}
		}
	}
	if len(s.orderExpress) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(s.orderExpress)
	}
	if s.offset != nil {
		sb.WriteString(" OFFSET ")
		sb.WriteString(strconv.Itoa(*s.offset))
	}
	if s.limit != nil {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.Itoa(*s.limit))
	}
	return sb.String(), nil
}

type OrderType string

const (
	ASC  OrderType = "ASC"
	DESC OrderType = "DESC"
)

type Order struct {
	column    string
	orderType string
}

func Asc(column string) *Order {
	return &Order{
		column:    column,
		orderType: string(ASC),
	}
}

func Desc(column string) *Order {
	return &Order{
		column:    column,
		orderType: string(DESC),
	}
}
