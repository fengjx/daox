package daox_test

import (
	"context"
	"testing"

	"github.com/fengjx/daox"
)

func TestQuery_ToSQLArgs(t *testing.T) {
	q := daox.Query{
		TableName: "users",
		Fields:    []string{"id", "name", "age", "ctime"},
		Conditions: []daox.Condition{
			{
				ConditionType: daox.ConditionTypeGte,
				Op:            daox.OpAnd,
				Field:         "age",
				Vals:          []any{18},
			},
			{
				ConditionType: daox.ConditionTypeLike,
				Op:            daox.OpAnd,
				Field:         "name",
				Vals:          []any{"%feng%"},
			},
		},
		OrderFields: []daox.OrderField{{Field: "ctime", OrderType: daox.OrderTypeDesc}},
		Page: &daox.Page{
			Offset: 0,
			Limit:  20,
		},
	}
	sql, args, err := q.ToSQLArgs()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("sql:", sql)
	t.Log("args:", args)

	countSQL, countArgs, err := q.ToCountSQLArgs()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("countSQL:", countSQL)
	t.Log("countArgs:", countArgs)
}

func TestFind(t *testing.T) {
	after(t)
	before(t)
	q := daox.Query{
		TableName: "demo_info",
		Fields:    []string{"id", "uid", "name", "sex", "login_time", "ctime"},
		Conditions: []daox.Condition{
			{
				ConditionType: daox.ConditionTypeGte,
				Op:            daox.OpAnd,
				Field:         "uid",
				Vals:          []any{100},
			},
		},
		OrderFields: []daox.OrderField{{Field: "id", OrderType: daox.OrderTypeDesc}},
		Page: &daox.Page{
			Offset: 0,
			Limit:  3,
		},
	}
	list, page, err := daox.Find[DemoInfo](context.Background(), newDb(), q)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list:", list)
	t.Log("page:", page)
	after(t)
}

func TestListMap(t *testing.T) {
	after(t)
	before(t)
	q := daox.Query{
		TableName: "demo_info",
		Fields:    []string{"id", "uid", "name", "sex", "login_time", "ctime"},
		Conditions: []daox.Condition{
			{
				ConditionType: daox.ConditionTypeGte,
				Op:            daox.OpAnd,
				Field:         "uid",
				Vals:          []any{100},
			},
		},
		OrderFields: []daox.OrderField{{Field: "id", OrderType: daox.OrderTypeDesc}},
		Page: &daox.Page{
			Offset: 0,
			Limit:  2,
		},
	}
	list, page, err := daox.FindListMap(context.Background(), newDb(), q)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list:", list)
	t.Log("page:", page)
	after(t)
}
