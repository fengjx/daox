package daox_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fengjx/daox"
)

func TestInsert(t *testing.T) {
	ctx := context.Background()
	tableName := "test_insert_info"
	before(t, tableName)
	nowSec := time.Now().Unix()
	record := daox.InsertRecord{
		TableName: tableName,
		Row: map[string]any{
			"uid":        1024,
			"name":       "test-insert",
			"sex":        "male",
			"login_time": nowSec,
			"utime":      nowSec,
			"ctime":      nowSec,
		},
	}
	id, err := daox.Insert(ctx, newDb(), record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("insert id:", id)
	assert.Equal(t, int64(11), id)
}

func TestInsertOptions(t *testing.T) {
	ctx := context.Background()
	tableName := "test_insert_opt_info"
	before(t, tableName)
	nowSec := time.Now().Unix()
	record := daox.InsertRecord{
		TableName: tableName,
		Row: map[string]any{
			"uid":        1024,
			"name":       "test-insert",
			"sex":        "male",
			"login_time": nowSec,
			"utime":      nowSec,
			"ctime":      nowSec,
		},
	}
	id, err := daox.Insert(ctx, newDb(), record, daox.WithInsertDataWrapper(func(ctx context.Context, src map[string]any) map[string]any {
		if name, ok := src["name"]; ok {
			src["name"] = fmt.Sprintf("opt_%s", name)
		}
		return src
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("insert id:", id)
	assert.Equal(t, int64(11), id)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	tableName := "test_update_info"
	before(t, tableName)
	nowSec := time.Now().Unix()
	record := daox.UpdateRecord{
		TableName: tableName,
		Fields: map[string]any{
			"login_time": nowSec,
			"utime":      nowSec,
			"ctime":      nowSec,
		},
		Conditions: []daox.Condition{
			{
				ConditionType: daox.ConditionTypeEq,
				Op:            daox.OpAnd,
				Field:         "uid",
				Vals:          []any{100},
			},
		},
	}
	affected, err := daox.Update(ctx, newDb(), record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update affected:", affected)
	assert.Equal(t, true, affected > 0)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	tableName := "test_delete_info"
	before(t, tableName)
	record := daox.DeleteRecord{
		TableName: tableName,
		Conditions: []daox.Condition{
			{
				Op:            daox.OpAnd,
				ConditionType: daox.ConditionTypeEq,
				Field:         "id",
				Vals:          []any{1},
			},
		},
	}
	affected, err := daox.Delete(ctx, newDb(), record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update affected:", affected)
	assert.Equal(t, true, affected > 0)
}
