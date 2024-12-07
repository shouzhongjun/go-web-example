package redis

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func ProvideRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
}
