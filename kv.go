package daox

type KV struct {
	Key   string
	Value interface{}
}

func Kv(key string, value interface{}) *KV {
	return &KV{
		Key:   key,
		Value: value,
	}
}

type MultiKV struct {
	Key    string
	Values []interface{}
}

func MultiKv(key string, values []interface{}) *MultiKV {
	return &MultiKV{
		Key:    key,
		Values: values,
	}
}
