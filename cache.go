package daox

import (
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type CacheHold struct {
	Data     string `json:"data"`
	ExpireAt int64  `json:"expire_at"`
}

// IsExpired 缓存是否过期
func (h *CacheHold) IsExpired() bool {
	return time.Now().Unix() <= h.ExpireAt
}

func (h *CacheHold) GetData(dest interface{}) (isExpired bool, err error) {
	isExpired = h.IsExpired()
	err = json.Unmarshal([]byte(h.Data), dest)
	return
}

type CacheTool struct {
	RedisClient    *redis.Client
	ExpireTime     time.Duration // 缓存时长
	CacheCleanTime time.Duration // 缓存清理时长
}

func NewCacheTool(redisCtl *redis.Client, expireTime, cacheCleanTime time.Duration) *CacheTool {
	return &CacheTool{
		RedisClient:    redisCtl,
		ExpireTime:     expireTime,
		CacheCleanTime: cacheCleanTime,
	}
}

func (c *CacheTool) Get(key string) (*CacheHold, error) {
	hold := &CacheHold{}
	result, err := c.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(result), hold)
	if err != nil {
		return nil, err
	}
	return hold, nil
}

func (c *CacheTool) Set(key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	expireAt := time.Now().Unix() + c.ExpireTime.Milliseconds()
	hold := &CacheHold{
		Data:     string(jsonData),
		ExpireAt: expireAt,
	}
	_, err = c.RedisClient.Set(ctx, key, hold, c.CacheCleanTime).Result()
	return err
}

// Fetch
// invalidStale 当缓存过期时，是否使用旧值
func (c *CacheTool) Fetch(key string, dest interface{}, invalidStale bool) error {

	return nil
}
