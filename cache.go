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

type FillDataFun func(missItem string, dest interface{}) error
type BatchCreateDataFun func(missItems []string) (map[string]interface{}, error)

type CacheProvider struct {
	RedisClient *redis.Client
	ExpireTime  time.Duration // 缓存时长
}

func NewCacheProvider(redisCtl *redis.Client, expireTime time.Duration) *CacheProvider {
	return &CacheProvider{
		RedisClient: redisCtl,
		ExpireTime:  expireTime,
	}
}

func (c *CacheProvider) get(keyPrefix string, item string, dest interface{}) (bool, error) {
	key := c.genKey(keyPrefix, item)
	result, err := c.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if result == "" {
		return false, nil
	}
	err = json.Unmarshal([]byte(result), dest)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *CacheProvider) set(keyPrefix string, item string, data interface{}) error {
	key := c.genKey(keyPrefix, item)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.RedisClient.Set(ctx, key, jsonData, c.ExpireTime).Result()
	return err
}

func (c *CacheProvider) setAll(keyPrefix string, dataList map[string]interface{}) error {
	pipe := c.RedisClient.Pipeline()
	for item, data := range dataList {
		key := c.genKey(keyPrefix, item)
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		pipe.SetNX(ctx, key, jsonData, c.ExpireTime)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *CacheProvider) Del(keyPrefix string, item string) error {
	key := c.genKey(keyPrefix, item)
	_, err := c.RedisClient.Del(ctx, key).Result()
	return err
}

func (c *CacheProvider) BatchDel(keyPrefix string, items []string) error {
	pipe := c.RedisClient.Pipeline()
	for _, item := range items {
		key := c.genKey(keyPrefix, item)
		pipe.Del(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Fetch
// invalidStale 当缓存过期时，是否使用旧值
func (c *CacheProvider) Fetch(keyPrefix string, item string, dest interface{}, fun FillDataFun) error {
	value := reflect.ValueOf(dest)
	err := c.CheckPointer(value)
	if err != nil {
		return err
	}
	exist, err := c.get(keyPrefix, item, dest)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	err = fun(item, dest)
	if err != nil {
		return err
	}
	err = c.set(keyPrefix, item, dest)
	if err != nil {
		return err
	}
	return nil
}

// BatchFetch
// dest: must a slice
// fun: to create miss data
func (c *CacheProvider) BatchFetch(keyPrefix string, items []string, dest interface{}, fun BatchCreateDataFun) error {
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
		keys[i] = c.genKey(keyPrefix, items[i])
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
	for key, val := range list {
		redisKey := c.genKey(keyPrefix, key)
		dataList[redisKey] = val
		direct.Set(reflect.Append(direct, reflect.ValueOf(val)))
	}
	return c.setAll(keyPrefix, dataList)
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

func (c *CacheProvider) genKey(prefix string, item string) string {
	return fmt.Sprintf("{%s}_%s", prefix, item)
}
