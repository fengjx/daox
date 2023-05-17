package daox

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/redis/go-redis/v9"
	"reflect"
	"time"
)

type CreateDataFun func() (interface{}, error)
type BatchCreateDataFun func(missItem []string) (map[string]interface{}, error)

type CacheTool struct {
	RedisClient *redis.Client
	ExpireTime  time.Duration // 缓存时长
}

func NewCacheTool(redisCtl *redis.Client, expireTime time.Duration) *CacheTool {
	return &CacheTool{
		RedisClient: redisCtl,
		ExpireTime:  expireTime,
	}
}

func (c *CacheTool) Get(key string, dest interface{}) (bool, error) {
	result, err := c.RedisClient.Get(ctx, key).Result()
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

func (c *CacheTool) Set(key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.RedisClient.Set(ctx, key, jsonData, c.ExpireTime).Result()
	return err
}

func (c *CacheTool) Del(key string) error {
	_, err := c.RedisClient.Del(ctx, key).Result()
	return err
}

// Fetch
// invalidStale 当缓存过期时，是否使用旧值
func (c *CacheTool) Fetch(key string, dest interface{}, fun CreateDataFun) error {
	ok, err := c.Get(key, dest)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	data, err := fun()
	if err != nil {
		return err
	}
	err = c.Set(key, data)
	if err != nil {
		return err
	}
	dest = data
	return nil
}

func (c *CacheTool) BatchFetch(keyPrefix string, items []string, dest interface{}, fun BatchCreateDataFun) error {
	var v, vp reflect.Value
	value := reflect.ValueOf(dest)
	// json.Unmarshal returns errors for these
	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value")
	}
	if value.IsNil() {
		return errors.New("must not a nil pointer")
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
		v = reflect.Indirect(v)

		err = json.Unmarshal([]byte(jsonStr), vp)
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
	for key, val := range list {
		err = c.Set(c.genKey(keyPrefix, key), val)
		if err != nil {
			return err
		}
		direct.Set(reflect.Append(direct, reflect.ValueOf(val)))
	}
	return nil
}

func (c *CacheTool) genKey(prefix string, item string) string {
	return fmt.Sprintf("{%s}_%s", prefix, item)
}
