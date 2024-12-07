package redis

import (
	"github.com/go-redis/redis/v8"
)

func ProvideRedisClient(config *RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
}
