package daox

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx/reflectx"
	"github.com/redis/go-redis/v9"
)

type BatchCreateDataFun func(missItems []string) (map[string]interface{}, error)

type CacheProvider struct {
	RedisClient *redis.Client
	Version     string
	KeyPrefix   string
	ExpireTime  time.Duration // 缓存时长
}

func NewCacheProvider(redisCtl *redis.Client, keyPrefix string, version string, expireTime time.Duration) *CacheProvider {
	return &CacheProvider{
		RedisClient: redisCtl,
		Version:     version,
		KeyPrefix:   keyPrefix,
		ExpireTime:  expireTime,
	}
}

func (c *CacheProvider) setAll(key string, dataList map[string]interface{}) error {
	pipe := c.RedisClient.Pipeline()
	for item, data := range dataList {
		cacheKey := c.genKey(key, item)
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		pipe.SetNX(ctx, cacheKey, jsonData, c.ExpireTime)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *CacheProvider) Del(key string, item string) error {
	cacheKey := c.genKey(key, item)
	_, err := c.RedisClient.Del(ctx, cacheKey).Result()
	return err
}

func (c *CacheProvider) BatchDel(key string, items []string) error {
	pipe := c.RedisClient.Pipeline()
	for _, item := range items {
		cacheKey := c.genKey(key, item)
		pipe.Del(ctx, cacheKey)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// BatchFetch
// dest: must a slice
// fun: to create miss data
func (c *CacheProvider) BatchFetch(key string, items []string, dest interface{}, fun BatchCreateDataFun) error {
	var v, vp reflect.Value
	value := reflect.ValueOf(dest)
	err := c.CheckPointer(value)
	if err != nil {
		return err
	}
	direct := reflect.Indirect(value)
	slice, err := baseType(value.Type(), reflect.Slice)
	if err != nil {
		return err
	}
	direct.SetLen(0)

	isPtr := slice.Elem().Kind() == reflect.Ptr
	base := reflectx.Deref(slice.Elem())

	keys := make([]string, len(items))
	for i := 0; i < len(items); i++ {
		keys[i] = c.genKey(key, items[i])
	}
	result, err := c.RedisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}
	missItems := make([]string, 0)
	for i, item := range result {
		jsonStr, ok := item.(string)
		if !ok || jsonStr == "" {
			missItems = append(missItems, items[i])
			continue
		}
		// create a new struct type (which returns PtrTo) and indirect it
		vp = reflect.New(base)
		v = reflect.Indirect(vp)

		err = json.Unmarshal([]byte(jsonStr), vp.Interface())
		if err != nil {
			return err
		}
		if isPtr {
			direct.Set(reflect.Append(direct, vp))
		} else {
			direct.Set(reflect.Append(direct, v))
		}
	}
	if len(missItems) == 0 {
		return nil
	}
	list, err := fun(missItems)
	if err != nil {
		return err
	}
	dataList := make(map[string]interface{}, len(list))
	for k, val := range list {
		cacheKey := c.genKey(key, k)
		dataList[cacheKey] = val
		direct.Set(reflect.Append(direct, reflect.ValueOf(val)))
	}
	return c.setAll(key, dataList)
}

func (c *CacheProvider) CheckPointer(value reflect.Value) error {
	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value")
	}
	if value.IsNil() {
		return errors.New("must not a nil pointer")
	}
	return nil
}

func (c *CacheProvider) genKey(key, item string) string {
	return fmt.Sprintf("{%s}_%s_%s_%s", c.KeyPrefix, c.Version, key, item)
}
