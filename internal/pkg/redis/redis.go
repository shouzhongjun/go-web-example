package redis

import (
	"github.com/go-redis/redis/v8"
	"goWebExample/internal/configs"
)

func NewRedisClient(config *configs.AllConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Redis.RedisAddr(),
		Password: config.Redis.Password,
		DB:       config.Redis.Db,
	})
}
