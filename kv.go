package daox

type KV struct {
	Key   string
	Value interface{}
}

func OfKv(key string, value interface{}) *KV {
	return &KV{
		Key:   key,
		Value: value,
	}
}

type MultiKV struct {
	Key    string
	Values []interface{}
}

func OfMultiKv(key string, values ...interface{}) *MultiKV {
	return &MultiKV{
		Key:    key,
		Values: values,
	}
}

func (kv *MultiKV) AddValue(val string) *MultiKV {
	kv.AddValue(val)
	return kv
}
