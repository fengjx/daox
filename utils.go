package daox

import (
	"encoding/json"
	"fmt"

	"github.com/fengjx/daox/sqlbuilder"
)

func isIDEmpty(id interface{}) bool {
	idStr := toString(id)
	return idStr == "" || idStr == "0"
}

func toString(src interface{}) string {
	if src == nil {
		return ""
	}

	switch v := src.(type) {
	case string:
		return src.(string)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", src)
	case float32, float64:
		bs, _ := json.Marshal(v)
		return string(bs)
	case bool:
		if b, ok := src.(bool); ok && b {
			return "true"
		} else {
			return "false"
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}

func containsString(collection []string, element string) bool {
	return sqlbuilder.ContainsString(collection, element)
}

func ModelListToMap(src []Model) map[interface{}]Model {
	if len(src) == 0 {
		return make(map[interface{}]Model, 0)
	}
	resMap := make(map[interface{}]Model, 0)
	for _, m := range src {
		resMap[m.GetID()] = m
	}
	return resMap
}
