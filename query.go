package daox

import (
	"context"
	"reflect"

	"github.com/jmoiron/sqlx"

	"github.com/fengjx/daox/sqlbuilder"
	"github.com/fengjx/daox/sqlbuilder/ql"
	"github.com/fengjx/daox/types"
	"github.com/fengjx/daox/utils"
)

const (
	OpAnd Op = "and"
	OpOr  Op = "or"

	ConditionTypeEq      ConditionType = "eq"       // 等于
	ConditionTypeNotEq   ConditionType = "not_eq"   // 不等于
	ConditionTypeLike    ConditionType = "like"     // 模糊匹配
	ConditionTypeNotLike ConditionType = "not_like" // 不包含
	ConditionTypeIn      ConditionType = "in"       // in
	ConditionTypeNotIn   ConditionType = "not_in"   // not in
	ConditionTypeGt      ConditionType = "gt"       // 大于
	ConditionTypeLt      ConditionType = "lt"       // 小于
	ConditionTypeGte     ConditionType = "gte"      // 大于等于
	ConditionTypeLte     ConditionType = "lte"      // 小于等于

	OrderTypeAsc  OrderType = "asc"  // 升序
	OrderTypeDesc OrderType = "desc" // 降序
)

// Page 分页参数
type Page struct {
	Offset     int64 `json:"offset"`      // 游标起始位置
	Limit      int64 `json:"limit"`       // 每页记录数
	HasNext    bool  `json:"has_next"`    // 是否有下一页
	Count      int64 `json:"count"`       // 总记录数
	QueryCount bool  `json:"query_count"` // 是否查询总数
}

// QueryRecord 查询参数
type QueryRecord struct {
	TableName   string       `json:"table_name"`             // 查询表
	Fields      []string     `json:"fields"`                 // 投影字段
	Conditions  []Condition  `json:"conditions,omitempty"`   // 查找字段
	OrderFields []OrderField `json:"order_fields,omitempty"` // 排序字段
	Page        *Page        `json:"page,omitempty"`         // 分页参数
}

// OrderField 排序字段
type OrderField struct {
	Field     string    `json:"field"`
	OrderType OrderType `json:"order_type"`
}

type OrderType string

// Condition 条件语句
type Condition struct {
	Disable       bool          `json:"disable"`        // true 禁用该条件
	Op            Op            `json:"op"`             // and or 连接符
	Field         string        `json:"field"`          // 查询条件字段
	Vals          []any         `json:"vals"`           // 查询字段值
	ConditionType ConditionType `json:"condition_type"` // 查找类型
}

type ConditionType string

// Op and or连接符
type Op string

// ToSQLArgs 返回 sql 语句和参数
func (q QueryRecord) ToSQLArgs() (sql string, args []any, err error) {
	selector := q.buildSelector()
	return selector.SQLArgs()
}

// ToCountSQLArgs 返回 count 查询 sql 语句和参数
func (q QueryRecord) ToCountSQLArgs() (sql string, args []any, err error) {
	selector := q.buildSelector()
	return selector.CountSQLArgs()
}

func (q QueryRecord) buildSelector() *sqlbuilder.Selector {
	selector := sqlbuilder.NewSelector(q.TableName)
	selector.Columns(q.Fields...)
	selector.Where(buildCondition(q.Conditions))
	if q.Page != nil {
		selector.Offset(q.Page.Offset).Limit(q.Page.Limit)
	}
	if len(q.OrderFields) > 0 {
		var orderBy []sqlbuilder.OrderBy
		for _, orderField := range q.OrderFields {
			switch orderField.OrderType {
			case OrderTypeAsc:
				orderBy = append(orderBy, ql.Asc(orderField.Field))
			case OrderTypeDesc:
				orderBy = append(orderBy, ql.Desc(orderField.Field))
			}
		}
		selector.OrderBy(orderBy...)
	}
	return selector
}

func buildCondition(conditions []Condition) sqlbuilder.ConditionBuilder {
	where := ql.C()
	for _, c := range conditions {
		if c.Disable {
			continue
		}
		switch {
		case c.ConditionType == ConditionTypeEq && c.Op == OpAnd:
			where.And(ql.Col(c.Field).EQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeEq && c.Op == OpOr:
			where.Or(ql.Col(c.Field).EQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeNotEq && c.Op == OpAnd:
			where.And(ql.Col(c.Field).NotEQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeNotEq && c.Op == OpOr:
			where.Or(ql.Col(c.Field).NotEQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeLike && c.Op == OpAnd:
			where.And(ql.Col(c.Field).Like(c.Vals[0]))
		case c.ConditionType == ConditionTypeLike && c.Op == OpOr:
			where.Or(ql.Col(c.Field).Like(c.Vals[0]))
		case c.ConditionType == ConditionTypeNotLike && c.Op == OpAnd:
			where.And(ql.Col(c.Field).NotLike(c.Vals[0]))
		case c.ConditionType == ConditionTypeNotLike && c.Op == OpOr:
			where.Or(ql.Col(c.Field).NotLike(c.Vals[0]))
		case c.ConditionType == ConditionTypeIn && c.Op == OpAnd:
			where.And(ql.Col(c.Field).In(c.Vals...))
		case c.ConditionType == ConditionTypeIn && c.Op == OpOr:
			where.Or(ql.Col(c.Field).In(c.Vals...))
		case c.ConditionType == ConditionTypeNotIn && c.Op == OpAnd:
			where.And(ql.Col(c.Field).NotIn(c.Vals...))
		case c.ConditionType == ConditionTypeNotIn && c.Op == OpOr:
			where.Or(ql.Col(c.Field).NotIn(c.Vals...))
		case c.ConditionType == ConditionTypeGt && c.Op == OpAnd:
			where.And(ql.Col(c.Field).GT(c.Vals[0]))
		case c.ConditionType == ConditionTypeGt && c.Op == OpOr:
			where.Or(ql.Col(c.Field).GT(c.Vals[0]))
		case c.ConditionType == ConditionTypeLt && c.Op == OpAnd:
			where.And(ql.Col(c.Field).LT(c.Vals[0]))
		case c.ConditionType == ConditionTypeLt && c.Op == OpOr:
			where.Or(ql.Col(c.Field).LT(c.Vals[0]))
		case c.ConditionType == ConditionTypeGte && c.Op == OpAnd:
			where.And(ql.Col(c.Field).GTEQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeGte && c.Op == OpOr:
			where.Or(ql.Col(c.Field).GTEQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeLte && c.Op == OpAnd:
			where.And(ql.Col(c.Field).LTEQ(c.Vals[0]))
		case c.ConditionType == ConditionTypeLte && c.Op == OpOr:
			where.Or(ql.Col(c.Field).LTEQ(c.Vals[0]))
		}
	}
	return where
}

// Find 通用查询封装
func Find[T any](ctx context.Context, dbx *sqlx.DB, query QueryRecord, opts ...SelectOption) (list []T, page *Page, err error) {
	opt := &SelectOptions{}
	for _, option := range opts {
		option(opt)
	}
	if opt.FieldsFilter != nil {
		query.Fields = filterSelectFields(ctx, opt.FieldsFilter, query.Fields)
	}
	sql, args, err := query.ToSQLArgs()
	if err != nil {
		return nil, query.Page, err
	}
	err = dbx.SelectContext(ctx, &list, sql, args...)
	if err != nil {
		return nil, query.Page, err
	}
	if opt.ResultWrapper != nil {
		for i, item := range list {
			list[i] = opt.ResultWrapper(ctx, item).(T)
		}
	}
	page = query.Page
	page.Offset += int64(len(list))
	if query.Page != nil && query.Page.QueryCount {
		count, err := getCount(ctx, dbx, query)
		if err != nil {
			return nil, query.Page, err
		}
		page.Count = count
		page.HasNext = count > page.Offset
	}
	return
}

// FindListMap 通用查询封装，返回 map 类型
func FindListMap(ctx context.Context, dbx *sqlx.DB, query QueryRecord, opts ...SelectOption) (list []map[string]any, page *Page, err error) {
	opt := &SelectOptions{}
	for _, option := range opts {
		option(opt)
	}
	if opt.FieldsFilter != nil {
		query.Fields = filterSelectFields(ctx, opt.FieldsFilter, query.Fields)
	}
	sql, args, err := query.ToSQLArgs()
	if err != nil {
		return nil, query.Page, err
	}
	rows, err := dbx.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, query.Page, err
	}
	defer rows.Close()
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return nil, query.Page, err
		}
		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return nil, query.Page, err
		}
		values := make([]interface{}, len(colTypes))
		for i, typ := range colTypes {
			t := types.SQLType2GolangType(typ.DatabaseTypeName())
			values[i] = reflect.New(t).Interface()
		}
		err = rows.Scan(values...)
		if err != nil {
			return nil, query.Page, err
		}
		data := make(map[string]any)
		for i, column := range columns {
			data[column] = reflect.ValueOf(values[i]).Elem().Interface()
		}
		if opt.ResultWrapper != nil {
			data = opt.ResultWrapper(ctx, data).(map[string]any)
		}
		list = append(list, data)
	}
	page = query.Page
	page.Offset += int64(len(list))
	if query.Page != nil && query.Page.QueryCount {
		count, err := getCount(ctx, dbx, query)
		if err != nil {
			return nil, query.Page, err
		}
		page.Count = count
		page.HasNext = count > page.Offset
	}
	return
}

func getCount(ctx context.Context, dbx *sqlx.DB, query QueryRecord) (int64, error) {
	var count int64
	if query.Page != nil && query.Page.QueryCount {
		countSQL, countArgs, err := query.ToCountSQLArgs()
		if err != nil {
			return 0, err
		}
		err = dbx.GetContext(ctx, &count, countSQL, countArgs...)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// GetRecord 单条记录查询
type GetRecord struct {
	TableName  string      `json:"table_name"`           // 查询表
	Fields     []string    `json:"fields"`               // 投影字段
	Conditions []Condition `json:"conditions,omitempty"` // 查找字段
}

// ToSQLArgs 返回 sql 语句和参数
func (r GetRecord) ToSQLArgs() (sql string, args []any, err error) {
	selector := sqlbuilder.NewSelector(r.TableName)
	selector.Columns(r.Fields...)
	selector.Where(buildCondition(r.Conditions))
	return selector.SQLArgs()
}

// Get 查询单条记录
func Get[T any](ctx context.Context, dbx *sqlx.DB, record GetRecord, opts ...SelectOption) (*T, error) {
	opt := &SelectOptions{}
	for _, option := range opts {
		option(opt)
	}
	if opt.FieldsFilter != nil {
		record.Fields = filterSelectFields(ctx, opt.FieldsFilter, record.Fields)
	}
	sql, args, err := record.ToSQLArgs()
	if err != nil {
		return nil, err
	}
	data := new(T)
	err = dbx.GetContext(ctx, data, sql, args...)
	if err != nil {
		return nil, err
	}
	if opt.ResultWrapper != nil {
		data = opt.ResultWrapper(ctx, data).(*T)
	}
	return data, nil
}

// GetMap 查询单条记录，返回 map
func GetMap(ctx context.Context, dbx *sqlx.DB, record GetRecord, opts ...SelectOption) (map[string]any, error) {
	opt := &SelectOptions{}
	for _, option := range opts {
		option(opt)
	}
	if opt.FieldsFilter != nil {
		record.Fields = filterSelectFields(ctx, opt.FieldsFilter, record.Fields)
	}
	sql, args, err := record.ToSQLArgs()
	if err != nil {
		return nil, err
	}
	row := dbx.QueryRowxContext(ctx, sql, args...)
	columns, err := row.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, err := row.ColumnTypes()
	if err != nil {
		return nil, err
	}
	values := make([]any, len(colTypes))
	for i, typ := range colTypes {
		t := types.SQLType2GolangType(typ.DatabaseTypeName())
		values[i] = reflect.New(t).Interface()
	}
	err = row.Scan(values...)
	if err != nil {
		return nil, err
	}
	data := make(map[string]any)
	for i, column := range columns {
		data[column] = reflect.ValueOf(values[i]).Elem().Interface()
	}
	if opt.ResultWrapper != nil {
		data = opt.ResultWrapper(ctx, data).(map[string]any)
	}
	return data, nil
}

func filterSelectFields(ctx context.Context, filter FieldsFilter, src []string) []string {
	disableFields := filter(ctx)
	var fields []string
	for _, field := range src {
		if utils.ContainsString(disableFields, field) {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}
