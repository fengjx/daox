package daox

// KV 表示键值对结构
type KV struct {
	Key   string // 字段名
	Value any    // 字段值，支持任意类型
}

// OfKv 创建一个新的键值对
// key: 字段名
// value: 字段值
func OfKv(key string, value any) *KV {
	return &KV{
		Key:   key,
		Value: value,
	}
}

// MultiKV 表示一个字段对应多个值的结构，用于 IN 查询等场景
type MultiKV struct {
	Key    string // 字段名
	Values []any  // 字段值列表，支持任意类型
}

// OfMultiKv 创建一个新的多值键值对
// key: 字段名
// values: 字段值列表
func OfMultiKv(key string, values ...any) *MultiKV {
	return &MultiKV{
		Key:    key,
		Values: values,
	}
}

// AddValue 向多值键值对中添加一个值
// val: 要添加的值
func (kv *MultiKV) AddValue(val any) *MultiKV {
	kv.Values = append(kv.Values, val)
	return kv
}
