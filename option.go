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

func WithRedis(redisClient *redis.Client) Option {
	return func(p *Dao) {
		p.Redis = redisClient
	}
}
