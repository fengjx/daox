package daox

type KV struct {
	Key   string
	Value any
}

func OfKv(key string, value any) *KV {
	return &KV{
		Key:   key,
		Value: value,
	}
}

type MultiKV struct {
	Key    string
	Values []any
}

func OfMultiKv(key string, values ...any) *MultiKV {
	return &MultiKV{
		Key:    key,
		Values: values,
	}
}

func (kv *MultiKV) AddValue(val string) *MultiKV {
	kv.AddValue(val)
	return kv
}
