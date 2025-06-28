package daox_test

import (
	"reflect"
	"testing"

	"github.com/fengjx/daox/v2"
)

type Profile struct {
	ID     int
	UserID int
}

type Order struct {
	ID     int64
	UserID int64
}

type Role struct {
	ID int64
}

type User struct {
	ID      int64    `json:"id"`
	Profile *Profile `daox:"relType:o2o,fromKey:UserID,ref:Profile,joinKey:ID"`
	Orders  []*Order `daox:"relType:o2m,fromKey:ID,ref:Order,joinKey:UserID"`
	Roles   []*Role  `daox:"relType:m2m,fromKey:ID,ref:Role,joinKey:ID,joinTable:user_role,joinFromKey:UserID,joinJoinKey:RoleID"`
	NoTag   string
}

func TestParseDaoxTag(t *testing.T) {
	typ := reflect.TypeOf(User{})

	cases := []struct {
		fieldName string
		want      *daox.RelationInfo
		wantOk    bool
	}{
		{
			"Profile",
			&daox.RelationInfo{
				RelType: "o2o",
				FromKey: "UserID",
				Ref:     "Profile",
				JoinKey: "ID",
			},
			true,
		},
		{
			"Orders",
			&daox.RelationInfo{
				RelType: "o2m",
				FromKey: "ID",
				Ref:     "Order",
				JoinKey: "UserID",
			},
			true,
		},
		{
			"Roles",
			&daox.RelationInfo{
				RelType: "m2m",
				FromKey: "ID",
				Ref:     "Role",
				JoinKey: "ID",
			},
			true,
		},
		{
			"NoTag",
			nil,
			false,
		},
		{
			"Invalid",
			&daox.RelationInfo{},
			true, // 取决于你的实现，若无有效键值可为 false
		},
	}

	for _, c := range cases {
		field, _ := typ.FieldByName(c.fieldName)
		got, ok := daox.ParseRelation(field)
		if ok != c.wantOk {
			t.Errorf("%s: want ok=%v, got %v", c.fieldName, c.wantOk, ok)
		}
		if c.wantOk && got != nil {
			if *got != *c.want {
				t.Errorf("%s: want %+v, got %+v", c.fieldName, c.want, got)
			}
		}
	}
}

func TestParseAllDaoxRelations(t *testing.T) {
	relations := daox.ParseRelations(reflect.TypeOf(User{}))
	if len(relations) != 3 {
		t.Errorf("期望3个关联字段，实际: %d", len(relations))
	}

	tests := map[string]*daox.RelationInfo{
		"Profile": {
			RelType: "o2o",
			FromKey: "UserID",
			Ref:     "Profile",
			JoinKey: "ID",
		},
		"Orders": {
			RelType: "o2m",
			FromKey: "ID",
			Ref:     "Order",
			JoinKey: "UserID",
		},
		"Roles": {
			RelType: "m2m",
			FromKey: "ID",
			Ref:     "Role",
			JoinKey: "ID",
		},
	}

	for field, want := range tests {
		got, ok := relations[field]
		if !ok {
			t.Errorf("缺少字段: %s", field)
			continue
		}
		if *got != *want {
			t.Errorf("字段 %s 解析不符，期望: %+v，实际: %+v", field, want, got)
		}
	}

	// 检查未带 tag 的字段不会被解析
	if _, ok := relations["NoTag"]; ok {
		t.Errorf("NoTag 字段不应被解析为关联")
	}
}
