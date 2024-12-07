package redis

import "github.com/google/wire"

var RedisSet = wire.NewSet(
	ProvideRedisConfig,
	ProvideRedisClient,
)
