//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"goWebExample/internal/pkg/db"
	"goWebExample/internal/pkg/kafka"
	"goWebExample/internal/pkg/redis"
	"goWebExample/internal/service"
)

// InitializeApp 使用 Wire 自动生成依赖绑定

var ApplicationProviderSet = wire.NewSet(db.DBSet, kafka.KafkaSet, redis.RedisSet, NewApp)

var BusinessProviderSet = wire.NewSet(service.ServicesProvider)

func InitializeApp() (*App, error) {
	wire.Build(
		ApplicationProviderSet,
		BusinessProviderSet,
	)
	return nil, nil
}
