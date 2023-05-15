package daox

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Option func(*Dao)

func WithDBRead(read *sqlx.DB) Option {
	return func(p *Dao) {
		p.DBRead = read
	}
}

func WithCache(redisClient *redis.Client, cacheMeta *CacheMeta) Option {
	return func(d *Dao) {
		d.Redis = redisClient
		d.CacheMeta = cacheMeta
	}
}

func IsAutoIncrement() Option {
	return func(dao *Dao) {
		dao.TableMeta.IsAutoIncrement = true
	}
}
