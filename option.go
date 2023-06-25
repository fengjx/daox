package daox

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Option func(*Dao)

func WithDBRead(read *sqlx.DB) Option {
	return func(p *Dao) {
		p.DBRead = read
	}
}

func WithCacheExpireTime(cacheExpireTime time.Duration) Option {
	return func(d *Dao) {
		d.CacheProvider.ExpireTime = cacheExpireTime
	}
}

func WithCacheVersion(cacheVersion string) Option {
	return func(d *Dao) {
		d.CacheProvider.Version = cacheVersion
	}
}

func WithCache(redisClient *redis.Client) Option {
	return func(d *Dao) {
		d.RedisClient = redisClient
	}
}

func IsAutoIncrement() Option {
	return func(dao *Dao) {
		dao.TableMeta.IsAutoIncrement = true
	}
}
