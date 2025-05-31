package daox_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fengjx/daox/v2"
)

func TestQuery_ToSQLArgs(t *testing.T) {
	q := daox.QueryRecord{
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
	ctx := context.Background()
	tableName := "test_find_info"
	before(t, tableName)
	q := daox.QueryRecord{
		TableName: tableName,
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
	list, page, err := daox.Find[DemoInfo](ctx, newDb(), q)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list:", list)
	t.Log("page:", page)
}

func TestListMap(t *testing.T) {
	ctx := context.Background()
	tableName := "test_find_map_info"
	before(t, tableName)
	q := daox.QueryRecord{
		TableName: tableName,
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
	list, page, err := daox.FindListMap(ctx, newDb(), q)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list:", list)
	t.Log("page:", page)
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	tableName := "test_get_info"
	before(t, tableName)
	record := daox.GetRecord{
		TableName: tableName,
		Fields:    []string{"id", "uid", "name", "sex", "login_time", "ctime"},
		Conditions: []daox.Condition{
			{
				Op:            daox.OpAnd,
				ConditionType: daox.ConditionTypeEq,
				Field:         "id",
				Vals:          []any{1},
			},
		},
	}
	data, err := daox.Get[DemoInfo](ctx, newDb(), record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("data", data)
	assert.Equal(t, int64(1), data.ID)
	assert.Equal(t, int64(100), data.UID)
}

func TestGetMap(t *testing.T) {
	ctx := context.Background()
	tableName := "test_get_map_info"
	before(t, tableName)
	record := daox.GetRecord{
		TableName: tableName,
		Fields:    []string{"id", "uid", "name", "sex", "login_time", "ctime"},
		Conditions: []daox.Condition{
			{
				Op:            daox.OpAnd,
				ConditionType: daox.ConditionTypeEq,
				Field:         "id",
				Vals:          []any{1},
			},
		},
	}
	data, err := daox.GetMap(ctx, newDb(), record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("data", data)
	assert.Equal(t, int32(1), data["id"])
	assert.Equal(t, int32(100), data["uid"])
}
